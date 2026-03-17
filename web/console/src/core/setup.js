import { CONSOLE_LOCAL_ENDPOINT_REF, visibleEndpoints } from "./endpoints";

function normalizeEndpointItem(item) {
  return {
    endpoint_ref: typeof item?.endpoint_ref === "string" ? item.endpoint_ref.trim() : "",
    name: typeof item?.name === "string" ? item.name : "",
    url: typeof item?.url === "string" ? item.url : "",
    mode: typeof item?.mode === "string" ? item.mode : "",
    connected: item?.connected === true,
    can_submit: item?.can_submit === true,
    agent_name: typeof item?.agent_name === "string" ? item.agent_name : "",
    submit_endpoint_ref:
      typeof item?.submit_endpoint_ref === "string" ? item.submit_endpoint_ref.trim() : "",
  };
}

function buildConsoleSetupState(items) {
  const endpoints = visibleEndpoints(items).map(normalizeEndpointItem);
  const connectedEndpoints = endpoints.filter((item) => item.connected === true);
  const chatReadyEndpoints = endpoints.filter((item) => item.connected === true && item.can_submit === true);
  const consoleLocalEndpoint =
    endpoints.find((item) => item.endpoint_ref === CONSOLE_LOCAL_ENDPOINT_REF) || null;
  const requiresSetup =
    chatReadyEndpoints.length === 0 &&
    consoleLocalEndpoint?.connected === true &&
    consoleLocalEndpoint?.can_submit !== true;

  return {
    endpoints,
    connectedEndpoints,
    chatReadyEndpoints,
    consoleLocalEndpoint,
    requiresSetup,
    primaryChatReadyEndpoint: chatReadyEndpoints[0] || null,
  };
}

export { buildConsoleSetupState };
