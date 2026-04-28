package application

import (
	"context"
	"errors"
	"mayfly-go/internal/ai/agent"
	"mayfly-go/internal/ai/application/dto"
	"mayfly-go/internal/ai/domain/entity"
	"mayfly-go/internal/ai/domain/repository"
	"mayfly-go/internal/ai/session"
	"mayfly-go/pkg/base"
	"mayfly-go/pkg/errorx"
	"mayfly-go/pkg/model"
	"mayfly-go/pkg/utils/collx"
	"mayfly-go/pkg/utils/jsonx"

	"github.com/cloudwego/eino/schema"
)

type Session interface {
	base.App[*entity.Session]
	session.Store

	// ListSessions 列出会话
	ListSessions(ctx context.Context, query *dto.SessionQuery) ([]*entity.Session, error)

	ListSessionMessages(ctx context.Context, query *dto.SessionMessageQuery) ([]*entity.SessionMessage, error)
}

type sessionAppImpl struct {
	base.AppImpl[*entity.Session, repository.Session]

	sessionMessageRepo repository.SessionMessage `inject:"T"`
}

var _ session.Store = (*sessionAppImpl)(nil)
var _ Session = (*sessionAppImpl)(nil)

func (s *sessionAppImpl) ListSessions(ctx context.Context, query *dto.SessionQuery) ([]*entity.Session, error) {
	cond := model.NewCond().
		Eq("creatorId", query.UserId).
		OrderByDesc("id")
	return s.ListByCond(cond)
}

func (s *sessionAppImpl) ListSessionMessages(ctx context.Context, query *dto.SessionMessageQuery) ([]*entity.SessionMessage, error) {
	cond := model.NewCond().
		Eq("sessionKey", query.SessionKey).
		OrderByAsc("id")
	return s.sessionMessageRepo.SelectByCond(cond)
}

// GetMessage 根据查询条件获取单条消息
func (s *sessionAppImpl) GetMessage(ctx context.Context, query *session.MessageQuery) ([]*session.Message, error) {
	msgs, err := s.sessionMessageRepo.SelectByCond(model.NewCond().
		Eq("msgType", query.MessageType).
		Eq("actionId", query.ActionId).
		Eq("turnId", query.TurnId).
		Eq("toolCallId", query.ToolCallId).
		OrderByDesc("id"),
	)
	if err != nil {
		return nil, err
	}
	if len(msgs) == 0 {
		return nil, errorx.NewBiz("no message")
	}

	return collx.ArrayMap(msgs, func(sm *entity.SessionMessage) *session.Message {
		return toSessionMessage(sm)
	}), nil
}

// UpdateMessage 更新单条消息
func (s *sessionAppImpl) UpdateMessage(ctx context.Context, msg *session.Message) error {
	if msg.Id == 0 {
		return nil
	}

	em := toEntityMessage("", msg)
	return s.sessionMessageRepo.Save(ctx, em)
}

// AppendMsgs 追加消息到会话历史
func (s *sessionAppImpl) AppendMsgs(ctx context.Context, sessionKey string, msgs ...*session.Message) error {
	if len(msgs) == 0 {
		return nil
	}

	sessionMessages := collx.ArrayMap(msgs, func(msg *session.Message) *entity.SessionMessage {
		return toEntityMessage(sessionKey, msg)
	})

	return s.sessionMessageRepo.BatchInsert(ctx, sessionMessages)
}

// GetHistory 获取会话历史消息
func (s *sessionAppImpl) GetHistory(ctx context.Context, sessionKey string, limit int) ([]*session.Message, error) {
	if limit <= 0 {
		limit = 1000
	}
	messages, err := s.sessionMessageRepo.SelectHistory(ctx, sessionKey, limit)
	if err != nil {
		return nil, err
	}
	return collx.ArrayMap(messages, func(msg *entity.SessionMessage) *session.Message {
		return toSessionMessage(msg)
	}), nil
}

// ClearHistory 清空会话历史消息
func (s *sessionAppImpl) ClearHistory(ctx context.Context, sessionKey string) error {
	return s.sessionMessageRepo.DeleteByCond(ctx, &entity.SessionMessage{SessionKey: sessionKey})
}

