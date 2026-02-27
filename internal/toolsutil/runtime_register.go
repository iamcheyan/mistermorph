package toolsutil

import (
	"github.com/quailyquaily/mistermorph/llm"
	"github.com/quailyquaily/mistermorph/tools"
)

type RuntimeToolsRegisterConfig struct {
	PlanCreate PlanCreateRegisterConfig
	TodoUpdate TodoUpdateRegisterConfig
}

func LoadRuntimeToolsRegisterConfigFromViper() RuntimeToolsRegisterConfig {
	return RuntimeToolsRegisterConfig{
		PlanCreate: LoadPlanCreateRegisterConfigFromViper(),
		TodoUpdate: LoadTodoUpdateRegisterConfigFromViper(),
	}
}

func RegisterRuntimeTools(reg *tools.Registry, cfg RuntimeToolsRegisterConfig, client llm.Client, model string) {
	if reg == nil {
		return
	}
	RegisterPlanTool(reg, cfg.PlanCreate, client, model)
	RegisterTodoUpdateTool(reg, cfg.TodoUpdate, client, model)
}
