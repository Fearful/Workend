import { f as fail, r as redirect } from './index-BkmUvga9.js';
import { a as apiFetch } from './api-chcxdA-p.js';

const SESSION_COOKIE = "workend_session";
const load = async ({ params, locals, cookies }) => {
  if (!locals.user) throw redirect(303, "/login");
  const cookie = cookies.get(SESSION_COOKIE);
  const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
  const boardResult = await apiFetch(`/api/projects/${params.id}/board`, { cookie: cookieHeader });
  return {
    boardResp: boardResult.ok ? boardResult.data ?? { configured: false } : { configured: false },
    boardError: boardResult.ok ? null : boardResult.error || null
  };
};
const actions = {
  setup: async ({ params, request, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
    const data = await request.formData();
    const columns = data.getAll("columns").map((v) => String(v)).filter((s) => s.length > 0);
    if (columns.length === 0) return fail(400, { error: "pick at least one label as a column" });
    const r = await apiFetch(
      `/api/projects/${params.id}/board`,
      { method: "POST", body: { columns }, cookie: cookieHeader }
    );
    if (!r.ok) return fail(r.status, { error: r.error || "setup failed" });
    return { ok: true, syncError: r.data?.sync_error };
  },
  sync: async ({ params, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
    const r = await apiFetch(`/api/projects/${params.id}/board/sync`, {
      method: "POST",
      cookie: cookieHeader
    });
    if (!r.ok) return fail(r.status, { error: r.error || "sync failed" });
    return { synced: true };
  }
};

var _page_server_ts = /*#__PURE__*/Object.freeze({
  __proto__: null,
  actions: actions,
  load: load
});

const index = 15;
let component_cache;
const component = async () => component_cache ??= (await import('./_page.svelte-DZ3AL8b7.js')).default;
const server_id = "src/routes/projects/[id]/board/+page.server.ts";
const imports = ["_app/immutable/nodes/15.DCfAA06p.js","_app/immutable/chunks/CWj6FrbW.js","_app/immutable/chunks/BCGREYIq.js","_app/immutable/chunks/glOr0qzD.js","_app/immutable/chunks/BcfnGdFX.js","_app/immutable/chunks/D8xXPBaV.js","_app/immutable/chunks/BS2x50Pa.js","_app/immutable/chunks/vbEortrD.js","_app/immutable/chunks/DuSA90i8.js","_app/immutable/chunks/Di6_auBm.js","_app/immutable/chunks/DdjgLl7P.js","_app/immutable/chunks/C-j7aE8P.js","_app/immutable/chunks/Bq1FU2Pv.js","_app/immutable/chunks/wpI5rr64.js","_app/immutable/chunks/B8IgtE21.js","_app/immutable/chunks/B5rGamLH.js","_app/immutable/chunks/DZ5gTY8C.js","_app/immutable/chunks/B2UVO_nM.js"];
const stylesheets = ["_app/immutable/assets/Panel.qbNo31SY.css","_app/immutable/assets/SectionHeader.Bb9ZRXXP.css","_app/immutable/assets/Tooltip.BrYf-THh.css","_app/immutable/assets/TimeAgo.BYfizd1M.css","_app/immutable/assets/15.Btu-YQeZ.css"];
const fonts = [];

export { component, fonts, imports, index, _page_server_ts as server, server_id, stylesheets };
//# sourceMappingURL=15-BNerK48_.js.map
