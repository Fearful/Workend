import { fail, redirect, error } from "@sveltejs/kit";
import { a as apiFetch } from "../../../../chunks/api.js";
const SESSION_COOKIE = "workend_session";
const load = async ({ params, locals, cookies }) => {
  if (!locals.user) throw redirect(303, "/login");
  const cookie = cookies.get(SESSION_COOKIE);
  const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
  const r = await apiFetch(`/api/issues/${params.id}`, { cookie: cookieHeader });
  if (r.status === 404) throw error(404, "issue not found");
  if (!r.ok || !r.data) throw error(500, r.error || "failed to load");
  return { issue: r.data };
};
const actions = {
  comment: async ({ params, request, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
    const data = await request.formData();
    const body = String(data.get("body") || "").trim();
    if (!body) return fail(400, { commentError: "body required", draft: body });
    const r = await apiFetch(`/api/issues/${params.id}/comments`, {
      method: "POST",
      body: { body },
      cookie: cookieHeader
    });
    if (!r.ok) return fail(r.status, { commentError: r.error || "comment failed", draft: body });
    return { commented: true };
  },
  close: async ({ params, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
    const r = await apiFetch(`/api/issues/${params.id}/close`, {
      method: "POST",
      cookie: cookieHeader
    });
    if (!r.ok) return fail(r.status, { closeError: r.error || "close failed" });
    return { closed: true };
  }
};
export {
  actions,
  load
};
