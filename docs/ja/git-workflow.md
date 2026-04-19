# Git ブランチ運用ガイド

> 本リポジトリ: `iamcheyan/mistermorph` (fork from `quailyquaily/mistermorph`)
> 対象読者: 自分一人で開発・ upstream へ PR を出す開発者

---

## 1. リモート構成

```
origin    → git@github.com:iamcheyan/mistermorph.git   (自分の fork)
upstream  → git@github.com:quailyquaily/mistermorph.git  (本家)
```

確認コマンド:

```bash
git remote -v
```

---

## 2. ブランチの役割

| ブランチ | 役割 | 派生元 | マージ先 |
|---------|------|--------|---------|
| `main` | 自分の開発の主軸。常に動作する状態を保つ | — | — |
| `backup-main-YYYYMMDD` | main の手動バックアップ。大きな upstream merge や実験の前に作成 | `main` | なし（保存用） |
| `feature/xxx` | 日常の機能開発・バグ修正 | `main` | `main` |
| `pr/xxx` | upstream への PR 用。クリーンな状態を保つ | `upstream/master` | `upstream/master` (PR) |
| `upstream-master` | upstream のローカルミラー。**原則手を加えない** | `upstream/master` | `main` (merge のみ) |

---

## 3. 日常開発フロー

### 3.1 新機能開発

```bash
# 1. main を最新にする
git checkout main
git pull origin main

# 2. feature ブランチを切る
git checkout -b feature/cli-compact-mode

# 3. 開発・コミット（Conventional Commits）
git add .
git commit -m "feat(cli): add compact mode for minimal prompt display"

# 4. main にマージ
git checkout main
git merge feature/cli-compact-mode --no-ff

# 5. リモートにプッシュ
git push origin main

# 6. feature ブランチを削除（任意）
git branch -d feature/cli-compact-mode
```

### 3.2 コミットメッセージ規約

```
<type>(<scope>): <subject>

<body>  # 必要なら
```

| type | 用途 |
|------|------|
| `feat` | 新機能 |
| `fix` | バグ修正 |
| `docs` | ドキュメントのみ |
| `refactor` | リファクタリング（動作変更なし） |
| `chore` | ビルド・依存関係など |

例:

```
feat(cli): add TAB auto-completion for built-in commands
fix(cli): clear multi-line spinner output correctly
docs: add git workflow guide
```

---

## 4. upstream との同期

### 4.1 upstream の変更を取り込む

```bash
# 1. upstream をフェッチ
git fetch upstream

# 2. upstream-master を最新にする（原則手を加えない・fast-forward のみ）
git checkout upstream-master
git merge --ff-only upstream/master

# 3. バックアップを作成（大きな差分がある場合）
git checkout -b backup-main-$(date +%Y%m%d)
git checkout main

# 4. upstream-master を main にマージ
git merge upstream-master --no-ff

# 5. コンフリクトがあれば解消してコミット
#    エディタで解消 → git add . → git commit

# 6. 自分のリモートにプッシュ
git push origin main
```

### 4.2 マージ戦略の理由

- **rebase は使わない**: upstream と自分の main の差分が数百コミットある場合、rebase はコミットを再適用する必要があり破壊的で危険
- **merge を使う**: 履歴を保ちつつ、安全に upstream の変更を取り込める
- **`--no-ff`**: マージコミットを残して、いつ upstream の変更を取り込んだか分かりやすくする

---

## 5. upstream へ PR を出す

### 5.1 前提

- PR 用ブランチは `upstream/master` から切る（自分の `main` から切らない）
- upstream には自分専用の変更（個人設定・実験コードなど）を含めない
- 1 PR = 1 テーマ（レビューしやすい粒度）

### 5.2 PR ブランチ作成手順

```bash
# 1. upstream を最新にする
git fetch upstream

# 2. upstream/master から PR ブランチを切る
git checkout -b pr/cli-enhancements upstream/master

# 3. 自分の main から必要なコミットを cherry-pick する
#    まず対象コミットを確認
git log main --oneline --not upstream/master

# 4. cherry-pick（順番に適用）
git cherry-pick <commit-hash-1>
git cherry-pick <commit-hash-2>
# ... 必要なコミットを順番に

# 5. コンフリクトがあれば解消
#    解消後: git add . → git cherry-pick --continue

# 6. 自分の fork にプッシュ（PR 用ブランチ）
git push origin pr/cli-enhancements

# 7. GitHub で PR を作成
#    base: quailyquaily/mistermorph:master
#    compare: iamcheyan/mistermorph:pr/cli-enhancements
```

### 5.3 cherry-pick が面倒な場合（コミットが多い・履歴が汚い）

