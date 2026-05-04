import { f as fail, r as redirect } from './index-BkmUvga9.js';
import { a as apiFetch } from './api-chcxdA-p.js';

const SESSION_COOKIE = "workend_session";
const load = async ({ locals, cookies, url }) => {
  if (!locals.user) throw redirect(303, "/login");
  const cookie = cookies.get(SESSION_COOKIE);
  const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
  const [connResult, sshResult, patResult] = await Promise.all([
    apiFetch("/api/me/connections", { cookie: cookieHeader }),
    apiFetch("/api/me/ssh-keys", { cookie: cookieHeader }),
    apiFetch("/api/me/pat-credentials", { cookie: cookieHeader })
  ]);
  const providersConfigured = connResult.status !== 404;
  return {
    connections: connResult.ok ? connResult.data ?? [] : [],
    providersConfigured,
    sshKeys: sshResult.ok ? sshResult.data ?? [] : [],
    patCredentials: patResult.ok ? patResult.data ?? [] : [],
    flash: {
      connected: url.searchParams.get("connected"),
      error: url.searchParams.get("conn_error")
    }
  };
};
const actions = {
  disconnect: async ({ request, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const id = String((await request.formData()).get("id") || "");
    if (!id) return fail(400, { error: "id required" });
    const result = await apiFetch(`/api/me/connections/${id}`, {
      method: "DELETE",
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : void 0
    });
    if (!result.ok) return fail(result.status, { error: result.error || "disconnect failed" });
    return { disconnected: true };
  },
  addSSHKey: async ({ request, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const data = await request.formData();
    const name = String(data.get("name") || "").trim();
    const privateKey = String(data.get("private_key") || "");
    if (!name || privateKey.length < 50) {
      return fail(400, { sshError: "name and a PEM private key are required", sshName: name });
    }
    const result = await apiFetch("/api/me/ssh-keys", {
      method: "POST",
      body: { name, private_key: privateKey },
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : void 0
    });
    if (!result.ok) return fail(result.status, { sshError: result.error || "failed", sshName: name });
    return { sshAdded: true };
  },
  deleteSSHKey: async ({ request, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const id = String((await request.formData()).get("id") || "");
    if (!id) return fail(400, { sshError: "id required" });
    const result = await apiFetch(`/api/me/ssh-keys/${id}`, {
      method: "DELETE",
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : void 0
    });
    if (!result.ok) return fail(result.status, { sshError: result.error || "delete failed" });
    return { sshDeleted: true };
  },
  addPAT: async ({ request, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const data = await request.formData();
    const host = String(data.get("host") || "").trim();
    const label = String(data.get("label") || "").trim();
    const token = String(data.get("token") || "").trim();
    if (!host || !label || token.length < 8) {
      return fail(400, {
        patError: "host, label and a token (8+ chars) are required",
        patHost: host,
        patLabel: label
      });
    }
    const result = await apiFetch("/api/me/pat-credentials", {
      method: "POST",
      body: { host, label, token },
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : void 0
    });
    if (!result.ok) {
      return fail(result.status, {
        patError: result.error || "failed",
        patHost: host,
        patLabel: label
      });
    }
    return { patAdded: true };
  },
  deletePAT: async ({ request, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const id = String((await request.formData()).get("id") || "");
    if (!id) return fail(400, { patError: "id required" });
    const result = await apiFetch(`/api/me/pat-credentials/${id}`, {
      method: "DELETE",
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : void 0
    });
    if (!result.ok) return fail(result.status, { patError: result.error || "delete failed" });
    return { patDeleted: true };
  }
};

var _page_server_ts = /*#__PURE__*/Object.freeze({
  __proto__: null,
  actions: actions,
  load: load
});

const index = 21;
let component_cache;
const component = async () => component_cache ??= (await import('./_page.svelte-fKZXz85p.js')).default;
const server_id = "src/routes/settings/+page.server.ts";
const imports = ["_app/immutable/nodes/21.mayRjJjU.js","_app/immutable/chunks/C4JmIpy3.js","_app/immutable/chunks/apxFJXnE.js","_app/immutable/chunks/CvynrFVy.js","_app/immutable/chunks/CvI2x4sI.js","_app/immutable/chunks/Bn0VCKPQ.js","_app/immutable/chunks/cLUGwKvc.js","_app/immutable/chunks/DC4q0MGt.js","_app/immutable/chunks/Dx6EV4QR.js","_app/immutable/chunks/BZxKEKJE.js","_app/immutable/chunks/ChBkQGhY.js","_app/immutable/chunks/BiQCKUjL.js","_app/immutable/chunks/BO09PiA2.js","_app/immutable/chunks/Cxf1vVyU.js"];
const stylesheets = ["_app/immutable/assets/PageHeader.Ct_CmObI.css","_app/immutable/assets/Panel.Dd2v_-2B.css","_app/immutable/assets/Badge.Cj2_Wv2F.css","_app/immutable/assets/21.DI48QSgC.css"];
const fonts = [];

export { component, fonts, imports, index, _page_server_ts as server, server_id, stylesheets };
//# sourceMappingURL=21-AhDar-Ro.js.map
