package grouptrigger

import (
	_ "embed"
	"encoding/json"
	"strings"
	"text/template"

	"github.com/quailyquaily/mistermorph/internal/chathistory"
	"github.com/quailyquaily/mistermorph/internal/prompttmpl"
)

//go:embed prompts/addressing_system.md
var addressingSystemPromptTemplateSource string

//go:embed prompts/addressing_user.md
var addressingUserPromptTemplateSource string

var addressingPromptTemplateFuncs = template.FuncMap{
	"toJSON": func(v any) (string, error) {
		b, err := json.Marshal(v)
		if err != nil {
			return "", err
		}
		return string(b), nil
	},
}

var addressingSystemPromptTemplate = prompttmpl.MustParse(
	"grouptrigger_addressing_system",
	addressingSystemPromptTemplateSource,
	nil,
)

var addressingUserPromptTemplate = prompttmpl.MustParse(
	"grouptrigger_addressing_user",
	addressingUserPromptTemplateSource,
	addressingPromptTemplateFuncs,
)

const AddressingPersonaFallback = "You are MisterMorph, a general-purpose AI agent that can use tools to complete tasks."

type addressingSystemPromptData struct {
	PersonaIdentity string
}

type addressingUserPromptData struct {
	CurrentMessage      any
	ChatHistoryMessages []chathistory.ChatHistoryItem
}

func RenderAddressingPrompts(personaIdentity string, currentMessage any, historyMessages []chathistory.ChatHistoryItem) (string, string, error) {
	personaIdentity = strings.TrimSpace(personaIdentity)
	if personaIdentity == "" {
		personaIdentity = AddressingPersonaFallback
	}

	systemPrompt, err := prompttmpl.Render(addressingSystemPromptTemplate, addressingSystemPromptData{
		PersonaIdentity: personaIdentity,
	})
	if err != nil {
		return "", "", err
	}
	userPrompt, err := prompttmpl.Render(addressingUserPromptTemplate, addressingUserPromptData{
		CurrentMessage:      currentMessage,
		ChatHistoryMessages: historyMessages,
	})
	if err != nil {
		return "", "", err
	}
	return systemPrompt, userPrompt, nil
}
