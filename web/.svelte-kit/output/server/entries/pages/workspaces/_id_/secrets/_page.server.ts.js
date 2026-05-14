import { fail, redirect, error } from "@sveltejs/kit";
import { a as apiFetch } from "../../../../../chunks/api.js";
const SESSION_COOKIE = "workend_session";
const load = async ({ params, locals, cookies }) => {
  if (!locals.user) throw redirect(303, "/login");
  const cookie = cookies.get(SESSION_COOKIE);
  const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
  const [wsResp, secretsResp] = await Promise.all([
    apiFetch(`/api/workspaces/${params.id}`, { cookie: cookieHeader }),
    apiFetch(`/api/workspaces/${params.id}/secrets`, { cookie: cookieHeader })
  ]);
  if (wsResp.status === 404) throw error(404, "workspace not found");
  if (!wsResp.ok || !wsResp.data) throw error(500, wsResp.error || "failed to load workspace");
  return {
    workspace: wsResp.data,
    secrets: secretsResp.ok ? secretsResp.data ?? [] : [],
    secretsError: secretsResp.ok ? null : secretsResp.error || "failed to load secrets"
  };
};
const actions = {
  createSecret: async ({ request, params, cookies }) => {
    const data = await request.formData();
    const name = String(data.get("name") || "").trim();
    const value = String(data.get("value") || "");
    if (!name || !value) return fail(400, { error: "Name and value are required", name });
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
    const result = await apiFetch(`/api/workspaces/${params.id}/secrets`, {
      method: "POST",
      body: { name, value },
      cookie: cookieHeader
    });
    if (!result.ok) return fail(result.status, { error: result.error || "Failed to create secret", name });
    return { created: true };
  },
  updateSecret: async ({ request, params, cookies }) => {
    const data = await request.formData();
    const secretId = String(data.get("secret_id") || "");
    const value = String(data.get("value") || "");
    if (!secretId || !value) return fail(400, { error: "Secret ID and value are required" });
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
    const result = await apiFetch(`/api/workspaces/${params.id}/secrets/${secretId}`, {
      method: "PUT",
      body: { value },
      cookie: cookieHeader
    });
    if (!result.ok) return fail(result.status, { error: result.error || "Failed to update secret" });
    return { updated: true };
  },
  deleteSecret: async ({ request, params, cookies }) => {
    const data = await request.formData();
    const secretId = String(data.get("secret_id") || "");
    if (!secretId) return fail(400, { error: "Secret ID required" });
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
    const result = await apiFetch(`/api/workspaces/${params.id}/secrets/${secretId}`, {
      method: "DELETE",
      cookie: cookieHeader
    });
    if (!result.ok) return fail(result.status, { error: result.error || "Failed to delete secret" });
    return { deleted: true };
  }
};
export {
  actions,
  load
};
