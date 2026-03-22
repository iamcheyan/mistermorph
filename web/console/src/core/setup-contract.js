import heartbeatTemplate from "../../../../assets/config/HEARTBEAT.md?raw";
import identityTemplate from "../../../../assets/config/IDENTITY.md?raw";
import scriptsTemplate from "../../../../assets/config/SCRIPTS.md?raw";
import soulTemplate from "../../../../assets/config/SOUL.md?raw";
import todoDoneTemplate from "../../../../assets/config/TODO.DONE.md?raw";
import todoTemplate from "../../../../assets/config/TODO.md?raw";

const SETUP_PROVIDER_OPENAI_COMPATIBLE = "openai_compatible";
const SETUP_PROVIDER_GEMINI = "gemini";
const SETUP_PROVIDER_ANTHROPIC = "anthropic";
const SETUP_PROVIDER_CLOUDFLARE = "cloudflare";

const SETUP_PROVIDER_OPTIONS = [
  { title: "OpenAI Compatible", value: SETUP_PROVIDER_OPENAI_COMPATIBLE },
  { title: "Gemini", value: SETUP_PROVIDER_GEMINI },
  { title: "Anthropic", value: SETUP_PROVIDER_ANTHROPIC },
  { title: "Cloudflare", value: SETUP_PROVIDER_CLOUDFLARE },
];

const SETUP_REQUIRED_MARKDOWN_FILES = [
  { name: "HEARTBEAT.md", content: heartbeatTemplate },
  { name: "SCRIPTS.md", content: scriptsTemplate },
  { name: "TODO.md", content: todoTemplate },
  { name: "TODO.DONE.md", content: todoDoneTemplate },
  { name: "IDENTITY.md", content: identityTemplate },
  { name: "SOUL.md", content: soulTemplate },
];

function normalizeSetupProviderChoice(provider) {
  switch (String(provider || "").trim().toLowerCase()) {
    case SETUP_PROVIDER_GEMINI:
      return SETUP_PROVIDER_GEMINI;
    case SETUP_PROVIDER_ANTHROPIC:
      return SETUP_PROVIDER_ANTHROPIC;
    case SETUP_PROVIDER_CLOUDFLARE:
      return SETUP_PROVIDER_CLOUDFLARE;
    default:
      return SETUP_PROVIDER_OPENAI_COMPATIBLE;
  }
}

function defaultEndpointForSetupProvider(choice) {
  switch (normalizeSetupProviderChoice(choice)) {
    case SETUP_PROVIDER_GEMINI:
      return "https://generativelanguage.googleapis.com";
    case SETUP_PROVIDER_ANTHROPIC:
      return "https://api.anthropic.com";
    case SETUP_PROVIDER_CLOUDFLARE:
      return "https://api.cloudflare.com/client/v4";
    default:
      return "https://api.openai.com";
  }
}

function isOfficialOpenAIEndpoint(endpoint) {
  const value = String(endpoint || "").trim().replace(/\/+$/, "").toLowerCase();
  return (
    value === "" ||
    value === "https://api.openai.com" ||
    value === "https://api.openai.com/v1" ||
    value === "http://api.openai.com" ||
    value === "http://api.openai.com/v1"
  );
}

function normalizeSetupProviderForSave(choice, endpoint) {
  switch (normalizeSetupProviderChoice(choice)) {
    case SETUP_PROVIDER_GEMINI:
      return SETUP_PROVIDER_GEMINI;
    case SETUP_PROVIDER_ANTHROPIC:
      return SETUP_PROVIDER_ANTHROPIC;
    case SETUP_PROVIDER_CLOUDFLARE:
      return SETUP_PROVIDER_CLOUDFLARE;
    default:
      return isOfficialOpenAIEndpoint(endpoint) ? "openai" : "openai_custom";
  }
}

export {
  SETUP_PROVIDER_ANTHROPIC,
  SETUP_PROVIDER_CLOUDFLARE,
  SETUP_PROVIDER_GEMINI,
  SETUP_PROVIDER_OPENAI_COMPATIBLE,
  SETUP_PROVIDER_OPTIONS,
  SETUP_REQUIRED_MARKDOWN_FILES,
  defaultEndpointForSetupProvider,
  normalizeSetupProviderChoice,
  normalizeSetupProviderForSave,
};
