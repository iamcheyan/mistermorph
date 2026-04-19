import { computed, onMounted, ref, watch } from "vue";
import "./StatsView.css";

import AppKicker from "../components/AppKicker";
import AppPage from "../components/AppPage";
import { endpointState, formatTime, runtimeApiFetch, translate } from "../core/context";

function hasMetricValue(totals, key) {
  return Boolean(totals) && Object.prototype.hasOwnProperty.call(totals, key);
}

function formatNumber(value) {
  const n = Number(value || 0);
  if (!Number.isFinite(n)) {
    return "0";
  }
  return Math.trunc(n).toLocaleString();
}

function formatCost(value, currency = "USD") {
  const n = Number(value);
  if (!Number.isFinite(n)) {
    return "-";
  }
  try {
    return new Intl.NumberFormat(undefined, {
      style: "currency",
      currency: String(currency || "USD").toUpperCase(),
      minimumFractionDigits: Math.abs(n) > 0 && Math.abs(n) < 1 ? 4 : 2,
      maximumFractionDigits: 6,
    }).format(n);
  } catch {
    return `${String(currency || "USD").toUpperCase()} ${n.toFixed(4)}`;
  }
}

function formatFixedCost(value, currency = "USD", fractionDigits = 6) {
  const n = Number(value);
  if (!Number.isFinite(n)) {
    return "-";
  }
  try {
    return new Intl.NumberFormat(undefined, {
      style: "currency",
      currency: String(currency || "USD").toUpperCase(),
      minimumFractionDigits: fractionDigits,
      maximumFractionDigits: fractionDigits,
    }).format(n);
  } catch {
    return `${String(currency || "USD").toUpperCase()} ${n.toFixed(fractionDigits)}`;
  }
}

function summaryLeadMetric(t, totals) {
  const costCurrency = typeof totals?.cost_currency === "string" ? totals.cost_currency : "USD";
  if (hasMetricValue(totals, "total_cost")) {
    return {
      key: "total_cost",
      label: t("stats_total_cost"),
      value: formatCost(totals.total_cost, costCurrency),
    };
  }
  return {
    key: "total_tokens",
    label: t("stats_total_tokens"),
    value: formatNumber(totals.total_tokens),
  };
}

function summaryPrimaryMetrics(t, totals) {
  const leadKey = summaryLeadMetric(t, totals).key;
  return [
    { key: "total_tokens", label: t("stats_total_tokens"), value: formatNumber(totals.total_tokens) },
    { key: "input_tokens", label: t("stats_input_tokens"), value: formatNumber(totals.input_tokens) },
    { key: "output_tokens", label: t("stats_output_tokens"), value: formatNumber(totals.output_tokens) },
    { key: "requests", label: t("stats_requests"), value: formatNumber(totals.requests) },
    { key: "cached_input_tokens", label: t("stats_cached_input_tokens"), value: formatNumber(totals.cached_input_tokens) },
    {
      key: "cache_creation_input_tokens",
      label: t("stats_cache_creation_input_tokens"),
      value: formatNumber(totals.cache_creation_input_tokens),
    },
  ].filter((item) => item.key !== leadKey);
}

function summaryCostMetrics(t, totals) {
  const costCurrency = typeof totals?.cost_currency === "string" ? totals.cost_currency : "USD";
  return [
    {
      key: "input_cost",
      label: t("stats_input_cost"),
      value: formatCost(totals?.input_cost, costCurrency),
      available: hasMetricValue(totals, "input_cost"),
    },
    {
      key: "output_cost",
      label: t("stats_output_cost"),
      value: formatCost(totals?.output_cost, costCurrency),
      available: hasMetricValue(totals, "output_cost"),
    },
    {
      key: "cached_input_cost",
      label: t("stats_cached_input_cost"),
      value: formatCost(totals?.cached_input_cost, costCurrency),
      available: hasMetricValue(totals, "cached_input_cost"),
    },
    {
      key: "cache_creation_input_cost",
      label: t("stats_cache_creation_input_cost"),
      value: formatCost(totals?.cache_creation_input_cost, costCurrency),
      available: hasMetricValue(totals, "cache_creation_input_cost"),
    },
  ].filter((item) => item.available);
}

