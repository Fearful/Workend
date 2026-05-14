import { r as redirect, e as error } from './index-BkmUvga9.js';
import { a as apiFetch } from './api-chcxdA-p.js';

const SESSION_COOKIE = "workend_session";
const load = async ({ params, url, locals, cookies }) => {
  if (!locals.user) throw redirect(303, "/login");
  const cookie = cookies.get(SESSION_COOKIE);
  const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
  const before = url.searchParams.get("before") || "";
  const limit = 50;
  const feedUrl = before ? `/api/workspaces/${params.id}/feed?limit=${limit}&before=${before}` : `/api/workspaces/${params.id}/feed?limit=${limit}`;
  const [wsResp, feedResp] = await Promise.all([
    apiFetch(`/api/workspaces/${params.id}`, { cookie: cookieHeader }),
    apiFetch(feedUrl, { cookie: cookieHeader })
  ]);
  if (wsResp.status === 404) throw error(404, "workspace not found");
  if (!wsResp.ok || !wsResp.data) throw error(500, wsResp.error || "failed to load workspace");
  const events = feedResp.ok ? feedResp.data ?? [] : [];
  const nextCursor = events.length >= limit ? String(events[events.length - 1].id) : null;
  return {
    workspace: wsResp.data,
    events,
    feedError: feedResp.ok ? null : feedResp.error || "failed to load activity feed",
    nextCursor,
    currentBefore: before || null
  };
};

var _page_server_ts = /*#__PURE__*/Object.freeze({
  __proto__: null,
  load: load
});

const index = 32;
let component_cache;
const component = async () => component_cache ??= (await import('./_page.svelte-D81iNYHs.js')).default;
const server_id = "src/routes/workspaces/[id]/activity/+page.server.ts";
const imports = ["_app/immutable/nodes/32.vAFmEN8q.js","_app/immutable/chunks/CWj6FrbW.js","_app/immutable/chunks/BCGREYIq.js","_app/immutable/chunks/glOr0qzD.js","_app/immutable/chunks/BcfnGdFX.js","_app/immutable/chunks/D8xXPBaV.js","_app/immutable/chunks/Bx7T9aDp.js","_app/immutable/chunks/YSzAdjSj.js","_app/immutable/chunks/BS2x50Pa.js","_app/immutable/chunks/wpI5rr64.js","_app/immutable/chunks/Bq1FU2Pv.js","_app/immutable/chunks/CHpGRhK0.js","_app/immutable/chunks/B5rGamLH.js","_app/immutable/chunks/DZ5gTY8C.js","_app/immutable/chunks/Di6_auBm.js","_app/immutable/chunks/vbEortrD.js","_app/immutable/chunks/B2UVO_nM.js"];
const stylesheets = ["_app/immutable/assets/Breadcrumb.BmuUfcGr.css","_app/immutable/assets/EmptyState.ByauYn1J.css","_app/immutable/assets/Badge.Cj2_Wv2F.css","_app/immutable/assets/Tooltip.BrYf-THh.css","_app/immutable/assets/TimeAgo.BYfizd1M.css","_app/immutable/assets/32.BfB2o6re.css"];
const fonts = [];

export { component, fonts, imports, index, _page_server_ts as server, server_id, stylesheets };
//# sourceMappingURL=32-28IPNEKV.js.map
