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

const index = 21;
let component_cache;
const component = async () => component_cache ??= (await import('./_page.svelte-CIi-93Wc.js')).default;
const server_id = "src/routes/projects/[id]/runs/+page.server.ts";
const imports = ["_app/immutable/nodes/21.Bll1q6W8.js","_app/immutable/chunks/CWj6FrbW.js","_app/immutable/chunks/BCGREYIq.js","_app/immutable/chunks/glOr0qzD.js","_app/immutable/chunks/BcfnGdFX.js","_app/immutable/chunks/D8xXPBaV.js","_app/immutable/chunks/BS2x50Pa.js","_app/immutable/chunks/DuSA90i8.js","_app/immutable/chunks/Di6_auBm.js","_app/immutable/chunks/DdjgLl7P.js","_app/immutable/chunks/CxpEhr-0.js","_app/immutable/chunks/BoB1CPyg.js","_app/immutable/chunks/YSzAdjSj.js","_app/immutable/chunks/Cz5G4ddd.js","_app/immutable/chunks/Bq1FU2Pv.js","_app/immutable/chunks/B5rGamLH.js","_app/immutable/chunks/DZ5gTY8C.js","_app/immutable/chunks/vbEortrD.js","_app/immutable/chunks/B2UVO_nM.js","_app/immutable/chunks/B8IgtE21.js"];
const stylesheets = ["_app/immutable/assets/EmptyState.ByauYn1J.css","_app/immutable/assets/StatusPill.DBhu1QxG.css","_app/immutable/assets/Tooltip.BrYf-THh.css","_app/immutable/assets/TimeAgo.BYfizd1M.css","_app/immutable/assets/SectionHeader.Bb9ZRXXP.css","_app/immutable/assets/21.CsVL3I7B.css"];
const fonts = [];

export { component, fonts, imports, index, _page_server_ts as server, server_id, stylesheets };
//# sourceMappingURL=21-ef2cMoIL.js.map
