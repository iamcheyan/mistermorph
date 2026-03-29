import { computed } from "vue";
import "./AppKicker.css";

const AppKicker = {
  props: {
    as: {
      type: String,
      default: "p",
    },
    left: {
      type: String,
      default: "",
    },
    right: {
      type: String,
      default: "",
    },
  },
  setup(props) {
    const leftText = computed(() => String(props.left || "").trim());
    const rightText = computed(() => String(props.right || "").trim());
    const hasRight = computed(() => rightText.value !== "");
    const ariaLabel = computed(() => [leftText.value, rightText.value].filter(Boolean).join(" "));

    return {
      leftText,
      rightText,
      hasRight,
      ariaLabel,
    };
  },
  template: `
    <component :is="as" class="ui-kicker hero-install-kicker" :aria-label="ariaLabel">
      <span class="hero-install-kicker-bracket hero-install-kicker-bracket-open">[</span>
      <span v-if="leftText" class="hero-install-kicker-label">{{ leftText }}</span>
      <span v-if="hasRight" class="hero-install-kicker-sep">//</span>
      <span v-if="hasRight" class="hero-install-kicker-label">{{ rightText }}</span>
      <span class="hero-install-kicker-bracket hero-install-kicker-bracket-close">]</span>
    </component>
  `,
};

export default AppKicker;
