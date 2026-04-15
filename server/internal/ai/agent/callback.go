package agent

import (
	"context"

	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components/model"
	template "github.com/cloudwego/eino/utils/callbacks"
)

var logCallback = template.NewHandlerHelper().ChatModel(&template.ModelCallbackHandler{
	OnStart: func(ctx context.Context, info *callbacks.RunInfo, input *model.CallbackInput) context.Context {
		// logx.DebugfContext(ctx, "ChatModel %s started with input: %v\n", info.Name, input)
		return ctx
	},
}).Handler()
