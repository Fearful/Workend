import { r as redirect, e as error } from './index-BkmUvga9.js';
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

var _page_server_ts = /*#__PURE__*/Object.freeze({
  __proto__: null,
  load: load
});

const index = 4;
let component_cache;
const component = async () => component_cache ??= (await import('./_page.svelte-Cs-wP7q9.js')).default;
const server_id = "src/routes/admin/+page.server.ts";
const imports = ["_app/immutable/nodes/4.D7SDRFkU.js","_app/immutable/chunks/C4JmIpy3.js","_app/immutable/chunks/apxFJXnE.js","_app/immutable/chunks/CvynrFVy.js","_app/immutable/chunks/CvI2x4sI.js","_app/immutable/chunks/Bn0VCKPQ.js","_app/immutable/chunks/DC4q0MGt.js","_app/immutable/chunks/Dx6EV4QR.js","_app/immutable/chunks/BO09PiA2.js","_app/immutable/chunks/ChBkQGhY.js","_app/immutable/chunks/BiQCKUjL.js"];
const stylesheets = ["_app/immutable/assets/PageHeader.Ct_CmObI.css","_app/immutable/assets/Badge.Cj2_Wv2F.css","_app/immutable/assets/4.ChWhzFTL.css"];
const fonts = [];

export { component, fonts, imports, index, _page_server_ts as server, server_id, stylesheets };
//# sourceMappingURL=4-OU6ySHMX.js.map
