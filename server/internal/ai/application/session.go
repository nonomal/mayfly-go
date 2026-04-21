package application

import (
	"cmp"
	"context"
	"errors"
	"mayfly-go/internal/ai/agent"
	"mayfly-go/internal/ai/application/dto"
	"mayfly-go/internal/ai/domain/entity"
	"mayfly-go/internal/ai/domain/repository"
	"mayfly-go/internal/ai/session"
	"mayfly-go/pkg/base"
	"mayfly-go/pkg/model"
	"mayfly-go/pkg/utils/collx"
	"mayfly-go/pkg/utils/jsonx"
	"mayfly-go/pkg/utils/stringx"

	"github.com/cloudwego/eino/adk"
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

// AppendMsgs 追加消息到会话历史
func (s *sessionAppImpl) AppendMsgs(ctx context.Context, sessionKey string, msgs ...adk.Message) error {
	if len(msgs) == 0 {
		return nil
	}

	messages := collx.ArrayMap(msgs, func(msg adk.Message) *entity.SessionMessage {
		sm := &entity.SessionMessage{
			SessionKey: sessionKey,
			TurnId:     cmp.Or(agent.GetTurnId(msg), stringx.RandUUID()),
			Role:       string(msg.Role),
			Content:    msg.Content,
			ToolCalls:  jsonx.ToStr(msg.ToolCalls),
			ActionId:   agent.GetActionId(msg),
		}
		extra := collx.M(msg.Extra)
		if msg.Role == schema.Tool {
			extra.Set("toolName", msg.ToolName)
		}
		extra.Delete("reasoning-content") // 思考内容可能很多
		sm.Extra = extra
		return sm
	})

	return s.sessionMessageRepo.BatchInsert(ctx, messages)
}

// GetHistory 获取会话历史消息
func (s *sessionAppImpl) GetHistory(ctx context.Context, sessionKey string, limit int) ([]adk.Message, error) {
	if limit <= 0 {
		limit = 1000
	}
	messages, err := s.sessionMessageRepo.SelectHistory(ctx, sessionKey, limit)
	if err != nil {
		return nil, err
	}
	return collx.ArrayMap(messages, func(msg *entity.SessionMessage) adk.Message {
		sm := &schema.Message{
			Role:    schema.RoleType(msg.Role),
			Content: msg.Content,
		}
		if msg.Role == string(schema.Tool) {
			sm.ToolCallID = msg.ActionId
		}
		if msg.ToolCalls != "" {
			tollcalls, _ := jsonx.ToByStr[[]schema.ToolCall](msg.ToolCalls)
			sm.ToolCalls = *tollcalls
		}
		return sm
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