// ListMetas 列出所有会话元信息
func (s *sessionAppImpl) ListMetas(ctx context.Context) ([]*session.SessionMeta, error) {
	return nil, errors.New("not implemented")
}

// GetMeta 获取会话元信息
func (s *sessionAppImpl) GetMeta(ctx context.Context, sessionKey string) (*session.SessionMeta, error) {
	sessionMeta := &entity.Session{SessionKey: sessionKey}
	err := s.GetByCond(sessionMeta)
	if err != nil {
		return nil, nil
	}

	return &session.SessionMeta{
		Key:        sessionMeta.SessionKey,
		Summary:    sessionMeta.Summary,
		Count:      sessionMeta.MessageCount,
		TokenCount: sessionMeta.TokenCount,
		CreatedAt:  *sessionMeta.CreateTime,
		UpdatedAt:  *sessionMeta.UpdateTime,
		Skip:       sessionMeta.GetExtraInt("skip"),
	}, nil
}

// SaveMeta 保存会话元信息
func (s *sessionAppImpl) SaveMeta(ctx context.Context, meta *session.SessionMeta) error {
	session := &entity.Session{
		SessionKey:   meta.Key,
		Summary:      meta.Summary,
		MessageCount: meta.Count,
		TokenCount:   meta.TokenCount,
	}
	session.SetExtraValue("skip", meta.Skip)

	// 检查是否存在，存在则更新，不存在则创建
	existing := &entity.Session{SessionKey: meta.Key}
	err := s.GetByCond(existing)
	if err == nil {
		session.Id = existing.Id
	} else {
		session.Title = meta.Extra.GetStr("title")
	}

	return s.Save(ctx, session)
}

// DeleteMeta 删除会话元信息
func (s *sessionAppImpl) DeleteMeta(ctx context.Context, sessionKey string) error {
	return s.DeleteByCond(ctx, &entity.Session{SessionKey: sessionKey})
}

// toEntityMessage 将 session.Message 转换为 entity.SessionMessage
func toEntityMessage(sessionKey string, msg *session.Message) *entity.SessionMessage {
	// 根据角色和 toolcalls 判断消息类型
	msgType := getMessageType(msg)

	sm := &entity.SessionMessage{
		SessionKey: sessionKey,
		TurnId:     msg.TurnId,
		Role:       string(msg.Role),
		MsgType:    msgType,
		Content:    msg.Content,
		ToolCalls:  jsonx.ToStr(msg.ToolCalls),
		ToolCallId: msg.ToolCallId,
		ActionId:   msg.ActionId,
	}
	sm.Id = uint64(msg.Id)
	extra := collx.M(msg.Extra)
	if msg.Role == schema.Tool {
		extra.Set("toolName", msg.ToolName)
	}
	sm.Extra = extra
	return sm
}

// toSessionMessage 将 entity.SessionMessage 转换为 session.Message
func toSessionMessage(msg *entity.SessionMessage) *session.Message {
	sm := &session.Message{
		Id:         int64(msg.Id),
		Role:       schema.RoleType(msg.Role),
		Content:    msg.Content,
		ToolCalls:  nil,
		ToolCallId: msg.ToolCallId,
		ActionId:   msg.ActionId,
		MsgType:    msg.MsgType,
	}
	if msg.ToolCalls != "" {
		tollcalls, _ := jsonx.ToByStr[[]schema.ToolCall](msg.ToolCalls)
		if tollcalls != nil {
			sm.ToolCalls = *tollcalls
		}
	}
	if msg.Extra != nil {
		sm.Extra = msg.Extra
	}
	return sm
}

// getMessageType 根据 session.Message 判断消息类型
func getMessageType(msg *session.Message) string {
	switch msg.Role {
	case schema.User:
		return entity.MsgTypeUser
	case schema.Tool:
		return entity.MsgTypeToolResult
	case agent.RoleInternal:
		if mt := msg.Extra.GetStr("type"); mt != "" {
			return mt
		}
		return entity.MsgTypeInternal
	case schema.Assistant:
		// assistant 带 tool_calls 为工具调用，否则为普通回复
		if len(msg.ToolCalls) > 0 {
			return entity.MsgTypeToolCall
		}
		return entity.MsgTypeAssistant
	default:
		return entity.MsgTypeAssistant
	}
}