```bash
# 方法B: 差分をパッチとして適用

# 1. PR ブランチを作成
git checkout -b pr/cli-enhancements upstream/master

# 2. 対象ファイルの差分を取得
git diff upstream/master..main -- cmd/mistermorph/clicmd/cli.go \
  internal/configdefaults/defaults.go \
  assets/config/config.example.yaml > /tmp/cli-changes.patch

# 3. パッチを適用
git apply /tmp/cli-changes.patch

# 4. 整理してコミット（1つにまとめるか、論理的に分ける）
git add .
git commit -m "feat(cli): add compact mode, spinner fixes, and readline integration"

# 5. プッシュ
git push origin pr/cli-enhancements
```

### 5.4 PR 作成後のクリーンアップ

```bash
# PR がマージされたら
git checkout main
git fetch upstream
git merge upstream/master --no-ff   # upstream の変更を自分の main に反映
git push origin main

# PR ブランチを削除
git branch -d pr/cli-enhancements
git push origin --delete pr/cli-enhancements
```

---

## 6. プライバシー確認チェックリスト

PR を出す前・リポジトリを公開する前に必ず確認:

```bash
# 自分の main と upstream/master の差分ファイル一覧
git diff upstream/master --name-only

# 各ファイルの差分を確認（機密情報が含まれていないか）
git diff upstream/master -- <file>

# 特に注意が必要なファイル・ディレクトリ:
# - .morph/memory.md        → 個人の会話履歴
# - Application/            → 組織固有の設定・AWS profile 名など
# - deploy/*/env.example.sh → 実際の値が入っていないか
# - secrets/                → 認証情報
# - docs/gemini-cli-auth.md → 実際のトークンが含まれていないか
```

### 6.1 含めてはいけないもの

- API キー・トークン・パスワード
- 組織名・プロジェクト名（一般的でなければ）
- 個人のメールアドレス・ユーザー名
- 内部ネットワーク構成・証明書

### 6.2 含めてよいもの

- ドキュメント・README
- 一般的な設定例（`example.com`, `your-api-key` などのプレースホルダー）
- オープンソースとして有用な機能改善

---

## 7. よくあるシナリオ

### 7.1 upstream の変更と自分の変更が競合した

```bash
git checkout main
git fetch upstream
git merge upstream/master

# コンフリクト発生時:
# 1. コンフリクトファイルを開いて編集
# 2. git add .
# 3. git commit（マージコミットが完成）
```

### 7.2 PR ブランチを修正したい（レビュー指摘対応）

```bash
git checkout pr/cli-enhancements
# 修正を加える
git add .
git commit --amend          # 直前のコミットを修正
git push origin pr/cli-enhancements --force-with-lease
```

### 7.3 実験的な変更をしたい（main を汚したくない）

```bash
# 実験ブランチを切る（merge しない前提）
git checkout -b experiment/some-idea main
# 自由に開発・壊しても OK
# うまくいったら、クリーンな feature ブランチからやり直す
```

---

## 8. コマンド早見表

| 目的 | コマンド |
|------|---------|
| リモート確認 | `git remote -v` |
| upstream フェッチ | `git fetch upstream` |
| バックアップ作成 | `git checkout -b backup-main-$(date +%Y%m%d)` |
| upstream 同期 | `git merge upstream-master --no-ff` |
| feature ブランチ作成 | `git checkout -b feature/xxx main` |
| PR ブランチ作成 | `git checkout -b pr/xxx upstream/master` |
| cherry-pick | `git cherry-pick <commit>` |
| 差分パッチ作成 | `git diff A..B -- <files> > patch` |
| パッチ適用 | `git apply patch` |
| ブランチ一覧 | `git branch -a` |
| コミット履歴 | `git log --oneline --graph` |

---

## 9. 図解

```
upstream/master  ─────●─────●─────●─────●─────●─────●───
                      ↑                           ↑
                      │                    (定期的に fetch)
                      │                           │
                      └──────────┬────────────────┘
                                 │ merge
                                 ↓
main  ──────────────────────────●─────●─────●─────●─────
                                ↑     ↑     ↑     ↑
                                │     │     │     │
                                └─────┘     └─────┘
                                feature/A   feature/B
                                (merge 後   (merge 後
                                 削除可)     削除可)

PR 用:
upstream/master  ─────●─────────────────────────────────
                      ↑
                      │ checkout -b pr/xxx
                      ↓
pr/xxx             ───●─────●─────●─────  (cherry-pick した変更)
                      ↑
                      │ push → GitHub PR
                      ↓
                   (upstream に merge される)
```

---

*最終更新: 2026-04-19*
