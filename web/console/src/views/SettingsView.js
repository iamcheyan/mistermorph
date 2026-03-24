import { computed, onMounted, onUnmounted, reactive, ref, watch } from "vue";
import { useRouter } from "vue-router";
import "./SettingsView.css";

import AppPage from "../components/AppPage";
import SetupConnectionTestDialog from "../components/SetupConnectionTestDialog";
import SetupPickerDialog from "../components/SetupPickerDialog";
import {
  apiFetch,
  applyLanguageChange,
  clearAuth,
  endpointState,
  loadEndpoints,
  localeState,
  runtimeEndpointByRef,
  translate,
} from "../core/context";
import {
  OPENAI_COMPATIBLE_API_BASE_OPTIONS,
  defaultEndpointForSetupProvider,
  normalizeSetupProviderChoice,
  normalizeSetupProviderForSave,
  resolveSetupAPIKeyHelp,
  SETUP_PROVIDER_CLOUDFLARE,
  SETUP_PROVIDER_OPTIONS,
  setupProviderRequiresAPIKey,
  setupProviderSupportsModelLookup,
} from "../core/setup-contract";

function tuiKicker(left, right) {
  const lhs = String(left || "").trim();
  const rhs = String(right || "").trim();
  if (lhs && rhs) {
    return `[ ${lhs.toUpperCase()} // ${rhs.toUpperCase()} ]`;
  }
  return `[ ${(lhs || rhs).toUpperCase()} ]`;
}

const MULTIMODAL_SOURCES = [
  { id: "telegram", titleKey: "settings_multimodal_source_telegram", noteKey: "settings_multimodal_note_telegram" },
  { id: "slack", titleKey: "settings_multimodal_source_slack", noteKey: "settings_multimodal_note_slack" },
  { id: "line", titleKey: "settings_multimodal_source_line", noteKey: "settings_multimodal_note_line" },
  {
    id: "remote_download",
    titleKey: "settings_multimodal_source_remote_download",
    noteKey: "settings_multimodal_note_remote_download",
  },
];

const TOOL_ITEMS = [
  { id: "write_file", titleKey: "settings_tool_write_file", noteKey: "settings_tool_note_write_file" },
  { id: "contacts_send", titleKey: "settings_tool_contacts_send", noteKey: "settings_tool_note_contacts_send" },
  { id: "todo_update", titleKey: "settings_tool_todo_update", noteKey: "settings_tool_note_todo_update" },
  { id: "plan_create", titleKey: "settings_tool_plan_create", noteKey: "settings_tool_note_plan_create" },
  { id: "url_fetch", titleKey: "settings_tool_url_fetch", noteKey: "settings_tool_note_url_fetch" },
  { id: "web_search", titleKey: "settings_tool_web_search", noteKey: "settings_tool_note_web_search" },
  { id: "bash", titleKey: "settings_tool_bash", noteKey: "settings_tool_note_bash" },
];

const MANAGED_RUNTIME_ITEMS = [
  { id: "telegram", titleKey: "settings_console_runtime_telegram", noteKey: "settings_console_runtime_note_telegram" },
  { id: "slack", titleKey: "settings_console_runtime_slack", noteKey: "settings_console_runtime_note_slack" },
];

const LOCAL_CONSOLE_ENDPOINT_REF = "ep_console_local";
function buildAgentSnapshot(state) {
  return JSON.stringify({
    llm: {
      provider: String(state.llm.provider || "").trim(),
      endpoint: String(state.llm.endpoint || "").trim(),
      model: String(state.llm.model || "").trim(),
      api_key: String(state.llm.api_key || "").trim(),
      cloudflare_api_token: String(state.llm.cloudflare_api_token || "").trim(),
      cloudflare_account_id: String(state.llm.cloudflare_account_id || "").trim(),
      reasoning_effort: String(state.llm.reasoning_effort || "").trim(),
      tools_emulation_mode: String(state.llm.tools_emulation_mode || "").trim(),
    },
    multimodal: {
      telegram: !!state.multimodal.telegram,
      slack: !!state.multimodal.slack,
      line: !!state.multimodal.line,
      remote_download: !!state.multimodal.remote_download,
    },
    tools: {
      write_file: !!state.tools.write_file,
      contacts_send: !!state.tools.contacts_send,
      todo_update: !!state.tools.todo_update,
      plan_create: !!state.tools.plan_create,
      url_fetch: !!state.tools.url_fetch,
      web_search: !!state.tools.web_search,
      bash: !!state.tools.bash,
    },
  });
}

function buildConsoleSnapshot(state) {
  return JSON.stringify({
    managed_runtimes: {
      telegram: !!state.managedRuntimes.telegram,
      slack: !!state.managedRuntimes.slack,
    },
  });
}

