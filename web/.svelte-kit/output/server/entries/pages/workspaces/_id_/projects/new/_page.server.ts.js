import { fail, redirect, error } from "@sveltejs/kit";
import { a as apiFetch } from "../../../../../../chunks/api.js";
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
export {
  actions,
  load
};
