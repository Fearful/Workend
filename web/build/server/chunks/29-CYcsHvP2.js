import { f as fail, r as redirect } from './index-BkmUvga9.js';
import { a as apiFetch } from './api-chcxdA-p.js';

const SESSION_COOKIE = "workend_session";
const load = async ({ locals, cookies }) => {
  if (!locals.user) throw redirect(303, "/login");
  const cookie = cookies.get(SESSION_COOKIE);
  const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
  const [notifsResp, subsResp, wsResp] = await Promise.all([
    apiFetch("/api/me/notifications", { cookie: cookieHeader }),
    apiFetch("/api/me/notification-subscriptions", { cookie: cookieHeader }),
    apiFetch("/api/workspaces", { cookie: cookieHeader })
  ]);
  const workspaces = wsResp.ok ? wsResp.data ?? [] : [];
  const projectsByWS = {};
  await Promise.all(workspaces.map(async (w) => {
    const r = await apiFetch(`/api/workspaces/${w.id}/projects`, { cookie: cookieHeader });
    if (r.ok && r.data) projectsByWS[w.id] = r.data;
  }));
  return {
    notifications: notifsResp.ok ? notifsResp.data ?? [] : [],
    subscriptions: subsResp.ok ? subsResp.data ?? [] : [],
    workspaces,
    projectsByWS
  };
};
const actions = {
  create: async ({ request, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const data = await request.formData();
    const kind = String(data.get("kind") || "");
    const target = String(data.get("target") || "").trim();
    const trigger = String(data.get("trigger") || "on_failure");
    if (!kind || !target) return fail(400, { error: "kind and target required", target });
    const result = await apiFetch("/api/me/notifications", {
      method: "POST",
      body: { kind, target, trigger },
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : void 0
    });
    if (!result.ok) return fail(result.status, { error: result.error || "create failed", target });
    return { created: true };
  },
  delete: async ({ request, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const id = String((await request.formData()).get("id") || "");
    const result = await apiFetch(`/api/me/notifications/${id}`, {
      method: "DELETE",
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : void 0
    });
    if (!result.ok) return fail(result.status, { error: result.error || "delete failed" });
    return { deleted: true };
  },
  test: async ({ request, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const id = String((await request.formData()).get("id") || "");
    const result = await apiFetch(`/api/me/notifications/${id}/test`, {
      method: "POST",
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : void 0
    });
    if (!result.ok) return fail(result.status, { error: result.error || "test send failed" });
    return { tested: true };
  },
  subscribe: async ({ request, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const data = await request.formData();
    const config_id = String(data.get("config_id") || "");
    const scope_type = String(data.get("scope_type") || "global");
    const scope_id = String(data.get("scope_id") || "").trim();
    const severity = String(data.get("severity") || "failures");
    const result = await apiFetch("/api/me/notification-subscriptions", {
      method: "POST",
      body: { config_id, scope_type, scope_id, severity },
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : void 0
    });
    if (!result.ok) return fail(result.status, { error: result.error || "subscribe failed" });
    return { subscribed: true };
  },
  unsubscribe: async ({ request, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const id = String((await request.formData()).get("id") || "");
    const result = await apiFetch(`/api/me/notification-subscriptions/${id}`, {
      method: "DELETE",
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : void 0
    });
    if (!result.ok) return fail(result.status, { error: result.error || "unsubscribe failed" });
    return { unsubscribed: true };
  }
};

var _page_server_ts = /*#__PURE__*/Object.freeze({
  __proto__: null,
  actions: actions,
  load: load
});

const index = 29;
let component_cache;
const component = async () => component_cache ??= (await import('./_page.svelte-BGKzw1f4.js')).default;
const server_id = "src/routes/settings/notifications/+page.server.ts";
const imports = ["_app/immutable/nodes/29.C0yOeIdY.js","_app/immutable/chunks/CWj6FrbW.js","_app/immutable/chunks/BCGREYIq.js","_app/immutable/chunks/glOr0qzD.js","_app/immutable/chunks/BcfnGdFX.js","_app/immutable/chunks/D8xXPBaV.js","_app/immutable/chunks/Dv-DGh57.js","_app/immutable/chunks/Bx7T9aDp.js","_app/immutable/chunks/CT9QbQKf.js","_app/immutable/chunks/BS2x50Pa.js","_app/immutable/chunks/C-j7aE8P.js","_app/immutable/chunks/Bq1FU2Pv.js","_app/immutable/chunks/CHpGRhK0.js","_app/immutable/chunks/YSzAdjSj.js","_app/immutable/chunks/wpI5rr64.js"];
const stylesheets = ["_app/immutable/assets/Breadcrumb.BmuUfcGr.css","_app/immutable/assets/PageHeader.Ct_CmObI.css","_app/immutable/assets/Panel.qbNo31SY.css","_app/immutable/assets/Badge.Cj2_Wv2F.css","_app/immutable/assets/EmptyState.ByauYn1J.css","_app/immutable/assets/29.jxliwq_n.css"];
const fonts = [];

export { component, fonts, imports, index, _page_server_ts as server, server_id, stylesheets };
//# sourceMappingURL=29-CYcsHvP2.js.map
