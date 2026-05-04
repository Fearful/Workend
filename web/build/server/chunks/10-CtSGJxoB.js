import { f as fail, r as redirect } from './index-BkmUvga9.js';
import { a as apiFetch } from './api-chcxdA-p.js';

const SESSION_COOKIE = "workend_session";
const load = async ({ locals, cookies, url }) => {
  if (!locals.user) throw redirect(303, "/login");
  const cookie = cookies.get(SESSION_COOKIE);
  const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
  const unreadOnly = url.searchParams.get("unread") === "true";
  const r = await apiFetch(
    `/api/me/mentions?limit=100${unreadOnly ? "&unread_only=true" : ""}`,
    { cookie: cookieHeader }
  );
  return {
    mentions: r.ok ? r.data ?? [] : [],
    unreadOnly
  };
};
const actions = {
  markRead: async ({ request, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
    const data = await request.formData();
    const idRaw = data.get("id");
    const body = idRaw ? { id: Number(idRaw) } : {};
    const r = await apiFetch("/api/me/mentions/read", {
      method: "POST",
      body,
      cookie: cookieHeader
    });
    if (!r.ok) return fail(r.status, { error: r.error || "failed" });
    return { ok: true };
  }
};

var _page_server_ts = /*#__PURE__*/Object.freeze({
  __proto__: null,
  actions: actions,
  load: load
});

const index = 10;
let component_cache;
const component = async () => component_cache ??= (await import('./_page.svelte-CxDz2miZ.js')).default;
const server_id = "src/routes/mentions/+page.server.ts";
const imports = ["_app/immutable/nodes/10.nJ6TdmDR.js","_app/immutable/chunks/C4JmIpy3.js","_app/immutable/chunks/apxFJXnE.js","_app/immutable/chunks/CvynrFVy.js","_app/immutable/chunks/CvI2x4sI.js","_app/immutable/chunks/Bn0VCKPQ.js","_app/immutable/chunks/cLUGwKvc.js","_app/immutable/chunks/ChBkQGhY.js","_app/immutable/chunks/Dx6EV4QR.js","_app/immutable/chunks/8aH_EF2d.js","_app/immutable/chunks/DC4q0MGt.js","_app/immutable/chunks/Di6J7QjO.js"];
const stylesheets = ["_app/immutable/assets/PageHeader.Ct_CmObI.css","_app/immutable/assets/EmptyState.ByauYn1J.css","_app/immutable/assets/10.BxgXEenA.css"];
const fonts = [];

export { component, fonts, imports, index, _page_server_ts as server, server_id, stylesheets };
//# sourceMappingURL=10-CtSGJxoB.js.map
