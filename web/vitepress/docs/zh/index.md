---
layout: home

title: Mister Morph 文档
description: 包含 CLI、通道、运行模式和 Go 嵌入的文档。

hero:
  text: "Mister Morph 文档"
  tagline: "先看快速开始，再按需要看配置、运行模式或 Go 嵌入。"
  actions:
    - theme: brand
      text: 快速开始
      link: /zh/guide/quickstart-cli
---

<section class="morph-home-section morph-home-start">
  <header class="morph-home-section-heading">
    <MorphKicker text="[ START // PATHS ]" />
    <h2>先走最短路径</h2>
    <p>文档入口分成三条最常见路线。先选眼前的任务，再继续往更深层的 guide 展开。</p>
  </header>
  <div class="morph-home-start-grid">
    <article class="morph-home-route">
      <div class="morph-home-route-head">
        <h3 class="morph-home-route-title">配置 provider</h3>
        <p class="morph-home-route-copy">先把模型 key、环境变量和默认值配好，再进入其它运行或嵌入文档。</p>
      </div>
      <div class="morph-home-route-actions">
        <a class="morph-home-route-primary" href="/zh/guide/install-and-config">安装与配置</a>
      </div>
    </article>
    <article class="morph-home-route">
      <div class="morph-home-route-head">
        <h3 class="morph-home-route-title">了解运行模式</h3>
        <p class="morph-home-route-copy">先看 CLI、通道、memory 和 guard 怎么配合，再决定如何自动化或长期运行。</p>
      </div>
      <div class="morph-home-route-actions">
        <a class="morph-home-route-primary" href="/zh/guide/runtime-modes">Runtime 模式</a>
      </div>
    </article>
    <article class="morph-home-route">
      <div class="morph-home-route-head">
        <h3 class="morph-home-route-title">在 Go 中嵌入 core</h3>
        <p class="morph-home-route-copy">当你需要自定义前端或宿主层，但仍想复用同一套 agent runtime，就从这里开始。</p>
      </div>
      <div class="morph-home-route-actions">
        <a class="morph-home-route-primary" href="/zh/guide/build-agent-with-core">用 Core 搭建 Agent</a>
      </div>
    </article>
  </div>
</section>
