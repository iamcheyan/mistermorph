import { computed, onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import "./SetupView.css";

import AppPage from "../components/AppPage";
import { endpointDisplayItem } from "../core/endpoints";
import { endpointState, loadEndpoints, setSelectedEndpointRef, translate } from "../core/context";
import { buildConsoleSetupState } from "../core/setup";

function tuiKicker(left, right) {
  const lhs = String(left || "").trim();
  const rhs = String(right || "").trim();
  if (lhs && rhs) {
    return `[ ${lhs} // ${rhs} ]`;
  }
  return `[ ${lhs || rhs} ]`;
}

function normalizeRedirectTarget(raw) {
  const value = String(raw || "").trim();
  if (!value || value === "/setup" || value === "/login") {
    return "";
  }
  return value;
}

const SetupView = {
  components: {
    AppPage,
  },
  setup() {
    const t = translate;
    const route = useRoute();
    const router = useRouter();
    const loading = ref(false);
    const err = ref("");
    const setup = computed(() => buildConsoleSetupState(endpointState.items));
    const endpointRows = computed(() =>
      setup.value.endpoints.map((item) => ({
        ...item,
        ...endpointDisplayItem(item, t),
      }))
    );

    function nextReadyRoute() {
      const redirect = normalizeRedirectTarget(route.query.redirect);
      if (redirect) {
        return redirect;
      }
      return setup.value.primaryChatReadyEndpoint ? "/chat" : "/overview";
    }

    function leaveIfReady() {
      if (setup.value.requiresSetup) {
        return false;
      }
      if (setup.value.primaryChatReadyEndpoint?.endpoint_ref) {
        setSelectedEndpointRef(setup.value.primaryChatReadyEndpoint.endpoint_ref);
      }
      router.replace(nextReadyRoute());
      return true;
    }

    async function refreshStatus() {
      if (loading.value) {
        return;
      }
      loading.value = true;
      err.value = "";
      try {
        await loadEndpoints();
        leaveIfReady();
      } catch (e) {
        err.value = e.message || t("msg_load_failed");
      } finally {
        loading.value = false;
      }
    }

    function openSettings() {
      router.push("/settings");
    }

    onMounted(() => {
      if (endpointState.items.length === 0) {
        void refreshStatus();
        return;
      }
      leaveIfReady();
    });

    return {
      t,
      err,
      loading,
      endpointRows,
      setup,
      tuiKicker,
      refreshStatus,
      openSettings,
    };
  },
  template: `
    <AppPage :title="t('setup_title')" :showMobileNavTrigger="false">
      <section class="setup-page">
        <QFence v-if="err" type="danger" icon="QIconCloseCircle" :text="err" />
        <section class="setup-hero stat-item">
          <p class="ui-kicker">{{ tuiKicker(t("endpoint_channel_console"), t("setup_title")) }}</p>
          <h1 class="setup-hero-title">{{ t("setup_local_title") }}</h1>
          <p class="setup-hero-text muted">{{ t("setup_local_body") }}</p>
          <div class="setup-hero-actions">
            <QButton class="primary" @click="openSettings">{{ t("setup_action_open_settings") }}</QButton>
            <QButton class="outlined" :loading="loading" @click="refreshStatus">{{ t("setup_action_refresh") }}</QButton>
          </div>
        </section>

        <div class="setup-grid">
          <section class="setup-panel ui-track-panel">
            <p class="ui-kicker">{{ tuiKicker(t("setup_title"), t("setup_minimum_title")) }}</p>
            <h2 class="setup-panel-title">{{ t("setup_minimum_heading") }}</h2>
            <p class="setup-panel-text muted">{{ t("setup_minimum_body") }}</p>
            <div class="setup-requirements">
              <div class="setup-requirement">
                <span class="setup-requirement-name">{{ t("setup_requirement_provider_label") }}</span>
                <span class="setup-requirement-desc">{{ t("setup_requirement_provider_desc") }}</span>
              </div>
              <div class="setup-requirement">
                <span class="setup-requirement-name">{{ t("setup_requirement_model_label") }}</span>
                <span class="setup-requirement-desc">{{ t("setup_requirement_model_desc") }}</span>
              </div>
              <div class="setup-requirement">
                <span class="setup-requirement-name">{{ t("setup_requirement_api_key_label") }}</span>
                <span class="setup-requirement-desc">{{ t("setup_requirement_api_key_desc") }}</span>
              </div>
            </div>
          </section>

          <section class="setup-panel ui-track-panel">
            <p class="ui-kicker">{{ tuiKicker(t("runtime_title"), t("group_endpoints")) }}</p>
            <h2 class="setup-panel-title">{{ t("setup_status_heading") }}</h2>
            <div class="setup-status-list">
              <div v-for="item in endpointRows" :key="item.endpoint_ref" class="setup-status-row">
                <span class="setup-status-dot">
                  <QBadge :type="item.connected ? 'success' : 'default'" size="md" variant="filled" :dot="true" />
                </span>
                <div class="setup-status-main">
                  <span class="setup-status-name"><code>{{ item.title }}</code></span>
                  <span class="setup-status-meta">{{ item.location }}</span>
                  <span class="setup-status-meta">
                    {{
                      item.can_submit
                        ? t("setup_status_submit_ready")
                        : (item.connected ? t("setup_status_submit_missing") : t("setup_status_endpoint_offline"))
                    }}
                  </span>
                </div>
              </div>
            </div>
            <p class="setup-note">
              {{ t("setup_status_note") }}
            </p>
          </section>
        </div>
      </section>
    </AppPage>
  `,
};

export default SetupView;
