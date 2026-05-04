import { r as redirect, e as error } from './index-BkmUvga9.js';
import { a as apiFetch } from './api-chcxdA-p.js';

const SESSION_COOKIE = "workend_session";
const load = async ({ params, locals, cookies }) => {
  if (!locals.user) throw redirect(303, "/login");
  const cookie = cookies.get(SESSION_COOKIE);
  const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
  const projResult = await apiFetch(`/api/projects/${params.id}`, { cookie: cookieHeader });
  if (projResult.status === 404) throw error(404, "project not found");
  if (!projResult.ok || !projResult.data) throw error(500, projResult.error || "failed to load");
  const wsResult = await apiFetch(
    `/api/workspaces/${projResult.data.workspace_id}`,
    { cookie: cookieHeader }
  );
  return {
    project: projResult.data,
    workspace: wsResult.ok ? wsResult.data ?? null : null
  };
};

var _layout_server_ts = /*#__PURE__*/Object.freeze({
  __proto__: null,
  load: load
});

const index = 2;
let component_cache;
const component = async () => component_cache ??= (await import('./_layout.svelte-CUy9vLBM.js')).default;
const server_id = "src/routes/projects/[id]/+layout.server.ts";
const imports = ["_app/immutable/nodes/2.-ik0hqHR.js","_app/immutable/chunks/C4JmIpy3.js","_app/immutable/chunks/apxFJXnE.js","_app/immutable/chunks/CvynrFVy.js","_app/immutable/chunks/Dx6EV4QR.js","_app/immutable/chunks/CvI2x4sI.js","_app/immutable/chunks/CgWWeIEW.js","_app/immutable/chunks/Dd7pVJ5Y.js","_app/immutable/chunks/8aH_EF2d.js","_app/immutable/chunks/VBdYAuAB.js","_app/immutable/chunks/Bn0VCKPQ.js","_app/immutable/chunks/cLUGwKvc.js","_app/immutable/chunks/ChBkQGhY.js","_app/immutable/chunks/059uK4jn.js","_app/immutable/chunks/BvqIPo9c.js","_app/immutable/chunks/CYuiovvJ.js","_app/immutable/chunks/BiQCKUjL.js","_app/immutable/chunks/m1uqkJwC.js","_app/immutable/chunks/Cxf1vVyU.js"];
const stylesheets = ["_app/immutable/assets/Breadcrumb.BmuUfcGr.css","_app/immutable/assets/Modal.CHlCiKdq.css","_app/immutable/assets/2.CbY1DSR2.css"];
const fonts = [];

export { component, fonts, imports, index, _layout_server_ts as server, server_id, stylesheets };
//# sourceMappingURL=2-BH813f9E.js.map
