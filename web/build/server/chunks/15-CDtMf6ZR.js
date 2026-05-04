import { r as redirect } from './index-BkmUvga9.js';
import { a as apiFetch } from './api-chcxdA-p.js';

const SESSION_COOKIE = "workend_session";
const load = async ({ params, locals, cookies }) => {
  if (!locals.user) throw redirect(303, "/login");
  const cookie = cookies.get(SESSION_COOKIE);
  const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
  const imagesResult = await apiFetch(`/api/projects/${params.id}/images`, { cookie: cookieHeader });
  return {
    images: imagesResult.ok ? imagesResult.data ?? [] : []
  };
};

var _page_server_ts = /*#__PURE__*/Object.freeze({
  __proto__: null,
  load: load
});

const index = 15;
let component_cache;
const component = async () => component_cache ??= (await import('./_page.svelte-2OjSY747.js')).default;
const server_id = "src/routes/projects/[id]/images/+page.server.ts";
const imports = ["_app/immutable/nodes/15.nrqvQV8i.js","_app/immutable/chunks/C4JmIpy3.js","_app/immutable/chunks/apxFJXnE.js","_app/immutable/chunks/CvynrFVy.js","_app/immutable/chunks/CvI2x4sI.js","_app/immutable/chunks/Bn0VCKPQ.js","_app/immutable/chunks/CYuiovvJ.js","_app/immutable/chunks/Dx6EV4QR.js","_app/immutable/chunks/8aH_EF2d.js","_app/immutable/chunks/Di6J7QjO.js","_app/immutable/chunks/cLUGwKvc.js"];
const stylesheets = ["_app/immutable/assets/EmptyState.ByauYn1J.css","_app/immutable/assets/15.DsAOZ_S2.css"];
const fonts = [];

export { component, fonts, imports, index, _page_server_ts as server, server_id, stylesheets };
//# sourceMappingURL=15-CDtMf6ZR.js.map
