package imsg

import "mayfly-go/pkg/i18n"

var Zh_CN = map[i18n.MsgId]string{
	LogSave:       "Milvus-保存",
	LogDelete:     "Milvus-删除",
	LogGetConn:    "Milvus-创建连接",
	LogUpdateDocs: "Milvus-更新文档",
	LogDelDocs:    "Milvus-删除文档",
	LogInsertDocs: "Milvus-插入文档",

	ErrInfoExist: "该信息已存在（host + ssh + username）",
}
