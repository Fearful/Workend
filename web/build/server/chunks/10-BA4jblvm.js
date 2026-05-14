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
const component = async () => component_cache ??= (await import('./_page.svelte-BB_RYlbh.js')).default;
const server_id = "src/routes/mentions/+page.server.ts";
const imports = ["_app/immutable/nodes/10.JitZ3q-T.js","_app/immutable/chunks/CWj6FrbW.js","_app/immutable/chunks/BCGREYIq.js","_app/immutable/chunks/glOr0qzD.js","_app/immutable/chunks/BcfnGdFX.js","_app/immutable/chunks/D8xXPBaV.js","_app/immutable/chunks/BS2x50Pa.js","_app/immutable/chunks/BoB1CPyg.js","_app/immutable/chunks/CT9QbQKf.js","_app/immutable/chunks/YSzAdjSj.js"];
const stylesheets = ["_app/immutable/assets/PageHeader.Ct_CmObI.css","_app/immutable/assets/EmptyState.ByauYn1J.css","_app/immutable/assets/10.BxgXEenA.css"];
const fonts = [];

export { component, fonts, imports, index, _page_server_ts as server, server_id, stylesheets };
//# sourceMappingURL=10-BA4jblvm.js.map
