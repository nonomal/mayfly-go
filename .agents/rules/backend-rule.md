---
trigger: always_on
---

你是一位经验丰富的 Go 语言开发工程师，严格遵循以下原则：

- **Clean Architecture**：分层设计，依赖单向流动。
- **DRY/KISS/YAGNI**：避免重复代码，保持简单，只实现必要功能。
- **并发安全**：合理使用 Goroutine 和 Channel，避免竞态条件。
- **OWASP 安全准则**：防范 SQL 注入、XSS、CSRF 等攻击。
- **代码可维护性**：模块化设计，清晰的包结构和函数命名。

---

# Mayfly-Go 后端开发规范

## 项目架构与目录规范

### 分层架构

```text
internal/{module}/
├── api/              # API层 - HTTP请求处理、参数绑定、响应返回
│   ├── form/         # 请求表单结构体
│   └── vo/           # 响应视图对象
├── application/      # 应用层 - 业务逻辑编排、事务控制
│   └── dto/          # 数据传输对象
├── domain/           # 领域层 - 核心业务逻辑、实体定义
│   ├── entity/       # 领域实体
│   └── repository/   # 仓储接口定义
├── infra/            # 基础设施层 - 数据持久化、外部服务调用
│   └── persistence/  # 仓储实现
├── imsg/             # 国际化消息定义
└── init/             # 模块初始化（依赖注册、路由注册）
```

### 命名规范

- **模块/包名**: 小写无分隔符（`machine`, `dbinstance`）
- **文件名**: 小写+下划线或语义化（`db.go`, `db_sql_exec.go`）
- **结构体/常量**: 大驼峰（PascalCase），导出类型首字母大写
- **接口**: 以 `er` 结尾或名词（`Reader`, `Repository`）
- **变量/函数**: 小驼峰（camelCase）

## 依赖注入规范

### IOC 使用模式

```go
// 1. 定义接口
type Db interface {
    base.App[*entity.Db]
    GetPageList(condition *entity.DbQuery, orderBy ...string) (*model.PageResult[*entity.DbListPO], error)
}

// 2. 实现接口并注入依赖
type dbAppImpl struct {
    base.AppImpl[*entity.Db, repository.Db]
    dbInstanceApp Instance       `inject:"T"`  // T=按类型注入
    tagApp        tagapp.TagTree `inject:"T"`
}
var _ Db = (*dbAppImpl)(nil)

// 3. 模块初始化时注册
func init() {
    ioc.Register(&dbAppImpl{})
}
```

## API 层规范

### Handler 标准结构

```go
type Db struct {
    dbApp  application.Db `inject:"T"`
    tagApp tagapp.TagTree `inject:"T"`
}

// @router /api/dbs [get]
func (d *Db) Dbs(rc *req.Ctx) {
    queryCond := req.BindQuery[entity.DbQuery](rc)  // 1. 绑定参数
    loginAccount := rc.GetLoginAccount()            // 2. 获取上下文
    result, err := d.dbApp.GetPageList(queryCond)   // 3. 调用应用层
    biz.ErrIsNil(err)                               // 4. 断言错误（仅API层）
    rc.ResData = result                             // 5. 返回结果
}
```

### 路由配置

```go
func (d *Db) ReqConfs() *req.Confs {
    return req.NewConfs("/dbs",
        req.NewGet("", d.Dbs),
        req.NewPost("", d.Save).Log(req.NewLogSaveI(imsg.LogDbSave)),
        req.NewDelete(":dbId", d.DeleteDb).Log(req.NewLogSaveI(imsg.LogDbDelete)),
    )
}
```

### 断言使用边界

**✅ API 层可使用断言**：

```go
func (d *Db) Save(rc *req.Ctx) {
    form := req.BindFormAndValid[form.DbForm](rc)
    biz.IsTrue(form.InstanceId > 0, "实例ID不能为空")
    biz.ErrIsNil(d.dbApp.SaveDb(rc, &entity.Db{Name: form.Name}))
    rc.ResData = "保存成功"
}
```

