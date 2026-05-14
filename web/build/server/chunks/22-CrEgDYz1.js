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

const index = 22;
let component_cache;
const component = async () => component_cache ??= (await import('./_page.svelte-CjdW2Wkm.js')).default;
const server_id = "src/routes/projects/[id]/schedules/+page.server.ts";
const imports = ["_app/immutable/nodes/22.BDmRP1wD.js","_app/immutable/chunks/CWj6FrbW.js","_app/immutable/chunks/BCGREYIq.js","_app/immutable/chunks/glOr0qzD.js","_app/immutable/chunks/BcfnGdFX.js","_app/immutable/chunks/D8xXPBaV.js","_app/immutable/chunks/C-j7aE8P.js","_app/immutable/chunks/BS2x50Pa.js","_app/immutable/chunks/Bq1FU2Pv.js","_app/immutable/chunks/CHpGRhK0.js","_app/immutable/chunks/YSzAdjSj.js","_app/immutable/chunks/wpI5rr64.js","_app/immutable/chunks/ilPNOASy.js","_app/immutable/chunks/vbEortrD.js","_app/immutable/chunks/B8IgtE21.js","_app/immutable/chunks/B5rGamLH.js","_app/immutable/chunks/DZ5gTY8C.js","_app/immutable/chunks/Di6_auBm.js","_app/immutable/chunks/B2UVO_nM.js"];
const stylesheets = ["_app/immutable/assets/Panel.qbNo31SY.css","_app/immutable/assets/Badge.Cj2_Wv2F.css","_app/immutable/assets/EmptyState.ByauYn1J.css","_app/immutable/assets/Modal.CHlCiKdq.css","_app/immutable/assets/SectionHeader.Bb9ZRXXP.css","_app/immutable/assets/Tooltip.BrYf-THh.css","_app/immutable/assets/TimeAgo.BYfizd1M.css","_app/immutable/assets/22.sHRtwR1W.css"];
const fonts = [];

export { component, fonts, imports, index, _page_server_ts as server, server_id, stylesheets };
//# sourceMappingURL=22-CrEgDYz1.js.map
