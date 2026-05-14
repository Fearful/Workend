import { f as fail, r as redirect, e as error } from './index-BkmUvga9.js';
import { a as apiFetch } from './api-chcxdA-p.js';

const SESSION_COOKIE = "workend_session";
const load = async ({ params, locals, cookies }) => {
  if (!locals.user) throw redirect(303, "/login");
  const cookie = cookies.get(SESSION_COOKIE);
  const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
  const sandboxResult = await apiFetch(`/api/sandboxes/${params.id}`, { cookie: cookieHeader });
  if (sandboxResult.status === 404) throw error(404, "sandbox not found");
  if (!sandboxResult.ok || !sandboxResult.data) throw error(500, sandboxResult.error || "failed to load");
  const sessionsResult = await apiFetch(
    `/api/sandboxes/${params.id}/shell-sessions`,
    { cookie: cookieHeader }
  );
  const sessions = sessionsResult.ok ? sessionsResult.data ?? [] : [];
  return { sandbox: sandboxResult.data, shellSessions: sessions };
};
const actions = {
  createShell: async ({ params, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
    const r = await apiFetch(`/api/sandboxes/${params.id}/shell`, {
      method: "POST",
      cookie: cookieHeader
    });
    if (!r.ok) return fail(r.status, { error: r.error || "failed to create shell session" });
    return { shellCreated: true, sessionId: r.data?.id };
  },
  exec: async ({ request, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
    const fd = await request.formData();
    const sessionId = String(fd.get("session_id") || "");
    const command = String(fd.get("command") || "").trim();
    if (!sessionId) return fail(400, { error: "session_id required" });
    if (!command) return fail(400, { error: "command required", sessionId });
    const r = await apiFetch(
      `/api/shell-sessions/${sessionId}/exec`,
      {
        method: "POST",
        body: { command },
        cookie: cookieHeader
      }
    );
    if (!r.ok) return fail(r.status, { error: r.error || "exec failed", sessionId });
    return { execResult: r.data, sessionId, command };
  },
  closeShell: async ({ request, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
    const fd = await request.formData();
    const sessionId = String(fd.get("session_id") || "");
    if (!sessionId) return fail(400, { error: "session_id required" });
    const r = await apiFetch(`/api/shell-sessions/${sessionId}/close`, {
      method: "POST",
      cookie: cookieHeader
    });
    if (!r.ok) return fail(r.status, { error: r.error || "failed to close session" });
    return { shellClosed: true };
  },
  extend: async ({ params, request, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
    const fd = await request.formData();
    const hoursRaw = String(fd.get("hours") || "").trim();
    const hours = parseInt(hoursRaw, 10);
    if (!Number.isFinite(hours) || hours < 1 || hours > 72) {
      return fail(400, { error: "hours must be 1-72" });
    }
    const r = await apiFetch(`/api/sandboxes/${params.id}/extend`, {
      method: "POST",
      body: { hours },
      cookie: cookieHeader
    });
    if (!r.ok) return fail(r.status, { error: r.error || "failed to extend" });
    return { extended: true };
  },
  destroy: async ({ params, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
    const r = await apiFetch(`/api/sandboxes/${params.id}/destroy`, {
      method: "POST",
      cookie: cookieHeader
    });
    if (!r.ok) return fail(r.status, { error: r.error || "failed to destroy" });
    throw redirect(303, "/sandboxes");
  }
};

var _page_server_ts = /*#__PURE__*/Object.freeze({
  __proto__: null,
  actions: actions,
  load: load
});

const index = 27;
let component_cache;
const component = async () => component_cache ??= (await import('./_page.svelte-FNkAs7B6.js')).default;
const server_id = "src/routes/sandboxes/[id]/+page.server.ts";
const imports = ["_app/immutable/nodes/27.DWOPQBlL.js","_app/immutable/chunks/CWj6FrbW.js","_app/immutable/chunks/BCGREYIq.js","_app/immutable/chunks/glOr0qzD.js","_app/immutable/chunks/BcfnGdFX.js","_app/immutable/chunks/D8xXPBaV.js","_app/immutable/chunks/BS2x50Pa.js","_app/immutable/chunks/C5qpB67O.js","_app/immutable/chunks/DdjgLl7P.js","_app/immutable/chunks/Di6_auBm.js","_app/immutable/chunks/Bx7T9aDp.js","_app/immutable/chunks/C-j7aE8P.js","_app/immutable/chunks/Bq1FU2Pv.js","_app/immutable/chunks/Cz5G4ddd.js","_app/immutable/chunks/ilPNOASy.js","_app/immutable/chunks/vbEortrD.js","_app/immutable/chunks/B5rGamLH.js","_app/immutable/chunks/DZ5gTY8C.js","_app/immutable/chunks/B2UVO_nM.js","_app/immutable/chunks/wpI5rr64.js"];
const stylesheets = ["_app/immutable/assets/Breadcrumb.BmuUfcGr.css","_app/immutable/assets/Panel.qbNo31SY.css","_app/immutable/assets/StatusPill.DBhu1QxG.css","_app/immutable/assets/Modal.CHlCiKdq.css","_app/immutable/assets/Tooltip.BrYf-THh.css","_app/immutable/assets/TimeAgo.BYfizd1M.css","_app/immutable/assets/27.DDwFDYKx.css","_app/immutable/assets/Badge.Cj2_Wv2F.css"];
const fonts = [];

export { component, fonts, imports, index, _page_server_ts as server, server_id, stylesheets };
//# sourceMappingURL=27-BdgfH1NX.js.map
