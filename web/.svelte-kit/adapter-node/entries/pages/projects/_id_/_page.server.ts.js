import { fail, redirect } from "@sveltejs/kit";
import { a as apiFetch } from "../../../../chunks/api.js";
const SESSION_COOKIE = "workend_session";
const load = async ({ params, locals, cookies }) => {
  if (!locals.user) throw redirect(303, "/login");
  const cookie = cookies.get(SESSION_COOKIE);
  const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
  const [tasksResult, runsResult, statsResult, pinsResult, pipelinesResult, ciConfigsResult, composeResult] = await Promise.all([
    apiFetch(`/api/projects/${params.id}/tasks`, { cookie: cookieHeader }),
    apiFetch(`/api/projects/${params.id}/runs`, { cookie: cookieHeader }),
    apiFetch(`/api/projects/${params.id}/stats`, { cookie: cookieHeader }),
    apiFetch(`/api/me/pinned-tasks`, { cookie: cookieHeader }),
    apiFetch(`/api/projects/${params.id}/remote-pipelines`, { cookie: cookieHeader }),
    apiFetch(`/api/projects/${params.id}/pipeline-configs`, { cookie: cookieHeader }),
    apiFetch(`/api/projects/${params.id}/compose`, { cookie: cookieHeader })
  ]);
  return {
    tasks: tasksResult.ok ? tasksResult.data ?? [] : [],
    runs: runsResult.ok ? runsResult.data ?? [] : [],
    stats: statsResult.ok ? statsResult.data ?? null : null,
    pinnedTaskIDs: pinsResult.ok ? (pinsResult.data ?? []).map((p) => p.task_id) : [],
    remotePipelines: pipelinesResult.ok ? pipelinesResult.data ?? [] : [],
    pipelineConfigs: ciConfigsResult.ok ? ciConfigsResult.data ?? [] : [],
    compose: composeResult.ok ? composeResult.data ?? null : null
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
  }
};
export {
  actions,
  load
};
