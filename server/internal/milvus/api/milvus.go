package api

import (
	"mayfly-go/internal/milvus/api/form"
	"mayfly-go/internal/milvus/api/vo"
	"mayfly-go/internal/milvus/application"
	"mayfly-go/internal/milvus/domain/entity"
	"mayfly-go/internal/milvus/imsg"
	"mayfly-go/internal/pkg/consts"
	tagapp "mayfly-go/internal/tag/application"
	tagentity "mayfly-go/internal/tag/domain/entity"
	"mayfly-go/pkg/biz"
	"mayfly-go/pkg/model"
	"mayfly-go/pkg/req"
	"mayfly-go/pkg/utils/collx"
	"strings"

	"github.com/spf13/cast"
)

// Milvus Milvus API
type Milvus struct {
	milvusApp  application.Milvus `inject:"T"`
	tagTreeApp tagapp.TagTree     `inject:"T"`
}

// ReqConfs 注册路由
func (m *Milvus) ReqConfs() *req.Confs {
	reqs := [...]*req.Conf{
		// 获取所有 milvus 列表
		req.NewGet("", m.Milvuses),

		// 测试连接
		req.NewPost("/test-conn", m.TestConn),

		// 保存
		req.NewPost("", m.Save).Log(req.NewLogSaveI(imsg.LogSave)),

		// 删除
		req.NewDelete(":id", m.DeleteById).Log(req.NewLogSaveI(imsg.LogDelete)),
	}

	return req.NewConfs("milvus", reqs[:]...)
}

// Milvuses 获取 Milvus 列表
func (m *Milvus) Milvuses(rc *req.Ctx) {
	queryCond := req.BindQuery[entity.MilvusQuery](rc)

	// 不存在可访问标签 id，即没有可操作数据
	tags := m.tagTreeApp.GetAccountTags(rc.GetLoginAccount().Id, &tagentity.TagTreeQuery{
		TypePaths:     collx.AsArray(tagentity.NewTypePaths(tagentity.TagTypeMilvus)),
		CodePathLikes: []string{queryCond.TagPath},
	})
	if len(tags) == 0 {
		rc.ResData = model.NewEmptyPageResult[any]()
		return
	}
	queryCond.Codes = tags.GetCodes()

	res, err := m.milvusApp.GetPageList(queryCond)
	biz.ErrIsNil(err)

	resVo := model.PageResultConv[*entity.Milvus, *vo.Milvus](res)
	vos := resVo.List

	// 填充标签信息
	m.tagTreeApp.FillTagInfo(tagentity.TagType(consts.ResourceTypeMilvus), collx.ArrayMap(vos, func(mvo *vo.Milvus) tagentity.ITagResource {
		return mvo
	})...)

	rc.ResData = resVo
}

// TestConn 测试连接
func (m *Milvus) TestConn(rc *req.Ctx) {
	_, milvus := req.BindJsonAndCopyTo[form.Milvus, entity.Milvus](rc)
	biz.ErrIsNilAppendErr(m.milvusApp.TestConn(milvus), "connection error: %s")
}

// Save 保存
func (m *Milvus) Save(rc *req.Ctx) {
	f, milvus := req.BindJsonAndCopyTo[form.Milvus, entity.Milvus](rc)

	// 密码脱敏记录日志
	f.Password = "***"
	rc.ReqParam = f

	biz.ErrIsNil(m.milvusApp.SaveMilvus(rc.MetaCtx, milvus, f.TagCodePaths...))
}

// DeleteById 删除
func (m *Milvus) DeleteById(rc *req.Ctx) {
	idsStr := rc.PathParam("id")
	rc.ReqParam = idsStr
	ids := strings.Split(idsStr, ",")

	for _, v := range ids {
		m.milvusApp.Delete(rc.MetaCtx, cast.ToUint64(v))
	}
}
