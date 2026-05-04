import { fail, redirect } from "@sveltejs/kit";
import { a as apiFetch } from "../../../../chunks/api.js";
const SESSION_COOKIE = "workend_session";
const load = async ({ locals, cookies }) => {
  if (!locals.user) throw redirect(303, "/login");
  const cookie = cookies.get(SESSION_COOKIE);
  const result = await apiFetch("/api/me/notifications", {
    cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : void 0
  });
  return { notifications: result.ok ? result.data ?? [] : [] };
};
const actions = {
  create: async ({ request, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const data = await request.formData();
    const kind = String(data.get("kind") || "");
    const target = String(data.get("target") || "").trim();
    const trigger = String(data.get("trigger") || "on_failure");
    if (!kind || !target) return fail(400, { error: "kind and target required", target });
    const result = await apiFetch("/api/me/notifications", {
      method: "POST",
      body: { kind, target, trigger },
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : void 0
    });
    if (!result.ok) return fail(result.status, { error: result.error || "create failed", target });
    return { created: true };
  },
  delete: async ({ request, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const id = String((await request.formData()).get("id") || "");
    const result = await apiFetch(`/api/me/notifications/${id}`, {
      method: "DELETE",
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : void 0
    });
    if (!result.ok) return fail(result.status, { error: result.error || "delete failed" });
    return { deleted: true };
  },
  test: async ({ request, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const id = String((await request.formData()).get("id") || "");
    const result = await apiFetch(`/api/me/notifications/${id}/test`, {
      method: "POST",
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : void 0
    });
    if (!result.ok) return fail(result.status, { error: result.error || "test send failed" });
    return { tested: true };
  }
};
export {
  actions,
  load
};
