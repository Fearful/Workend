import { r as redirect } from './index-BkmUvga9.js';
import { a as apiFetch } from './api-chcxdA-p.js';

const SESSION_COOKIE = "workend_session";
const allowedStatuses = /* @__PURE__ */ new Set(["succeeded", "failed", "cancelled", "running", "queued"]);
const load = async ({ params, url, locals, cookies }) => {
  if (!locals.user) throw redirect(303, "/login");
  const cookie = cookies.get(SESSION_COOKIE);
  const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
  const status = url.searchParams.get("status") || "";
  const filter = allowedStatuses.has(status) ? status : "";
  const path = filter ? `/api/projects/${params.id}/runs?status=${encodeURIComponent(filter)}` : `/api/projects/${params.id}/runs`;
  const runsResult = await apiFetch(path, { cookie: cookieHeader });
  return {
    runs: runsResult.ok ? runsResult.data ?? [] : [],
    filter
  };
};

var _page_server_ts = /*#__PURE__*/Object.freeze({
  __proto__: null,
  load: load
});

const index = 16;
let component_cache;
const component = async () => component_cache ??= (await import('./_page.svelte-BX9yJIio.js')).default;
const server_id = "src/routes/projects/[id]/runs/+page.server.ts";
const imports = ["_app/immutable/nodes/16.CNoxma9g.js","_app/immutable/chunks/C4JmIpy3.js","_app/immutable/chunks/apxFJXnE.js","_app/immutable/chunks/CvynrFVy.js","_app/immutable/chunks/CvI2x4sI.js","_app/immutable/chunks/Bn0VCKPQ.js","_app/immutable/chunks/cLUGwKvc.js","_app/immutable/chunks/ChBkQGhY.js","_app/immutable/chunks/Dx6EV4QR.js","_app/immutable/chunks/CYuiovvJ.js","_app/immutable/chunks/CgWWeIEW.js","_app/immutable/chunks/Dd7pVJ5Y.js","_app/immutable/chunks/059uK4jn.js","_app/immutable/chunks/8aH_EF2d.js","_app/immutable/chunks/Di6J7QjO.js"];
const stylesheets = ["_app/immutable/assets/EmptyState.ByauYn1J.css","_app/immutable/assets/16.BCsy4Npt.css"];
const fonts = [];

export { component, fonts, imports, index, _page_server_ts as server, server_id, stylesheets };
//# sourceMappingURL=16-Cb0KM0cx.js.map
