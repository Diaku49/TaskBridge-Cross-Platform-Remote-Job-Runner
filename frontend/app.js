const runtimeConfig = window.__TASKBRIDGE_CONFIG__ || {};
const defaultAPIBase = runtimeConfig.apiBase || "http://localhost:8080";
const defaultJobTargetBase = (runtimeConfig.jobTargetBase || defaultAPIBase).replace(/\/$/, "");

const state = {
  apiBase: defaultAPIBase,
  view: "jobs",
  jobs: [],
  agents: [],
  selectedKind: null,
  selectedId: null,
  toastTimer: null,
};

const payloadTemplates = {
  wait: {
    duration_seconds: 5,
  },
  http_check: {
    url: `${defaultJobTargetBase}/health`,
    expected_status: 200,
  },
  tcp_check: {
    address: defaultJobTargetBase.replace(/^https?:\/\//, ""),
  },
  file_exists: {
    path: "/tmp/taskbridge-demo/input.txt",
  },
  checksum: {
    path: "/tmp/taskbridge-demo/input.txt",
    algorithm: "sha256",
  },
  write_file: {
    path: "/tmp/taskbridge-demo/output.txt",
    content: "hello from taskbridge\n",
    create_dirs: true,
  },
};

const els = {
  apiBaseInput: document.querySelector("#apiBaseInput"),
  healthBadge: document.querySelector("#healthBadge"),
  refreshButton: document.querySelector("#refreshButton"),
  jobsTab: document.querySelector("#jobsTab"),
  agentsTab: document.querySelector("#agentsTab"),
  filterInput: document.querySelector("#filterInput"),
  jobsView: document.querySelector("#jobsView"),
  agentsView: document.querySelector("#agentsView"),
  jobsTable: document.querySelector("#jobsTable"),
  agentsTable: document.querySelector("#agentsTable"),
  detailBody: document.querySelector("#detailBody"),
  createJobForm: document.querySelector("#createJobForm"),
  createJobName: document.querySelector("#createJobName"),
  createJobType: document.querySelector("#createJobType"),
  createJobTimeout: document.querySelector("#createJobTimeout"),
  createJobRetries: document.querySelector("#createJobRetries"),
  createJobPayload: document.querySelector("#createJobPayload"),
  createJobButton: document.querySelector("#createJobButton"),
  jobLookupForm: document.querySelector("#jobLookupForm"),
  jobLookupInput: document.querySelector("#jobLookupInput"),
  toast: document.querySelector("#toast"),
  metricJobs: document.querySelector("#metricJobs"),
  metricRunning: document.querySelector("#metricRunning"),
  metricPending: document.querySelector("#metricPending"),
  metricAgents: document.querySelector("#metricAgents"),
};

function apiBase() {
  return state.apiBase.replace(/\/$/, "");
}

function apiPath(path) {
  return `${apiBase()}${path}`;
}

async function fetchJSON(path) {
  const response = await fetch(apiPath(path), {
    headers: { Accept: "application/json" },
  });

  if (response.status === 204) {
    return null;
  }

  let data = null;
  try {
    data = await response.json();
  } catch {
    data = null;
  }

  if (!response.ok) {
    const message = data && data.message ? data.message : response.statusText;
    throw new Error(message);
  }

  return data;
}

async function sendJSON(path, payload) {
  const response = await fetch(apiPath(path), {
    method: "POST",
    headers: {
      Accept: "application/json",
      "Content-Type": "application/json",
    },
    body: JSON.stringify(payload),
  });

  let data = null;
  try {
    data = await response.json();
  } catch {
    data = null;
  }

  if (!response.ok) {
    const message = data && data.message ? data.message : response.statusText;
    throw new Error(message);
  }

  return data;
}

async function refreshAll() {
  setLoading(true);
  try {
    const [health, jobs, agents] = await Promise.all([
      fetchJSON("/health"),
      fetchJSON("/jobs"),
      fetchJSON("/agents"),
    ]);

    state.jobs = Array.isArray(jobs) ? jobs : [];
    state.agents = Array.isArray(agents) ? agents : [];
    setHealth(true, health && health.service ? health.service : "online");
    render();
  } catch (error) {
    setHealth(false, error.message);
    showToast(error.message);
  } finally {
    setLoading(false);
  }
}

async function createJob() {
  const name = els.createJobName.value.trim();
  if (!name) {
    showToast("Job name is required");
    return;
  }

  let payload;
  try {
    payload = JSON.parse(els.createJobPayload.value || "{}");
  } catch (error) {
    showToast(`Invalid payload JSON: ${error.message}`);
    return;
  }

  const request = {
    name,
    type: els.createJobType.value,
    payload,
    timeout_seconds: Number(els.createJobTimeout.value || 0),
    max_retries: Number(els.createJobRetries.value || 0),
  };

  els.createJobButton.disabled = true;
  try {
    const job = await sendJSON("/jobs", request);
    const index = state.jobs.findIndex((item) => item.id === job.id);
    if (index >= 0) {
      state.jobs[index] = job;
    } else {
      state.jobs.unshift(job);
    }

    state.view = "jobs";
    state.selectedKind = "job";
    state.selectedId = job.id;
    els.filterInput.value = "";
    render();
    showToast("Job created");
  } catch (error) {
    showToast(error.message);
  } finally {
    els.createJobButton.disabled = false;
  }
}

function setPayloadTemplate(type, force = false) {
  if (!force && els.createJobPayload.value.trim()) {
    return;
  }

  const payload = payloadTemplates[type] || {};
  els.createJobPayload.value = formatJSON(payload);
}

function setLoading(isLoading) {
  els.refreshButton.disabled = isLoading;
  els.refreshButton.classList.toggle("is-loading", isLoading);
}

function setHealth(isOnline, label) {
  els.healthBadge.classList.toggle("is-online", isOnline);
  els.healthBadge.classList.toggle("is-offline", !isOnline);
  els.healthBadge.classList.remove("is-muted");
  els.healthBadge.lastChild.textContent = isOnline ? "Online" : "Offline";
  els.healthBadge.title = label || "";
}

function render() {
  renderMetrics();
  renderTabs();
  renderJobs();
  renderAgents();
  renderSelected();
}

function renderMetrics() {
  const jobs = state.jobs;
  els.metricJobs.textContent = jobs.length;
  els.metricRunning.textContent = jobs.filter((job) => job.status === "RUNNING").length;
  els.metricPending.textContent = jobs.filter((job) => job.status === "PENDING" || job.status === "RETRYING").length;
  els.metricAgents.textContent = state.agents.filter((agent) => agent.status === "online").length;
}

function renderTabs() {
  const showingJobs = state.view === "jobs";
  els.jobsTab.classList.toggle("is-active", showingJobs);
  els.agentsTab.classList.toggle("is-active", !showingJobs);
  els.jobsTab.setAttribute("aria-selected", String(showingJobs));
  els.agentsTab.setAttribute("aria-selected", String(!showingJobs));
  els.jobsView.classList.toggle("is-hidden", !showingJobs);
  els.agentsView.classList.toggle("is-hidden", showingJobs);
}

function renderJobs() {
  const filter = normalizedFilter();
  const jobs = state.jobs.filter((job) => {
    return matchesFilter(filter, [
      job.id,
      job.name,
      job.type,
      job.status,
      job.assigned_agent_id,
    ]);
  });

  if (jobs.length === 0) {
    els.jobsTable.innerHTML = emptyRow("No jobs", 6);
    return;
  }

  els.jobsTable.innerHTML = jobs.map((job) => `
    <tr class="clickable ${isSelected("job", job.id) ? "is-selected" : ""}" data-kind="job" data-id="${escapeAttr(job.id)}">
      <td>
        <div class="name-cell">
          <strong>${escapeHTML(job.name || "Unnamed job")}</strong>
        </div>
      </td>
      <td>${statusPill(job.status)}</td>
      <td><span class="mono">${escapeHTML(job.type || "")}</span></td>
      <td><span class="mono">${escapeHTML(job.assigned_agent_id || "-")}</span></td>
      <td>${Number(job.attempt_count || 0)} / ${Number(job.max_retries || 0)}</td>
      <td><span class="row-action"><svg><use href="#icon-chevron"></use></svg></span></td>
    </tr>
  `).join("");
}

function renderAgents() {
  const filter = normalizedFilter();
  const agents = state.agents.filter((agent) => {
    return matchesFilter(filter, [
      agent.id,
      agent.hostname,
      agent.os,
      agent.arch,
      agent.status,
      ...(agent.capabilities || []),
    ]);
  });

  if (agents.length === 0) {
    els.agentsTable.innerHTML = emptyRow("No agents", 6);
    return;
  }

  els.agentsTable.innerHTML = agents.map((agent) => `
    <tr class="clickable ${isSelected("agent", agent.id) ? "is-selected" : ""}" data-kind="agent" data-id="${escapeAttr(agent.id)}">
      <td><strong>${escapeHTML(agent.id || "Unknown")}</strong></td>
      <td>${statusPill(agent.status)}</td>
      <td>${escapeHTML(agent.hostname || "-")}</td>
      <td><span class="mono">${escapeHTML([agent.os, agent.arch].filter(Boolean).join(" / ") || "-")}</span></td>
      <td>${capabilityList(agent.capabilities || [])}</td>
      <td><span class="row-action"><svg><use href="#icon-chevron"></use></svg></span></td>
    </tr>
  `).join("");
}

function renderSelected() {
  if (!state.selectedKind || !state.selectedId) {
    renderEmptyDetail();
    return;
  }

  const item = state.selectedKind === "job"
    ? state.jobs.find((job) => job.id === state.selectedId)
    : state.agents.find((agent) => agent.id === state.selectedId);

  if (!item) {
    renderEmptyDetail();
    return;
  }

  if (state.selectedKind === "job") {
    renderJobDetail(item);
  } else {
    renderAgentDetail(item);
  }
}

function renderJobDetail(job) {
  els.detailBody.className = "detail-body";
  els.detailBody.innerHTML = `
    <div class="detail-title">
      <div>
        <h3>${escapeHTML(job.name || "Unnamed job")}</h3>
        <p class="mono">${escapeHTML(job.id)}</p>
      </div>
      ${statusPill(job.status)}
    </div>
    <div class="detail-grid">
      ${field("Type", job.type || "-")}
      ${field("Assigned Agent", job.assigned_agent_id || "-")}
      ${field("Attempts", `${Number(job.attempt_count || 0)} / ${Number(job.max_retries || 0)}`)}
      ${field("Timeout", `${Number(job.timeout_seconds || 0)}s`)}
      ${field("Created", formatDate(job.created_at))}
      ${field("Finished", formatDate(job.finished_at))}
    </div>
    <span class="eyebrow">Payload</span>
    <pre class="json-block">${escapeHTML(formatJSON(job.payload || {}))}</pre>
    <span class="eyebrow">Result</span>
    <pre class="json-block">${escapeHTML(formatJSON(job.result || {}))}</pre>
    <span class="eyebrow">Logs</span>
    <pre class="log-list">${escapeHTML((job.logs || []).join("\n") || "-")}</pre>
    ${job.error ? `<span class="eyebrow">Error</span><pre class="log-list">${escapeHTML(job.error)}</pre>` : ""}
  `;
}

function renderAgentDetail(agent) {
  els.detailBody.className = "detail-body";
  els.detailBody.innerHTML = `
    <div class="detail-title">
      <div>
        <h3>${escapeHTML(agent.id || "Unknown agent")}</h3>
        <p>${escapeHTML(agent.hostname || "-")}</p>
      </div>
      ${statusPill(agent.status)}
    </div>
    <div class="detail-grid">
      ${field("OS", agent.os || "-")}
      ${field("Arch", agent.arch || "-")}
      ${field("Version", agent.version || "-")}
      ${field("Last Seen", formatDate(agent.last_seen))}
    </div>
    <span class="eyebrow">Capabilities</span>
    <div class="cap-list">${capabilityList(agent.capabilities || [])}</div>
  `;
}

function renderEmptyDetail() {
  els.detailBody.className = "detail-body empty-state";
  els.detailBody.innerHTML = `
    <svg><use href="#icon-database"></use></svg>
    <p>No selection</p>
  `;
}

function selectRow(kind, id) {
  state.selectedKind = kind;
  state.selectedId = id;
  if (kind === "agent") {
    state.view = "agents";
  }
  render();
}

async function lookupJob(id) {
  if (!id.trim()) {
    showToast("Job ID is required");
    return;
  }

  try {
    const job = await fetchJSON(`/jobs/${encodeURIComponent(id.trim())}`);
    const index = state.jobs.findIndex((item) => item.id === job.id);
    if (index >= 0) {
      state.jobs[index] = job;
    } else {
      state.jobs.unshift(job);
    }
    state.view = "jobs";
    selectRow("job", job.id);
  } catch (error) {
    showToast(error.message);
  }
}

function normalizedFilter() {
  return els.filterInput.value.trim().toLowerCase();
}

function matchesFilter(filter, values) {
  if (!filter) {
    return true;
  }
  return values.some((value) => String(value || "").toLowerCase().includes(filter));
}

function isSelected(kind, id) {
  return state.selectedKind === kind && state.selectedId === id;
}

function statusPill(status) {
  const value = String(status || "unknown");
  return `<span class="status-pill ${escapeAttr(value.toLowerCase())}">${escapeHTML(value)}</span>`;
}

function capabilityList(capabilities) {
  if (!capabilities.length) {
    return `<span class="cap">none</span>`;
  }
  return capabilities.map((capability) => `<span class="cap">${escapeHTML(capability)}</span>`).join("");
}

function field(label, value) {
  return `
    <div class="field">
      <span>${escapeHTML(label)}</span>
      <strong>${escapeHTML(value)}</strong>
    </div>
  `;
}

function emptyRow(label, colspan) {
  return `<tr><td colspan="${colspan}">${escapeHTML(label)}</td></tr>`;
}

function formatDate(value) {
  if (!value) {
    return "-";
  }
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }
  return new Intl.DateTimeFormat(undefined, {
    month: "short",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
  }).format(date);
}

