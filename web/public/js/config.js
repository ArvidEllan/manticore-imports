window.MANTICORE_CONFIG = window.MANTICORE_CONFIG || {
  apiBaseUrl: window.API_BASE_URL || "",
  stage: "local"
};

function getApiBaseUrl() {
  return (window.MANTICORE_CONFIG && window.MANTICORE_CONFIG.apiBaseUrl) || "";
}
