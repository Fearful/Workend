import { fail, redirect } from "@sveltejs/kit";
import { a as apiFetch } from "../../../../../chunks/api.js";
const SESSION_COOKIE = "workend_session";
const load = async ({ params, locals, cookies }) => {
  if (!locals.user) throw redirect(303, "/login");
  const cookie = cookies.get(SESSION_COOKIE);
  const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
  const [pipelinesResult, tasksResult] = await Promise.all([
    apiFetch(`/api/projects/${params.id}/pipelines`, { cookie: cookieHeader }),
    apiFetch(`/api/projects/${params.id}/tasks`, { cookie: cookieHeader })
  ]);
  return {
    pipelines: pipelinesResult.ok ? pipelinesResult.data ?? [] : [],
    tasks: tasksResult.ok ? tasksResult.data ?? [] : []
  };
};
const actions = {
  create: async ({ params, request, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const data = await request.formData();
    const name = String(data.get("name") || "").trim();
    const taskIDs = data.getAll("task_id").map((v) => String(v)).filter(Boolean);
    if (!name) return fail(400, { error: "name required", name });
    if (taskIDs.length === 0) return fail(400, { error: "pick at least one task", name });
    const result = await apiFetch(`/api/projects/${params.id}/pipelines`, {
      method: "POST",
      body: { name, task_ids: taskIDs },
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : void 0
    });
    if (!result.ok) return fail(result.status, { error: result.error || "create failed", name });
    return { created: true };
  },
  delete: async ({ request, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const data = await request.formData();
    const id = String(data.get("id") || "");
    if (!id) return fail(400, { error: "id required" });
    const result = await apiFetch(`/api/pipelines/${id}`, {
      method: "DELETE",
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : void 0
    });
    if (!result.ok) return fail(result.status, { error: result.error || "delete failed" });
    return { deleted: true };
  },
  run: async ({ request, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const data = await request.formData();
    const id = String(data.get("id") || "");
    if (!id) return fail(400, { error: "id required" });
    const result = await apiFetch(`/api/pipelines/${id}/runs`, {
      method: "POST",
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : void 0
    });
    if (!result.ok || !result.data) {
      return fail(result.status, { error: result.error || "pipeline run failed to start" });
    }
    throw redirect(303, `/pipeline-runs/${result.data.pipeline_run_id}`);
  }
};
export {
  actions,
  load
};
