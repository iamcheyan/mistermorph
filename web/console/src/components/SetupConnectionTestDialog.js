import { computed } from "vue";
import { translate } from "../core/context";
import "./SetupConnectionTestDialog.css";

const TEST_CONNECTION_BENCHMARK_IDS = ["text_reply", "json_response", "tool_calling"];

const SetupConnectionTestDialog = {
  props: {
    modelValue: Boolean,
    loading: Boolean,
    error: {
      type: String,
      default: "",
    },
    benchmarks: {
      type: Array,
      default: () => [],
    },
    provider: {
      type: String,
      default: "",
    },
    model: {
      type: String,
      default: "",
    },
    showIntro: {
      type: Boolean,
      default: true,
    },
  },
  emits: ["update:modelValue", "retry"],
  setup(props, { emit }) {
    const t = translate;

    const hasBenchmarks = computed(() => Array.isArray(props.benchmarks) && props.benchmarks.length > 0);
    const visibleBenchmarks = computed(() => {
      if (props.loading && !hasBenchmarks.value) {
        return TEST_CONNECTION_BENCHMARK_IDS.map((id) => ({
          id,
          ok: false,
          running: true,
          duration_ms: 0,
          detail: "",
          error: "",
        }));
      }
      return (Array.isArray(props.benchmarks) ? props.benchmarks : []).map((item) => ({
        ...item,
        running: false,
      }));
    });
    const showBenchmarks = computed(() => props.loading || hasBenchmarks.value);

    function close() {
      emit("update:modelValue", false);
    }

    function retry() {
      emit("retry");
    }

    function formatBenchmarkSeconds(durationMS) {
      const ms = Number(durationMS || 0);
      if (!Number.isFinite(ms) || ms <= 0) {
        return "0s";
      }
      const seconds = ms / 1000;
      const digits = seconds >= 10 ? 1 : 2;
      return `${seconds
        .toFixed(digits)
        .replace(/\.0+$/, "")
        .replace(/(\.\d*[1-9])0+$/, "$1")}s`;
    }

    return {
      t,
      visibleBenchmarks,
      showBenchmarks,
      close,
      retry,
      formatBenchmarkSeconds,
    };
  },
  template: `
    <QDialog
      :modelValue="modelValue"
      width="560px"
      @update:modelValue="$emit('update:modelValue', $event)"
      @close="close"
    >
      <section class="connection-test-dialog">
        <header class="connection-test-head">
          <div class="connection-test-copy">
            <h3 class="connection-test-title">{{ t("setup_llm_test_title") }}</h3>
            <p v-if="showIntro" class="connection-test-intro">{{ t("setup_llm_test_intro") }}</p>
          </div>
        </header>

        <QFence
          v-if="error"
          type="danger"
          icon="QIconCloseCircle"
          :text="error"
        />

        <div v-if="showBenchmarks" class="connection-test-result">
          <p class="connection-test-result-label">
            {{ t("setup_llm_test_success", { provider: provider || t('ttl_unknown'), model: model || t('ttl_unknown') }) }}
          </p>
          <div class="connection-test-benchmark-list">
            <article
              v-for="item in visibleBenchmarks"
              :key="item.id"
              :class="['connection-test-benchmark', { 'is-ok': item.ok, 'is-failed': !item.ok, 'is-loading': item.running }]"
            >
              <div class="connection-test-benchmark-main">
                <p class="connection-test-benchmark-title">{{ t('setup_llm_test_benchmark_' + item.id) }}</p>
                <p v-if="item.running" class="connection-test-benchmark-detail">{{ t("setup_llm_test_running") }}</p>
                <p v-else-if="item.ok && item.detail" class="connection-test-benchmark-detail">{{ item.detail }}</p>
                <p v-else-if="item.error" class="connection-test-benchmark-error">{{ item.error }}</p>
              </div>
              <div class="connection-test-benchmark-side">
                <div class="connection-test-benchmark-status-row">
                  <span v-if="item.running" class="connection-test-benchmark-spinner" aria-hidden="true"></span>
                  <strong v-else class="connection-test-benchmark-status">
                    {{ item.ok ? t("setup_llm_test_status_ok") : t("setup_llm_test_status_failed") }}
                  </strong>
                </div>
                <span v-if="!item.running" class="connection-test-benchmark-time">{{ formatBenchmarkSeconds(item.duration_ms) }}</span>
              </div>
            </article>
          </div>
        </div>

        <div class="connection-test-actions">
          <QButton class="outlined" @click="close">{{ t("action_close") }}</QButton>
          <QButton class="primary" :loading="loading" @click="retry">
            {{ loading ? t("setup_llm_test_running") : t("setup_llm_test_retry") }}
          </QButton>
        </div>
      </section>
    </QDialog>
  `,
};

export default SetupConnectionTestDialog;
