import { onBeforeUnmount, onMounted, ref, watch } from "vue";

let rendererModulePromise = null;

async function loadRendererModule() {
  rendererModulePromise ||= Promise.all([
    import("../vendor/markdown-renderer/index.js"),
    import("../vendor/markdown-renderer/index.css"),
  ]).then(([module]) => module);
  return rendererModulePromise;
}

const MarkdownContent = {
  emits: ["rendered"],
  props: {
    source: {
      type: String,
      default: "",
    },
    format: {
      type: String,
      default: "auto",
    },
    theme: {
      type: String,
      default: "paper",
    },
  },
  setup(props, { emit }) {
    const host = ref(null);
    const renderer = ref(null);

    async function syncRenderer() {
      const element = host.value;
      if (!element) {
        return;
      }
      const { MarkdownRenderer } = await loadRendererModule();
      if (!host.value || host.value !== element) {
        return;
      }
      if (!renderer.value) {
        renderer.value = new MarkdownRenderer(element, {
          format: props.format,
          theme: props.theme,
        });
      }
      await renderer.value.update(props.source, {
        format: props.format,
        theme: props.theme,
      });
      if (host.value === element) {
        emit("rendered");
      }
    }

    onMounted(() => {
      void syncRenderer();
    });

    onBeforeUnmount(() => {
      renderer.value?.destroy();
      renderer.value = null;
    });

    watch(
      () => [props.source, props.format, props.theme],
      () => {
        void syncRenderer();
      }
    );

    return {
      host,
    };
  },
  template: `<div ref="host" class="chat-markdown-content"></div>`,
};

export default MarkdownContent;
