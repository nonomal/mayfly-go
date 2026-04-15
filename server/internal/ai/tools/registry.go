package tools

import (
	"context"
	"mayfly-go/pkg/utils/collx"

	"github.com/cloudwego/eino/components/tool"
)

type ToolType string

const (
	ToolTypeDb   ToolType = "db"
	ToolTypeFlow ToolType = "flow"
)

// Registry 工具注册中心
type Registry struct {
	tools collx.SM[string, tool.BaseTool] // 工具注册表
}

func NewRegistry() *Registry {
	return &Registry{}
}

// DefaultRegistry 默认工具注册中心实例
var DefaultRegistry = NewRegistry()

// Register 注册工具到注册中心
func (r *Registry) Register(tools ...tool.BaseTool) error {
	for _, t := range tools {
		ti, err := t.Info(context.Background())
		if err != nil {
			return err
		}
		r.tools.Store(ti.Name, t)
	}
	return nil
}

// Get 获取工具
func (r *Registry) Get(name string) (tool.BaseTool, bool) {
	return r.tools.Load(name)
}

// GetAll 获取所有工具
func (r *Registry) GetAll() []tool.BaseTool {
	return r.tools.Values()
}

// Clear 清空工具注册表
func (r *Registry) Clear() {
	r.tools.Clear()
}

// toolRegistry 工具注册中心，一个ToolType对应多个工具
var toolRegistry collx.SM[ToolType, []tool.BaseTool]

// RegisterTool 注册agent工具
func RegisterTool(toolType ToolType, tool ...tool.BaseTool) {
	if tools, exist := toolRegistry.Load(toolType); exist {
		toolRegistry.Store(toolType, append(tools, tool...))
	} else {
		toolRegistry.Store(toolType, tool)
	}
}

// GetTools 获取指定类型的所有工具
func GetTools(toolType ToolType) ([]tool.BaseTool, bool) {
	return toolRegistry.Load(toolType)
}

// GetAllTools 获取所有已注册的工具
func GetAllTools() []tool.BaseTool {
	var allTools []tool.BaseTool
	toolRegistry.Range(func(key ToolType, tools []tool.BaseTool) bool {
		allTools = append(allTools, tools...)
		return true
	})
	return allTools
}

// GetToolsByTypes 获取指定类型的多个工具
func GetToolsByTypes(types []ToolType) map[ToolType][]tool.BaseTool {
	result := make(map[ToolType][]tool.BaseTool)
	for _, t := range types {
		if tools, exists := toolRegistry.Load(t); exists {
			result[t] = tools
		}
	}
	return result
}

// ClearTools 清空指定类型的工具
func ClearTools(toolType ToolType) {
	toolRegistry.Delete(toolType)
}
