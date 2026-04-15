package migrations

import (
	aientity "mayfly-go/internal/ai/domain/entity"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func V1_11() []*gormigrate.Migration {
	var migrations []*gormigrate.Migration
	migrations = append(migrations, V1_11_0()...)
	return migrations
}

func V1_11_0() []*gormigrate.Migration {
	return []*gormigrate.Migration{
		{
			ID: "v1.11.0",
			Migrate: func(tx *gorm.DB) error {
				err := tx.AutoMigrate(&aientity.Session{},
					&aientity.SessionMessage{},
				)
				if err != nil {
					return err
				}
				tx.Exec(`INSERT INTO t_sys_resource (id, pid, ui_path, type, status, name, code, weight, meta, creator_id, creator, modifier_id, modifier, create_time, update_time, is_deleted, delete_time) VALUES (1775967861, 0, 'lUgXhO96/', 1, 1, 'AI', '/ai', 20000000, '{"icon":"icon ai/ai","isKeepAlive":true,"routeName":"ai"}', 1, 'admin', 1, 'admin', '2026-04-12 12:24:22', '2026-04-12 17:08:47', 0, NULL)`)
				tx.Exec(`INSERT INTO t_sys_resource (id, pid, ui_path, type, status, name, code, weight, meta, creator_id, creator, modifier_id, modifier, create_time, update_time, is_deleted, delete_time) VALUES (1775967903, 1775967861, 'lUgXhO96/VPfzI5pQ/', 1, 1, 'menu.aiAssistant', 'assistant', 1775967903, '{"icon":"icon ai/assistant","isKeepAlive":true,"routeName":"AiAssistant"}', 1, 'admin', 1, 'admin', '2026-04-12 12:25:03', '2026-04-12 17:05:49', 0, NULL)`)
				return nil
			},
			Rollback: func(tx *gorm.DB) error {
				return nil
			},
		},
	}
}
