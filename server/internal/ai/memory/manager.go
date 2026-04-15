package memory

type Manager struct {
	store Store
}

func NewManager(store Store) *Manager {
	return &Manager{
		store: store,
	}
}
