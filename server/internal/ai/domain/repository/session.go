package repository

import (
	"mayfly-go/internal/ai/domain/entity"
	"mayfly-go/pkg/base"
)

type Session interface {
	base.Repo[*entity.Session]
}
