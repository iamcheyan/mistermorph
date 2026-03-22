import { computed } from "vue";
import { translate } from "../core/context";
import "./RawTextEditorDialog.css";

const RawTextEditorDialog = {
  emits: ["close", "save", "update:modelValue"],
  props: {
    open: {
      type: Boolean,
      default: false,
    },
    modelValue: {
      type: String,
      default: "",
    },
    title: {
      type: String,
      default: "",
    },
    path: {
      type: String,
      default: "",
    },
    loading: {
      type: Boolean,
      default: false,
    },
    saving: {
      type: Boolean,
      default: false,
    },
  },
  setup(props, { emit }) {
    const t = translate;
    const resolvedTitle = computed(() => props.title || t("repair_editor_title"));

    function close() {
      emit("close");
    }

    function save() {
      emit("save");
    }

    function onInput(value) {
      emit("update:modelValue", String(value || ""));
    }

    return {
      t,
      close,
      save,
      onInput,
      resolvedTitle,
    };
  },
  template: `
    <div v-if="open" class="raw-text-overlay" @click.self="close">
      <section class="raw-text-dialog frame">
        <header class="raw-text-head">
          <div class="raw-text-copy">
            <h3 class="raw-text-title">{{ resolvedTitle }}</h3>
            <code v-if="path" class="raw-text-path">{{ path }}</code>
          </div>
          <div class="raw-text-actions">
            <QButton class="plain sm" :disabled="saving" @click="close">{{ t("action_close") }}</QButton>
            <QButton class="primary sm" :loading="saving" :disabled="loading" @click="save">{{ t("action_save") }}</QButton>
          </div>
        </header>
        <QProgress v-if="loading" :infinite="true" />
        <QTextarea
          v-else
          class="raw-text-body"
          :modelValue="modelValue"
          :rows="20"
          :disabled="saving"
          @update:modelValue="onInput"
        />
      </section>
    </div>
  `,
};

export default RawTextEditorDialog;
