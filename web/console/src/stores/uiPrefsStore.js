import { reactive } from "vue";

const UI_PREFS_STORAGE_KEY = "mistermorph_console_ui_prefs_v1";
const DEFAULT_CHAT_MARKDOWN_THEME = "console";
const CHAT_MARKDOWN_THEME_IDS = Object.freeze([
  "console",
  "paper",
  "folio",
  "blueprint",
]);

const uiPrefsState = reactive({
  chatMarkdownTheme: DEFAULT_CHAT_MARKDOWN_THEME,
});

function normalizeChatMarkdownTheme(value) {
  const next = typeof value === "string" ? value.trim().toLowerCase() : "";
  return CHAT_MARKDOWN_THEME_IDS.includes(next) ? next : DEFAULT_CHAT_MARKDOWN_THEME;
}

function saveUIPreferences() {
  localStorage.setItem(
    UI_PREFS_STORAGE_KEY,
    JSON.stringify({
      chat_markdown_theme: uiPrefsState.chatMarkdownTheme,
    })
  );
}

function setChatMarkdownTheme(value) {
  uiPrefsState.chatMarkdownTheme = normalizeChatMarkdownTheme(value);
  saveUIPreferences();
}

function hydrateUIPreferences() {
  const raw = localStorage.getItem(UI_PREFS_STORAGE_KEY);
  if (!raw) {
    uiPrefsState.chatMarkdownTheme = DEFAULT_CHAT_MARKDOWN_THEME;
    saveUIPreferences();
    return;
  }
  try {
    const parsed = JSON.parse(raw);
    uiPrefsState.chatMarkdownTheme = normalizeChatMarkdownTheme(parsed.chat_markdown_theme);
  } catch {
    uiPrefsState.chatMarkdownTheme = DEFAULT_CHAT_MARKDOWN_THEME;
    saveUIPreferences();
  }
}

export {
  CHAT_MARKDOWN_THEME_IDS,
  DEFAULT_CHAT_MARKDOWN_THEME,
  uiPrefsState,
  setChatMarkdownTheme,
  hydrateUIPreferences,
};
