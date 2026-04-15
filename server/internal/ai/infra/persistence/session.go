package persistence

import (
	"mayfly-go/internal/ai/domain/entity"
	"mayfly-go/internal/ai/domain/repository"
	"mayfly-go/pkg/base"
)

type sessionRepoImpl struct {
	base.RepoImpl[*entity.Session]
}

func newSessionRepo() repository.Session {
	return &sessionRepoImpl{}
}
