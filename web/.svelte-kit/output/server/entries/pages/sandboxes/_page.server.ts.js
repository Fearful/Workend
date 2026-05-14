import { fail, redirect } from "@sveltejs/kit";
import { a as apiFetch } from "../../../chunks/api.js";
const SESSION_COOKIE = "workend_session";
const load = async ({ locals, cookies }) => {
  if (!locals.user) throw redirect(303, "/login");
  const cookie = cookies.get(SESSION_COOKIE);
  const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
  const result = await apiFetch("/api/me/sandboxes", { cookie: cookieHeader });
  const sandboxes = result.ok ? result.data ?? [] : [];
  return { sandboxes };
};
const actions = {
  create: async ({ request, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
    const fd = await request.formData();
    const project_id = String(fd.get("project_id") || "").trim();
    const branch = String(fd.get("branch") || "").trim() || void 0;
    const expiresRaw = String(fd.get("expires_hours") || "").trim();
    if (!project_id) return fail(400, { error: "project_id is required" });
    const expires_hours = expiresRaw ? parseInt(expiresRaw, 10) : void 0;
    if (expires_hours !== void 0 && (!Number.isFinite(expires_hours) || expires_hours < 1 || expires_hours > 72)) {
      return fail(400, { error: "expires_hours must be 1-72" });
    }
    const r = await apiFetch("/api/sandboxes", {
      method: "POST",
      body: { project_id, branch, expires_hours },
      cookie: cookieHeader
    });
    if (!r.ok) return fail(r.status, { error: r.error || "failed to create sandbox" });
    if (r.data?.id) throw redirect(303, `/sandboxes/${r.data.id}`);
    return { created: true };
  },
  destroy: async ({ request, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
    const fd = await request.formData();
    const id = String(fd.get("id") || "");
    if (!id) return fail(400, { error: "id required" });
    const r = await apiFetch(`/api/sandboxes/${id}/destroy`, {
      method: "POST",
      cookie: cookieHeader
    });
    if (!r.ok) return fail(r.status, { error: r.error || "failed to destroy sandbox" });
    return { destroyed: true };
  },
  extend: async ({ request, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
    const fd = await request.formData();
    const id = String(fd.get("id") || "");
    const hoursRaw = String(fd.get("hours") || "").trim();
    if (!id) return fail(400, { error: "id required" });
    const hours = parseInt(hoursRaw, 10);
    if (!Number.isFinite(hours) || hours < 1 || hours > 72) {
      return fail(400, { error: "hours must be 1-72" });
    }
    const r = await apiFetch(`/api/sandboxes/${id}/extend`, {
      method: "POST",
      body: { hours },
      cookie: cookieHeader
    });
    if (!r.ok) return fail(r.status, { error: r.error || "failed to extend sandbox" });
    return { extended: true };
  }
};
export {
  actions,
  load
};
