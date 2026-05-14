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
const component = async () => component_cache ??= (await import('./_layout.svelte-CaDVYT2z.js')).default;
const server_id = "src/routes/projects/[id]/+layout.server.ts";
const imports = ["_app/immutable/nodes/2.D_-oxuV4.js","_app/immutable/chunks/CWj6FrbW.js","_app/immutable/chunks/BCGREYIq.js","_app/immutable/chunks/BS2x50Pa.js","_app/immutable/chunks/glOr0qzD.js","_app/immutable/chunks/D8xXPBaV.js","_app/immutable/chunks/DuSA90i8.js","_app/immutable/chunks/Di6_auBm.js","_app/immutable/chunks/DdjgLl7P.js","_app/immutable/chunks/BoB1CPyg.js","_app/immutable/chunks/Bx7T9aDp.js","_app/immutable/chunks/BcfnGdFX.js","_app/immutable/chunks/CxpEhr-0.js","_app/immutable/chunks/Cz5G4ddd.js","_app/immutable/chunks/Bq1FU2Pv.js","_app/immutable/chunks/B5rGamLH.js","_app/immutable/chunks/DZ5gTY8C.js","_app/immutable/chunks/vbEortrD.js","_app/immutable/chunks/B2UVO_nM.js","_app/immutable/chunks/ilPNOASy.js","_app/immutable/chunks/wpI5rr64.js"];
const stylesheets = ["_app/immutable/assets/Breadcrumb.BmuUfcGr.css","_app/immutable/assets/StatusPill.DBhu1QxG.css","_app/immutable/assets/Tooltip.BrYf-THh.css","_app/immutable/assets/TimeAgo.BYfizd1M.css","_app/immutable/assets/Modal.CHlCiKdq.css","_app/immutable/assets/2.DW8Ekywe.css"];
const fonts = [];

export { component, fonts, imports, index, _layout_server_ts as server, server_id, stylesheets };
//# sourceMappingURL=2-CZ6qpr9O.js.map
