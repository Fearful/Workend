import { r as redirect, e as error } from './index-BkmUvga9.js';
import { a as apiFetch } from './api-chcxdA-p.js';

const SESSION_COOKIE = "workend_session";
const load = async ({ params, locals, cookies }) => {
  if (!locals.user) throw redirect(303, "/login");
  const cookie = cookies.get(SESSION_COOKIE);
  const result = await apiFetch(
    `/api/runs/${params.id}/compare?to=${encodeURIComponent(params.other)}`,
    { cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : void 0 }
  );
  if (!result.ok || !result.data) throw error(result.status, result.error || "compare failed");
  return result.data;
};

var _page_server_ts = /*#__PURE__*/Object.freeze({
  __proto__: null,
  load: load
});

const index = 25;
let component_cache;
const component = async () => component_cache ??= (await import('./_page.svelte-BVpaetxH.js')).default;
const server_id = "src/routes/runs/[id]/compare/[other]/+page.server.ts";
const imports = ["_app/immutable/nodes/25.BOSskJoq.js","_app/immutable/chunks/CWj6FrbW.js","_app/immutable/chunks/BCGREYIq.js","_app/immutable/chunks/BcfnGdFX.js","_app/immutable/chunks/BS2x50Pa.js","_app/immutable/chunks/glOr0qzD.js","_app/immutable/chunks/BoB1CPyg.js","_app/immutable/chunks/Bx7T9aDp.js","_app/immutable/chunks/D8xXPBaV.js","_app/immutable/chunks/CT9QbQKf.js","_app/immutable/chunks/vbEortrD.js","_app/immutable/chunks/Bq1FU2Pv.js"];
const stylesheets = ["_app/immutable/assets/Breadcrumb.BmuUfcGr.css","_app/immutable/assets/PageHeader.Ct_CmObI.css","_app/immutable/assets/25.hCPyTSHD.css"];
const fonts = [];

export { component, fonts, imports, index, _page_server_ts as server, server_id, stylesheets };
//# sourceMappingURL=25-fy7zB2mu.js.map
