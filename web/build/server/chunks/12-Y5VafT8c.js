import { r as redirect, e as error } from './index-BkmUvga9.js';
import { a as apiFetch } from './api-chcxdA-p.js';

const SESSION_COOKIE = "workend_session";
const load = async ({ params, locals, cookies }) => {
  if (!locals.user) throw redirect(303, "/login");
  const cookie = cookies.get(SESSION_COOKIE);
  const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
  const prResult = await apiFetch(`/api/pipeline-runs/${params.id}`, { cookie: cookieHeader });
  if (prResult.status === 404) throw error(404, "pipeline run not found");
  if (!prResult.ok || !prResult.data) throw error(500, prResult.error || "failed to load");
  let pipeline = null;
  return { pipelineRun: prResult.data, pipeline };
};

var _page_server_ts = /*#__PURE__*/Object.freeze({
  __proto__: null,
  load: load
});

const index = 12;
let component_cache;
const component = async () => component_cache ??= (await import('./_page.svelte-BTB_SzBh.js')).default;
const server_id = "src/routes/pipeline-runs/[id]/+page.server.ts";
const imports = ["_app/immutable/nodes/12.DdHCTVCQ.js","_app/immutable/chunks/CWj6FrbW.js","_app/immutable/chunks/Di6_auBm.js","_app/immutable/chunks/BCGREYIq.js","_app/immutable/chunks/glOr0qzD.js","_app/immutable/chunks/BcfnGdFX.js","_app/immutable/chunks/D8xXPBaV.js","_app/immutable/chunks/DuSA90i8.js","_app/immutable/chunks/DdjgLl7P.js","_app/immutable/chunks/Bx7T9aDp.js","_app/immutable/chunks/Cz5G4ddd.js","_app/immutable/chunks/BS2x50Pa.js","_app/immutable/chunks/Bq1FU2Pv.js","_app/immutable/chunks/B5rGamLH.js","_app/immutable/chunks/DZ5gTY8C.js","_app/immutable/chunks/vbEortrD.js","_app/immutable/chunks/B2UVO_nM.js","_app/immutable/chunks/BoB1CPyg.js"];
const stylesheets = ["_app/immutable/assets/Breadcrumb.BmuUfcGr.css","_app/immutable/assets/StatusPill.DBhu1QxG.css","_app/immutable/assets/Tooltip.BrYf-THh.css","_app/immutable/assets/TimeAgo.BYfizd1M.css","_app/immutable/assets/12.XHB6JYK8.css"];
const fonts = [];

export { component, fonts, imports, index, _page_server_ts as server, server_id, stylesheets };
//# sourceMappingURL=12-Y5VafT8c.js.map