function costMetrics(t, totals) {
  const costCurrency = typeof totals?.cost_currency === "string" ? totals.cost_currency : "USD";
  return [
    {
      key: "total_cost",
      label: t("stats_total_cost"),
      value: hasMetricValue(totals, "total_cost") ? formatCost(totals.total_cost, costCurrency) : "-",
      unavailable: !hasMetricValue(totals, "total_cost"),
    },
    {
      key: "input_cost",
      label: t("stats_input_cost"),
      value: hasMetricValue(totals, "input_cost") ? formatCost(totals.input_cost, costCurrency) : "-",
      unavailable: !hasMetricValue(totals, "input_cost"),
    },
    {
      key: "output_cost",
      label: t("stats_output_cost"),
      value: hasMetricValue(totals, "output_cost") ? formatCost(totals.output_cost, costCurrency) : "-",
      unavailable: !hasMetricValue(totals, "output_cost"),
    },
    {
      key: "cached_input_cost",
      label: t("stats_cached_input_cost"),
      value: hasMetricValue(totals, "cached_input_cost") ? formatCost(totals.cached_input_cost, costCurrency) : "-",
      unavailable: !hasMetricValue(totals, "cached_input_cost"),
    },
    {
      key: "cache_creation_input_cost",
      label: t("stats_cache_creation_input_cost"),
      value: hasMetricValue(totals, "cache_creation_input_cost")
        ? formatCost(totals.cache_creation_input_cost, costCurrency)
        : "-",
      unavailable: !hasMetricValue(totals, "cache_creation_input_cost"),
    },
  ].filter((item) => item.key === "total_cost" || !item.unavailable);
}

function tokenMetrics(t, totals) {
  return [
    { key: "total_tokens", label: t("stats_total_tokens"), value: formatNumber(totals.total_tokens) },
    { key: "input_tokens", label: t("stats_input_tokens"), value: formatNumber(totals.input_tokens) },
    { key: "output_tokens", label: t("stats_output_tokens"), value: formatNumber(totals.output_tokens) },
    { key: "cached_input_tokens", label: t("stats_cached_input_tokens"), value: formatNumber(totals.cached_input_tokens) },
    {
      key: "cache_creation_input_tokens",
      label: t("stats_cache_creation_input_tokens"),
      value: formatNumber(totals.cache_creation_input_tokens),
    },
  ];
}

function visibleModelCostColumns(t, rows) {
  const columns = [
    { key: "total_cost", label: t("stats_total_cost"), kind: "cost" },
    { key: "input_cost", label: t("stats_input_cost"), kind: "cost" },
    { key: "output_cost", label: t("stats_output_cost"), kind: "cost" },
    { key: "cached_input_cost", label: t("stats_cached_input_cost"), kind: "cost" },
    { key: "cache_creation_input_cost", label: t("stats_cache_creation_input_cost"), kind: "cost" },
  ];
  if (!rows.some((row) => columns.some((column) => hasMetricValue(row, column.key)))) {
    return [];
  }
  return columns.filter((column) => rows.some((row) => hasMetricValue(row, column.key)));
}

function visibleModelTokenColumns(t, rows) {
  const columns = [
    { key: "total_tokens", label: t("stats_total_tokens"), kind: "token", always: true },
    { key: "input_tokens", label: t("stats_input_tokens"), kind: "token", always: true },
    { key: "output_tokens", label: t("stats_output_tokens"), kind: "token", always: true },
    { key: "cached_input_tokens", label: t("stats_cached_input_tokens"), kind: "token" },
    { key: "cache_creation_input_tokens", label: t("stats_cache_creation_input_tokens"), kind: "token" },
  ];
  return columns.filter(
    (column) => column.always || rows.some((row) => hasMetricValue(row, column.key) || Number(row?.[column.key] || 0) > 0)
  );
}

function formatModelLedgerValue(row, column) {
  if (column.kind === "cost") {
    const currency = typeof row?.cost_currency === "string" ? row.cost_currency : "USD";
    return hasMetricValue(row, column.key) ? formatFixedCost(row[column.key], currency) : "-";
  }
  return formatNumber(row?.[column.key]);
}

function isModelLedgerValueUnavailable(row, column) {
  return column.kind === "cost" && !hasMetricValue(row, column.key);
}