function formatJSON(value) {
  return JSON.stringify(value, null, 2);
}

function escapeHTML(value) {
  return String(value)
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#039;");
}

function escapeAttr(value) {
  return escapeHTML(value);
}

function showToast(message) {
  window.clearTimeout(state.toastTimer);
  els.toast.textContent = message;
  els.toast.classList.add("is-visible");
  state.toastTimer = window.setTimeout(() => {
    els.toast.classList.remove("is-visible");
  }, 3200);
}

els.refreshButton.addEventListener("click", refreshAll);

els.apiBaseInput.addEventListener("change", () => {
  state.apiBase = els.apiBaseInput.value.trim() || defaultAPIBase;
  refreshAll();
});

els.jobsTab.addEventListener("click", () => {
  state.view = "jobs";
  render();
});

els.agentsTab.addEventListener("click", () => {
  state.view = "agents";
  render();
});

els.filterInput.addEventListener("input", render);

els.createJobType.addEventListener("change", () => {
  setPayloadTemplate(els.createJobType.value, true);
});

els.createJobForm.addEventListener("submit", (event) => {
  event.preventDefault();
  createJob();
});

els.jobsTable.addEventListener("click", (event) => {
  const row = event.target.closest("tr[data-kind='job']");
  if (row) {
    selectRow("job", row.dataset.id);
  }
});

els.agentsTable.addEventListener("click", (event) => {
  const row = event.target.closest("tr[data-kind='agent']");
  if (row) {
    selectRow("agent", row.dataset.id);
  }
});

els.jobLookupForm.addEventListener("submit", (event) => {
  event.preventDefault();
  lookupJob(els.jobLookupInput.value);
});

setPayloadTemplate(els.createJobType.value, true);
els.apiBaseInput.value = state.apiBase;
refreshAll();
