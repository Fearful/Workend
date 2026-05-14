import { r as redirect } from './index-BkmUvga9.js';
import { a as apiFetch } from './api-chcxdA-p.js';

const SESSION_COOKIE = "workend_session";
const load = async ({ locals, cookies }) => {
  if (!locals.user) throw redirect(303, "/login");
  const token = cookies.get(SESSION_COOKIE);
  const cookieHeader = token ? `${SESSION_COOKIE}=${token}` : void 0;
  const [workspacesResult, runsResult, healthResult, comparisonResult, incidentsResult, storageResult] = await Promise.all([
    apiFetch("/api/workspaces", { cookie: cookieHeader }),
    apiFetch("/api/me/runs", { cookie: cookieHeader }),
    apiFetch("/api/widgets/workspaces/health", { cookie: cookieHeader }),
    apiFetch("/api/widgets/workspaces/comparison?days=7", { cookie: cookieHeader }),
    apiFetch("/api/widgets/workspaces/incidents", { cookie: cookieHeader }),
    apiFetch("/api/widgets/workspaces/storage", { cookie: cookieHeader })
  ]);
  return {
    workspaces: workspacesResult.ok ? workspacesResult.data ?? [] : [],
    recentRuns: runsResult.ok ? runsResult.data ?? [] : [],
    error: workspacesResult.ok ? null : workspacesResult.error || "failed to load workspaces",
    health: healthResult.ok ? healthResult.data ?? [] : [],
    comparison: comparisonResult.ok ? comparisonResult.data ?? [] : [],
    incidents: incidentsResult.ok ? incidentsResult.data ?? [] : [],
    storage: storageResult.ok ? storageResult.data ?? [] : []
  };
};

var _page_server_ts = /*#__PURE__*/Object.freeze({
  __proto__: null,
  load: load
});

const index = 3;
let component_cache;
const component = async () => component_cache ??= (await import('./_page.svelte-_3rbK2cF.js')).default;
const server_id = "src/routes/+page.server.ts";
const imports = ["_app/immutable/nodes/3.BBoYvJCP.js","_app/immutable/chunks/CWj6FrbW.js","_app/immutable/chunks/BCGREYIq.js","_app/immutable/chunks/glOr0qzD.js","_app/immutable/chunks/BcfnGdFX.js","_app/immutable/chunks/D8xXPBaV.js","_app/immutable/chunks/BS2x50Pa.js","_app/immutable/chunks/vbEortrD.js","_app/immutable/chunks/CT9QbQKf.js","_app/immutable/chunks/C-j7aE8P.js","_app/immutable/chunks/Bq1FU2Pv.js","_app/immutable/chunks/CHpGRhK0.js","_app/immutable/chunks/YSzAdjSj.js","_app/immutable/chunks/wpI5rr64.js","_app/immutable/chunks/Cz5G4ddd.js","_app/immutable/chunks/B5rGamLH.js","_app/immutable/chunks/DZ5gTY8C.js","_app/immutable/chunks/Di6_auBm.js","_app/immutable/chunks/B2UVO_nM.js","_app/immutable/chunks/B8IgtE21.js"];
const stylesheets = ["_app/immutable/assets/PageHeader.Ct_CmObI.css","_app/immutable/assets/Panel.qbNo31SY.css","_app/immutable/assets/Badge.Cj2_Wv2F.css","_app/immutable/assets/EmptyState.ByauYn1J.css","_app/immutable/assets/StatusPill.DBhu1QxG.css","_app/immutable/assets/Tooltip.BrYf-THh.css","_app/immutable/assets/TimeAgo.BYfizd1M.css","_app/immutable/assets/SectionHeader.Bb9ZRXXP.css","_app/immutable/assets/3.Dwh9h2PZ.css"];
const fonts = [];

export { component, fonts, imports, index, _page_server_ts as server, server_id, stylesheets };
//# sourceMappingURL=3-bRwly-vz.js.map
