import { f as fail, r as redirect } from './index-BkmUvga9.js';
import { a as apiFetch } from './api-chcxdA-p.js';

const SESSION_COOKIE = "workend_session";
const load = async ({ params, locals, cookies }) => {
  if (!locals.user) throw redirect(303, "/login");
  const cookie = cookies.get(SESSION_COOKIE);
  const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
  const blameResult = await apiFetch(
    `/api/projects/${params.id}/blame-timeline?limit=50`,
    { cookie: cookieHeader }
  );
  return {
    blameTimeline: blameResult.ok ? blameResult.data ?? [] : []
  };
};
const actions = {
  syncCommits: async ({ params, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
    const result = await apiFetch(
      `/api/projects/${params.id}/sync-commits`,
      { method: "POST", cookie: cookieHeader }
    );
    if (!result.ok) return fail(result.status, { error: result.error || "commit sync failed" });
    return { synced: result.data?.synced ?? 0, inserted: result.data?.inserted ?? 0 };
  }
};

var _page_server_ts = /*#__PURE__*/Object.freeze({
  __proto__: null,
  actions: actions,
  load: load
});

const index = 14;
let component_cache;
const component = async () => component_cache ??= (await import('./_page.svelte-DVGRtZDY.js')).default;
const server_id = "src/routes/projects/[id]/blame/+page.server.ts";
const imports = ["_app/immutable/nodes/14.DDHsI_YY.js","_app/immutable/chunks/CWj6FrbW.js","_app/immutable/chunks/BCGREYIq.js","_app/immutable/chunks/glOr0qzD.js","_app/immutable/chunks/BcfnGdFX.js","_app/immutable/chunks/D8xXPBaV.js","_app/immutable/chunks/BS2x50Pa.js","_app/immutable/chunks/DuSA90i8.js","_app/immutable/chunks/Di6_auBm.js","_app/immutable/chunks/DdjgLl7P.js","_app/immutable/chunks/Cz5G4ddd.js","_app/immutable/chunks/Bq1FU2Pv.js","_app/immutable/chunks/ilPNOASy.js","_app/immutable/chunks/vbEortrD.js","_app/immutable/chunks/B5rGamLH.js","_app/immutable/chunks/DZ5gTY8C.js","_app/immutable/chunks/B2UVO_nM.js","_app/immutable/chunks/wpI5rr64.js","_app/immutable/chunks/BoB1CPyg.js"];
const stylesheets = ["_app/immutable/assets/StatusPill.DBhu1QxG.css","_app/immutable/assets/Modal.CHlCiKdq.css","_app/immutable/assets/Tooltip.BrYf-THh.css","_app/immutable/assets/TimeAgo.BYfizd1M.css","_app/immutable/assets/14.NdXPGzkd.css","_app/immutable/assets/Panel.qbNo31SY.css"];
const fonts = [];

export { component, fonts, imports, index, _page_server_ts as server, server_id, stylesheets };
//# sourceMappingURL=14-CLP2_SfO.js.map
