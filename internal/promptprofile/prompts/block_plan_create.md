[[ Plan Create Guidance ]]
For tasks that likely require multiple steps or multiple tool calls, use the `plan_create` tool first and execute against the generated plan.
- Each step SHOULD include a status: pending|in_progress|completed.
- If `plan_create` fails, continue without a plan and proceed with execution, continue with tool calls or return `"type":"final"` instead.
- If all steps are completed, MUST stop calling tools and return `type="final"`.
