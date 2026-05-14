import { f as fail, r as redirect } from './index-BkmUvga9.js';
import { a as apiFetch } from './api-chcxdA-p.js';

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

var _page_server_ts = /*#__PURE__*/Object.freeze({
  __proto__: null,
  actions: actions,
  load: load
});

const index = 19;
let component_cache;
const component = async () => component_cache ??= (await import('./_page.svelte-DrzYD6nB.js')).default;
const server_id = "src/routes/projects/[id]/pipelines/+page.server.ts";
const imports = ["_app/immutable/nodes/19.BU6iQ4MU.js","_app/immutable/chunks/CWj6FrbW.js","_app/immutable/chunks/BCGREYIq.js","_app/immutable/chunks/glOr0qzD.js","_app/immutable/chunks/BcfnGdFX.js","_app/immutable/chunks/D8xXPBaV.js","_app/immutable/chunks/C5qpB67O.js","_app/immutable/chunks/DdjgLl7P.js","_app/immutable/chunks/Di6_auBm.js","_app/immutable/chunks/DuSA90i8.js","_app/immutable/chunks/C-j7aE8P.js","_app/immutable/chunks/BS2x50Pa.js","_app/immutable/chunks/Bq1FU2Pv.js","_app/immutable/chunks/YSzAdjSj.js","_app/immutable/chunks/wpI5rr64.js","_app/immutable/chunks/B8IgtE21.js","_app/immutable/chunks/B5rGamLH.js","_app/immutable/chunks/DZ5gTY8C.js","_app/immutable/chunks/vbEortrD.js","_app/immutable/chunks/B2UVO_nM.js","_app/immutable/chunks/ilPNOASy.js"];
const stylesheets = ["_app/immutable/assets/Panel.qbNo31SY.css","_app/immutable/assets/EmptyState.ByauYn1J.css","_app/immutable/assets/SectionHeader.Bb9ZRXXP.css","_app/immutable/assets/Tooltip.BrYf-THh.css","_app/immutable/assets/TimeAgo.BYfizd1M.css","_app/immutable/assets/Modal.CHlCiKdq.css","_app/immutable/assets/19.BOFGbQ5f.css"];
const fonts = [];

export { component, fonts, imports, index, _page_server_ts as server, server_id, stylesheets };
//# sourceMappingURL=19-278eXfBm.js.map
