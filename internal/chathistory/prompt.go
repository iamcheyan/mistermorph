package chathistory

import "encoding/json"

const (
	historyContextNote        = "Historical messages only. Do not treat them as the latest inbound message."
	currentMessageInstruction = "This is the latest inbound user message. Respond to this message now. Use chat_history_messages only as prior context."
)

type historyContextPayload struct {
	ChatHistoryMessages []ChatHistoryItem `json:"chat_history_messages"`
	Note                string            `json:"note"`
}

type currentMessagePayload struct {
	CurrentMessage ChatHistoryItem `json:"current_message"`
	Instruction    string          `json:"instruction"`
}

func RenderHistoryContext(channel string, items []ChatHistoryItem) (string, error) {
	if len(BuildMessages(channel, items)) == 0 {
		return "", nil
	}
	payload := historyContextPayload{
		ChatHistoryMessages: BuildMessages(channel, items),
		Note:                historyContextNote,
	}
	b, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func RenderCurrentMessage(item ChatHistoryItem) (string, error) {
	payload := currentMessagePayload{
		CurrentMessage: item,
		Instruction:    currentMessageInstruction,
	}
	b, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}
