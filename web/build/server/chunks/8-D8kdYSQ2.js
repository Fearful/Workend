import { f as fail, r as redirect } from './index-BkmUvga9.js';
import { a as apiFetch, p as parseSetCookie } from './api-chcxdA-p.js';

const load = async ({ locals }) => {
  if (locals.user) throw redirect(303, "/");
  const authCfg = await apiFetch("/api/auth/config");
  return {
    oidcConfigured: authCfg.ok ? authCfg.data?.oidc_configured ?? false : false,
    oidcProviderName: authCfg.ok ? authCfg.data?.oidc_provider_name ?? "SSO" : "SSO"
  };
};
const actions = {
  default: async ({ request, cookies }) => {
    const data = await request.formData();
    const email = String(data.get("email") || "");
    const password = String(data.get("password") || "");
    if (!email || !password) {
      return fail(400, { email, error: "email and password required" });
    }
    const result = await apiFetch("/api/auth/login", {
      method: "POST",
      body: { email, password }
    });
    if (!result.ok) {
      return fail(result.status, { email, error: result.error || "login failed" });
    }
    if (result.setCookie) {
      const parsed = parseSetCookie(result.setCookie);
      if (parsed) cookies.set(parsed.name, parsed.value, parsed.options);
    }
    throw redirect(303, "/");
  }
};

var _page_server_ts = /*#__PURE__*/Object.freeze({
  __proto__: null,
  actions: actions,
  load: load
});

const index = 8;
let component_cache;
const component = async () => component_cache ??= (await import('./_page.svelte-BzLvKsqo.js')).default;
const server_id = "src/routes/login/+page.server.ts";
const imports = ["_app/immutable/nodes/8.gMpUN1DL.js","_app/immutable/chunks/CWj6FrbW.js","_app/immutable/chunks/BCGREYIq.js","_app/immutable/chunks/glOr0qzD.js","_app/immutable/chunks/D8xXPBaV.js","_app/immutable/chunks/CT9QbQKf.js","_app/immutable/chunks/BS2x50Pa.js","_app/immutable/chunks/wpI5rr64.js","_app/immutable/chunks/Bq1FU2Pv.js"];
const stylesheets = ["_app/immutable/assets/PageHeader.Ct_CmObI.css","_app/immutable/assets/8.DvihiADC.css"];
const fonts = [];

export { component, fonts, imports, index, _page_server_ts as server, server_id, stylesheets };
//# sourceMappingURL=8-D8kdYSQ2.js.map
