package grouptrigger

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/quailyquaily/mistermorph/internal/jsonutil"
	"github.com/quailyquaily/mistermorph/llm"
	"github.com/quailyquaily/mistermorph/tools"
)

type Decision struct {
	Reason            string
	UsedAddressingLLM bool

	AddressingLLMAttempted bool
	AddressingLLMOK        bool
	Addressing             Addressing
}

type Addressing struct {
	Addressed      bool
	Confidence     float64
	WannaInterject bool
	Interject      float64
	Impulse        float64
	IsLightweight  bool
	Reason         string
}

type AddressingFunc func(ctx context.Context) (Addressing, bool, error)

type DecideOptions struct {
	Mode                     string
	ConfidenceThreshold      float64
	InterjectThreshold       float64
	ExplicitReason           string
	ExplicitMatched          bool
	AddressingFallbackReason string
	AddressingTimeout        time.Duration
	Addressing               AddressingFunc
}

type LLMDecisionOptions struct {
	Client         llm.Client
	Model          string
	SystemPrompt   string
	UserPrompt     string
	AddressingTool tools.Tool
	MaxToolRounds  int
}

func Decide(ctx context.Context, opts DecideOptions) (Decision, bool, error) {
	mode := strings.ToLower(strings.TrimSpace(opts.Mode))
	if mode == "" {
		mode = "smart"
	}

	confidenceThreshold := clamp01(opts.ConfidenceThreshold)

	interjectThreshold := clamp01(opts.InterjectThreshold)

	if opts.ExplicitMatched {
		return Decision{
			Reason: strings.TrimSpace(opts.ExplicitReason),
			Addressing: Addressing{
				Impulse: 1,
			},
		}, true, nil
	}

	if mode != "talkative" && mode != "smart" {
		return Decision{}, false, nil
	}

	dec := Decision{
		AddressingLLMAttempted: true,
		Reason:                 strings.TrimSpace(opts.AddressingFallbackReason),
	}
	if opts.Addressing == nil {
		return dec, false, nil
	}

	addrCtx := ctx
	if addrCtx == nil {
		addrCtx = context.Background()
	}
	cancel := func() {}
	if opts.AddressingTimeout > 0 {
		addrCtx, cancel = context.WithTimeout(addrCtx, opts.AddressingTimeout)
	}
	llmDec, llmOK, llmErr := opts.Addressing(addrCtx)
	cancel()
	if llmErr != nil {
		return dec, false, llmErr
	}
	llmDec = normalizeAddressing(llmDec)

	dec.AddressingLLMOK = llmOK
	dec.Addressing = llmDec
	if llmDec.Reason != "" {
		dec.Reason = llmDec.Reason
	}
	if !llmOK {
		return dec, false, nil
	}

	switch mode {
	case "smart":
		if llmDec.Addressed && llmDec.Confidence >= confidenceThreshold {
			dec.UsedAddressingLLM = true
			return dec, true, nil
		}
	case "talkative":
		if llmDec.WannaInterject && llmDec.Interject > interjectThreshold {
			dec.UsedAddressingLLM = true
			return dec, true, nil
		}
	}
	return dec, false, nil
}

