# markdown-renderer

`web/markdown-renderer` is a standalone browser-side markdown renderer used by
the console frontend. It does not depend on Vue. The console consumes its built
artifacts from `web/console/src/vendor/markdown-renderer/` and wraps them with a
thin Vue component.

## Scope

- Markdown + GFM
- Inline HTML
- Syntax highlighting
- KaTeX math rendering
- Mermaid
- Graphviz
- Infographic
- Theme system shared across markdown, diagrams, and console embedding

## Supported Syntax

### Markdown

- Paragraphs, headings, lists, tables, task lists
- Inline code and fenced code blocks
- Inline HTML

### Math

- Inline math: `$E=mc^2$`
- Display math:

  ```md
  $$
  E=mc^2
  $$
  ```

- Fenced math blocks are also normalized into display math:

  ````md
  ```
  $$
  E=mc^2
  $$
  ```
  ````

  ````md
  ```math
  E=mc^2
  ```
  ````

Supported math fence languages: `math`, `latex`, `tex`, `katex`.

### Diagrams

- `mermaid`
- `mmd`
- `graphviz`
- `dot`
- `gv`
- `infographic`

Pure-source auto detection currently supports:

- Mermaid source
- Graphviz DOT
- Infographic source

## Themes

Built-in themes:

- `paper`
- `console`
- `folio`
- `blueprint`

Each theme drives:

- CSS variables for markdown content
- Mermaid theme variables
- Graphviz render attributes
- Infographic theme config

The renderer exports `supportedThemes`, `themeCatalog`, `themes`, and
`resolveTheme`.

## Public API

Main exports from `src/index.js`:

- `MarkdownRenderer`
- `mountMarkdownRenderer(root, source, options)`
- `supportedFenceLanguages`
- `supportedThemes`
- `themeCatalog`
- `themes`
- `resolveTheme`

Renderer options:

- `format`: default `auto`
- `theme`: default `paper`

## Build

Install dependencies:

```sh
pnpm install
```

Build the standalone bundle:

```sh
pnpm run build
```

Build and copy artifacts into the console vendor directory:

```sh
pnpm run build-console
```

## Console Integration

The console wrapper lives at:

- `web/console/src/components/MarkdownContent.js`

It lazy-loads:

- `web/console/src/vendor/markdown-renderer/index.js`
- `web/console/src/vendor/markdown-renderer/index.css`

Current console chat uses the `console` theme for agent output.

## Notes

- The renderer is intentionally framework-agnostic.
- Streaming updates currently re-render the whole markdown payload on each
  source change.
- Mermaid, Graphviz, and Infographic remain the heavy parts of the bundle.
