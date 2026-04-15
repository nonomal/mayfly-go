package persistence

import (
	"context"
	"mayfly-go/internal/ai/domain/entity"
	"mayfly-go/internal/ai/domain/repository"
	"mayfly-go/pkg/base"
)

type sessionMessageRepoImpl struct {
	base.RepoImpl[*entity.SessionMessage]
}

func newSessionMessageRepo() repository.SessionMessage {
	return &sessionMessageRepoImpl{}
}

func (s *sessionMessageRepoImpl) SelectHistory(ctx context.Context, sessionKey string, limit int) ([]*entity.SessionMessage, error) {
	var messages []*entity.SessionMessage
	if err := s.SelectBySql("select * from t_ai_session_message where session_key = ? order by id asc limit ?", &messages, sessionKey, limit); err != nil {
		return nil, err
	}
	return messages, nil
}
