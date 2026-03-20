const BODY_FONT = '"IBM Plex Sans", "Noto Sans SC", "Noto Sans JP", sans-serif';
const MONO_FONT = '"IBM Plex Mono", monospace';

function createTheme(definition) {
  return Object.freeze({
    id: definition.id,
    label: definition.label,
    cssVars: Object.freeze({
      "--mmr-bg": definition.bg,
      "--mmr-surface": definition.surface,
      "--mmr-surface-strong": definition.surfaceStrong,
      "--mmr-line": definition.line,
      "--mmr-line-strong": definition.lineStrong,
      "--mmr-text": definition.text,
      "--mmr-text-muted": definition.textMuted,
      "--mmr-accent": definition.accent,
      "--mmr-shadow": definition.shadow,
      "--mmr-radius": definition.radius || "10px",
      "--mmr-code-bg": definition.codeBg,
      "--mmr-code-bg-start": definition.codeBgStart,
      "--mmr-code-bg-end": definition.codeBgEnd,
      "--mmr-quote-bg-start": definition.quoteBgStart,
      "--mmr-quote-bg-end": definition.quoteBgEnd,
      "--mmr-danger": definition.danger,
      "--mmr-font-body": BODY_FONT,
      "--mmr-font-mono": MONO_FONT,
    }),
    mermaid: Object.freeze({
      startOnLoad: false,
      securityLevel: "strict",
      theme: "base",
      fontFamily: BODY_FONT,
      themeVariables: Object.freeze({
        background: definition.mermaidBackground,
        primaryColor: definition.mermaidPrimary,
        primaryBorderColor: definition.mermaidBorder,
        primaryTextColor: definition.text,
        secondaryColor: definition.mermaidSecondary,
        tertiaryColor: definition.mermaidTertiary,
        lineColor: definition.mermaidLine,
        textColor: definition.text,
        clusterBkg: definition.mermaidClusterBackground,
        clusterBorder: definition.mermaidBorder,
        mainBkg: definition.mermaidPrimary,
      }),
    }),
    graphviz: Object.freeze({
      graphAttributes: Object.freeze({
        bgcolor: "transparent",
      }),
      nodeAttributes: Object.freeze({
        fontname: "IBM Plex Sans",
        fontsize: "12",
        color: definition.graphvizBorder,
        fontcolor: definition.text,
      }),
      edgeAttributes: Object.freeze({
        color: definition.graphvizLine,
        fontname: "IBM Plex Sans",
        fontsize: "11",
        fontcolor: definition.textMuted,
      }),
    }),
    infographic: Object.freeze({
      theme: definition.infographicTheme,
      themeConfig: Object.freeze({
        palette: definition.infographicPalette,
        colorBg: definition.infographicBackground,
        colorPrimary: definition.accent,
        base: Object.freeze({
          text: Object.freeze({
            "font-family": BODY_FONT,
            fill: definition.text,
          }),
          shape: Object.freeze({
            stroke: definition.graphvizBorder,
          }),
          global: Object.freeze({
            fill: definition.infographicBackground,
          }),
        }),
      }),
    }),
  });
}

