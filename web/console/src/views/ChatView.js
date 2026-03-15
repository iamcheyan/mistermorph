import { computed, nextTick, onMounted, onUnmounted, ref, watch } from "vue";
import "./ChatView.css";

import AppPage from "../components/AppPage";
import { endpointChannelLabel } from "../core/endpoints";
import {
  endpointState,
  runtimeApiFetchForEndpoint,
  runtimeEndpointByRef,
  translate,
} from "../core/context";

const POLL_INTERVAL_MS = 1200;
const COMPOSER_MAX_ROWS = 5;

function normalizeTaskStatus(raw) {
  const value = String(raw || "").trim().toLowerCase();
  switch (value) {
    case "queued":
    case "running":
    case "pending":
    case "done":
    case "failed":
    case "canceled":
      return value;
    default:
      return "queued";
  }
}

function isTerminalStatus(status) {
  return status === "done" || status === "failed" || status === "canceled";
}

function stringifyResult(result) {
  if (typeof result === "string") {
    return result.trim();
  }
  if (result === undefined || result === null) {
    return "";
  }
  try {
    return JSON.stringify(result, null, 2);
  } catch {
    return String(result);
  }
}

function newHistoryID() {
  return `${Date.now()}_${Math.random().toString(16).slice(2, 10)}`;
}

