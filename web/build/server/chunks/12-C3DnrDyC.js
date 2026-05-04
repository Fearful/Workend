import { f as fail, r as redirect } from './index-BkmUvga9.js';
import { a as apiFetch } from './api-chcxdA-p.js';

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

var _page_server_ts = /*#__PURE__*/Object.freeze({
  __proto__: null,
  actions: actions,
  load: load
});

const index = 12;
let component_cache;
const component = async () => component_cache ??= (await import('./_page.svelte-DCt_JGOp.js')).default;
const server_id = "src/routes/projects/[id]/+page.server.ts";
const imports = ["_app/immutable/nodes/12.DZfNTlMI.js","_app/immutable/chunks/C4JmIpy3.js","_app/immutable/chunks/apxFJXnE.js","_app/immutable/chunks/Dd7pVJ5Y.js","_app/immutable/chunks/CvynrFVy.js","_app/immutable/chunks/CvI2x4sI.js","_app/immutable/chunks/Bn0VCKPQ.js","_app/immutable/chunks/cLUGwKvc.js","_app/immutable/chunks/ChBkQGhY.js","_app/immutable/chunks/Dx6EV4QR.js","_app/immutable/chunks/CYuiovvJ.js","_app/immutable/chunks/Dmb8Bvf1.js","_app/immutable/chunks/CgWWeIEW.js","_app/immutable/chunks/8aH_EF2d.js","_app/immutable/chunks/BZxKEKJE.js","_app/immutable/chunks/BiQCKUjL.js","_app/immutable/chunks/BvqIPo9c.js","_app/immutable/chunks/BO09PiA2.js","_app/immutable/chunks/Cxf1vVyU.js","_app/immutable/chunks/m1uqkJwC.js"];
const stylesheets = ["_app/immutable/assets/Panel.Dd2v_-2B.css","_app/immutable/assets/Badge.Cj2_Wv2F.css","_app/immutable/assets/Modal.CHlCiKdq.css","_app/immutable/assets/12.CCYAIcq8.css"];
const fonts = [];

export { component, fonts, imports, index, _page_server_ts as server, server_id, stylesheets };
//# sourceMappingURL=12-C3DnrDyC.js.map
