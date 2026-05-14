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

const index = 17;
let component_cache;
const component = async () => component_cache ??= (await import('./_page.svelte-IcRxDYFO.js')).default;
const server_id = "src/routes/projects/[id]/images/+page.server.ts";
const imports = ["_app/immutable/nodes/17.Di9fS7sB.js","_app/immutable/chunks/CWj6FrbW.js","_app/immutable/chunks/BCGREYIq.js","_app/immutable/chunks/glOr0qzD.js","_app/immutable/chunks/BcfnGdFX.js","_app/immutable/chunks/D8xXPBaV.js","_app/immutable/chunks/vbEortrD.js","_app/immutable/chunks/BS2x50Pa.js","_app/immutable/chunks/BoB1CPyg.js","_app/immutable/chunks/YSzAdjSj.js","_app/immutable/chunks/B8IgtE21.js","_app/immutable/chunks/Bq1FU2Pv.js","_app/immutable/chunks/B5rGamLH.js","_app/immutable/chunks/DZ5gTY8C.js","_app/immutable/chunks/Di6_auBm.js","_app/immutable/chunks/B2UVO_nM.js"];
const stylesheets = ["_app/immutable/assets/EmptyState.ByauYn1J.css","_app/immutable/assets/SectionHeader.Bb9ZRXXP.css","_app/immutable/assets/Tooltip.BrYf-THh.css","_app/immutable/assets/TimeAgo.BYfizd1M.css","_app/immutable/assets/17.DwCgkU9H.css"];
const fonts = [];

export { component, fonts, imports, index, _page_server_ts as server, server_id, stylesheets };
//# sourceMappingURL=17-DKoaDdN8.js.map
