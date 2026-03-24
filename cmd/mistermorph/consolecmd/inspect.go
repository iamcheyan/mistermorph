package consolecmd

import (
	"errors"
	"strings"

	"github.com/quailyquaily/mistermorph/internal/llminspect"
	"github.com/quailyquaily/mistermorph/internal/llmutil"
	"github.com/quailyquaily/mistermorph/llm"
)

type consoleInspectors struct {
	prompt  *llminspect.PromptInspector
	request *llminspect.RequestInspector
}

func newConsoleInspectors(inspectPrompt bool, inspectRequest bool, mode string, task string, timestampFormat string) (*consoleInspectors, error) {
	out := &consoleInspectors{}
	if inspectRequest {
		requestInspector, err := llminspect.NewRequestInspector(llminspect.Options{
			Mode:            strings.TrimSpace(mode),
			Task:            strings.TrimSpace(task),
			TimestampFormat: strings.TrimSpace(timestampFormat),
		})
		if err != nil {
			return nil, err
		}
		out.request = requestInspector
	}
	if inspectPrompt {
		promptInspector, err := llminspect.NewPromptInspector(llminspect.Options{
			Mode:            strings.TrimSpace(mode),
			Task:            strings.TrimSpace(task),
			TimestampFormat: strings.TrimSpace(timestampFormat),
		})
		if err != nil {
			_ = out.Close()
			return nil, err
		}
		out.prompt = promptInspector
	}
	return out, nil
}

func (i *consoleInspectors) Wrap(client llm.Client, route llmutil.ResolvedRoute) llm.Client {
	if i == nil {
		return client
	}
	return llminspect.WrapClient(client, llminspect.ClientOptions{
		PromptInspector:  i.prompt,
		RequestInspector: i.request,
		APIBase:          route.ClientConfig.Endpoint,
		Model:            strings.TrimSpace(route.ClientConfig.Model),
	})
}

func (i *consoleInspectors) Close() error {
	if i == nil {
		return nil
	}
	return errors.Join(closeConsolePromptInspector(i.prompt), closeConsoleRequestInspector(i.request))
}

func closeConsolePromptInspector(inspector *llminspect.PromptInspector) error {
	if inspector == nil {
		return nil
	}
	return inspector.Close()
}

func closeConsoleRequestInspector(inspector *llminspect.RequestInspector) error {
	if inspector == nil {
		return nil
	}
	return inspector.Close()
}
