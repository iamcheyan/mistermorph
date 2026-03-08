---
date: 2026-03-08
title: Prompt Token Reduction Review
status: completed
---

# Prompt Token Reduction Review

## 1) Goal

- Inspect recent request dumps under `dump/`
- Identify the largest token-cost drivers in the main runtime prompt path
- Propose the minimum high-yield reductions before doing any broader prompt redesign

## 2) Sample Reviewed

Reviewed files:

- `dump/prompt_telegram_20260308_221933.md`
- `dump/request_telegram_20260308_221933.md`

This review is based on one recent Telegram sample:

- user message: `Hi`
- runtime: Telegram private chat
- model: `x-ai/grok-4.1-fast-reasoning`

Important limitation:

- this is only one sample
- conclusions are still useful because the request is extremely simple, which makes structural waste easy to see

## 3) High-Level Result

For a simple `Hi`, the runtime consumed approximately:

- main reply request: `6955` total tokens
- memory draft request: `1474` total tokens

Combined:

- about `8429` total tokens

This is disproportionate for such a trivial turn.

The main conclusion:

- the largest waste is not `meta`
- the largest waste is not `current_message`
- the largest waste is:
  - oversized `system` prompt
  - duplicated tool information
  - unnecessary memory-draft follow-up for trivial turns

## 4) Main Request Breakdown

From the first main request payload:

- message count: `3`
- `system` chars: about `15687`
- `meta` chars: about `319`
- `current_message` chars: about `518`
- tool count: `11`
- tool schema JSON chars: about `8572`
- total message chars: about `16524`

Implication:

- `meta + current_message` together are only about `837` chars
- the dominant costs are:
  - `system`
  - tool schemas

So the recent runtime message-order cleanup was still correct, but it does not address the largest token cost.

## 5) System Prompt Hotspots

Measured approximately from `dump/prompt_telegram_20260308_221933.md`:

- `## Rules`: about `2917` chars
- `[[ Telegram Policies ]] + [[ Telegram Output Guide ]]`: about `2441` chars
- `## Persona`: about `2062` chars
- `## Available Skills`: about `1813` chars
- `## Available Tools`: about `1521` chars
- `## TODO Workflow`: about `1411` chars
- `[[ Local Scripts ]]`: about `1144` chars
- `[[ Plan Create Guidance ]]`: about `874` chars
- `## Reference Format`: about `862` chars
- `## Response Format`: about `631` chars

This means the main prompt is currently carrying a lot of policy and catalog material even for a one-word greeting.

## 6) Duplicate Information

There is one especially obvious duplication:

- tools are described in the system prompt under `## Available Tools`
- the same tools are also sent via API `tools` schema

Relevant code:

