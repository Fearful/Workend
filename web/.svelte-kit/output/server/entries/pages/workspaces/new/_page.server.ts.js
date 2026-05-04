import { fail, redirect } from "@sveltejs/kit";
import { a as apiFetch } from "../../../../chunks/api.js";
const SESSION_COOKIE = "workend_session";
const load = async ({ locals }) => {
  if (!locals.user) throw redirect(303, "/login");
  return {};
};
const actions = {
  default: async ({ request, cookies }) => {
    const data = await request.formData();
    const name = String(data.get("name") || "").trim();
    const description = String(data.get("description") || "").trim();
    if (!name) {
      return fail(400, { name, description, error: "name required" });
    }
    const token = cookies.get(SESSION_COOKIE);
    const result = await apiFetch("/api/workspaces", {
      method: "POST",
      body: { name, description },
      cookie: token ? `${SESSION_COOKIE}=${token}` : void 0
    });
    if (!result.ok) {
      return fail(result.status, { name, description, error: result.error || "create failed" });
    }
    throw redirect(303, "/");
  }
};
export {
  actions,
  load
};
