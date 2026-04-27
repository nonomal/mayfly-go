package application

import (
	"context"
	"mayfly-go/internal/milvus/domain/entity"
	"mayfly-go/internal/milvus/domain/repository"
	"mayfly-go/internal/milvus/imsg"
	"mayfly-go/internal/milvus/mvm"
	tagapp "mayfly-go/internal/tag/application"
	tagdto "mayfly-go/internal/tag/application/dto"
	tagentity "mayfly-go/internal/tag/domain/entity"
	"mayfly-go/pkg/base"
	"mayfly-go/pkg/biz"
	"mayfly-go/pkg/errorx"
	"mayfly-go/pkg/model"
	"mayfly-go/pkg/req"
	"mayfly-go/pkg/utils/stringx"
)

type Milvus interface {
	base.App[*entity.Milvus]

	// 分页获取机器脚本信息列表
	GetPageList(condition *entity.MilvusQuery, orderBy ...string) (*model.PageResult[*entity.Milvus], error)

	TestConn(entity *entity.Milvus) error

	SaveMilvus(ctx context.Context, entity *entity.Milvus, tagCodePaths ...string) error

	Delete(ctx context.Context, id uint64) error

	// 获取Milvus连接实例
	GetMilvusConn(rc *req.Ctx) (*mvm.MilvusConn, error)
}

// MilvusApp Milvus 应用服务
type milvusAppImpl struct {
	base.AppImpl[*entity.Milvus, repository.Milvus]
	tagTreeApp tagapp.TagTree `inject:"T"`
}

var _ Milvus = (*milvusAppImpl)(nil)

// GetPageList 获取分页列表
func (a *milvusAppImpl) GetPageList(query *entity.MilvusQuery, orderBy ...string) (*model.PageResult[*entity.Milvus], error) {
	return a.GetRepo().GetList(query, orderBy...)
}

// TestConn 测试连接
func (a *milvusAppImpl) TestConn(milvus *entity.Milvus) error {
	conn, err := milvus.ToMilvusInfo().Conn()
	if err != nil {
		return err
	}
	// 尝试获取数据库列表
	_, err = conn.ListDatabases()
	if err != nil {
		return err
	}
	return nil
}

// SaveMilvus 保存 Milvus 实例

func (a *milvusAppImpl) SaveMilvus(ctx context.Context, m *entity.Milvus, tagCodePaths ...string) error {
	old := &entity.Milvus{Host: m.Host, SshTunnelMachineId: m.SshTunnelMachineId, Username: m.Username}
	err := a.GetByCond(old)

	if m.Id == 0 {
		if err == nil {
			return errorx.NewBizI(ctx, imsg.ErrInfoExist)
		}
		// 生成随机编号
		m.Code = stringx.Rand(10)

		return a.Tx(ctx,
			func(ctx context.Context) error {
				return a.Insert(ctx, m)
			},
			func(ctx context.Context) error {
				return a.tagTreeApp.SaveResourceTag(ctx, &tagdto.SaveResourceTag{
					ResourceTag: &tagdto.ResourceTag{
						Type: tagentity.TagTypeMilvus,
						Code: m.Code,
						Name: m.Name,
					},
					ParentTagCodePaths: tagCodePaths,
				})
			})
	}

	// 如果存在该库，则校验修改的库是否为该库
	if err == nil && old.Id != m.Id {
		return errorx.NewBizI(ctx, imsg.ErrInfoExist)
	}
	// 如果调整了ssh等会查不到旧数据，故需要根据id获取旧信息将code赋值给标签进行关联
	if old.Code == "" {
		old, _ = a.GetById(m.Id)
	}

	// 先关闭连接
	mvm.CloseAll(m.Id)
	m.Code = ""
	return a.Tx(ctx, func(ctx context.Context) error {
		return a.UpdateById(ctx, m)
	}, func(ctx context.Context) error {
		if old.Name != m.Name {
			if err := a.tagTreeApp.UpdateTagName(ctx, tagentity.TagTypeMilvus, old.Code, m.Name); err != nil {
				return err
			}
		}
		return a.tagTreeApp.SaveResourceTag(ctx, &tagdto.SaveResourceTag{
			ResourceTag: &tagdto.ResourceTag{
				Type: tagentity.TagTypeMilvus,
				Code: old.Code,
			},
			ParentTagCodePaths: tagCodePaths,
		})
	})
}

// Delete 删除 Milvus 实例
func (a *milvusAppImpl) Delete(ctx context.Context, id uint64) error {
	milvusEntity, err := a.GetById(id)
	if err != nil {
		return errorx.NewBiz("milvus not found")
	}

	mvm.CloseAll(milvusEntity.Id)
	return a.Tx(ctx,
		func(ctx context.Context) error {
			return a.DeleteById(ctx, id)
		},
		func(ctx context.Context) error {
			return a.tagTreeApp.SaveResourceTag(ctx, &tagdto.SaveResourceTag{ResourceTag: &tagdto.ResourceTag{
				Type: tagentity.TagTypeMilvus,
				Code: milvusEntity.Code,
			}})
		})
}

// GetMilvusClient 获取 Milvus 客户端
func (a *milvusAppImpl) GetMilvusConn(rc *req.Ctx) (*mvm.MilvusConn, error) {
	id := rc.PathParamInt("id")
	biz.IsTrue(id > 0, "milvusId error")

	db := rc.Query("db")
	if db == "" {
		m, err := a.GetById(uint64(id))
		if err != nil {
			return nil, err
		}
		db = m.Database
	}
	if db == "" {
		db = "default"
	}
	return mvm.GetMilvusConn(rc, uint64(id), db, func() (*mvm.MilvusInfo, error) {
		me, err := a.GetById(uint64(id))
		me.Database = db
		if err != nil {
			return nil, errorx.NewBiz("milvus not found")
		}
		return me.ToMilvusInfo(), nil
	})

}
