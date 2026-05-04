import { f as fail, r as redirect, e as error } from './index-BkmUvga9.js';
import { a as apiFetch } from './api-chcxdA-p.js';

const SESSION_COOKIE = "workend_session";
const load = async ({ params, locals, cookies }) => {
  if (!locals.user) throw redirect(303, "/login");
  const cookie = cookies.get(SESSION_COOKIE);
  const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
  const r = await apiFetch(`/api/issues/${params.id}`, { cookie: cookieHeader });
  if (r.status === 404) throw error(404, "issue not found");
  if (!r.ok || !r.data) throw error(500, r.error || "failed to load");
  return { issue: r.data };
};
const actions = {
  comment: async ({ params, request, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
    const data = await request.formData();
    const body = String(data.get("body") || "").trim();
    if (!body) return fail(400, { commentError: "body required", draft: body });
    const r = await apiFetch(`/api/issues/${params.id}/comments`, {
      method: "POST",
      body: { body },
      cookie: cookieHeader
    });
    if (!r.ok) return fail(r.status, { commentError: r.error || "comment failed", draft: body });
    return { commented: true };
  },
  close: async ({ params, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
    const r = await apiFetch(`/api/issues/${params.id}/close`, {
      method: "POST",
      cookie: cookieHeader
    });
    if (!r.ok) return fail(r.status, { closeError: r.error || "close failed" });
    return { closed: true };
  }
};

var _page_server_ts = /*#__PURE__*/Object.freeze({
  __proto__: null,
  actions: actions,
  load: load
});

const index = 7;
let component_cache;
const component = async () => component_cache ??= (await import('./_page.svelte-D1AVicPM.js')).default;
const server_id = "src/routes/issues/[id]/+page.server.ts";
const imports = ["_app/immutable/nodes/7.BYJEXiFz.js","_app/immutable/chunks/C4JmIpy3.js","_app/immutable/chunks/apxFJXnE.js","_app/immutable/chunks/CvynrFVy.js","_app/immutable/chunks/CvI2x4sI.js","_app/immutable/chunks/Bn0VCKPQ.js","_app/immutable/chunks/cLUGwKvc.js","_app/immutable/chunks/Dmb8Bvf1.js","_app/immutable/chunks/8aH_EF2d.js","_app/immutable/chunks/BZxKEKJE.js","_app/immutable/chunks/Dx6EV4QR.js","_app/immutable/chunks/ChBkQGhY.js","_app/immutable/chunks/BiQCKUjL.js","_app/immutable/chunks/BO09PiA2.js","_app/immutable/chunks/Cxf1vVyU.js"];
const stylesheets = ["_app/immutable/assets/Panel.Dd2v_-2B.css","_app/immutable/assets/Badge.Cj2_Wv2F.css","_app/immutable/assets/7.Bh465XZN.css"];
const fonts = [];

export { component, fonts, imports, index, _page_server_ts as server, server_id, stylesheets };
//# sourceMappingURL=7-B0Pu7-mx.js.map