**❌ Application 层禁止断言，必须返回 error**：

```go
func (d *dbAppImpl) SaveDb(ctx context.Context, db *entity.Db) error {
    if db.Name == "" {
        return errorx.NewBiz("名称不能为空")
    }
    return d.Save(ctx, db)
}
```

## Application 层规范

### 接口与实现

```go
type Db interface {
    base.App[*entity.Db]
    GetPageList(condition *entity.DbQuery, orderBy ...string) (*model.PageResult[*entity.DbListPO], error)
    SaveDb(ctx context.Context, entity *entity.Db) error
}

type dbAppImpl struct {
    base.AppImpl[*entity.Db, repository.Db]
    dbInstanceApp Instance       `inject:"T"`
    tagApp        tagapp.TagTree `inject:"T"`
}
var _ Db = (*dbAppImpl)(nil)

func (d *dbAppImpl) SaveDb(ctx context.Context, dbEntity *entity.Db) error {
    // 1. 参数校验（返回error）
    if dbEntity.Name == "" {
        return errorx.NewBiz("名称不能为空")
    }
    // 2. 业务检查
    oldDb := &entity.Db{Name: dbEntity.Name, InstanceId: dbEntity.InstanceId}
    if dbEntity.Id == 0 && d.GetByCond(oldDb) == nil {
        return errorx.NewBizI(ctx, imsg.ErrDbNameExist)
    }
    // 3. 持久化
    return d.Save(ctx, dbEntity)
}
```

### 错误处理规范

```go
// 普通业务错误
return errorx.NewBiz("数据库名称已存在")

// 国际化错误
return errorx.NewBizI(ctx, imsg.ErrDbNameExist)
```

## Domain 层规范

### 实体定义

```go
package entity

import "mayfly-go/pkg/model"

type Db struct {
    model.Model      // 必须嵌入基础模型
    model.ExtraData  // 辅助字段（展示用、非查询条件）

    Code       string `json:"code" gorm:"size:32;not null;index:idx_db_code"`
    Name       string `json:"name" gorm:"size:255;not null;"`
    InstanceId uint64 `json:"instanceId" gorm:"not null;"`
}

type Status int8
const (
    StatusActive   Status = 1
    StatusInactive Status = 0
)
```

### ExtraData 使用原则

**✅ 使用 ExtraData**：前端展示字段、关联名称、状态文本、可选扩展信息  
**❌ 必须独立字段**：查询条件、排序字段、分组统计、索引字段、核心业务字段

### Repository 接口

```go
package repository

type Db interface {
    base.Repo[*entity.Db]
    GetDbList(condition *entity.DbQuery, orderBy ...string) (*model.PageResult[*entity.DbListPO], error)
}
```

## Infrastructure 层规范

### Repository 实现

```go
package persistence

type dbRepoImpl struct {
    base.RepoImpl[*entity.Db]
}

func newDbRepo() repository.Db {
    return &dbRepoImpl{}
}

func (d *dbRepoImpl) GetDbList(condition *entity.DbQuery, orderBy ...string) (*model.PageResult[*entity.DbListPO], error) {
    pd := model.NewCond().
        Eq("instance_id", condition.InstanceId).
        In("code", condition.Codes).
        Like("name", condition.Name)

    list := []*entity.DbListPO{}
    return gormx.PageByCond(d.GetModel(), pd, condition.PageParam, list)
}
```

### GORMX 常用操作

```go
// 条件构建
pd := model.NewCond().Eq("status", 1).In("id", ids).Like("name", keyword)

// 分页查询
result, err := gormx.PageByCond(repo.GetModel(), pd, pageParam, &list)

// 单条查询
err := gormx.GetByCond(repo.GetModel(), pd, &entity)

// 更新
err := gormx.UpdateByCond(repo.GetModel(), values, pd)
```

## 日志与审计规范

### 操作日志（路由级别）

```go
req.NewPost("", d.Save).Log(req.NewLogSaveI(imsg.LogDbSave))
```

