import { fail, redirect, error } from "@sveltejs/kit";
import { a as apiFetch } from "../../../chunks/api.js";
const SESSION_COOKIE = "workend_session";
const load = async ({ locals, cookies }) => {
  if (!locals.user) throw redirect(303, "/login");
  if (!locals.user.is_admin) throw error(403, "admin only");
  const cookie = cookies.get(SESSION_COOKIE);
  const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
  const [usersResult, auditResult] = await Promise.all([
    apiFetch("/api/admin/users", { cookie: cookieHeader }),
    apiFetch("/api/admin/audit-log", { cookie: cookieHeader })
  ]);
  return {
    users: usersResult.ok ? usersResult.data ?? [] : [],
    audit: auditResult.ok ? auditResult.data ?? [] : []
  };
};
const actions = {
  setRetention: async ({ request, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
    const data = await request.formData();
    const days = parseInt(String(data.get("days") || ""), 10);
    if (isNaN(days) || days < 1) {
      return fail(400, { retentionError: "A valid number of days is required (minimum 1)" });
    }
    const result = await apiFetch("/api/admin/audit-log/retention", {
      method: "POST",
      body: { days },
      cookie: cookieHeader
    });
    if (!result.ok) return fail(result.status, { retentionError: result.error || "failed to set retention" });
    return { retentionSet: true, retentionDays: days };
  }
};
export {
  actions,
  load
};
