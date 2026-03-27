---
title: Overview
description: What Mister Morph is, and the fastest path through these docs.
---

# Overview

Mister Morph is a unified agent project with two primary usage patterns:

- CLI-first workflow (`mistermorph run`, `telegram`, `slack`, `console serve`)
- Embedded Go core (`integration` package)

## Choose Your Path

- Building automation quickly: start with [Quickstart (CLI)](/guide/quickstart-cli)
- Embedding in your Go project: go to [Build an Agent with Core](/guide/build-agent-with-core)
- Running long-lived channels: read [Runtime Modes](/guide/runtime-modes)
- Production hardening: read [Security and Guard](/guide/security-and-guard)

## Repository Structure Snapshot

- CLI entry: `cmd/mistermorph/`
- Agent engine: `agent/`
- Embedding core: `integration/`
- Built-in tools: `tools/`
- Provider backends: `providers/`
- Detailed docs: `docs/`