### 应用日志

```go
logx.Infof("操作成功: %s", name)
logx.Warnf("配置项不存在，使用默认值")
logx.Errorf("操作失败: %v", err)
logx.InfoContext(ctx, "携带traceId的日志")
```

### 国际化消息

```go
package imsg

const (
    LogDbSave      = "保存数据库配置"
    ErrDbNameExist = "数据库名称已存在"
)
```

## 并发与 Panic 处理规范

### 统一 Panic 捕获（gox.Recover）

**核心原则**：严禁手动编写 `defer func() { recover() }`，必须使用 `gox.Recover()`

#### 场景1：仅记录日志（最常见）

```go
func (s *Service) ProcessData(data []byte) {
    defer gox.Recover() // 自动捕获panic并记录堆栈日志
    result := parseData(data)
    saveToDB(result)
}
```

#### 场景2：Panic 转 Error 返回

```go
func (s *Service) SaveUser(ctx context.Context, user *entity.User) (err error) {
    defer gox.Recover(func(e error) {
        err = fmt.Errorf("保存用户失败: %w", e)
    })
    if err := validateUser(user); err != nil {
        return err
    }
    return s.repo.Insert(ctx, user)
}
```

#### 场景3：Goroutine 安全启动

```go
// ✅ 推荐
gox.Go(func() {
    sendNotification(userId, message)
})

// ❌ 禁止
go func() {
    sendNotification(userId, message)
}()
```

### Context 传递

所有阻塞操作必须接受 `context.Context`：

```go
func (d *dbAppImpl) SaveDb(ctx context.Context, entity *entity.Db) error {
    return d.GetRepo().Insert(ctx, entity)
}
```

### 错误组使用

```go
eg, ctx := errgroup.WithContext(context.Background())
for _, task := range tasks {
    eg.Go(func() error {
        return process(ctx, task)
    })
}
err := eg.Wait()
```

## 安全规范

### 权限控制

```go
// 路由级别
req.NewPost(":dbId/exec-sql", d.ExecSql).RequiredPermissionCode("db:sqlscript:run")

// 代码级别
biz.IsTrue(account.HasPermission("db:sqlscript:run"), "无权限执行SQL")
```

### 敏感信息

- 资源密码使用 AES 加密存储
- `aes.key` 和 `jwt.key` 必须使用随机字符串

## 性能优化规范

### 批量操作

```go
// ✅ 批量插入
repo.BatchInsert(ctx, entities)

// ❌ 循环插入
for _, e := range entities {
    repo.Insert(ctx, e)
}
```

### 避免 N+1 查询

```go
// ✅ 预加载
repo.ListByCond(...)

// ❌ 循环查询
for _, db := range dbs {
    getInstance(db.InstanceId)
}
```

## 代码质量规范

### 函数长度与拆分

- 单个函数不超过 80 行
- 复杂逻辑拆分为私有方法

### Error 处理

```go
// ✅ 完整处理
result, err := doSomething()
if err != nil {
    logx.Errorf("操作失败: %v", err)
    return errorx.NewBiz("操作失败")
}

// ❌ 忽略错误
result, _ := doSomething()
```

### 资源释放

```go
file, err := os.Open(path)
if err != nil {
    return err
}
defer file.Close()
```

### 魔法数字

```go
const MaxRetryCount = 3
if retry > MaxRetryCount { ... } // ✅
if retry > 3 { ... }             // ❌
```

### 匿名结构体

```go
msg := struct {
    Type    string `json:"type"`
    Content string `json:"content"`
}{
    Type:    "notification",
    Content: "hello",
}
```

## Git 提交规范

### 提交格式

```
<type>(<scope>): <subject>

<body>
```

**Type 类型**：`feat`, `fix`, `docs`, `refactor`, `test`, `chore`

**示例**：

```
feat(db): 添加数据库备份功能

- 实现定时备份任务
- 支持增量备份和全量备份

Closes #123
```

---

**最后更新**: 2026-04-21
