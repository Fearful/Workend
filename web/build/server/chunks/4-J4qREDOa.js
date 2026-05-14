import { f as fail, r as redirect, e as error } from './index-BkmUvga9.js';
import { a as apiFetch } from './api-chcxdA-p.js';

const SESSION_COOKIE = "workend_session";
const load = async ({ locals, cookies }) => {
  if (!locals.user) throw redirect(303, "/login");
  if (!locals.user.is_admin) throw error(403, "admin only");
  const cookie = cookies.get(SESSION_COOKIE);
  const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
  const [usersResult, auditResult] = await Promise.all([
    apiFetch("/api/admin/users", { cookie: cookieHeader }),
    apiFetch("/api/admin/audit-log", { cookie: cookieHeader })
  ]);
  return {
    users: usersResult.ok ? usersResult.data ?? [] : [],
    audit: auditResult.ok ? auditResult.data ?? [] : []
  };
};
const actions = {
  setRetention: async ({ request, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
    const data = await request.formData();
    const days = parseInt(String(data.get("days") || ""), 10);
    if (isNaN(days) || days < 1) {
      return fail(400, { retentionError: "A valid number of days is required (minimum 1)" });
    }
    const result = await apiFetch("/api/admin/audit-log/retention", {
      method: "POST",
      body: { days },
      cookie: cookieHeader
    });
    if (!result.ok) return fail(result.status, { retentionError: result.error || "failed to set retention" });
    return { retentionSet: true, retentionDays: days };
  }
};

var _page_server_ts = /*#__PURE__*/Object.freeze({
  __proto__: null,
  actions: actions,
  load: load
});

const index = 4;
let component_cache;
const component = async () => component_cache ??= (await import('./_page.svelte-CBJSiN3O.js')).default;
const server_id = "src/routes/admin/+page.server.ts";
const imports = ["_app/immutable/nodes/4.CtiVQOGN.js","_app/immutable/chunks/CWj6FrbW.js","_app/immutable/chunks/BCGREYIq.js","_app/immutable/chunks/glOr0qzD.js","_app/immutable/chunks/BcfnGdFX.js","_app/immutable/chunks/D8xXPBaV.js","_app/immutable/chunks/C5qpB67O.js","_app/immutable/chunks/CT9QbQKf.js","_app/immutable/chunks/BS2x50Pa.js","_app/immutable/chunks/C-j7aE8P.js","_app/immutable/chunks/Bq1FU2Pv.js","_app/immutable/chunks/CHpGRhK0.js"];
const stylesheets = ["_app/immutable/assets/PageHeader.Ct_CmObI.css","_app/immutable/assets/Panel.qbNo31SY.css","_app/immutable/assets/Badge.Cj2_Wv2F.css","_app/immutable/assets/4.jiiUQkUp.css"];
const fonts = [];

export { component, fonts, imports, index, _page_server_ts as server, server_id, stylesheets };
//# sourceMappingURL=4-J4qREDOa.js.map