- system prompt injects tool summaries via [prompt_template.go](/home/lyric/Codework/arch/mistermorph/agent/prompt_template.go#L90)
- tool summaries come from [registry.go](/home/lyric/Codework/arch/mistermorph/tools/registry.go#L44)

First-principles judgment:

- the model already receives the tool names, descriptions, and schemas in the tool API payload
- keeping a second tool catalog in the system prompt is low-value duplication

## 7) Memory Draft Waste

The sample also triggered a second request for memory draft generation:

- prompt tokens: `878`
- total tokens: `1474`

Yet the resulting summary was just:

- `[Lyric](tg:@ballcatcat) greeted with 'Hi' and agent responded casually.`

Relevant code path:

- main runtime always records memory after a published text reply:
  - [runtime_task.go](/home/lyric/Codework/arch/mistermorph/internal/channelruntime/telegram/runtime_task.go#L198)
- memory draft is built from a dedicated LLM request:
  - [memory_flow.go](/home/lyric/Codework/arch/mistermorph/internal/channelruntime/telegram/memory_flow.go#L277)

Current memory draft request also sets:

- `max_tokens = 10240`
  - [memory_flow.go](/home/lyric/Codework/arch/mistermorph/internal/channelruntime/telegram/memory_flow.go#L294)

First-principles judgment:

- for trivial turns, this memory pass is usually not worth the cost
- even when memory draft is worth doing, `10240` is far too loose for this JSON shape

## 8) Root Causes

### A. System prompt is too eager

The system prompt currently bundles:

- persona
- tool list
- skill list
- todo workflow
- reference format
- response format
- local scripts notes
- plan-create guidance
- Telegram policies
- memory summaries
- global rules

That is too much for every turn, especially for lightweight chat.

### B. Too many blocks are unconditional

The following are currently good candidates for conditional injection:

- local scripts notes
- TODO workflow
- detailed reference format guide
- long skill catalog
- some Telegram-specific output detail

### C. Memory summarization lacks a cheap gate

The current path does not first ask:

- did this turn contain any durable information?
- was a tool used?
- was there a decision, artifact, identifier, URL, commitment, or reminder worth storing?

Without that gate, trivial chat causes avoidable second-pass token spend.

## 9) Recommended Optimizations

Ordered by expected value and implementation safety.

### 1. Remove `## Available Tools` from the system prompt

Why:

- tool API schema already provides the real tool contract
- system-level duplication is expensive and low-yield

Expected effect:

- immediate prompt reduction with low behavioral risk

### 2. Add a trivial-turn gate before memory draft generation

Example conditions for skipping memory draft:

- inbound text is very short
- no tool call happened
- no file / URL / identifier / reminder / decision / commitment
- final output is lightweight small talk or acknowledgement

Expected effect:

- saves the entire second request for many chatty turns

### 3. Reduce memory draft `max_tokens`

Current:

- `10240`

Suggested first reduction:

- `512` or `768`

Why:

- output shape is fixed JSON
- current limit is far beyond the realistic response size

### 4. Compress `Telegram Policies` and `Telegram Output Guide`

Most expensive low-value part inside this block:

- the very long full emoji allow-list

First-principles judgment:

- this list should not live in the prompt if the tool already enforces it
- the prompt only needs the semantic rule:
  - reaction only for lightweight acknowledgement
  - do not answer a real question with reaction only

### 5. Compress `## Rules`

Many rules are valid, but the current wording is too verbose.

Likely reductions:

- merge overlapping file/path/tool-safety rules
- shorten repeated “must not / do not” wording
- move tool-specific transport details into tool descriptions or tool implementations where possible

### 6. Make `Local Scripts`, `TODO Workflow`, and skill catalog more conditional

Suggested principle:

- do not inject these blocks unless the current task actually touches that domain

Examples:

- `Local Scripts` block only for script/automation tasks
- `TODO Workflow` only when the turn has reminder/todo semantics
- skills block only for sticky or task-relevant skills

### 7. Consider shrinking `Response Format`

The current response-format section is correct but long.

Possible reduction:

- keep only the two JSON shapes and remove extra prose around them

## 10) Estimated Savings

These are rough sample-based estimates, not billing-grade measurements.

They are derived from the reviewed `Hi` sample only, so they should be used for prioritization, not for exact forecasting.

| Optimization | Estimated savings | Notes |
| --- | --- | --- |
| Remove `## Available Tools` from system prompt | about `350-500` prompt tokens per main request | There is already a full tool schema payload in the API request. This estimate is based on the current `## Available Tools` section size. |
| Skip memory draft for trivial turns | about `1474` total tokens per skipped trivial turn | In the sample, the memory draft request consumed `878` prompt tokens and `1474` total tokens. This is the single highest-yield change for lightweight chat. |
| Compress `Telegram Policies` and `Telegram Output Guide` | about `500-800` prompt tokens per Telegram request | The current block is heavily inflated by output-policy detail and the long emoji allow-list. |
| Compress `## Rules` | about `700-900` prompt tokens per request | This estimate assumes a substantial wording pass, not just cosmetic trimming. |
| Make `Local Scripts`, `TODO Workflow`, and skill catalog conditional | about `900-1300` prompt tokens on turns where they are irrelevant | This only applies when those blocks are currently being injected but do not matter for the turn. |
| Reduce memory draft `max_tokens` | little or no guaranteed average-token savings by itself | This is still worth doing, but mainly as a safety and upper-bound control. It does not automatically reduce real usage unless completions are currently running too long. |

Combined first-pass expectation for a trivial turn like this sample:

- apply `remove Available Tools`
- apply `skip trivial-turn memory draft`
- apply `compress Telegram policies`
- apply `compress Rules`

Conservative outcome:

- from about `8429` total tokens down to about `5000-6000`

If conditional prompt blocks are also cleaned up well:

- simple chat turns may have a path toward about `4000-5000`

This is still not "cheap", but it is materially better than the current shape.

## 11) What Not To Prioritize First

Based on this dump, these are not the biggest wins:

- further shrinking `mister_morph_meta`
- further shrinking `current_message`
- changing history/current ordering again

These areas are already relatively small compared with:

- system prompt size
- tool schema duplication
- unnecessary memory draft calls

## 12) Concrete Priority Order

If implementing in chunks, the best order is:

1. remove `Available Tools` from system prompt
2. add trivial-turn skip for memory draft
3. reduce memory draft `max_tokens`
4. compress Telegram policy/output block
5. compress general `Rules`
6. make scripts / TODO / skills blocks conditional

## 13) First-Principles Conclusion

The current system is not primarily spending tokens on the new runtime message structure.

It is spending tokens on:

- oversized always-on instructions
- duplicated tool descriptions
- memory post-processing that is too eager for trivial turns

So the right next optimization direction is:

- reduce unconditional prompt mass
- remove duplicate instruction channels
- avoid second-pass LLM work when there is nothing meaningful to remember
