package api

import (
	"context"
	"mayfly-go/internal/ai/agent"
	"mayfly-go/internal/ai/api/form"
	"mayfly-go/internal/ai/api/vo"
	"mayfly-go/internal/ai/application"
	"mayfly-go/internal/ai/application/dto"
	"mayfly-go/internal/ai/domain/entity"
	"mayfly-go/internal/ai/tools"
	"mayfly-go/pkg/biz"
	"mayfly-go/pkg/logx"
	"mayfly-go/pkg/req"
	"mayfly-go/pkg/utils/anyx"
	"mayfly-go/pkg/utils/collx"
	"mayfly-go/pkg/utils/jsonx"
	"mayfly-go/pkg/validatorx"
	"mayfly-go/pkg/ws"
	"time"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
	"github.com/gorilla/websocket"
)

// Ai API结构体，用于处理AI相关请求
type Ai struct {
	sessionApp application.Session `inject:"T"`
}

// ReqConfs 获取AI相关的请求配置
func (a *Ai) ReqConfs() *req.Confs {
	reqs := [...]*req.Conf{
		req.NewGet("/chat", a.Chat).NoRes(),
		req.NewGet("/chat/sessions", a.ChatSessions),
		req.NewGet("/chat/messages", a.ChatMessages),
		req.NewDelete("/chat/sessions/:sessionKey", a.DeleteSession),
		req.NewPost("/chat/sessions/rename", a.RenameSessionTitle),
	}

	return req.NewConfs("/ai", reqs[:]...)
}

// ChatSessions 获取会话列表
func (a *Ai) ChatSessions(rc *req.Ctx) {
	sessions, err := a.sessionApp.ListSessions(rc.MetaCtx, &dto.SessionQuery{
		UserId: rc.GetLoginAccount().Id,
	})
	biz.ErrIsNil(err)
	rc.ResData = sessions
}

func (a *Ai) DeleteSession(rc *req.Ctx) {
	a.sessionApp.DeleteMeta(rc.MetaCtx, rc.PathParam("sessionKey"))
}

func (a *Ai) RenameSessionTitle(rc *req.Ctx) {
	rename := req.BindJson[struct {
		SessionKey string `json:"sessionKey"  binding:"required"`
		Title      string `json:"title"  binding:"required"`
	}](rc)
	biz.ErrIsNil(a.sessionApp.UpdateByCond(rc.MetaCtx, &entity.Session{Title: rename.Title}, &entity.Session{SessionKey: rename.SessionKey}))
}

// ChatMessages 获取会话消息
func (a *Ai) ChatMessages(rc *req.Ctx) {
	messages, err := a.sessionApp.ListSessionMessages(rc.MetaCtx, &dto.SessionMessageQuery{
		SessionKey: rc.Query("sessionKey"),
	})
	biz.ErrIsNil(err)
	rc.ResData = collx.ArrayMap(messages, func(msg *entity.SessionMessage) *vo.ChatMsg {
		cm := &vo.ChatMsg{
			TurnId:           msg.TurnId,
			Content:          msg.Content,
			ReasoningContent: msg.GetExtraString("reasoningContent"),
			Time:             *msg.CreateTime,
			Role:             msg.Role,
			ActionId:         msg.ActionId,
			Extra:            msg.Extra,
		}
		if msg.ToolCalls != "" {
			tollcalls, _ := jsonx.ToByStr[[]schema.ToolCall](msg.ToolCalls)
			cm.ToolCalls = *tollcalls
		}
		return cm
	})
}

// Chat 聊天
func (a *Ai) Chat(rc *req.Ctx) {
	wsConn, err := ws.Upgrader.Upgrade(rc.GetWriter(), rc.GetRequest(), nil)
	defer func() {
		if wsConn != nil {
			if err := recover(); err != nil {
				wsConn.WriteMessage(websocket.TextMessage, []byte(jsonx.ToStr(&vo.ChatMsg{Role: string(agent.RoleInternal), Type: "error", Content: anyx.ToString(err), Time: time.Now()})))
			}
			wsConn.Close()
		}
	}()
	biz.ErrIsNilAppendErr(err, "Upgrade websocket fail: %s")
	// 权限校验
	// rc = rc.WithRequiredPermission(req.NewPermission("ai:chat"))
	err = req.PermissionHandler(rc)
	biz.ErrIsNil(err)

	ctx := rc.MetaCtx
	ag, err := agent.GetDefaultAgent(ctx)
	biz.ErrIsNilAppendErr(err, "get agent error: %s")
	for {
		messageType, message, err := wsConn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logx.Debugf("error: %v", err)
			}
			break
		}

		now := time.Now()

		if messageType == websocket.TextMessage {
			chatMsg, err := jsonx.To[form.ChatMsg](message)
			biz.ErrIsNilAppendErr(err, "parse chat message error: %s")

			var userMessage []adk.Message
			agentRunOptions := []agent.RunOption{
				agent.WithRunSessionKey(chatMsg.SessionId),
				agent.WithOnChunk(func(ctx context.Context, m adk.Message) error {
					// 工具调用，不推送前端，等完整事件处理结束后推送，不然前端没法获取完整工具调用内容
					if len(m.ToolCalls) > 0 || m.Role == schema.Tool {
						return nil
					}
					respMsg := vo.ChatMsg{
						Type:             "chunk",
						Time:             now,
						Role:             string(m.Role),
						Content:          m.Content,
						ReasoningContent: m.ReasoningContent,
						ToolCalls:        m.ToolCalls,
					}
					wsConn.WriteMessage(websocket.TextMessage, []byte(jsonx.ToStr(respMsg)))
					return nil
				}),
				agent.WithOnEvent(func(ctx context.Context, ae *adk.AgentEvent, m adk.Message) error {
					if len(m.ToolCalls) > 0 || m.Role == schema.Tool || m.Role == agent.RoleInternal {
						respMsg := vo.ChatMsg{
							Type:             "tool",
							TurnId:           agent.GetTurnId(m),
							Time:             now,
							Role:             string(m.Role),
							Content:          m.Content,
							ReasoningContent: m.ReasoningContent,
							ToolCalls:        m.ToolCalls,
							ActionId:         agent.GetActionId(m),
							Extra:            m.Extra,
						}
						wsConn.WriteMessage(websocket.TextMessage, []byte(jsonx.ToStr(respMsg)))
					}
					return nil
				}),
			}

			if chatMsg.Type == form.ChatMsgTypeInterruptResume {
				ir, err := jsonx.ToByStr[tools.InterruptResume](chatMsg.Content)
				biz.ErrIsNil(err)
				biz.ErrIsNil(validatorx.Validate(ir))
				agentRunOptions = append(agentRunOptions,
					agent.WithResumeParams(ir),
					agent.WithTurnId(ir.TurnId),
				)
			} else {
				userMessage = collx.AsArray(schema.UserMessage(chatMsg.Content))
			}

			_, err = ag.Run(ctx, userMessage, agentRunOptions...)

			endMsg := &vo.ChatMsg{Role: string(agent.RoleInternal), Time: now}
			if err != nil {
				endMsg.Type = "error"
				endMsg.Content = err.Error()
			} else {
				endMsg.Type = "end"
			}
			wsConn.WriteMessage(websocket.TextMessage, []byte(jsonx.ToStr(endMsg)))
		}
	}
}
