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
const component = async () => component_cache ??= (await import('./_page.svelte-C5RQbsse.js')).default;
const server_id = "src/routes/login/+page.server.ts";
const imports = ["_app/immutable/nodes/8.D-wk3Dv6.js","_app/immutable/chunks/C4JmIpy3.js","_app/immutable/chunks/apxFJXnE.js","_app/immutable/chunks/CvynrFVy.js","_app/immutable/chunks/CvI2x4sI.js","_app/immutable/chunks/cLUGwKvc.js","_app/immutable/chunks/DC4q0MGt.js","_app/immutable/chunks/Dx6EV4QR.js","_app/immutable/chunks/Cxf1vVyU.js","_app/immutable/chunks/ChBkQGhY.js","_app/immutable/chunks/BiQCKUjL.js"];
const stylesheets = ["_app/immutable/assets/PageHeader.Ct_CmObI.css","_app/immutable/assets/8.DvihiADC.css"];
const fonts = [];

export { component, fonts, imports, index, _page_server_ts as server, server_id, stylesheets };
//# sourceMappingURL=8-CKu3S1Gr.js.map
