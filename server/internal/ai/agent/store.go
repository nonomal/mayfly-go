package agent

import (
	"context"

	"github.com/cloudwego/eino/compose"
)

type CheckPointStore interface {
	compose.CheckPointStore

	Delete(ctx context.Context, key string) error
}

var checkPointStore CheckPointStore

func GetDefaultCheckPointStore() CheckPointStore {
	if checkPointStore != nil {
		return checkPointStore
	}
	checkPointStore = NewInMemoryStore()
	return checkPointStore
}

func NewInMemoryStore() CheckPointStore {
	return &inMemoryStore{
		mem: map[string][]byte{},
	}
}

type inMemoryStore struct {
	mem map[string][]byte
}

var _ CheckPointStore = (*inMemoryStore)(nil)

func (i *inMemoryStore) Set(ctx context.Context, key string, value []byte) error {
	i.mem[key] = value
	return nil
}

func (i *inMemoryStore) Get(ctx context.Context, key string) ([]byte, bool, error) {
	v, ok := i.mem[key]
	return v, ok, nil
}

func (i *inMemoryStore) Delete(ctx context.Context, key string) error {
	delete(i.mem, key)
	return nil
}
