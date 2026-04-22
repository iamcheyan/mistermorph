package chatcmd

import (
	"fmt"
	"io"
	"os"
	"os/user"
	"regexp"
	"strings"

	"golang.org/x/term"
)

const chatBanner = `

▄▄   ▄▄  ▄▄▄  ▄▄▄▄  ▄▄▄▄  ▄▄ ▄▄
██▀▄▀██ ██▀██ ██▄█▄ ██▄█▀ ██▄██
██   ██ ▀███▀ ██ ██ ██    ██ ██
`

func buildUserName() string {
	userName := ""
	if u, err := user.Current(); err == nil && u != nil {
		userName = strings.TrimSpace(u.Username)
	}
	if userName == "" {
		userName = strings.TrimSpace(os.Getenv("USER"))
	}
	if userName == "" {
		userName = "you"
	}
	return userName
}

func buildUserPrompt(compactMode bool, userName string) string {
	if compactMode {
		return "\033[32m• \033[0m"
	}
	return fmt.Sprintf("\033[42m\033[30m %s> \033[0m ", userName)
}

func thinkingAnimation(writer io.Writer) (stop func(), setMessage func(msg string)) {
	if !isTerminalWriter(writer) {
		return func() {}, func(string) {}
	}

	currentMsg := "assistant is thinking..."
	render := func(msg string) {
		renderMsg := formatThinkingMessageForTerminal(msg)
		_, _ = fmt.Fprintf(writer, "\r\033[K\033[90m%s\033[0m", renderMsg)
	}

	render(currentMsg)
	stop = func() {
		_, _ = fmt.Fprint(writer, "\r\033[K")
	}
	setMessage = func(newMsg string) {
		normalized := normalizeThinkingMessage(newMsg)
		if normalized == currentMsg {
			return
		}
		currentMsg = normalized
		render(currentMsg)
	}
	return stop, setMessage
}

var thinkingWhitespaceRe = regexp.MustCompile(`\s+`)

type fdWriter interface {
	Fd() uintptr
}

func isTerminalWriter(writer io.Writer) bool {
	fw, ok := writer.(fdWriter)
	if !ok {
		return false
	}
	return term.IsTerminal(int(fw.Fd()))
}

func normalizeThinkingMessage(msg string) string {
	msg = strings.TrimSpace(msg)
	if msg == "" {
		return "assistant is thinking..."
	}
	msg = thinkingWhitespaceRe.ReplaceAllString(msg, " ")
	return strings.TrimSpace(msg)
}

func formatThinkingMessageForTerminal(msg string) string {
	msg = normalizeThinkingMessage(msg)
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || width <= 0 {
		width = 80
	}

	available := width - 4 // spinner + space + a little margin to avoid terminal wrapping
	if available < 20 {
		available = 20
	}
	return truncateDisplayWidth(msg, available)
}

func printChatSessionHeader(writer io.Writer, compactMode bool, model string, fileCacheDir string) {
	if !compactMode {
		_, _ = fmt.Fprint(writer, chatBanner)
	}
	if model != "" {
		_, _ = fmt.Fprintf(writer, "model=%s\n", displayModelName(model))
	}
	if fileCacheDir != "" {
		_, _ = fmt.Fprintf(writer, "file_cache_dir=%s\n", fileCacheDir)
	}
	if !compactMode {
		_, _ = fmt.Fprintln(writer, "\033[90mInteractive chat started. Press Ctrl+C or type /exit to quit.\033[0m")
	}
}

func displayModelName(model string) string {
	model = strings.TrimSpace(model)
	if model == "" {
		return ""
	}
	if idx := strings.LastIndex(model, "/"); idx >= 0 && idx+1 < len(model) {
		return strings.TrimSpace(model[idx+1:])
	}
	return model
}
