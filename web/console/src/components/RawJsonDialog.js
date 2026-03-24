import { translate } from "../core/context";
import "./RawJsonDialog.css";

const RawJsonDialog = {
  emits: ["close"],
  props: {
    open: {
      type: Boolean,
      default: false,
    },
    title: {
      type: String,
      default: "",
    },
    json: {
      type: String,
      default: "",
    },
  },
  setup(props, { emit }) {
    const t = translate;

    function close() {
      emit("close");
    }

    async function copy() {
      const text = String(props.json || "");
      if (!text) {
        return;
      }
      try {
        if (navigator?.clipboard?.writeText) {
          await navigator.clipboard.writeText(text);
          return;
        }
      } catch {}
      const textarea = document.createElement("textarea");
      textarea.value = text;
      textarea.setAttribute("readonly", "true");
      textarea.style.position = "fixed";
      textarea.style.opacity = "0";
      textarea.style.pointerEvents = "none";
      document.body.appendChild(textarea);
      textarea.select();
      try {
        document.execCommand("copy");
      } finally {
        document.body.removeChild(textarea);
      }
    }

    return {
      t,
      close,
      copy,
    };
  },
  template: `
    <QDialog
      :modelValue="open"
      width="860px"
      @update:modelValue="!$event && close()"
      @close="close"
    >
      <template #header>
        <header class="raw-json-head">
          <h3 class="raw-json-title">{{ title || 'RAW JSON' }}</h3>
        </header>
      </template>

      <section class="raw-json-dialog">
        <div class="raw-json-codebox">
          <div class="raw-json-codebox-tools">
            <QButton class="plain xs" @click="copy">{{ t("action_copy") }}</QButton>
          </div>
          <pre class="raw-json-body">{{ json }}</pre>
        </div>
      </section>
    </QDialog>
  `,
};

export default RawJsonDialog;
