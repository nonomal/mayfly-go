package repository

import (
	"context"
	"mayfly-go/internal/ai/domain/entity"
	"mayfly-go/pkg/base"
)

type SessionMessage interface {
	base.Repo[*entity.SessionMessage]

	SelectHistory(ctx context.Context, sessionKey string, limit int) ([]*entity.SessionMessage, error)
}
