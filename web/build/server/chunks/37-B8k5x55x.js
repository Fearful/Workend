import { f as fail, r as redirect } from './index-BkmUvga9.js';
import { a as apiFetch } from './api-chcxdA-p.js';

const SESSION_COOKIE = "workend_session";
const load = async ({ locals }) => {
  if (!locals.user) throw redirect(303, "/login");
  return {};
};
const actions = {
  default: async ({ request, cookies }) => {
    const data = await request.formData();
    const name = String(data.get("name") || "").trim();
    const description = String(data.get("description") || "").trim();
    if (!name) {
      return fail(400, { name, description, error: "name required" });
    }
    const token = cookies.get(SESSION_COOKIE);
    const result = await apiFetch("/api/workspaces", {
      method: "POST",
      body: { name, description },
      cookie: token ? `${SESSION_COOKIE}=${token}` : void 0
    });
    if (!result.ok) {
      return fail(result.status, { name, description, error: result.error || "create failed" });
    }
    throw redirect(303, "/");
  }
};

var _page_server_ts = /*#__PURE__*/Object.freeze({
  __proto__: null,
  actions: actions,
  load: load
});

const index = 37;
let component_cache;
const component = async () => component_cache ??= (await import('./_page.svelte-BLH2r-I3.js')).default;
const server_id = "src/routes/workspaces/new/+page.server.ts";
const imports = ["_app/immutable/nodes/37.Dk1XDMx2.js","_app/immutable/chunks/CWj6FrbW.js","_app/immutable/chunks/BCGREYIq.js","_app/immutable/chunks/glOr0qzD.js","_app/immutable/chunks/D8xXPBaV.js","_app/immutable/chunks/CT9QbQKf.js","_app/immutable/chunks/BS2x50Pa.js","_app/immutable/chunks/wpI5rr64.js","_app/immutable/chunks/Bq1FU2Pv.js"];
const stylesheets = ["_app/immutable/assets/PageHeader.Ct_CmObI.css","_app/immutable/assets/37._Y1PPEyG.css"];
const fonts = [];

export { component, fonts, imports, index, _page_server_ts as server, server_id, stylesheets };
//# sourceMappingURL=37-B8k5x55x.js.map
