import { r as redirect } from './index-BkmUvga9.js';
import { a as apiFetch } from './api-chcxdA-p.js';

const SESSION_COOKIE = "workend_session";
const load = async ({ locals, cookies }) => {
  if (!locals.user) throw redirect(303, "/login");
  const token = cookies.get(SESSION_COOKIE);
  const cookieHeader = token ? `${SESSION_COOKIE}=${token}` : void 0;
  const [workspacesResult, runsResult] = await Promise.all([
    apiFetch("/api/workspaces", { cookie: cookieHeader }),
    apiFetch("/api/me/runs", { cookie: cookieHeader })
  ]);
  return {
    workspaces: workspacesResult.ok ? workspacesResult.data ?? [] : [],
    recentRuns: runsResult.ok ? runsResult.data ?? [] : [],
    error: workspacesResult.ok ? null : workspacesResult.error || "failed to load workspaces"
  };
};

var _page_server_ts = /*#__PURE__*/Object.freeze({
  __proto__: null,
  load: load
});

const index = 3;
let component_cache;
const component = async () => component_cache ??= (await import('./_page.svelte-CG3RtZnl.js')).default;
const server_id = "src/routes/+page.server.ts";
const imports = ["_app/immutable/nodes/3.Zxb6yjsC.js","_app/immutable/chunks/C4JmIpy3.js","_app/immutable/chunks/apxFJXnE.js","_app/immutable/chunks/CvynrFVy.js","_app/immutable/chunks/CvI2x4sI.js","_app/immutable/chunks/Bn0VCKPQ.js","_app/immutable/chunks/cLUGwKvc.js","_app/immutable/chunks/DC4q0MGt.js","_app/immutable/chunks/Dx6EV4QR.js","_app/immutable/chunks/Di6J7QjO.js","_app/immutable/chunks/Cxf1vVyU.js","_app/immutable/chunks/ChBkQGhY.js","_app/immutable/chunks/BiQCKUjL.js","_app/immutable/chunks/CYuiovvJ.js","_app/immutable/chunks/8aH_EF2d.js"];
const stylesheets = ["_app/immutable/assets/PageHeader.Ct_CmObI.css","_app/immutable/assets/EmptyState.ByauYn1J.css","_app/immutable/assets/3.9tb9nhd3.css"];
const fonts = [];

export { component, fonts, imports, index, _page_server_ts as server, server_id, stylesheets };
//# sourceMappingURL=3-yM9ZFa9_.js.map