const SettingsView = {
  components: {
    AppPage,
    SetupConnectionTestDialog,
    SetupPickerDialog,
  },
  setup() {
    const t = translate;
    const router = useRouter();
    const lang = computed(() => localeState.lang);
    const loggingOut = ref(false);
    const agentLoading = ref(false);
    const agentSaving = ref(false);
    const agentErr = ref("");
    const agentOk = ref("");
    const llmConfigPath = ref("");
    const loadedSnapshot = ref("");
    const llmEnvManaged = ref({});
    const consoleLoading = ref(false);
    const consoleSaving = ref(false);
    const consoleErr = ref("");
    const consoleOk = ref("");
    const consoleConfigPath = ref("");
    const loadedConsoleSnapshot = ref("");
    const selectedSectionID = ref("agent");
    const isMobile = ref(false);
    const mobilePanelVisible = ref(false);
    const apiBasePickerOpen = ref(false);
    const modelPickerOpen = ref(false);
    const modelPickerLoading = ref(false);
    const modelPickerError = ref("");
    const modelPickerItems = ref([]);
    const testConnectionOpen = ref(false);
    const testConnectionLoading = ref(false);
    const testConnectionError = ref("");
    const testConnectionBenchmarks = ref([]);
    const testConnectionMeta = reactive({
      provider: "",
      model: "",
    });

    const state = reactive({
      llm: {
        provider: "",
        endpoint: "",
        model: "",
        api_key: "",
        cloudflare_api_token: "",
        cloudflare_account_id: "",
        reasoning_effort: "",
        tools_emulation_mode: "",
      },
      multimodal: {
        telegram: false,
        slack: false,
        line: false,
        remote_download: false,
      },
      tools: {
        write_file: true,
        contacts_send: true,
        todo_update: true,
        plan_create: true,
        url_fetch: true,
        web_search: true,
        bash: true,
      },
      managedRuntimes: {
        telegram: false,
        slack: false,
      },
    });

    const providerItems = computed(() => SETUP_PROVIDER_OPTIONS);
    const providerItem = computed(
      () => providerItems.value.find((item) => item.value === state.llm.provider) || null
    );
    const showCloudflareAccountField = computed(
      () => normalizeSetupProviderChoice(llmFieldValue("provider")) === SETUP_PROVIDER_CLOUDFLARE
    );
    const credentialFieldName = computed(() => (showCloudflareAccountField.value ? "cloudflare_api_token" : "api_key"));
    const credentialLabelKey = computed(() =>
      showCloudflareAccountField.value ? "settings_agent_cloudflare_api_token_label" : "settings_agent_api_key_label"
    );
    const credentialPlaceholderKey = computed(() =>
      showCloudflareAccountField.value ? "settings_agent_cloudflare_api_token_placeholder" : "settings_agent_api_key_placeholder"
    );
    const credentialHintKey = computed(() =>
      showCloudflareAccountField.value ? "setup_llm_api_token_hint" : "setup_llm_api_key_hint"
    );
    const credentialHintPlainKey = computed(() =>
      showCloudflareAccountField.value ? "setup_llm_api_token_hint_plain" : "setup_llm_api_key_hint_plain"
    );
    const showOpenAICompatibleHelpers = computed(() => setupProviderSupportsModelLookup(llmFieldValue("provider")));
    const modelLookupDisabled = computed(
      () =>
        agentLoading.value ||
        agentSaving.value ||
        !showOpenAICompatibleHelpers.value ||
        !hasLLMFieldValue("api_key")
    );
    const apiBasePickerItems = computed(() =>
      OPENAI_COMPATIBLE_API_BASE_OPTIONS.map((item) => ({
        id: item.id,
        title: item.title,
        value: item.baseURL,
        note: "",
      }))
    );
    const credentialHelp = computed(() => {
      const provider = llmFieldValue("provider");
      if (provider === "" || isLLMFieldEnvManaged(credentialFieldName.value)) {
        return null;
      }
      return resolveSetupAPIKeyHelp(provider, llmFieldValue("endpoint"));
    });
    const credentialHelpParts = computed(() => {
      if (!credentialHelp.value) {
        return null;
      }
      const marker = "__PROVIDER__";
      const template = String(t(credentialHintKey.value, { provider: marker }) || "");
      const index = template.indexOf(marker);
      if (index === -1) {
        return {
          before: template.trim(),
          after: "",
        };
      }
      return {
        before: template.slice(0, index),
        after: template.slice(index + marker.length),
      };
    });
    const reasoningEffortItems = computed(() => [
      { title: t("settings_llm_reasoning_none"), value: "" },
      { title: t("settings_llm_reasoning_minimal"), value: "minimal" },
      { title: t("settings_llm_reasoning_low"), value: "low" },
      { title: t("settings_llm_reasoning_medium"), value: "medium" },
      { title: t("settings_llm_reasoning_high"), value: "high" },
      { title: t("settings_llm_reasoning_max"), value: "max" },
      { title: t("settings_llm_reasoning_xhigh"), value: "xhigh" },
    ]);
    const reasoningEffortItem = computed(
      () => reasoningEffortItems.value.find((item) => item.value === state.llm.reasoning_effort) || reasoningEffortItems.value[0]
    );
    const toolsEmulationItems = computed(() => [
      { title: t("settings_llm_tools_emulation_off"), value: "off" },
      { title: t("settings_llm_tools_emulation_fallback"), value: "fallback" },
      { title: t("settings_llm_tools_emulation_force"), value: "force" },
    ]);
    const toolsEmulationItem = computed(
      () =>
        toolsEmulationItems.value.find((item) => item.value === state.llm.tools_emulation_mode) ||
        toolsEmulationItems.value[0]
    );
    const multimodalItems = computed(() => MULTIMODAL_SOURCES);
    const toolItems = computed(() => TOOL_ITEMS);
    const managedRuntimeItems = computed(() => MANAGED_RUNTIME_ITEMS);
    const selectedEndpoint = computed(() => runtimeEndpointByRef(endpointState.selectedRef));
    const showConsoleManagedSettings = computed(
      () => String(selectedEndpoint.value?.endpoint_ref || "").trim() === LOCAL_CONSOLE_ENDPOINT_REF
    );

    const settingsSections = computed(() => {
      const items = [
        {
          id: "agent",
          title: t("settings_agent_block_title"),
          meta: t("settings_section_agent_meta"),
          groupTitle: t("settings_agent_title"),
          saveKind: "agent",
        },
        {
          id: "inputs",
          title: t("settings_multimodal_title"),
          meta: t("settings_section_inputs_meta"),
          groupTitle: t("settings_agent_title"),
          saveKind: "agent",
        },
        {
          id: "tools",
          title: t("settings_tools_title"),
          meta: t("settings_section_tools_meta"),
          groupTitle: t("settings_agent_title"),
          saveKind: "agent",
        },
      ];
      if (showConsoleManagedSettings.value) {
        items.push({
          id: "runtimes",
          title: t("settings_console_runtime_title"),
          meta: t("settings_section_runtimes_meta"),
          groupTitle: t("settings_console_title"),
          saveKind: "console",
        });
      }
      items.push({
        id: "console",
        title: t("settings_console_title"),
        meta: t("settings_section_console_meta"),
        groupTitle: t("settings_console_title"),
        saveKind: "",
      });
      return items;
    });

    const selectedSection = computed(
      () => settingsSections.value.find((item) => item.id === selectedSectionID.value) || settingsSections.value[0] || null
    );
    const activeSaveKind = computed(() => String(selectedSection.value?.saveKind || ""));
    const panelKicker = computed(() => {
      if (!selectedSection.value) {
        return "";
      }
      return tuiKicker(selectedSection.value.groupTitle, selectedSection.value.title);
    });
    const panelHint = computed(() => {
      switch (selectedSection.value?.id) {
        case "agent":
          return t("settings_agent_llm_hint", { path: llmConfigPath.value || "config.yaml" });
        case "inputs":
          return t("settings_multimodal_hint");
        case "tools":
          return t("settings_tools_hint");
        case "runtimes":
          return t("settings_console_runtime_hint", { path: consoleConfigPath.value || "config.yaml" });
        case "console":
          return t("settings_console_preferences_hint");
        default:
          return "";
      }
    });
    const showIndexPane = computed(() => !isMobile.value || !mobilePanelVisible.value);
    const showPanelPane = computed(() => !isMobile.value || mobilePanelVisible.value);
    const mobileShowBack = computed(() => isMobile.value && mobilePanelVisible.value);
    const mobileBarTitle = computed(() =>
      mobileShowBack.value ? selectedSection.value?.title || t("settings_title") : t("settings_title")
    );
    const pageClass = computed(() => (isMobile.value ? "settings-page settings-page-mobile-split" : "settings-page"));
    const testConnectionDisabled = computed(
      () =>
        testConnectionLoading.value ||
        agentLoading.value ||
        agentSaving.value ||
        !hasLLMFieldValue("provider") ||
        !hasLLMFieldValue("model") ||
        (setupProviderRequiresAPIKey(llmFieldValue("provider")) && !hasLLMFieldValue(credentialFieldName.value)) ||
        (showCloudflareAccountField.value && !hasLLMFieldValue("cloudflare_api_token")) ||
        (showCloudflareAccountField.value && !hasLLMFieldValue("cloudflare_account_id"))
    );
    const agentDirty = computed(() => buildAgentSnapshot(state) !== loadedSnapshot.value);
    const agentSaveDisabled = computed(
      () =>
        agentLoading.value ||
        agentSaving.value ||
        !hasLLMFieldValue("provider") ||
        !agentDirty.value ||
        (showCloudflareAccountField.value && !hasLLMFieldValue("cloudflare_api_token")) ||
        (showCloudflareAccountField.value && !hasLLMFieldValue("cloudflare_account_id"))
    );
    const consoleDirty = computed(() => buildConsoleSnapshot(state) !== loadedConsoleSnapshot.value);
    const consoleSaveDisabled = computed(() => consoleLoading.value || consoleSaving.value || !consoleDirty.value);

    function applyPayload(data) {
      const llm = data?.llm && typeof data.llm === "object" ? data.llm : {};
      const envManagedPayload = data?.env_managed && typeof data.env_managed === "object" ? data.env_managed : {};
      const llmEnvManagedPayload =
        envManagedPayload?.llm && typeof envManagedPayload.llm === "object" ? envManagedPayload.llm : {};
      const multimodal = data?.multimodal && typeof data.multimodal === "object" ? data.multimodal : {};
      const tools = data?.tools && typeof data.tools === "object" ? data.tools : {};
      const imageSources = Array.isArray(multimodal.image_sources) ? multimodal.image_sources : [];

      state.llm.provider = normalizeSetupProviderChoice(llm.provider);
      state.llm.endpoint = typeof llm.endpoint === "string" ? llm.endpoint : "";
      state.llm.model = typeof llm.model === "string" ? llm.model : "";
      state.llm.api_key = typeof llm.api_key === "string" ? llm.api_key : "";
      state.llm.cloudflare_api_token =
        typeof llm.cloudflare_api_token === "string" ? llm.cloudflare_api_token : "";
      state.llm.cloudflare_account_id = typeof llm.cloudflare_account_id === "string" ? llm.cloudflare_account_id : "";
      state.llm.reasoning_effort = typeof llm.reasoning_effort === "string" ? llm.reasoning_effort : "";
      state.llm.tools_emulation_mode =
        typeof llm.tools_emulation_mode === "string" ? llm.tools_emulation_mode : "off";
      for (const item of MULTIMODAL_SOURCES) {
        state.multimodal[item.id] = imageSources.includes(item.id);
      }
      state.tools.write_file = !!tools.write_file_enabled;
      state.tools.contacts_send = !!tools.contacts_send_enabled;
      state.tools.todo_update = !!tools.todo_update_enabled;
      state.tools.plan_create = !!tools.plan_create_enabled;
      state.tools.url_fetch = !!tools.url_fetch_enabled;
      state.tools.web_search = !!tools.web_search_enabled;
      state.tools.bash = !!tools.bash_enabled;
      llmEnvManaged.value = llmEnvManagedPayload;

      loadedSnapshot.value = buildAgentSnapshot(state);
    }

    function llmFieldEnvName(field) {
      const entry = llmEnvManaged.value && typeof llmEnvManaged.value === "object" ? llmEnvManaged.value[field] : null;
      if (!entry || typeof entry !== "object") {
        return "";
      }
      return typeof entry.env_name === "string" ? entry.env_name : "";
    }

    function llmFieldEnvValue(field) {
      const entry = llmEnvManaged.value && typeof llmEnvManaged.value === "object" ? llmEnvManaged.value[field] : null;
      if (!entry || typeof entry !== "object") {
        return "";
      }
      return typeof entry.value === "string" ? entry.value.trim() : "";
    }

    function isLLMFieldEnvManaged(field) {
      return llmFieldEnvName(field) !== "";
    }

    function llmFieldValue(field) {
      const key = String(field || "").trim();
      if (!key) {
        return "";
      }
      if (isLLMFieldEnvManaged(key)) {
        return llmFieldEnvValue(key);
      }
      const value = state.llm && typeof state.llm === "object" ? state.llm[key] : "";
      return typeof value === "string" ? value.trim() : "";
    }

    function hasLLMFieldValue(field) {
      return llmFieldValue(field) !== "" || isLLMFieldEnvManaged(field);
    }

    function llmFieldManagedDisplayValue(field) {
      const envValue = llmFieldEnvValue(field);
      if (envValue !== "") {
        return envValue;
      }
      if (["api_key", "cloudflare_api_token"].includes(String(field || "").trim())) {
        return "";
      }
      return isLLMFieldEnvManaged(field) ? llmFieldValue(field) : "";
    }

    function llmFieldManagedHeadline(field) {
      const envName = llmFieldEnvName(field);
      if (envName === "") {
        return "";
      }
      const value = llmFieldManagedDisplayValue(field);
      return value === "" ? envName : `${envName}=${value}`;
    }

    async function loadAgentSettings() {
      agentLoading.value = true;
      agentErr.value = "";
      agentOk.value = "";
      try {
        const data = await apiFetch("/settings/agent");
        llmConfigPath.value = typeof data.config_path === "string" ? data.config_path : "";
        applyPayload(data);
      } catch (e) {
        agentErr.value = e.message || t("msg_load_failed");
      } finally {
        agentLoading.value = false;
      }
    }

    function applyConsolePayload(data) {
      const values = Array.isArray(data?.managed_runtimes) ? data.managed_runtimes : [];
      for (const item of MANAGED_RUNTIME_ITEMS) {
        state.managedRuntimes[item.id] = values.includes(item.id);
      }
      loadedConsoleSnapshot.value = buildConsoleSnapshot(state);
    }

    async function loadConsoleSettings() {
      if (!showConsoleManagedSettings.value) {
        return;
      }
      consoleLoading.value = true;
      consoleErr.value = "";
      consoleOk.value = "";
      try {
        const data = await apiFetch("/settings/console");
        consoleConfigPath.value = typeof data.config_path === "string" ? data.config_path : "";
        applyConsolePayload(data);
      } catch (e) {
        consoleErr.value = e.message || t("msg_load_failed");
      } finally {
        consoleLoading.value = false;
      }
    }

    function buildSavePayload() {
      return {
        llm: buildLLMSettingsPayload(),
        multimodal: {
          image_sources: MULTIMODAL_SOURCES.filter((item) => state.multimodal[item.id]).map((item) => item.id),
        },
        tools: {
          write_file_enabled: state.tools.write_file,
          contacts_send_enabled: state.tools.contacts_send,
          todo_update_enabled: state.tools.todo_update,
          plan_create_enabled: state.tools.plan_create,
          url_fetch_enabled: state.tools.url_fetch,
          web_search_enabled: state.tools.web_search,
          bash_enabled: state.tools.bash,
        },
      };
    }

    function buildLLMSettingsPayload() {
      const payload = {};
      const provider = normalizeSetupProviderChoice(llmFieldValue("provider"), { allowEmpty: true });
      const useCloudflareCredentials = normalizeSetupProviderChoice(state.llm.provider) === SETUP_PROVIDER_CLOUDFLARE;
      if (!isLLMFieldEnvManaged("provider")) {
        payload.provider = normalizeSetupProviderForSave(state.llm.provider, state.llm.endpoint);
      }
      if (!isLLMFieldEnvManaged("endpoint")) {
        payload.endpoint = String(state.llm.endpoint || "").trim();
      }
      if (!isLLMFieldEnvManaged("model")) {
        payload.model = String(state.llm.model || "").trim();
      }
      if (provider === SETUP_PROVIDER_CLOUDFLARE) {
        if (!isLLMFieldEnvManaged("cloudflare_api_token")) {
          payload.cloudflare_api_token = useCloudflareCredentials ? String(state.llm.cloudflare_api_token || "").trim() : "";
        }
        if (!isLLMFieldEnvManaged("cloudflare_account_id")) {
          payload.cloudflare_account_id = String(state.llm.cloudflare_account_id || "").trim();
        }
      } else if (!isLLMFieldEnvManaged("api_key")) {
        payload.api_key = String(state.llm.api_key || "").trim();
      }
      if (!isLLMFieldEnvManaged("reasoning_effort")) {
        payload.reasoning_effort = String(state.llm.reasoning_effort || "").trim();
      }
      if (!isLLMFieldEnvManaged("tools_emulation_mode")) {
        payload.tools_emulation_mode = String(state.llm.tools_emulation_mode || "").trim();
      }
      return payload;
    }

    function buildLLMTestPayload() {
      const payload = {};
      const provider = normalizeSetupProviderChoice(llmFieldValue("provider"), { allowEmpty: true });
      if (!isLLMFieldEnvManaged("provider") && provider !== "") {
        payload.provider = normalizeSetupProviderForSave(state.llm.provider, state.llm.endpoint);
      }
      if (!isLLMFieldEnvManaged("endpoint")) {
        const endpoint = String(state.llm.endpoint || "").trim();
        if (endpoint !== "") {
          payload.endpoint = endpoint;
        }
      }
      if (!isLLMFieldEnvManaged("model")) {
        const model = String(state.llm.model || "").trim();
        if (model !== "") {
          payload.model = model;
        }
      }
      if (provider === SETUP_PROVIDER_CLOUDFLARE) {
        if (!isLLMFieldEnvManaged("cloudflare_api_token")) {
          const token = String(state.llm.cloudflare_api_token || "").trim();
          if (token !== "") {
            payload.cloudflare_api_token = token;
          }
        }
        if (!isLLMFieldEnvManaged("cloudflare_account_id")) {
          const accountID = String(state.llm.cloudflare_account_id || "").trim();
          if (accountID !== "") {
            payload.cloudflare_account_id = accountID;
          }
        }
      } else if (!isLLMFieldEnvManaged("api_key")) {
        const apiKey = String(state.llm.api_key || "").trim();
        if (apiKey !== "") {
          payload.api_key = apiKey;
        }
      }
      if (!isLLMFieldEnvManaged("reasoning_effort")) {
        const reasoningEffort = String(state.llm.reasoning_effort || "").trim();
        if (reasoningEffort !== "") {
          payload.reasoning_effort = reasoningEffort;
        }
      }
      if (!isLLMFieldEnvManaged("tools_emulation_mode")) {
        const toolsEmulationMode = String(state.llm.tools_emulation_mode || "").trim();
        if (toolsEmulationMode !== "") {
          payload.tools_emulation_mode = toolsEmulationMode;
        }
      }
      return payload;
    }

    function buildConsoleSavePayload() {
      return {
        managed_runtimes: MANAGED_RUNTIME_ITEMS.filter((item) => state.managedRuntimes[item.id]).map((item) => item.id),
      };
    }

    async function saveAgentSettings() {
      if (agentSaveDisabled.value) {
        return;
      }
      agentSaving.value = true;
      agentErr.value = "";
      agentOk.value = "";
      try {
        const payload = await apiFetch("/settings/agent", {
          method: "PUT",
          body: buildSavePayload(),
        });
        llmConfigPath.value = typeof payload.config_path === "string" ? payload.config_path : llmConfigPath.value;
        applyPayload(payload);
        await loadEndpoints();
        agentOk.value = t("msg_save_success");
      } catch (e) {
        agentErr.value = e.message || t("msg_save_failed");
      } finally {
        agentSaving.value = false;
      }
    }

    async function saveConsoleSettings() {
      if (consoleSaveDisabled.value || !showConsoleManagedSettings.value) {
        return;
      }
      consoleSaving.value = true;
      consoleErr.value = "";
      consoleOk.value = "";
      try {
        const payload = await apiFetch("/settings/console", {
          method: "PUT",
          body: buildConsoleSavePayload(),
        });
        consoleConfigPath.value =
          typeof payload.config_path === "string" ? payload.config_path : consoleConfigPath.value;
        applyConsolePayload(payload);
        consoleOk.value = t("msg_save_success");
      } catch (e) {
        consoleErr.value = e.message || t("msg_save_failed");
      } finally {
        consoleSaving.value = false;
      }
    }

    async function logout() {
      loggingOut.value = true;
      try {
        await apiFetch("/auth/logout", { method: "POST" });
      } catch {
        // ignore logout failure
      } finally {
        clearAuth();
        router.replace("/login");
        loggingOut.value = false;
      }
    }

    function onProviderChange(item) {
      if (!item || typeof item !== "object") {
        return;
      }
      const nextProvider = String(item.value || "").trim();
      const previousDefault = defaultEndpointForSetupProvider(state.llm.provider);
      state.llm.provider = nextProvider;
      if (String(state.llm.endpoint || "").trim() === "" || String(state.llm.endpoint || "").trim() === previousDefault) {
        state.llm.endpoint = defaultEndpointForSetupProvider(nextProvider);
      }
    }

    function openExternal(url) {
      const target = String(url || "").trim();
      if (!target) {
        return;
      }
      window.open(target, "_blank", "noopener,noreferrer");
    }

    function openAPIBasePicker() {
      if (!showOpenAICompatibleHelpers.value || agentLoading.value || agentSaving.value) {
        return;
      }
      apiBasePickerOpen.value = true;
    }

    function applyAPIBaseOption(item) {
      state.llm.endpoint = String(item?.value || "").trim();
    }

    async function openModelPicker() {
      if (modelLookupDisabled.value) {
        return;
      }
      modelPickerOpen.value = true;
      modelPickerLoading.value = true;
      modelPickerError.value = "";
      modelPickerItems.value = [];
      try {
        const payload = await apiFetch("/settings/agent/models", {
          method: "POST",
          body: {
            endpoint: llmFieldValue("endpoint"),
            api_key: llmFieldValue("api_key"),
          },
        });
        const items = Array.isArray(payload?.items) ? payload.items : [];
        modelPickerItems.value = items.map((value) => ({
          id: value,
          title: value,
          value,
          note: "",
        }));
      } catch (e) {
        modelPickerError.value = e.message || t("msg_load_failed");
      } finally {
        modelPickerLoading.value = false;
      }
    }

    function applyModelOption(item) {
      state.llm.model = String(item?.value || "").trim();
    }

    async function openTestConnection() {
      if (testConnectionDisabled.value) {
        return;
      }
      testConnectionOpen.value = true;
      await runConnectionTest();
    }

    async function runConnectionTest() {
      if (testConnectionLoading.value) {
        return;
      }
      const nextPayload = buildLLMTestPayload();
      testConnectionLoading.value = true;
      testConnectionError.value = "";
      testConnectionBenchmarks.value = [];
      testConnectionMeta.provider = normalizeSetupProviderForSave(llmFieldValue("provider"), llmFieldValue("endpoint"));
      testConnectionMeta.model = String(nextPayload.model || "").trim();
      try {
        const payload = await apiFetch("/settings/agent/test", {
          method: "POST",
          body: {
            llm: nextPayload,
          },
        });
        testConnectionMeta.provider = String(payload?.provider || "").trim();
        testConnectionMeta.model = String(payload?.model || "").trim();
        const items = Array.isArray(payload?.benchmarks) ? payload.benchmarks : [];
        testConnectionBenchmarks.value = items.map((item) => ({
          id: String(item?.id || "").trim(),
          ok: item?.ok === true,
          duration_ms: Number(item?.duration_ms || 0),
          detail: String(item?.detail || "").trim(),
          error: String(item?.error || "").trim(),
          raw_response: String(item?.raw_response || ""),
        }));
      } catch (e) {
        testConnectionError.value = e.message || t("msg_load_failed");
      } finally {
        testConnectionLoading.value = false;
      }
    }

    function onReasoningEffortChange(item) {
      if (!item || typeof item !== "object") {
        return;
      }
      state.llm.reasoning_effort = String(item.value || "").trim();
    }

    function onToolsEmulationChange(item) {
      if (!item || typeof item !== "object") {
        return;
      }
      state.llm.tools_emulation_mode = String(item.value || "").trim();
    }

    function setMultimodalSource(id, value) {
      if (!Object.prototype.hasOwnProperty.call(state.multimodal, id)) {
        return;
      }
      state.multimodal[id] = !!value;
    }

    function setToolEnabled(id, value) {
      if (!Object.prototype.hasOwnProperty.call(state.tools, id)) {
        return;
      }
      state.tools[id] = !!value;
    }

    function setManagedRuntimeEnabled(id, value) {
      if (!Object.prototype.hasOwnProperty.call(state.managedRuntimes, id)) {
        return;
      }
      state.managedRuntimes[id] = !!value;
    }

    function refreshMobileMode() {
      isMobile.value = typeof window !== "undefined" && window.innerWidth <= 920;
    }

    function showIndexView() {
      mobilePanelVisible.value = false;
    }

    function selectSection(id) {
      selectedSectionID.value = String(id || "").trim();
      if (isMobile.value) {
        mobilePanelVisible.value = true;
      }
    }

    function isSelectedSection(item) {
      return String(item?.id || "") === selectedSectionID.value;
    }

    function sectionClass(item) {
      const classes = ["settings-index-item", "workspace-sidebar-item"];
      if (isSelectedSection(item)) {
        classes.push("is-active");
      }
      return classes.join(" ");
    }

    onMounted(() => {
      window.addEventListener("resize", refreshMobileMode);
      refreshMobileMode();
      void loadAgentSettings();
    });

    onUnmounted(() => {
      window.removeEventListener("resize", refreshMobileMode);
    });

    watch(
      settingsSections,
      (items) => {
        if (!items.some((item) => item.id === selectedSectionID.value)) {
          selectedSectionID.value = items[0]?.id || "";
        }
      },
      { immediate: true }
    );

    watch(
      showConsoleManagedSettings,
      (enabled) => {
        consoleErr.value = "";
        consoleOk.value = "";
        if (enabled) {
          void loadConsoleSettings();
          return;
        }
        if (selectedSectionID.value === "runtimes") {
          selectedSectionID.value = "console";
        }
      },
      { immediate: true }
    );

    return {
      t,
      lang,
      loggingOut,
      agentLoading,
      agentSaving,
      agentErr,
      agentOk,
      consoleLoading,
      consoleSaving,
      consoleErr,
      consoleOk,
      state,
      llmEnvManaged,
      providerItems,
      providerItem,
      showCloudflareAccountField,
      reasoningEffortItems,
      reasoningEffortItem,
      toolsEmulationItems,
      toolsEmulationItem,
      showOpenAICompatibleHelpers,
      modelLookupDisabled,
      apiBasePickerItems,
      credentialLabelKey,
      credentialPlaceholderKey,
      credentialHelp,
      credentialHelpParts,
      credentialHintPlainKey,
      multimodalItems,
      toolItems,
      managedRuntimeItems,
      settingsSections,
      selectedSection,
      panelKicker,
      panelHint,
      activeSaveKind,
      showIndexPane,
      showPanelPane,
      mobileShowBack,
      mobileBarTitle,
      pageClass,
      agentSaveDisabled,
      consoleSaveDisabled,
      testConnectionDisabled,
      logout,
      saveAgentSettings,
      saveConsoleSettings,
      onProviderChange,
      onReasoningEffortChange,
      onToolsEmulationChange,
      llmFieldEnvName,
      llmFieldEnvValue,
      isLLMFieldEnvManaged,
      llmFieldValue,
      hasLLMFieldValue,
      llmFieldManagedDisplayValue,
      llmFieldManagedHeadline,
      openExternal,
      openAPIBasePicker,
      applyAPIBaseOption,
      openModelPicker,
      applyModelOption,
      openTestConnection,
      runConnectionTest,
      setMultimodalSource,
      setToolEnabled,
      setManagedRuntimeEnabled,
      selectSection,
      isSelectedSection,
      sectionClass,
      showIndexView,
      tuiKicker,
      apiBasePickerOpen,
      modelPickerOpen,
      modelPickerLoading,
      modelPickerError,
      modelPickerItems,
      testConnectionOpen,
      testConnectionLoading,
      testConnectionError,
      testConnectionBenchmarks,
      testConnectionMeta,
      onLanguageChange: applyLanguageChange,
    };
  },
  template: `
    <AppPage :title="t('settings_title')" :class="pageClass" :showMobileNavTrigger="!mobileShowBack">
      <template #leading>
        <div class="settings-page-bar">
          <QButton
            v-if="mobileShowBack"
            class="outlined xs icon settings-page-bar-back"
            :title="t('settings_title')"
            :aria-label="t('settings_title')"
            @click="showIndexView"
          >
            <QIconArrowLeft class="icon" />
          </QButton>
          <h2 class="page-title page-bar-title workspace-section-title">{{ mobileBarTitle }}</h2>
        </div>
      </template>
      <div class="settings-workbench">
        <aside v-if="showIndexPane" class="settings-index workspace-sidebar-section">
          <div class="settings-index-items workspace-sidebar-list">
            <button
              v-for="item in settingsSections"
              :key="item.id"
              type="button"
              :class="sectionClass(item)"
              :aria-current="isSelectedSection(item) ? 'page' : undefined"
              @click="selectSection(item.id)"
            >
              <span class="workspace-sidebar-item-copy">
                <span class="workspace-sidebar-item-title">{{ item.title }}</span>
                <span class="workspace-sidebar-item-meta">{{ item.meta }}</span>
              </span>
              <span class="workspace-sidebar-item-marker">
                <QBadge v-if="isSelectedSection(item)" dot type="primary" size="sm" />
              </span>
            </button>
          </div>
        </aside>

        <QCard v-if="showPanelPane && selectedSection" class="settings-panel-card" variant="default">
          <div class="settings-panel-shell">
            <header class="settings-panel-head">
              <div class="settings-panel-copy">
                <p class="ui-kicker">{{ panelKicker }}</p>
                <h3 class="settings-panel-title workspace-document-title">{{ selectedSection.title }}</h3>
                <p class="settings-panel-meta">{{ panelHint }}</p>
              </div>
              <div class="settings-panel-actions">
                <QButton
                  v-if="activeSaveKind === 'agent'"
                  class="primary"
                  :loading="agentSaving"
                  :disabled="agentSaveDisabled"
                  @click="saveAgentSettings"
                >
                  {{ t("action_save") }}
                </QButton>
                <QButton
                  v-else-if="activeSaveKind === 'console'"
                  class="primary"
                  :loading="consoleSaving"
                  :disabled="consoleSaveDisabled"
                  @click="saveConsoleSettings"
                >
                  {{ t("action_save") }}
                </QButton>
              </div>
            </header>

            <div class="settings-panel-notices">
              <QFence v-if="activeSaveKind === 'agent' && agentErr" type="danger" icon="QIconCloseCircle" :text="agentErr" />
              <QFence v-if="activeSaveKind === 'agent' && agentOk" type="success" icon="QIconCheckCircle" :text="agentOk" />
              <QFence
                v-if="activeSaveKind === 'console' && consoleErr"
                type="danger"
                icon="QIconCloseCircle"
                :text="consoleErr"
              />
              <QFence
                v-if="activeSaveKind === 'console' && consoleOk"
                type="success"
                icon="QIconCheckCircle"
                :text="consoleOk"
              />
            </div>

            <div class="settings-panel-body">
              <div v-if="selectedSection.id === 'agent'" class="settings-form-grid">
                <label class="settings-field is-wide">
                  <span class="settings-field-label">{{ t("settings_agent_provider_label") }}</span>
                  <div v-if="isLLMFieldEnvManaged('provider')" class="settings-env-managed">
                    <code class="settings-env-managed-env">{{ llmFieldManagedHeadline("provider") }}</code>
                    <p class="settings-env-managed-body">{{ t("settings_env_managed_body") }}</p>
                  </div>
                  <QDropdownMenu
                    v-else
                    :key="state.llm.provider || 'provider'"
                    :items="providerItems"
                    :initialItem="providerItem"
                    :placeholder="t('settings_agent_provider_placeholder')"
                    @change="onProviderChange"
                  />
                </label>
                <label v-if="!showCloudflareAccountField" class="settings-field is-wide">
                  <span class="settings-field-label">{{ t("settings_agent_endpoint_label") }}</span>
                  <div v-if="isLLMFieldEnvManaged('endpoint')" class="settings-env-managed">
                    <code class="settings-env-managed-env">{{ llmFieldManagedHeadline("endpoint") }}</code>
                    <p class="settings-env-managed-body">{{ t("settings_env_managed_body") }}</p>
                  </div>
                  <div v-else class="settings-field-control">
                    <QInput
                      v-model="state.llm.endpoint"
                      :placeholder="t('settings_agent_endpoint_placeholder')"
                      :disabled="agentLoading || agentSaving"
                    />
                    <QButton
                      type="button"
                      class="outlined icon settings-field-action"
                      :title="t('setup_llm_api_base_picker_title')"
                      :aria-label="t('setup_llm_api_base_picker_title')"
                      :disabled="!showOpenAICompatibleHelpers || agentLoading || agentSaving"
                      @click.prevent="openAPIBasePicker"
                    >
                      <QIconLink class="icon" />
                    </QButton>
                  </div>
                </label>
                <label v-if="showCloudflareAccountField" class="settings-field is-wide">
                  <span class="settings-field-label">{{ t("settings_agent_cloudflare_account_label") }}</span>
                  <div v-if="isLLMFieldEnvManaged('cloudflare_account_id')" class="settings-env-managed">
                    <code class="settings-env-managed-env">{{ llmFieldManagedHeadline("cloudflare_account_id") }}</code>
                    <p class="settings-env-managed-body">{{ t("settings_env_managed_body") }}</p>
                  </div>
                  <QInput
                    v-else
                    v-model="state.llm.cloudflare_account_id"
                    :placeholder="t('settings_agent_cloudflare_account_placeholder')"
                    :disabled="agentLoading || agentSaving"
                  />
                </label>
                <label class="settings-field is-wide">
                  <span class="settings-field-label">{{ t(credentialLabelKey) }}</span>
                  <div
                    v-if="showCloudflareAccountField ? isLLMFieldEnvManaged('cloudflare_api_token') : isLLMFieldEnvManaged('api_key')"
                    class="settings-env-managed"
                  >
                    <code class="settings-env-managed-env">{{ llmFieldManagedHeadline(showCloudflareAccountField ? "cloudflare_api_token" : "api_key") }}</code>
                    <p class="settings-env-managed-body">{{ t("settings_env_managed_body") }}</p>
                  </div>
                  <QInput
                    v-else-if="showCloudflareAccountField"
                    v-model="state.llm.cloudflare_api_token"
                    inputType="password"
                    :placeholder="t(credentialPlaceholderKey)"
                    :disabled="agentLoading || agentSaving"
                  />
                  <QInput
                    v-else
                    v-model="state.llm.api_key"
                    inputType="password"
                    :placeholder="t(credentialPlaceholderKey)"
                    :disabled="agentLoading || agentSaving"
                  />
                  <p v-if="credentialHelp" class="settings-field-hint">
                    <button v-if="credentialHelp.url" type="button" class="settings-field-link" @click="openExternal(credentialHelp.url)">
                      <span>{{ credentialHelpParts?.before }}</span>
                      <span class="settings-field-link-provider">{{ credentialHelp.title }}</span>
                      <span>{{ credentialHelpParts?.after }}</span>
                      <QIconArrowUpRight class="icon settings-field-link-icon" />
                    </button>
                    <span v-else class="settings-field-link is-static">
                      {{ t(credentialHintPlainKey, { provider: credentialHelp.title }) }}
                    </span>
                  </p>
                </label>
                <label class="settings-field is-wide">
                  <span class="settings-field-label">{{ t("settings_agent_model_label") }}</span>
                  <div v-if="isLLMFieldEnvManaged('model')" class="settings-env-managed">
                    <code class="settings-env-managed-env">{{ llmFieldManagedHeadline("model") }}</code>
                    <p class="settings-env-managed-body">{{ t("settings_env_managed_body") }}</p>
                  </div>
                  <div v-else class="settings-field-control">
                    <QInput
                      v-model="state.llm.model"
                      :placeholder="t('settings_agent_model_placeholder')"
                      :disabled="agentLoading || agentSaving"
                    />
                    <QButton
                      type="button"
                      class="outlined icon settings-field-action"
                      :title="t('setup_llm_model_picker_title')"
                      :aria-label="t('setup_llm_model_picker_title')"
                      :disabled="modelLookupDisabled"
                      @click.prevent="openModelPicker"
                    >
                      <QIconSearch class="icon" />
                    </QButton>
                  </div>
                </label>
                <label class="settings-field">
                  <span class="settings-field-label">{{ t("settings_llm_reasoning_label") }}</span>
                  <div v-if="isLLMFieldEnvManaged('reasoning_effort')" class="settings-env-managed">
                    <code class="settings-env-managed-env">{{ llmFieldManagedHeadline("reasoning_effort") }}</code>
                    <p class="settings-env-managed-body">{{ t("settings_env_managed_body") }}</p>
                  </div>
                  <QDropdownMenu
                    v-else
                    :key="state.llm.reasoning_effort || 'reasoning'"
                    :items="reasoningEffortItems"
                    :initialItem="reasoningEffortItem"
                    :placeholder="t('settings_llm_reasoning_placeholder')"
                    @change="onReasoningEffortChange"
                  />
                </label>
                <label class="settings-field">
                  <span class="settings-field-label">{{ t("settings_llm_tools_emulation_label") }}</span>
                  <div v-if="isLLMFieldEnvManaged('tools_emulation_mode')" class="settings-env-managed">
                    <code class="settings-env-managed-env">{{ llmFieldManagedHeadline("tools_emulation_mode") }}</code>
                    <p class="settings-env-managed-body">{{ t("settings_env_managed_body") }}</p>
                  </div>
                  <QDropdownMenu
                    v-else
                    :key="state.llm.tools_emulation_mode || 'tools-emulation'"
                    :items="toolsEmulationItems"
                    :initialItem="toolsEmulationItem"
                    :placeholder="t('settings_llm_tools_emulation_placeholder')"
                    @change="onToolsEmulationChange"
                  />
                </label>
                <div class="settings-agent-actions">
                  <QButton type="button" class="outlined settings-aux-action" :disabled="testConnectionDisabled" @click="openTestConnection">
                    {{ t("setup_llm_test_button") }}
                  </QButton>
                </div>
              </div>

              <div v-else-if="selectedSection.id === 'inputs'" class="settings-toggle-list">
                <div v-for="item in multimodalItems" :key="item.id" class="settings-toggle-row">
                  <div class="settings-toggle-copy">
                    <strong class="settings-toggle-title">{{ t(item.titleKey) }}</strong>
                    <span class="settings-toggle-note">{{ t(item.noteKey) }}</span>
                  </div>
                  <QSwitch
                    :modelValue="state.multimodal[item.id]"
                    :disabled="agentLoading || agentSaving"
                    @update:modelValue="setMultimodalSource(item.id, $event)"
                  />
                </div>
              </div>

              <div v-else-if="selectedSection.id === 'tools'" class="settings-toggle-list">
                <div v-for="item in toolItems" :key="item.id" class="settings-toggle-row">
                  <div class="settings-toggle-copy">
                    <strong class="settings-toggle-title">{{ t(item.titleKey) }}</strong>
                    <span class="settings-toggle-note">{{ t(item.noteKey) }}</span>
                  </div>
                  <QSwitch
                    :modelValue="state.tools[item.id]"
                    :disabled="agentLoading || agentSaving"
                    @update:modelValue="setToolEnabled(item.id, $event)"
                  />
                </div>
              </div>

              <div v-else-if="selectedSection.id === 'runtimes'" class="settings-toggle-list">
                <div v-for="item in managedRuntimeItems" :key="item.id" class="settings-toggle-row">
                  <div class="settings-toggle-copy">
                    <strong class="settings-toggle-title">{{ t(item.titleKey) }}</strong>
                    <span class="settings-toggle-note">{{ t(item.noteKey) }}</span>
                  </div>
                  <QSwitch
                    :modelValue="state.managedRuntimes[item.id]"
                    :disabled="consoleLoading || consoleSaving"
                    @update:modelValue="setManagedRuntimeEnabled(item.id, $event)"
                  />
                </div>
              </div>

              <div v-else class="settings-console-list">
                <div class="settings-console-row">
                  <div class="settings-card-copy">
                    <h4 class="settings-card-title">{{ t("settings_language_title") }}</h4>
                    <p class="settings-card-note">{{ t("settings_language_hint") }}</p>
                  </div>
                  <QLanguageSelector class="settings-console-control" :lang="lang" :presist="true" @change="onLanguageChange" />
                </div>
                <div class="settings-console-row settings-console-row-end">
                  <div class="settings-card-copy">
                    <h4 class="settings-card-title">{{ t("settings_session_title") }}</h4>
                    <p class="settings-card-note">{{ t("settings_session_hint") }}</p>
                  </div>
                  <QButton class="danger settings-console-control" :loading="loggingOut" @click="logout">
                    {{ t("action_logout") }}
                  </QButton>
                </div>
              </div>
            </div>
          </div>
        </QCard>
      </div>

      <SetupPickerDialog
        v-model="apiBasePickerOpen"
        :items="apiBasePickerItems"
        :loading="false"
        :error="''"
        :filterPlaceholder="t('setup_llm_api_base_picker_filter_placeholder')"
        :emptyText="t('setup_llm_api_base_picker_empty')"
        @select="applyAPIBaseOption"
      />

      <SetupPickerDialog
        v-model="modelPickerOpen"
        :items="modelPickerItems"
        :loading="modelPickerLoading"
        :error="modelPickerError"
        :filterPlaceholder="t('setup_llm_model_picker_filter_placeholder')"
        :emptyText="t('setup_llm_model_picker_empty')"
        :showValue="false"
        @select="applyModelOption"
      />

      <SetupConnectionTestDialog
        v-model="testConnectionOpen"
        :loading="testConnectionLoading"
        :error="testConnectionError"
        :benchmarks="testConnectionBenchmarks"
        :provider="testConnectionMeta.provider"
        :model="testConnectionMeta.model"
        @retry="runConnectionTest"
      />
    </AppPage>
  `,
};

export default SettingsView;
