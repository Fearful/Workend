import { f as fail, r as redirect, e as error } from './index-BkmUvga9.js';
import { a as apiFetch } from './api-chcxdA-p.js';

const SESSION_COOKIE = "workend_session";
const load = async ({ params, url, locals, cookies }) => {
  if (!locals.user) throw redirect(303, "/login");
  const cookie = cookies.get(SESSION_COOKIE);
  const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
  const wsResult = await apiFetch(`/api/workspaces/${params.id}`, { cookie: cookieHeader });
  if (wsResult.status === 404) throw error(404, "workspace not found");
  if (!wsResult.ok || !wsResult.data) throw error(500, wsResult.error || "failed to load workspace");
  const connectionsResult = await apiFetch("/api/me/connections", { cookie: cookieHeader });
  const connections = connectionsResult.ok ? connectionsResult.data ?? [] : [];
  const connectedProviders = connections.filter((c) => c.connected);
  const from = url.searchParams.get("from") || "";
  const page = Math.max(1, parseInt(url.searchParams.get("page") || "1", 10));
  let repos = [];
  let reposError = null;
  let activeProvider = null;
  if (from) {
    activeProvider = connectedProviders.find((c) => c.provider_id === from) ?? null;
    if (activeProvider) {
      const r = await apiFetch(`/api/me/connections/${from}/repos?page=${page}`, { cookie: cookieHeader });
      if (r.ok) {
        repos = r.data ?? [];
      } else {
        reposError = r.error || "failed to load repos";
      }
    }
  }
  return {
    workspace: wsResult.data,
    connectedProviders,
    activeProvider,
    repos,
    reposError,
    page
  };
};
const actions = {
  default: async ({ params, request, cookies }) => {
    const data = await request.formData();
    const name = String(data.get("name") || "").trim();
    const git_url = String(data.get("git_url") || "").trim();
    const branch = String(data.get("branch") || "").trim();
    if (!name || !git_url) {
      return fail(400, { name, git_url, branch, error: "name and git_url required" });
    }
    const cookie = cookies.get(SESSION_COOKIE);
    const result = await apiFetch(`/api/workspaces/${params.id}/projects`, {
      method: "POST",
      body: { name, git_url, branch },
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : void 0
    });
    if (!result.ok) {
      return fail(result.status, { name, git_url, branch, error: result.error || "create failed" });
    }
    throw redirect(303, `/workspaces/${params.id}`);
  }
};

var _page_server_ts = /*#__PURE__*/Object.freeze({
  __proto__: null,
  actions: actions,
  load: load
});

const index = 34;
let component_cache;
const component = async () => component_cache ??= (await import('./_page.svelte-D-Qs0R8h.js')).default;
const server_id = "src/routes/workspaces/[id]/projects/new/+page.server.ts";
const imports = ["_app/immutable/nodes/34.CnDTGnf2.js","_app/immutable/chunks/CWj6FrbW.js","_app/immutable/chunks/BCGREYIq.js","_app/immutable/chunks/glOr0qzD.js","_app/immutable/chunks/BcfnGdFX.js","_app/immutable/chunks/D8xXPBaV.js","_app/immutable/chunks/BS2x50Pa.js","_app/immutable/chunks/C5qpB67O.js","_app/immutable/chunks/B2UVO_nM.js","_app/immutable/chunks/DuSA90i8.js","_app/immutable/chunks/Di6_auBm.js","_app/immutable/chunks/DdjgLl7P.js","_app/immutable/chunks/BoB1CPyg.js","_app/immutable/chunks/Bx7T9aDp.js","_app/immutable/chunks/CT9QbQKf.js","_app/immutable/chunks/wpI5rr64.js","_app/immutable/chunks/Bq1FU2Pv.js"];
const stylesheets = ["_app/immutable/assets/Breadcrumb.BmuUfcGr.css","_app/immutable/assets/PageHeader.Ct_CmObI.css","_app/immutable/assets/34.Bns4E-Mf.css"];
const fonts = [];

export { component, fonts, imports, index, _page_server_ts as server, server_id, stylesheets };
//# sourceMappingURL=34-4XSUjw8h.js.map
