import { fail, redirect } from "@sveltejs/kit";
import { a as apiFetch } from "../../../../chunks/api.js";
const SESSION_COOKIE = "workend_session";
const load = async ({ params, locals, cookies }) => {
  if (!locals.user) throw redirect(303, "/login");
  const cookie = cookies.get(SESSION_COOKIE);
  const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
  const [
    tasksResult,
    runsResult,
    statsResult,
    pinsResult,
    pipelinesResult,
    ciConfigsResult,
    composeResult,
    overviewResult,
    favoritesResult,
    blameResult,
    packagesResult,
    previewsResult,
    impactResult,
    depsResult,
    liveRunResult,
    widgetAlertsResult,
    alertRulesResult
  ] = await Promise.all([
    apiFetch(`/api/projects/${params.id}/tasks`, { cookie: cookieHeader }),
    apiFetch(`/api/projects/${params.id}/runs`, { cookie: cookieHeader }),
    apiFetch(`/api/projects/${params.id}/stats`, { cookie: cookieHeader }),
    apiFetch(`/api/me/pinned-tasks`, { cookie: cookieHeader }),
    apiFetch(`/api/projects/${params.id}/remote-pipelines`, { cookie: cookieHeader }),
    apiFetch(`/api/projects/${params.id}/pipeline-configs`, { cookie: cookieHeader }),
    apiFetch(`/api/projects/${params.id}/compose`, { cookie: cookieHeader }),
    apiFetch(`/api/projects/${params.id}/overview`, { cookie: cookieHeader }),
    apiFetch(`/api/me/favorites`, { cookie: cookieHeader }),
    apiFetch(`/api/projects/${params.id}/blame-timeline?limit=50`, { cookie: cookieHeader }),
    apiFetch(`/api/projects/${params.id}/monorepo/packages`, { cookie: cookieHeader }),
    apiFetch(`/api/projects/${params.id}/previews`, { cookie: cookieHeader }),
    apiFetch(`/api/widgets/project/${params.id}/impact-radar`, { cookie: cookieHeader }),
    apiFetch(`/api/widgets/project/${params.id}/dependency-tree`, { cookie: cookieHeader }),
    apiFetch(`/api/widgets/project/${params.id}/live-run`, { cookie: cookieHeader }),
    apiFetch(`/api/widgets/project/${params.id}/alert-rules`, { cookie: cookieHeader }),
    apiFetch(`/api/me/alert-rules`, { cookie: cookieHeader })
  ]);
  const allFavorites = favoritesResult.ok ? favoritesResult.data ?? [] : [];
  const isFavorited = allFavorites.some((f) => f.project_id === params.id);
  return {
    tasks: tasksResult.ok ? tasksResult.data ?? [] : [],
    runs: runsResult.ok ? runsResult.data ?? [] : [],
    stats: statsResult.ok ? statsResult.data ?? null : null,
    pinnedTaskIDs: pinsResult.ok ? (pinsResult.data ?? []).map((p) => p.task_id) : [],
    remotePipelines: pipelinesResult.ok ? pipelinesResult.data ?? [] : [],
    pipelineConfigs: ciConfigsResult.ok ? ciConfigsResult.data ?? [] : [],
    compose: composeResult.ok ? composeResult.data ?? null : null,
    overview: overviewResult.ok ? overviewResult.data ?? null : null,
    isFavorited,
    blameTimeline: blameResult.ok ? blameResult.data ?? [] : [],
    monorepoPackages: packagesResult.ok ? packagesResult.data ?? [] : [],
    previews: previewsResult.ok ? previewsResult.data ?? [] : [],
    impactRadar: impactResult.ok ? impactResult.data ?? null : null,
    dependencyTree: depsResult.ok ? depsResult.data ?? null : null,
    liveRun: liveRunResult.ok ? liveRunResult.data ?? null : null,
    widgetAlertRules: widgetAlertsResult.ok ? widgetAlertsResult.data ?? [] : [],
    alertRules: (alertRulesResult.ok ? alertRulesResult.data ?? [] : []).filter(
      (r) => (tasksResult.ok ? tasksResult.data ?? [] : []).some((t) => t.id === r.task_id)
    )
  };
};
const actions = {
  run: async ({ request, cookies }) => {
    const data = await request.formData();
    const taskID = String(data.get("task_id") || "");
    if (!taskID) return fail(400, { error: "task_id required" });
    const cookie = cookies.get(SESSION_COOKIE);
    const result = await apiFetch(`/api/tasks/${taskID}/runs`, {
      method: "POST",
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : void 0
    });
    if (!result.ok || !result.data) {
      return fail(result.status, { error: result.error || "run failed to start" });
    }
    throw redirect(303, `/runs/${result.data.id}`);
  },
  runWithParams: async ({ request, cookies }) => {
    const data = await request.formData();
    const taskID = String(data.get("task_id") || "");
    if (!taskID) return fail(400, { error: "task_id required" });
    let env = {};
    let args = [];
    try {
      env = JSON.parse(String(data.get("env") || "{}"));
      args = JSON.parse(String(data.get("args") || "[]"));
    } catch {
      return fail(400, { error: "invalid env/args JSON" });
    }
    const cookie = cookies.get(SESSION_COOKIE);
    const result = await apiFetch(`/api/tasks/${taskID}/runs`, {
      method: "POST",
      body: { env, args },
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : void 0
    });
    if (!result.ok || !result.data) {
      return fail(result.status, { error: result.error || "run failed to start" });
    }
    throw redirect(303, `/runs/${result.data.id}`);
  },
  runOnBranch: async ({ request, cookies }) => {
    const data = await request.formData();
    const taskID = String(data.get("task_id") || "");
    const branch = String(data.get("branch") || "").trim();
    if (!taskID) return fail(400, { error: "task_id required" });
    if (!branch) return fail(400, { error: "branch required" });
    const cookie = cookies.get(SESSION_COOKIE);
    const result = await apiFetch(`/api/tasks/${taskID}/runs`, {
      method: "POST",
      body: { branch },
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : void 0
    });
    if (!result.ok || !result.data) {
      return fail(result.status, { error: result.error || "run failed to start" });
    }
    throw redirect(303, `/runs/${result.data.id}`);
  },
  setRetry: async ({ request, cookies }) => {
    const data = await request.formData();
    const taskID = String(data.get("task_id") || "");
    const maxRaw = String(data.get("retry_max") || "").trim();
    const backoffRaw = String(data.get("retry_backoff_sec") || "").trim();
    if (!taskID) return fail(400, { error: "task_id required" });
    const body = {};
    if (maxRaw !== "") {
      const n = parseInt(maxRaw, 10);
      if (!Number.isFinite(n) || n < 0 || n > 10) return fail(400, { error: "retry_max must be 0..10" });
      body.retry_max = n;
    }
    if (backoffRaw !== "") {
      const n = parseInt(backoffRaw, 10);
      if (!Number.isFinite(n) || n < 1 || n > 3600) return fail(400, { error: "retry_backoff_sec must be 1..3600" });
      body.retry_backoff_sec = n;
    }
    const cookie = cookies.get(SESSION_COOKIE);
    const result = await apiFetch(`/api/tasks/${taskID}`, {
      method: "PATCH",
      body,
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : void 0
    });
    if (!result.ok) return fail(result.status, { error: result.error || "failed" });
    return { ok: true };
  },
  toggleApproval: async ({ request, cookies }) => {
    const data = await request.formData();
    const taskID = String(data.get("task_id") || "");
    const next = String(data.get("next") || "") === "true";
    if (!taskID) return fail(400, { error: "task_id required" });
    const cookie = cookies.get(SESSION_COOKIE);
    const result = await apiFetch(`/api/tasks/${taskID}`, {
      method: "PATCH",
      body: { requires_approval: next },
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : void 0
    });
    if (!result.ok) return fail(result.status, { error: result.error || "failed" });
    return { ok: true };
  },
  togglePin: async ({ request, cookies }) => {
    const data = await request.formData();
    const taskID = String(data.get("task_id") || "");
    const pinned = String(data.get("pinned") || "") === "true";
    if (!taskID) return fail(400, { error: "task_id required" });
    const cookie = cookies.get(SESSION_COOKIE);
    const result = await apiFetch(`/api/tasks/${taskID}/pin`, {
      method: pinned ? "DELETE" : "POST",
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : void 0
    });
    if (!result.ok) return fail(result.status, { error: result.error || "pin failed" });
    return { pinned: !pinned };
  },
  setTimeout: async ({ request, cookies }) => {
    const data = await request.formData();
    const taskID = String(data.get("task_id") || "");
    const raw = String(data.get("timeout_seconds") || "").trim();
    if (!taskID) return fail(400, { error: "task_id required" });
    let body;
    if (raw === "") {
      body = { timeout_seconds: null };
    } else {
      const n = parseInt(raw, 10);
      if (!Number.isFinite(n) || n < 1 || n > 86400) {
        return fail(400, { error: "timeout_seconds must be 1..86400 or blank" });
      }
      body = { timeout_seconds: n };
    }
    const cookie = cookies.get(SESSION_COOKIE);
    const result = await apiFetch(`/api/tasks/${taskID}`, {
      method: "PATCH",
      body,
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : void 0
    });
    if (!result.ok) return fail(result.status, { error: result.error || "update failed" });
    return { updated: true };
  },
  setArtifactPatterns: async ({ request, cookies }) => {
    const data = await request.formData();
    const taskID = String(data.get("task_id") || "");
    const raw = String(data.get("artifact_patterns") || "").trim();
    if (!taskID) return fail(400, { error: "task_id required" });
    const patterns = raw === "" ? [] : raw.split("\n").map((s) => s.trim()).filter(Boolean);
    if (patterns.length > 32) return fail(400, { error: "too many patterns (max 32)" });
    const cookie = cookies.get(SESSION_COOKIE);
    const result = await apiFetch(`/api/tasks/${taskID}`, {
      method: "PATCH",
      body: { artifact_patterns: patterns },
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : void 0
    });
    if (!result.ok) return fail(result.status, { error: result.error || "update failed" });
    return { updated: true };
  },
  setNeedsServices: async ({ request, cookies }) => {
    const data = await request.formData();
    const taskID = String(data.get("task_id") || "");
    const raw = String(data.get("needs_services") || "").trim();
    if (!taskID) return fail(400, { error: "task_id required" });
    const services = raw === "" ? [] : raw.split(",").map((s) => s.trim()).filter(Boolean);
    if (services.length > 32) return fail(400, { error: "too many services (max 32)" });
    const cookie = cookies.get(SESSION_COOKIE);
    const result = await apiFetch(`/api/tasks/${taskID}`, {
      method: "PATCH",
      body: { needs_services: services },
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : void 0
    });
    if (!result.ok) return fail(result.status, { error: result.error || "update failed" });
    return { updated: true };
  },
  setConcurrency: async ({ request, cookies }) => {
    const data = await request.formData();
    const taskID = String(data.get("task_id") || "");
    const maxRaw = String(data.get("max_concurrency") || "").trim();
    const policy = String(data.get("supersede_policy") || "queue");
    if (!taskID) return fail(400, { error: "task_id required" });
    const max = parseInt(maxRaw, 10);
    if (!Number.isFinite(max) || max < 0 || max > 100) {
      return fail(400, { error: "max_concurrency must be 0..100" });
    }
    if (policy !== "queue" && policy !== "cancel-old" && policy !== "reject") {
      return fail(400, { error: "supersede_policy must be queue|cancel-old|reject" });
    }
    const cookie = cookies.get(SESSION_COOKIE);
    const result = await apiFetch(`/api/tasks/${taskID}`, {
      method: "PATCH",
      body: { max_concurrency: max, supersede_policy: policy },
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : void 0
    });
    if (!result.ok) return fail(result.status, { error: result.error || "update failed" });
    return { updated: true };
  },
  toggleFavorite: async ({ params, request, cookies }) => {
    const formData = await request.formData();
    const isFavorited = String(formData.get("is_favorited") || "") === "true";
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
    const result = await apiFetch(`/api/projects/${params.id}/favorite`, {
      method: isFavorited ? "DELETE" : "POST",
      cookie: cookieHeader
    });
    if (!result.ok) return fail(result.status, { error: result.error || "favorite toggle failed" });
    return { favorited: !isFavorited };
  },
  syncCommits: async ({ params, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
    const result = await apiFetch(
      `/api/projects/${params.id}/sync-commits`,
      { method: "POST", cookie: cookieHeader }
    );
    if (!result.ok) return fail(result.status, { error: result.error || "commit sync failed" });
    return { synced: result.data?.synced ?? 0, inserted: result.data?.inserted ?? 0 };
  },
  detectPackages: async ({ params, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
    const result = await apiFetch(
      `/api/projects/${params.id}/monorepo/detect`,
      { method: "POST", cookie: cookieHeader }
    );
    if (!result.ok) return fail(result.status, { error: result.error || "package detection failed" });
    return { detected: result.data?.detected ?? 0 };
  },
  autoMapTasks: async ({ params, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
    const result = await apiFetch(
      `/api/projects/${params.id}/monorepo/auto-map`,
      { method: "POST", cookie: cookieHeader }
    );
    if (!result.ok) return fail(result.status, { error: result.error || "auto-map failed" });
    return { mapped: result.data?.mapped ?? 0 };
  },
  createPreview: async ({ params, request, cookies }) => {
    const formData = await request.formData();
    const branch = String(formData.get("branch") || "").trim();
    if (!branch) return fail(400, { error: "branch required" });
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
    const result = await apiFetch(`/api/projects/${params.id}/previews`, {
      method: "POST",
      body: { branch },
      cookie: cookieHeader
    });
    if (!result.ok) return fail(result.status, { error: result.error || "preview creation failed" });
    return { previewCreated: true };
  },
  createAlertRule: async ({ request, cookies }) => {
    const formData = await request.formData();
    const taskID = String(formData.get("task_id") || "");
    const type = String(formData.get("type") || "");
    const thresholdRaw = String(formData.get("threshold") || "");
    const comparison = String(formData.get("comparison") || "");
    if (!taskID) return fail(400, { error: "task_id required" });
    if (!type) return fail(400, { error: "type required" });
    if (!comparison) return fail(400, { error: "comparison required" });
    const threshold = parseFloat(thresholdRaw);
    if (!Number.isFinite(threshold)) return fail(400, { error: "invalid threshold" });
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
    const result = await apiFetch(`/api/tasks/${taskID}/alert-rules`, {
      method: "POST",
      body: { type, threshold, comparison },
      cookie: cookieHeader
    });
    if (!result.ok) return fail(result.status, { error: result.error || "alert rule creation failed" });
    return { alertCreated: true };
  },
  deleteAlertRule: async ({ request, cookies }) => {
    const formData = await request.formData();
    const ruleID = String(formData.get("rule_id") || "");
    if (!ruleID) return fail(400, { error: "rule_id required" });
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
    const result = await apiFetch(`/api/alert-rules/${ruleID}`, {
      method: "DELETE",
      cookie: cookieHeader
    });
    if (!result.ok) return fail(result.status, { error: result.error || "delete failed" });
    return { alertDeleted: true };
  },
  toggleAlertRule: async ({ request, cookies }) => {
    const formData = await request.formData();
    const ruleID = String(formData.get("rule_id") || "");
    if (!ruleID) return fail(400, { error: "rule_id required" });
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
    const result = await apiFetch(`/api/alert-rules/${ruleID}/toggle`, {
      method: "POST",
      cookie: cookieHeader
    });
    if (!result.ok) return fail(result.status, { error: result.error || "toggle failed" });
    return { alertToggled: true };
  }
};
export {
  actions,
  load
};
