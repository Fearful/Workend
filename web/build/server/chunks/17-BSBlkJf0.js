import { f as fail, r as redirect } from './index-BkmUvga9.js';
import { a as apiFetch } from './api-chcxdA-p.js';

const SESSION_COOKIE = "workend_session";
const load = async ({ params, locals, cookies }) => {
  if (!locals.user) throw redirect(303, "/login");
  const cookie = cookies.get(SESSION_COOKIE);
  const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
  const [tasksResult, schedulesResult] = await Promise.all([
    apiFetch(`/api/projects/${params.id}/tasks`, { cookie: cookieHeader }),
    apiFetch(`/api/projects/${params.id}/schedules`, { cookie: cookieHeader })
  ]);
  return {
    tasks: tasksResult.ok ? tasksResult.data ?? [] : [],
    schedules: schedulesResult.ok ? schedulesResult.data ?? [] : []
  };
};
const actions = {
  create: async ({ params, request, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const data = await request.formData();
    const task_id = String(data.get("task_id") || "");
    const cron_expr = String(data.get("cron_expr") || "").trim();
    const enabled = data.get("enabled") === "on" || data.get("enabled") === "true";
    if (!task_id || !cron_expr) {
      return fail(400, { error: "task and cron expression required", cron_expr });
    }
    const result = await apiFetch(`/api/projects/${params.id}/schedules`, {
      method: "POST",
      body: { task_id, cron_expr, enabled },
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : void 0
    });
    if (!result.ok) return fail(result.status, { error: result.error || "create failed", cron_expr });
    return { created: true };
  },
  toggle: async ({ request, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const data = await request.formData();
    const id = String(data.get("id") || "");
    const result = await apiFetch(`/api/schedules/${id}/toggle`, {
      method: "POST",
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : void 0
    });
    if (!result.ok) return fail(result.status, { error: result.error || "toggle failed" });
    return { toggled: true };
  },
  delete: async ({ request, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const data = await request.formData();
    const id = String(data.get("id") || "");
    const result = await apiFetch(`/api/schedules/${id}`, {
      method: "DELETE",
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : void 0
    });
    if (!result.ok) return fail(result.status, { error: result.error || "delete failed" });
    return { deleted: true };
  }
};

var _page_server_ts = /*#__PURE__*/Object.freeze({
  __proto__: null,
  actions: actions,
  load: load
});

const index = 17;
let component_cache;
const component = async () => component_cache ??= (await import('./_page.svelte-FJWP7E_W.js')).default;
const server_id = "src/routes/projects/[id]/schedules/+page.server.ts";
const imports = ["_app/immutable/nodes/17.DuC2aKud.js","_app/immutable/chunks/C4JmIpy3.js","_app/immutable/chunks/apxFJXnE.js","_app/immutable/chunks/CvynrFVy.js","_app/immutable/chunks/CvI2x4sI.js","_app/immutable/chunks/Bn0VCKPQ.js","_app/immutable/chunks/cLUGwKvc.js","_app/immutable/chunks/BZxKEKJE.js","_app/immutable/chunks/Dx6EV4QR.js","_app/immutable/chunks/ChBkQGhY.js","_app/immutable/chunks/BiQCKUjL.js","_app/immutable/chunks/BO09PiA2.js","_app/immutable/chunks/Di6J7QjO.js","_app/immutable/chunks/Cxf1vVyU.js","_app/immutable/chunks/m1uqkJwC.js","_app/immutable/chunks/CYuiovvJ.js"];
const stylesheets = ["_app/immutable/assets/Panel.Dd2v_-2B.css","_app/immutable/assets/Badge.Cj2_Wv2F.css","_app/immutable/assets/EmptyState.ByauYn1J.css","_app/immutable/assets/Modal.CHlCiKdq.css","_app/immutable/assets/17.w0jJVxlJ.css"];
const fonts = [];

export { component, fonts, imports, index, _page_server_ts as server, server_id, stylesheets };
//# sourceMappingURL=17-BSBlkJf0.js.map