const StatsView = {
  components: {
    AppKicker,
    AppPage,
  },
  setup() {
    const t = translate;
    const loading = ref(false);
    const err = ref("");
    const activeTabID = ref("api_hosts");
    const payload = ref({
      updated_at: "",
      projected_records: 0,
      skipped_records: 0,
      summary: {},
      api_hosts: [],
      models: [],
    });

    const statsTabs = computed(() => [
      { id: "api_hosts", title: t("stats_group_api_hosts") },
      { id: "models", title: t("stats_group_models") },
    ]);
    const selectedStatsTab = computed(() => statsTabs.value.find((item) => item.id === activeTabID.value) || statsTabs.value[0] || null);

    const visibleHosts = computed(() => (Array.isArray(payload.value.api_hosts) ? payload.value.api_hosts : []));
    const visibleModels = computed(() => (Array.isArray(payload.value.models) ? payload.value.models : []));
    const primarySummaryMetric = computed(() => summaryLeadMetric(t, payload.value.summary || {}));
    const secondarySummaryMetrics = computed(() => summaryPrimaryMetrics(t, payload.value.summary || {}));
    const secondarySummaryCosts = computed(() => summaryCostMetrics(t, payload.value.summary || {}));
    const summaryMetaItems = computed(() => {
      const items = [];
      if (payload.value.updated_at) {
        items.push(`${t("stats_updated_at")}: ${formatTime(payload.value.updated_at)}`);
      }
      items.push(`${t("stats_projected_records")}: ${formatNumber(payload.value.projected_records)}`);
      if (Number(payload.value.skipped_records || 0) > 0) {
        items.push(`${t("stats_skipped_records")}: ${formatNumber(payload.value.skipped_records)}`);
      }
      return items;
    });
    async function load() {
      loading.value = true;
      err.value = "";
      try {
        const data = await runtimeApiFetch("/stats/llm/usage");
        payload.value = {
          updated_at: typeof data.updated_at === "string" ? data.updated_at : "",
          projected_records: Number(data.projected_records || 0),
          skipped_records: Number(data.skipped_records || 0),
          summary: data.summary && typeof data.summary === "object" ? data.summary : {},
          api_hosts: Array.isArray(data.api_hosts) ? data.api_hosts : [],
          models: Array.isArray(data.models) ? data.models : [],
        };
      } catch (e) {
        err.value = e.message || t("msg_load_failed");
      } finally {
        loading.value = false;
      }
    }

    function hostCostMetrics(item) {
      return costMetrics(t, item || {});
    }

    function hostTokenMetrics(item) {
      return tokenMetrics(t, item || {});
    }

    function modelLedgerCostColumns(items) {
      return visibleModelCostColumns(t, Array.isArray(items) ? items : []);
    }

    function modelLedgerTokenColumns(items) {
      return visibleModelTokenColumns(t, Array.isArray(items) ? items : []);
    }

    function onTabChange(detail) {
      const nextID = String(detail?.tab?.id || "").trim();
      activeTabID.value = nextID || "api_hosts";
    }

    onMounted(load);
    watch(
      () => endpointState.selectedRef,
      () => {
        void load();
      }
    );

    return {
      t,
      loading,
      err,
      payload,
      statsTabs,
      selectedStatsTab,
      visibleHosts,
      visibleModels,
      primarySummaryMetric,
      secondarySummaryMetrics,
      secondarySummaryCosts,
      summaryMetaItems,
      onTabChange,
      hostCostMetrics,
      hostTokenMetrics,
      modelLedgerCostColumns,
      modelLedgerTokenColumns,
      formatModelLedgerValue,
      isModelLedgerValueUnavailable,
      formatNumber,
    };
  },
  template: `
    <AppPage :title="t('stats_title')">
      <QProgress v-if="loading" :infinite="true" />
      <QFence v-if="err" type="danger" icon="QIconCloseCircle" :text="err" />

      <section class="stats-page">
        <header class="stats-hero">
          <div class="stats-hero-copy">
            <div class="stats-hero-head">
              <AppKicker as="h3" left="LLM" right="Usage" />
              <p v-if="summaryMetaItems.length > 0" class="stats-hero-meta">
                <span v-for="item in summaryMetaItems" :key="item" class="stats-hero-meta-item">{{ item }}</span>
              </p>
            </div>
            <div class="stats-lead-block">
              <span class="stats-lead-label">{{ primarySummaryMetric.label }}</span>
              <span class="stats-lead-value">{{ primarySummaryMetric.value }}</span>
            </div>
            <div v-if="secondarySummaryCosts.length > 0" class="stats-inline-meta stats-inline-meta-summary">
              <div v-for="item in secondarySummaryCosts" :key="item.key" class="stats-inline-meta-item">
                <span class="stats-inline-meta-label">{{ item.label }}</span>
                <span class="stats-inline-meta-value">{{ item.value }}</span>
              </div>
            </div>
          </div>
          <div class="stats-glance-grid">
            <div v-for="item in secondarySummaryMetrics" :key="item.key" class="stats-glance-item">
              <span class="stats-glance-label">{{ item.label }}</span>
              <span class="stats-glance-value">{{ item.value }}</span>
            </div>
          </div>
        </header>

        <section class="stats-section">
          <QTabs
            class="stats-section-tabs"
            :tabs="statsTabs"
            :modelValue="selectedStatsTab"
            variant="plain"
            @change="onTabChange"
          />

          <div v-if="selectedStatsTab && selectedStatsTab.id === 'api_hosts'" class="stats-section-panel">
            <div v-if="visibleHosts.length === 0" class="stats-empty">{{ t("stats_no_data") }}</div>
            <div v-else class="stats-host-list">
              <article v-for="host in visibleHosts" :key="host.api_host" class="stats-host-block">
                <header class="stats-host-head">
                  <div class="stats-host-ident">
                    <span class="stats-host-eyebrow">{{ t("stats_api_host") }}</span>
                    <code class="stats-host-name">{{ host.api_host }}</code>
                  </div>
                  <div class="stats-request-pill">
                    <span class="stats-request-pill-label">{{ t("stats_requests") }}</span>
                    <span class="stats-request-pill-value">{{ formatNumber(host.requests) }}</span>
                  </div>
                </header>

                <section class="stats-band stats-band-cost">
                  <header class="stats-band-head">
                    <QIconWallet class="stats-band-icon icon" />
                    <span class="stats-band-title">{{ t("stats_costs") }}</span>
                  </header>
                  <div class="stats-band-grid">
                    <div v-for="item in hostCostMetrics(host)" :key="host.api_host + ':cost:' + item.key" class="stats-band-cell">
                      <span class="stats-ledger-label">{{ item.label }}</span>
                      <span class="stats-ledger-value" :class="{ 'stats-ledger-value-unavailable': item.unavailable }">{{ item.value }}</span>
                    </div>
                  </div>
                </section>

                <section class="stats-band stats-band-token">
                  <header class="stats-band-head">
                    <QIconBarChart class="stats-band-icon icon" />
                    <span class="stats-band-title">{{ t("stats_tokens") }}</span>
                  </header>
                  <div class="stats-band-grid">
                    <div v-for="item in hostTokenMetrics(host)" :key="host.api_host + ':token:' + item.key" class="stats-band-cell">
                      <span class="stats-ledger-label">{{ item.label }}</span>
                      <span class="stats-ledger-value">{{ item.value }}</span>
                    </div>
                  </div>
                </section>

                <div v-if="Array.isArray(host.models) && host.models.length > 0" class="stats-model-table">
                  <div class="stats-model-ledger-scroll">
                    <table class="stats-model-ledger-table">
                      <thead>
                        <tr class="stats-model-ledger-group-row">
                          <th rowspan="2" class="stats-model-ledger-stub">{{ t("stats_model") }}</th>
                          <th rowspan="2" class="stats-model-ledger-stub stats-model-ledger-stub-requests">{{ t("stats_requests") }}</th>
                          <th
                            v-if="modelLedgerCostColumns(host.models).length > 0"
                            :colspan="modelLedgerCostColumns(host.models).length"
                            class="stats-model-ledger-group"
                          >
                            <span class="stats-model-ledger-group-copy">
                              <QIconWallet class="stats-model-ledger-group-icon icon" />
                              <span>{{ t("stats_costs") }}</span>
                            </span>
                          </th>
                          <th :colspan="modelLedgerTokenColumns(host.models).length" class="stats-model-ledger-group">
                            <span class="stats-model-ledger-group-copy">
                              <QIconBarChart class="stats-model-ledger-group-icon icon" />
                              <span>{{ t("stats_tokens") }}</span>
                            </span>
                          </th>
                        </tr>
                        <tr class="stats-model-ledger-column-row">
                          <th
                            v-for="column in modelLedgerCostColumns(host.models)"
                            :key="host.api_host + ':head:cost:' + column.key"
                            class="stats-model-ledger-column"
                          >
                            {{ column.label }}
                          </th>
                          <th
                            v-for="column in modelLedgerTokenColumns(host.models)"
                            :key="host.api_host + ':head:token:' + column.key"
                            class="stats-model-ledger-column"
                          >
                            {{ column.label }}
                          </th>
                        </tr>
                      </thead>
                      <tbody>
                        <tr v-for="model in host.models" :key="host.api_host + ':' + model.model" class="stats-model-ledger-row">
                          <th scope="row" class="stats-model-ledger-model">
                            <code class="stats-model-name">{{ model.model }}</code>
                          </th>
                          <td class="stats-model-ledger-value-cell stats-model-ledger-requests">{{ formatNumber(model.requests) }}</td>
                          <td
                            v-for="column in modelLedgerCostColumns(host.models)"
                            :key="host.api_host + ':' + model.model + ':cost:' + column.key"
                            class="stats-model-ledger-value-cell"
                            :class="{ 'stats-model-ledger-value-cell-unavailable': isModelLedgerValueUnavailable(model, column) }"
                          >
                            {{ formatModelLedgerValue(model, column) }}
                          </td>
                          <td
                            v-for="column in modelLedgerTokenColumns(host.models)"
                            :key="host.api_host + ':' + model.model + ':token:' + column.key"
                            class="stats-model-ledger-value-cell"
                          >
                            {{ formatModelLedgerValue(model, column) }}
                          </td>
                        </tr>
                      </tbody>
                    </table>
                  </div>
                </div>
              </article>
            </div>
          </div>

          <div v-else class="stats-section-panel">
            <div v-if="visibleModels.length === 0" class="stats-empty">{{ t("stats_no_data") }}</div>
            <div v-else class="stats-host-list">
              <section class="stats-host-block">
                <header class="stats-host-head">
                  <div class="stats-host-ident">
                    <span class="stats-host-eyebrow">{{ t("stats_group_models") }}</span>
                    <span class="stats-host-name">{{ t("stats_model") }}</span>
                  </div>
                </header>

                <div class="stats-model-table">
                  <div class="stats-model-ledger-scroll">
                    <table class="stats-model-ledger-table">
                      <thead>
                        <tr class="stats-model-ledger-group-row">
                          <th rowspan="2" class="stats-model-ledger-stub">{{ t("stats_model") }}</th>
                          <th rowspan="2" class="stats-model-ledger-stub stats-model-ledger-stub-requests">{{ t("stats_requests") }}</th>
                          <th
                            v-if="modelLedgerCostColumns(visibleModels).length > 0"
                            :colspan="modelLedgerCostColumns(visibleModels).length"
                            class="stats-model-ledger-group"
                          >
                            <span class="stats-model-ledger-group-copy">
                              <QIconWallet class="stats-model-ledger-group-icon icon" />
                              <span>{{ t("stats_costs") }}</span>
                            </span>
                          </th>
                          <th :colspan="modelLedgerTokenColumns(visibleModels).length" class="stats-model-ledger-group">
                            <span class="stats-model-ledger-group-copy">
                              <QIconBarChart class="stats-model-ledger-group-icon icon" />
                              <span>{{ t("stats_tokens") }}</span>
                            </span>
                          </th>
                        </tr>
                        <tr class="stats-model-ledger-column-row">
                          <th
                            v-for="column in modelLedgerCostColumns(visibleModels)"
                            :key="'models:head:cost:' + column.key"
                            class="stats-model-ledger-column"
                          >
                            {{ column.label }}
                          </th>
                          <th
                            v-for="column in modelLedgerTokenColumns(visibleModels)"
                            :key="'models:head:token:' + column.key"
                            class="stats-model-ledger-column"
                          >
                            {{ column.label }}
                          </th>
                        </tr>
                      </thead>
                      <tbody>
                        <tr v-for="model in visibleModels" :key="model.model" class="stats-model-ledger-row">
                          <th scope="row" class="stats-model-ledger-model">
                            <code class="stats-model-name">{{ model.model }}</code>
                          </th>
                          <td class="stats-model-ledger-value-cell stats-model-ledger-requests">{{ formatNumber(model.requests) }}</td>
                          <td
                            v-for="column in modelLedgerCostColumns(visibleModels)"
                            :key="model.model + ':cost:' + column.key"
                            class="stats-model-ledger-value-cell"
                            :class="{ 'stats-model-ledger-value-cell-unavailable': isModelLedgerValueUnavailable(model, column) }"
                          >
                            {{ formatModelLedgerValue(model, column) }}
                          </td>
                          <td
                            v-for="column in modelLedgerTokenColumns(visibleModels)"
                            :key="model.model + ':token:' + column.key"
                            class="stats-model-ledger-value-cell"
                          >
                            {{ formatModelLedgerValue(model, column) }}
                          </td>
                        </tr>
                      </tbody>
                    </table>
                  </div>
                </div>
              </section>
            </div>
          </div>
        </section>
      </section>
    </AppPage>
  `,
};

export default StatsView;
