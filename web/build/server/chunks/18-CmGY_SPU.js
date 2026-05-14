import { f as fail, r as redirect } from './index-BkmUvga9.js';
import { a as apiFetch } from './api-chcxdA-p.js';

const SESSION_COOKIE = "workend_session";
const load = async ({ params, locals, cookies }) => {
  if (!locals.user) throw redirect(303, "/login");
  const cookie = cookies.get(SESSION_COOKIE);
  const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
  const [packagesResult, tasksResult] = await Promise.all([
    apiFetch(
      `/api/projects/${params.id}/monorepo/packages`,
      { cookie: cookieHeader }
    ),
    apiFetch(
      `/api/projects/${params.id}/tasks`,
      { cookie: cookieHeader }
    )
  ]);
  return {
    packages: packagesResult.ok ? packagesResult.data ?? [] : [],
    tasks: tasksResult.ok ? tasksResult.data ?? [] : []
  };
};
const actions = {
  detectPackages: async ({ params, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
    const result = await apiFetch(
      `/api/projects/${params.id}/monorepo/detect`,
      { method: "POST", cookie: cookieHeader }
    );
    if (!result.ok) return fail(result.status, { error: result.error || "detection failed" });
    return { detected: result.data?.detected ?? 0 };
  },
  autoMap: async ({ params, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
    const result = await apiFetch(
      `/api/projects/${params.id}/monorepo/auto-map`,
      { method: "POST", cookie: cookieHeader }
    );
    if (!result.ok) return fail(result.status, { error: result.error || "auto-map failed" });
    return { mapped: result.data?.mapped ?? 0 };
  }
};

var _page_server_ts = /*#__PURE__*/Object.freeze({
  __proto__: null,
  actions: actions,
  load: load
});

const index = 18;
let component_cache;
const component = async () => component_cache ??= (await import('./_page.svelte-BGC4cwq3.js')).default;
const server_id = "src/routes/projects/[id]/monorepo/+page.server.ts";
const imports = ["_app/immutable/nodes/18.BsX2qSBf.js","_app/immutable/chunks/CWj6FrbW.js","_app/immutable/chunks/BCGREYIq.js","_app/immutable/chunks/glOr0qzD.js","_app/immutable/chunks/BcfnGdFX.js","_app/immutable/chunks/DuSA90i8.js","_app/immutable/chunks/Di6_auBm.js","_app/immutable/chunks/DdjgLl7P.js","_app/immutable/chunks/CHpGRhK0.js","_app/immutable/chunks/BS2x50Pa.js","_app/immutable/chunks/Bq1FU2Pv.js","_app/immutable/chunks/ilPNOASy.js","_app/immutable/chunks/D8xXPBaV.js","_app/immutable/chunks/vbEortrD.js","_app/immutable/chunks/wpI5rr64.js"];
const stylesheets = ["_app/immutable/assets/Badge.Cj2_Wv2F.css","_app/immutable/assets/Modal.CHlCiKdq.css","_app/immutable/assets/18.Dnd9vBfP.css","_app/immutable/assets/Panel.qbNo31SY.css"];
const fonts = [];

export { component, fonts, imports, index, _page_server_ts as server, server_id, stylesheets };
//# sourceMappingURL=18-CmGY_SPU.js.map
