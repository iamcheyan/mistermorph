package promptprofile

import (
	_ "embed"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/quailyquaily/mistermorph/agent"
	"github.com/quailyquaily/mistermorph/internal/prompttmpl"
	"github.com/quailyquaily/mistermorph/internal/statepaths"
	"github.com/quailyquaily/mistermorph/tools"
)

//go:embed prompts/block_plan_create.md
var planCreateBlockTemplateSource string

//go:embed prompts/block_local_tool_notes.md
var localToolNotesBlockTemplateSource string

//go:embed prompts/block_memory_summaries.md
var memorySummariesBlockTemplateSource string

//go:embed prompts/block_telegram_group_usernames.md
var groupUsernamesBlockTemplateSource string

//go:embed prompts/block_slack_mention_users.md
var slackMentionUsersBlockTemplateSource string

//go:embed prompts/block_telegram.md
var telegramRuntimePromptBlockTemplateSource string

//go:embed prompts/block_slack.md
var slackRuntimePromptBlockTemplateSource string

var localToolNotesBlockTemplate = prompttmpl.MustParse(
	"local_tool_notes_block",
	localToolNotesBlockTemplateSource,
	template.FuncMap{},
)

var memorySummariesBlockTemplate = prompttmpl.MustParse(
	"memory_summaries_block",
	memorySummariesBlockTemplateSource,
	template.FuncMap{},
)

var groupUsernamesBlockTemplate = prompttmpl.MustParse(
	"group_usernames_block",
	groupUsernamesBlockTemplateSource,
	template.FuncMap{},
)

var slackMentionUsersBlockTemplate = prompttmpl.MustParse(
	"slack_mention_users_block",
	slackMentionUsersBlockTemplateSource,
	template.FuncMap{},
)

var telegramRuntimePromptBlockTemplate = prompttmpl.MustParse(
	"telegram_runtime_block",
	telegramRuntimePromptBlockTemplateSource,
	template.FuncMap{},
)

var slackRuntimePromptBlockTemplate = prompttmpl.MustParse(
	"slack_runtime_block",
	slackRuntimePromptBlockTemplateSource,
	template.FuncMap{},
)

type telegramRuntimePromptBlockData struct {
	IsGroup   bool
	EmojiList string
}

type slackRuntimePromptBlockData struct {
	IsGroup   bool
	EmojiList string
}

type localToolNotesPromptBlockData struct {
	ScriptsNotes string
}

type memorySummariesPromptBlockData struct {
	Content string
}

type groupUsernamesPromptBlockData struct {
	UsernamesText string
}

type slackMentionUsersPromptBlockData struct {
	UserIDsText string
}

func AppendPlanCreateGuidanceBlock(spec *agent.PromptSpec, registry *tools.Registry) {
	if _, ok := registry.Get("plan_create"); !ok {
		return
	}
	content := strings.TrimSpace(planCreateBlockTemplateSource)
	if content == "" {
		return
	}
	spec.Blocks = append(spec.Blocks, agent.PromptBlock{
		Content: content,
	})
}

func AppendLocalToolNotesBlock(spec *agent.PromptSpec, log *slog.Logger) {
	if log == nil {
		log = slog.Default()
	}

	path := filepath.Join(statepaths.FileStateDir(), "SCRIPTS.md")
	raw, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Warn("prompt_local_tool_notes_load_failed", "path", path, "error", err.Error())
		}
		return
	}

	content := strings.TrimSpace(string(raw))
	if content == "" {
		content = "No local tool notes available."
	}
	content, err = prompttmpl.Render(localToolNotesBlockTemplate, localToolNotesPromptBlockData{
		ScriptsNotes: content,
	})
	if err != nil {
		log.Warn("prompt_local_tool_notes_render_failed", "path", path, "error", err.Error())
		return
	}
	content = strings.TrimSpace(content)
	if content == "" {
		return
	}

	spec.Blocks = append(spec.Blocks, agent.PromptBlock{
		Content: content,
	})
	log.Info("prompt_local_tool_notes_applied", "path", path, "size", len(content))
}

func AppendMemorySummariesBlock(spec *agent.PromptSpec, content string) {
	content = strings.TrimSpace(content)
	if content == "" {
		return
	}
	rendered, err := prompttmpl.Render(memorySummariesBlockTemplate, memorySummariesPromptBlockData{
		Content: content,
	})
	if err != nil {
		return
	}
	rendered = strings.TrimSpace(rendered)
	if rendered == "" {
		return
	}
	spec.Blocks = append(spec.Blocks, agent.PromptBlock{
		Content: rendered,
	})
}

func AppendTelegramRuntimeBlocks(spec *agent.PromptSpec, isGroup bool, mentionUsers []string, emojiList string) {
	content, err := prompttmpl.Render(telegramRuntimePromptBlockTemplate, telegramRuntimePromptBlockData{
		IsGroup:   isGroup,
		EmojiList: strings.TrimSpace(emojiList),
	})
	if err == nil {
		content = strings.TrimSpace(content)
		if content != "" {
			spec.Blocks = append(spec.Blocks, agent.PromptBlock{
				Content: content,
			})
		}
	}

	if !isGroup {
		return
	}
	if len(mentionUsers) > 0 {
		content, err := prompttmpl.Render(groupUsernamesBlockTemplate, groupUsernamesPromptBlockData{
			UsernamesText: strings.Join(mentionUsers, "\n"),
		})
		if err != nil {
			return
		}
		content = strings.TrimSpace(content)
		if content == "" {
			return
		}
		spec.Blocks = append(spec.Blocks, agent.PromptBlock{
			Content: content,
		})
	}
}

func AppendSlackRuntimeBlocks(spec *agent.PromptSpec, isGroup bool, mentionUsers []string, emojiList string) {
	content, err := prompttmpl.Render(slackRuntimePromptBlockTemplate, slackRuntimePromptBlockData{
		IsGroup:   isGroup,
		EmojiList: strings.TrimSpace(emojiList),
	})
	if err == nil {
		content = strings.TrimSpace(content)
		if content != "" {
			spec.Blocks = append(spec.Blocks, agent.PromptBlock{
				Content: content,
			})
		}
	}

	if !isGroup || len(mentionUsers) == 0 {
		return
	}
	content, err = prompttmpl.Render(slackMentionUsersBlockTemplate, slackMentionUsersPromptBlockData{
		UserIDsText: strings.Join(mentionUsers, "\n"),
	})
	if err != nil {
		return
	}
	content = strings.TrimSpace(content)
	if content == "" {
		return
	}
	spec.Blocks = append(spec.Blocks, agent.PromptBlock{
		Content: content,
	})
}
