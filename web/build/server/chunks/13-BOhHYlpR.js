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

const index = 13;
let component_cache;
const component = async () => component_cache ??= (await import('./_page.svelte-BzWNDfJ5.js')).default;
const server_id = "src/routes/projects/[id]/board/+page.server.ts";
const imports = ["_app/immutable/nodes/13.BwWbPEeC.js","_app/immutable/chunks/C4JmIpy3.js","_app/immutable/chunks/apxFJXnE.js","_app/immutable/chunks/CvynrFVy.js","_app/immutable/chunks/CvI2x4sI.js","_app/immutable/chunks/Bn0VCKPQ.js","_app/immutable/chunks/cLUGwKvc.js","_app/immutable/chunks/ChBkQGhY.js","_app/immutable/chunks/Dx6EV4QR.js","_app/immutable/chunks/CYuiovvJ.js","_app/immutable/chunks/CgWWeIEW.js","_app/immutable/chunks/Dd7pVJ5Y.js","_app/immutable/chunks/8aH_EF2d.js","_app/immutable/chunks/BZxKEKJE.js","_app/immutable/chunks/BiQCKUjL.js","_app/immutable/chunks/Cxf1vVyU.js"];
const stylesheets = ["_app/immutable/assets/Panel.Dd2v_-2B.css","_app/immutable/assets/13.8LCJYowV.css"];
const fonts = [];

export { component, fonts, imports, index, _page_server_ts as server, server_id, stylesheets };
//# sourceMappingURL=13-BOhHYlpR.js.map
