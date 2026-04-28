package tools

import (
	"github.com/cloudwego/eino/schema"
)

func init() {
	schema.Register[*ApprovalInfo]()
	schema.Register[*InterruptResume]()
	schema.Register[*ApprovalResume]()
	schema.Register[*ParamCompletionResume]()
	schema.Register[[]CompletionParamInfo]()
	schema.Register[*ParamCompletionInterruptInfo]()
}
