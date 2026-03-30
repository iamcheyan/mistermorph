---
layout: home

title: Mister Morph ドキュメント
description: CLI、チャネル、ランタイム、Go 組み込みを扱うドキュメント。

hero:
  text: "Mister Morph Docs"
  tagline: "まずクイックスタート。その後に設定、ランタイム、Go 組み込みを読む。"
  actions:
    - theme: brand
      text: クイックスタート
      link: /ja/guide/quickstart-cli
---

<section class="morph-home-section morph-home-start">
  <header class="morph-home-section-heading">
    <MorphKicker text="[ START // PATHS ]" />
    <h2>まず最短ルートから入る</h2>
    <p>入口は三つに絞る。いま必要な作業に合わせて入り、そのあとに deeper guide へ進む。</p>
  </header>
  <div class="morph-home-start-grid">
    <article class="morph-home-route">
      <div class="morph-home-route-head">
        <h3 class="morph-home-route-title">プロバイダを設定する</h3>
        <p class="morph-home-route-copy">モデル鍵、環境変数、既定値を先に固めてから、他の runtime や embedding ガイドへ進む。</p>
      </div>
      <div class="morph-home-route-actions">
        <a class="morph-home-route-primary" href="/ja/guide/install-and-config">インストールと設定</a>
      </div>
    </article>
    <article class="morph-home-route">
      <div class="morph-home-route-head">
        <h3 class="morph-home-route-title">Runtime モードを理解する</h3>
        <p class="morph-home-route-copy">CLI、チャネル、memory、guard の関係を先に掴み、どのように長寿命運用するか判断する。</p>
      </div>
      <div class="morph-home-route-actions">
        <a class="morph-home-route-primary" href="/ja/guide/runtime-modes">Runtime モード</a>
      </div>
    </article>
    <article class="morph-home-route">
      <div class="morph-home-route-head">
        <h3 class="morph-home-route-title">Go core を組み込む</h3>
        <p class="morph-home-route-copy">独自のホストや UI を持ちつつ、同じ agent runtime を再利用したいときの入口。</p>
      </div>
      <div class="morph-home-route-actions">
        <a class="morph-home-route-primary" href="/ja/guide/build-agent-with-core">Core で Agent を構築</a>
      </div>
    </article>
  </div>
</section>