// DecideViaLLM is a reusable addressing-LLM runner used by channel triggers.
// It executes optional tool-calls, parses strict JSON output, and normalizes fields
// into the shared Addressing shape.
func DecideViaLLM(ctx context.Context, opts LLMDecisionOptions) (Addressing, bool, error) {
	maxToolRounds := opts.MaxToolRounds
	if maxToolRounds <= 0 {
		maxToolRounds = 3
	}

	messages := []llm.Message{
		{Role: "system", Content: opts.SystemPrompt},
		{Role: "user", Content: opts.UserPrompt},
	}
	var llmTools []llm.Tool
	if t := llmToolFromTool(opts.AddressingTool); t != nil {
		llmTools = append(llmTools, *t)
	}

	reactionCalled := false
	var out addressingLLMOutput
	for round := 0; ; round++ {
		res, err := opts.Client.Chat(ctx, llm.Request{
			Model:     opts.Model,
			ForceJSON: true,
			Messages:  messages,
			Tools:     llmTools,
		})
		if err != nil {
			return Addressing{}, false, err
		}

		if len(res.ToolCalls) > 0 {
			if round >= maxToolRounds {
				return Addressing{}, false, fmt.Errorf("addressing_llm exceeded tool-call rounds")
			}
			messages = append(messages, llm.Message{
				Role:      "assistant",
				Content:   strings.TrimSpace(res.Text),
				ToolCalls: res.ToolCalls,
			})
			for _, tc := range res.ToolCalls {
				observation, executedOK := executeAddressingToolCall(ctx, opts.AddressingTool, tc)
				if executedOK {
					reactionCalled = true
				}
				if strings.TrimSpace(tc.ID) != "" {
					messages = append(messages, llm.Message{
						Role:       "tool",
						ToolCallID: tc.ID,
						Content:    observation,
					})
					continue
				}
				messages = append(messages, llm.Message{
					Role:    "user",
					Content: fmt.Sprintf("Tool Result (%s):\n%s", strings.TrimSpace(tc.Name), observation),
				})
			}
			continue
		}

		raw := strings.TrimSpace(res.Text)
		if raw == "" {
			return Addressing{}, false, fmt.Errorf("empty addressing_llm response")
		}
		if err := jsonutil.DecodeWithFallback(raw, &out); err != nil {
			return Addressing{}, false, fmt.Errorf("invalid addressing_llm json")
		}
		if out.IsLightweight && strings.TrimSpace(out.Reaction) != "" && !reactionCalled {
			toolName := ""
			if opts.AddressingTool != nil {
				toolName = strings.TrimSpace(opts.AddressingTool.Name())
			}
			observation, executedOK := executeAddressingToolCall(ctx, opts.AddressingTool, llm.ToolCall{
				Name:      toolName,
				Arguments: map[string]any{"emoji": strings.TrimSpace(out.Reaction)},
			})
			if !executedOK {
				return Addressing{}, false, fmt.Errorf("lightweight response requires reaction tool: %s", strings.TrimSpace(observation))
			}
		}
		break
	}

	wannaInterject := out.Interject > 0
	if out.WannaInterject != nil {
		wannaInterject = bool(*out.WannaInterject)
	}
	addressing := Addressing{
		Addressed:      out.Addressed,
		Confidence:     clamp01(out.Confidence),
		WannaInterject: wannaInterject,
		Interject:      clamp01(out.Interject),
		Impulse:        clamp01(out.Impulse),
		IsLightweight:  out.IsLightweight,
		Reason:         strings.TrimSpace(out.Reason),
	}
	return addressing, true, nil
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func normalizeAddressing(in Addressing) Addressing {
	in.Confidence = clamp01(in.Confidence)
	in.Interject = clamp01(in.Interject)
	in.Impulse = clamp01(in.Impulse)
	in.Reason = strings.TrimSpace(in.Reason)
	return in
}

type addressingLLMOutput struct {
	Addressed      bool                      `json:"addressed"`
	Confidence     float64                   `json:"confidence"`
	WannaInterject *addressingWannaInterject `json:"wanna_interject,omitempty"`
	Interject      float64                   `json:"interject"`
	Impulse        float64                   `json:"impulse"`
	IsLightweight  bool                      `json:"is_lightweight"`
	Reaction       string                    `json:"reaction"`
	Reason         string                    `json:"reason"`
}

type addressingWannaInterject bool

// addressingWannaInterject accepts bool, number, or numeric-string forms.
func (w *addressingWannaInterject) UnmarshalJSON(data []byte) error {
	raw := strings.TrimSpace(string(data))
	if raw == "" || raw == "null" {
		*w = false
		return nil
	}

	var b bool
	if err := json.Unmarshal(data, &b); err == nil {
		*w = addressingWannaInterject(b)
		return nil
	}

	var f float64
	if err := json.Unmarshal(data, &f); err == nil {
		*w = addressingWannaInterject(f >= 0.5)
		return nil
	}

	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		normalized := strings.ToLower(strings.TrimSpace(s))
		switch normalized {
		case "true":
			*w = true
			return nil
		case "false":
			*w = false
			return nil
		}
		fv, parseErr := strconv.ParseFloat(normalized, 64)
		if parseErr != nil {
			return fmt.Errorf("unsupported wanna_interject value: %q", s)
		}
		*w = addressingWannaInterject(fv >= 0.5)
		return nil
	}

	return fmt.Errorf("unsupported wanna_interject json: %s", raw)
}

func llmToolFromTool(t tools.Tool) *llm.Tool {
	if t == nil {
		return nil
	}
	name := strings.TrimSpace(t.Name())
	if name == "" {
		return nil
	}
	return &llm.Tool{
		Name:           name,
		Description:    strings.TrimSpace(t.Description()),
		ParametersJSON: strings.TrimSpace(t.ParameterSchema()),
	}
}

func executeAddressingToolCall(ctx context.Context, t tools.Tool, call llm.ToolCall) (string, bool) {
	name := strings.TrimSpace(call.Name)
	if !matchesAddressingToolName(t, name) {
		if name == "" {
			name = "<empty>"
		}
		return fmt.Sprintf("error: tool '%s' not found", name), false
	}
	observation, err := t.Execute(ctx, call.Arguments)
	if err != nil {
		if strings.TrimSpace(observation) == "" {
			observation = fmt.Sprintf("error: %s", err.Error())
		} else {
			observation = fmt.Sprintf("%s\n\nerror: %s", observation, err.Error())
		}
		return observation, false
	}
	if strings.TrimSpace(observation) == "" {
		return "ok", true
	}
	return observation, true
}

func matchesAddressingToolName(t tools.Tool, callName string) bool {
	if t == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(t.Name()), strings.TrimSpace(callName))
}