export const themes = Object.freeze({
  paper: createTheme({
    id: "paper",
    label: "Paper",
    bg: "#faf8f4",
    surface: "#f7f5f0",
    surfaceStrong: "#f0ece4",
    line: "rgba(77, 85, 99, 0.18)",
    lineStrong: "rgba(77, 85, 99, 0.28)",
    text: "#181714",
    textMuted: "#6a645e",
    accent: "#d96a2b",
    shadow: "rgba(20, 24, 32, 0.08)",
    codeBg: "#f5f0e8",
    codeBgStart: "rgba(255, 255, 255, 0.76)",
    codeBgEnd: "rgba(248, 246, 242, 0.94)",
    quoteBgStart: "rgba(255, 250, 245, 0.9)",
    quoteBgEnd: "rgba(247, 244, 239, 0.9)",
    danger: "#b44d3a",
    mermaidBackground: "#f4f2ee",
    mermaidPrimary: "#ebe5dc",
    mermaidSecondary: "#ece7df",
    mermaidTertiary: "#faf8f3",
    mermaidBorder: "#6e665f",
    mermaidLine: "#6a625d",
    mermaidClusterBackground: "#f8f4eb",
    graphvizBorder: "#6e665f",
    graphvizLine: "#6a625d",
    infographicTheme: "light",
    infographicPalette: "antv",
    infographicBackground: "#f7f5f0",
  }),
  console: createTheme({
    id: "console",
    label: "Console",
    bg: "#f6f6f6",
    surface: "#f2f2f0",
    surfaceStrong: "#ebebe8",
    line: "rgba(0, 0, 0, 0.12)",
    lineStrong: "rgba(0, 0, 0, 0.2)",
    text: "#141414",
    textMuted: "#6a6a6a",
    accent: "#d96a2b",
    shadow: "rgba(20, 24, 32, 0.06)",
    codeBg: "#efebe4",
    codeBgStart: "rgba(255, 255, 255, 0.76)",
    codeBgEnd: "rgba(244, 242, 238, 0.96)",
    quoteBgStart: "rgba(250, 246, 241, 0.92)",
    quoteBgEnd: "rgba(243, 239, 232, 0.94)",
    danger: "#bc4f37",
    mermaidBackground: "#f1f1ef",
    mermaidPrimary: "#e8e3da",
    mermaidSecondary: "#ebe6de",
    mermaidTertiary: "#f7f5f0",
    mermaidBorder: "#5f5f5f",
    mermaidLine: "#606060",
    mermaidClusterBackground: "#f2ede5",
    graphvizBorder: "#5f5f5f",
    graphvizLine: "#6a6a6a",
    infographicTheme: "light",
    infographicPalette: "antv",
    infographicBackground: "#f2f2f0",
  }),
  folio: createTheme({
    id: "folio",
    label: "Folio",
    bg: "#f8f1e5",
    surface: "#f3ebdf",
    surfaceStrong: "#e9dfcf",
    line: "rgba(89, 73, 52, 0.18)",
    lineStrong: "rgba(89, 73, 52, 0.28)",
    text: "#2f2419",
    textMuted: "#73614f",
    accent: "#8d6132",
    shadow: "rgba(43, 33, 21, 0.08)",
    codeBg: "#efe3d1",
    codeBgStart: "rgba(255, 252, 247, 0.8)",
    codeBgEnd: "rgba(244, 235, 223, 0.96)",
    quoteBgStart: "rgba(253, 248, 240, 0.9)",
    quoteBgEnd: "rgba(243, 235, 223, 0.96)",
    danger: "#a34734",
    mermaidBackground: "#f2e9db",
    mermaidPrimary: "#e6d8c4",
    mermaidSecondary: "#ecdfcc",
    mermaidTertiary: "#f7efe2",
    mermaidBorder: "#735b41",
    mermaidLine: "#7a6247",
    mermaidClusterBackground: "#efe2cf",
    graphvizBorder: "#735b41",
    graphvizLine: "#7a6247",
    infographicTheme: "light",
    infographicPalette: "antv",
    infographicBackground: "#f3ebdf",
  }),
  blueprint: createTheme({
    id: "blueprint",
    label: "Blueprint",
    bg: "#f1f5f8",
    surface: "#eef3f6",
    surfaceStrong: "#e4ebf1",
    line: "rgba(48, 74, 98, 0.18)",
    lineStrong: "rgba(48, 74, 98, 0.28)",
    text: "#122433",
    textMuted: "#5d7083",
    accent: "#2f6f9f",
    shadow: "rgba(16, 28, 40, 0.08)",
    codeBg: "#e8eef3",
    codeBgStart: "rgba(255, 255, 255, 0.8)",
    codeBgEnd: "rgba(233, 239, 244, 0.96)",
    quoteBgStart: "rgba(247, 251, 255, 0.92)",
    quoteBgEnd: "rgba(235, 242, 247, 0.96)",
    danger: "#b44d3a",
    mermaidBackground: "#edf2f6",
    mermaidPrimary: "#dbe6ee",
    mermaidSecondary: "#e3ecf3",
    mermaidTertiary: "#f5f8fb",
    mermaidBorder: "#40617c",
    mermaidLine: "#4f738f",
    mermaidClusterBackground: "#e6eef4",
    graphvizBorder: "#40617c",
    graphvizLine: "#4f738f",
    infographicTheme: "light",
    infographicPalette: "antv",
    infographicBackground: "#eef3f6",
  }),
});

export const themeNames = Object.freeze(Object.keys(themes));
export const themeCatalog = Object.freeze(
  themeNames.map((id) =>
    Object.freeze({
      id,
      label: themes[id].label,
    })
  )
);
export const themeCssVariableNames = Object.freeze(Object.keys(themes.paper.cssVars));

export function resolveTheme(raw) {
  const key = String(raw || "paper").trim().toLowerCase();
  return themes[key] || themes.paper;
}

export function applyThemeToElement(element, theme) {
  const resolved = resolveTheme(theme?.id || theme);
  element.dataset.mmrTheme = resolved.id;
  for (const name of themeCssVariableNames) {
    element.style.setProperty(name, resolved.cssVars[name]);
  }
  return resolved;
}
