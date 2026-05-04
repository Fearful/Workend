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
    const display_name = String(data.get("display_name") || "");
    if (!email || !password || !display_name) {
      return fail(400, { email, display_name, error: "all fields required" });
    }
    if (password.length < 8) {
      return fail(400, { email, display_name, error: "password must be at least 8 characters" });
    }
    const result = await apiFetch("/api/auth/signup", {
      method: "POST",
      body: { email, password, display_name }
    });
    if (!result.ok) {
      return fail(result.status, { email, display_name, error: result.error || "signup failed" });
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

const index = 23;
let component_cache;
const component = async () => component_cache ??= (await import('./_page.svelte-B9dDU0QF.js')).default;
const server_id = "src/routes/signup/+page.server.ts";
const imports = ["_app/immutable/nodes/23.C5FeMDk3.js","_app/immutable/chunks/C4JmIpy3.js","_app/immutable/chunks/apxFJXnE.js","_app/immutable/chunks/CvynrFVy.js","_app/immutable/chunks/CvI2x4sI.js","_app/immutable/chunks/cLUGwKvc.js","_app/immutable/chunks/DC4q0MGt.js","_app/immutable/chunks/Dx6EV4QR.js","_app/immutable/chunks/Cxf1vVyU.js","_app/immutable/chunks/ChBkQGhY.js","_app/immutable/chunks/BiQCKUjL.js"];
const stylesheets = ["_app/immutable/assets/PageHeader.Ct_CmObI.css","_app/immutable/assets/23.Bt31OwDb.css"];
const fonts = [];

export { component, fonts, imports, index, _page_server_ts as server, server_id, stylesheets };
//# sourceMappingURL=23-77IYp-fB.js.map