const ChatView = {
  components: {
    AppPage,
  },
  setup() {
    const t = translate;
    const chatHistoryItems = ref([]);
    const taskInput = ref("");
    const sending = ref(false);
    const err = ref("");
    const pollTimers = new Set();
    const composerField = ref(null);

    const selectedEndpoint = computed(() => runtimeEndpointByRef(endpointState.selectedRef));
    const submitEndpointRef = computed(() => {
      const selected = selectedEndpoint.value;
      if (!selected) {
        return "";
      }
      const mapped = String(selected.submit_endpoint_ref || "").trim();
      if (mapped) {
        return mapped;
      }
      return selected.can_submit ? String(selected.endpoint_ref || "").trim() : "";
    });
    const submitBlockedMessage = computed(() => {
      const selected = selectedEndpoint.value;
      if (!selected || !selected.connected) {
        return "";
      }
      if (submitEndpointRef.value) {
        return "";
      }
      return t("chat_submit_unsupported", {
        name: selected.name || selected.endpoint_ref || "-",
      });
    });
    const chatReadonly = computed(() => Boolean(submitBlockedMessage.value));
    const readonlyTitle = computed(() => {
      return t("chat_readonly_title", {
        channel: endpointChannelLabel(selectedEndpoint.value?.mode, t),
      });
    });
    const readonlyReason = computed(() => {
      const selected = selectedEndpoint.value;
      if (!selected) {
        return "";
      }
      return t("chat_readonly_reason", {
        name: selected.name || selected.endpoint_ref || "-",
        channel: endpointChannelLabel(selected.mode, t),
      });
    });
    const composerDisabled = computed(() => Boolean(submitBlockedMessage.value) || sending.value);
    const sendDisabled = computed(
      () => composerDisabled.value || String(taskInput.value || "").trim() === ""
    );

    function composerTextarea() {
      const root = composerField.value?.$el || composerField.value;
      if (!root || typeof root.querySelector !== "function") {
        return null;
      }
      return root.querySelector("textarea");
    }

    function syncComposerHeight() {
      void nextTick(() => {
        const textarea = composerTextarea();
        if (!textarea) {
          return;
        }
        const styles = window.getComputedStyle(textarea);
        const lineHeight = Number.parseFloat(styles.lineHeight) || 20;
        const paddingTop = Number.parseFloat(styles.paddingTop) || 0;
        const paddingBottom = Number.parseFloat(styles.paddingBottom) || 0;
        const borderTop = Number.parseFloat(styles.borderTopWidth) || 0;
        const borderBottom = Number.parseFloat(styles.borderBottomWidth) || 0;
        const minHeight = lineHeight + paddingTop + paddingBottom + borderTop + borderBottom;
        const maxHeight =
          lineHeight * COMPOSER_MAX_ROWS + paddingTop + paddingBottom + borderTop + borderBottom;

        textarea.style.height = "auto";
        const nextHeight = Math.max(minHeight, Math.min(textarea.scrollHeight, maxHeight));
        textarea.style.height = `${nextHeight}px`;
        textarea.style.overflowY = textarea.scrollHeight > maxHeight ? "auto" : "hidden";
      });
    }

    function clearPollTimers() {
      for (const timerID of pollTimers) {
        window.clearTimeout(timerID);
      }
      pollTimers.clear();
    }

    function roleText(role) {
      if (role === "user") {
        return t("chat_role_user");
      }
      if (role === "agent") {
        return t("chat_role_agent");
      }
      return t("chat_role_system");
    }

    function statusText(status) {
      switch (normalizeTaskStatus(status)) {
        case "running":
          return t("status_running");
        case "pending":
          return t("status_pending");
        case "done":
          return t("status_done");
        case "failed":
          return t("status_failed");
        case "canceled":
          return t("status_canceled");
        default:
          return t("status_queued");
      }
    }

    function historyClass(item) {
      const role = String(item?.role || "").trim().toLowerCase();
      if (role === "user") {
        return "chat-history-item chat-history-user";
      }
      if (role === "agent") {
        return "chat-history-item chat-history-agent";
      }
      return "chat-history-item chat-history-system";
    }

    function pushHistoryItem(partial) {
      const item = {
        id: newHistoryID(),
        role: String(partial?.role || "system"),
        text: String(partial?.text || ""),
        status: String(partial?.status || ""),
        taskId: String(partial?.taskId || ""),
      };
      chatHistoryItems.value = [...chatHistoryItems.value, item];
      return item.id;
    }

    function patchHistoryItem(id, patch) {
      const idx = chatHistoryItems.value.findIndex((item) => item.id === id);
      if (idx < 0) {
        return;
      }
      chatHistoryItems.value[idx] = {
        ...chatHistoryItems.value[idx],
        ...patch,
      };
    }

    function schedulePoll(fn) {
      const timerID = window.setTimeout(async () => {
        pollTimers.delete(timerID);
        await fn();
      }, POLL_INTERVAL_MS);
      pollTimers.add(timerID);
    }

    async function pollTask(taskID, historyID, endpointRef) {
      try {
        const detail = await runtimeApiFetchForEndpoint(endpointRef, `/tasks/${encodeURIComponent(taskID)}`);
        const status = normalizeTaskStatus(detail?.status);
        const hasResult = detail && detail.result !== undefined && detail.result !== null;
        const resultText = stringifyResult(detail?.result);
        const errorText = String(detail?.error || "").trim();
        const lines = [`${t("chat_task_prefix")} ${taskID}`];
        if (hasResult && resultText) {
          lines.push(resultText);
        } else if (errorText) {
          lines.push(errorText);
        } else if (isTerminalStatus(status)) {
          lines.push(t("chat_result_empty"));
        } else {
          lines.push(t("chat_polling_hint"));
        }
        patchHistoryItem(historyID, {
          status,
          text: lines.join("\n\n"),
        });
        if (!isTerminalStatus(status)) {
          schedulePoll(async () => {
            await pollTask(taskID, historyID, endpointRef);
          });
        }
      } catch (e) {
        patchHistoryItem(historyID, {
          status: "failed",
          text: e?.message || t("msg_load_failed"),
        });
      }
    }

    async function submitTask() {
      const task = String(taskInput.value || "").trim();
      if (!task || sending.value) {
        return;
      }
      const endpointRef = submitEndpointRef.value;
      if (!endpointRef) {
        err.value = submitBlockedMessage.value || t("msg_select_endpoint");
        return;
      }
      sending.value = true;
      err.value = "";
      taskInput.value = "";

      pushHistoryItem({
        role: "user",
        text: task,
      });
      const agentHistoryID = pushHistoryItem({
        role: "agent",
        text: t("chat_agent_waiting"),
        status: "queued",
      });

      try {
        const submitted = await runtimeApiFetchForEndpoint(endpointRef, "/tasks", {
          method: "POST",
          body: { task },
        });
        const taskID = String(submitted?.id || "").trim();
        const status = normalizeTaskStatus(submitted?.status);
        if (!taskID) {
          throw new Error(t("chat_missing_task_id"));
        }
        patchHistoryItem(agentHistoryID, {
          taskId: taskID,
          status,
          text: `${t("chat_task_prefix")} ${taskID}\n\n${t("chat_polling_hint")}`,
        });
        await pollTask(taskID, agentHistoryID, endpointRef);
      } catch (e) {
        const message = e?.message || t("msg_load_failed");
        err.value = message;
        patchHistoryItem(agentHistoryID, {
          status: "failed",
          text: message,
        });
      } finally {
        sending.value = false;
      }
    }

    onMounted(() => {
      pushHistoryItem({
        role: "system",
        text: t("chat_intro"),
      });
      syncComposerHeight();
    });
    onUnmounted(() => {
      clearPollTimers();
    });
    watch(
      () => endpointState.selectedRef,
      () => {
        clearPollTimers();
        err.value = "";
        syncComposerHeight();
      }
    );
    watch(taskInput, () => {
      syncComposerHeight();
    });

    return {
      t,
      chatHistoryItems,
      taskInput,
      sending,
      err,
      composerField,
      submitBlockedMessage,
      chatReadonly,
      readonlyTitle,
      readonlyReason,
      composerDisabled,
      sendDisabled,
      submitTask,
      roleText,
      statusText,
      historyClass,
    };
  },
  template: `
    <AppPage :title="t('chat_title')" class="chat-page">
      <QFence v-if="err" type="danger" icon="QIconCloseCircle" :text="err" />
      <section v-if="chatReadonly" class="chat-readonly frame">
        <h3 class="chat-readonly-title">{{ readonlyTitle }}</h3>
        <p class="chat-readonly-text">{{ readonlyReason }}</p>
      </section>
      <template v-else>
        <div class="chat-history">
          <article v-for="item in chatHistoryItems" :key="item.id" :class="historyClass(item)">
            <header class="chat-history-head">
              <code class="chat-history-role">{{ roleText(item.role) }}</code>
              <code v-if="item.status" class="chat-history-status">{{ statusText(item.status) }}</code>
            </header>
            <pre class="chat-history-body">{{ item.text }}</pre>
            <code v-if="item.taskId" class="chat-history-task">{{ t("chat_task_prefix") }} {{ item.taskId }}</code>
          </article>
          <p v-if="chatHistoryItems.length === 0" class="muted">{{ t("chat_empty") }}</p>
        </div>
        <div class="chat-composer">
          <QTextarea
            ref="composerField"
            v-model="taskInput"
            :rows="1"
            :disabled="composerDisabled"
            :placeholder="t('chat_input_placeholder')"
            @keydown.enter.exact.prevent="submitTask"
          >
            <template #append>
              <div class="chat-composer-append">
                <QButton
                  class="primary sm icon chat-composer-send"
                  :loading="sending"
                  :disabled="sendDisabled"
                  :title="t('chat_action_send')"
                  :aria-label="t('chat_action_send')"
                  @click="submitTask"
                >
                  <QIconSend class="icon" />
                </QButton>
              </div>
            </template>
          </QTextarea>
        </div>
      </template>
    </AppPage>
  `,
};

export default ChatView;
