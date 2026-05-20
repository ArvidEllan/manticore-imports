const apiBase = getApiBaseUrl();
document.getElementById("statusForm").addEventListener("submit", async (e) => {
  e.preventDefault();
  const form = new FormData(e.target);
  const reference = form.get("reference");
  const email = form.get("email");
  const res = await fetch(`${apiBase}/public/status/${encodeURIComponent(reference)}?email=${encodeURIComponent(email)}`);
  const data = await res.json();
  document.getElementById("result").textContent = JSON.stringify(data, null, 2);
});
