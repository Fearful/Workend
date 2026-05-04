import { fail, redirect } from "@sveltejs/kit";
import { a as apiFetch } from "../../../chunks/api.js";
const SESSION_COOKIE = "workend_session";
const load = async ({ locals, cookies, url }) => {
  if (!locals.user) throw redirect(303, "/login");
  const cookie = cookies.get(SESSION_COOKIE);
  const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
  const unreadOnly = url.searchParams.get("unread") === "true";
  const r = await apiFetch(
    `/api/me/mentions?limit=100${unreadOnly ? "&unread_only=true" : ""}`,
    { cookie: cookieHeader }
  );
  return {
    mentions: r.ok ? r.data ?? [] : [],
    unreadOnly
  };
};
const actions = {
  markRead: async ({ request, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
    const data = await request.formData();
    const idRaw = data.get("id");
    const body = idRaw ? { id: Number(idRaw) } : {};
    const r = await apiFetch("/api/me/mentions/read", {
      method: "POST",
      body,
      cookie: cookieHeader
    });
    if (!r.ok) return fail(r.status, { error: r.error || "failed" });
    return { ok: true };
  }
};
export {
  actions,
  load
};
