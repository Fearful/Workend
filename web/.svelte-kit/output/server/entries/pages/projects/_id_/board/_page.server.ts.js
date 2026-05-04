import { fail, redirect } from "@sveltejs/kit";
import { a as apiFetch } from "../../../../../chunks/api.js";
const SESSION_COOKIE = "workend_session";
const load = async ({ params, locals, cookies }) => {
  if (!locals.user) throw redirect(303, "/login");
  const cookie = cookies.get(SESSION_COOKIE);
  const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
  const boardResult = await apiFetch(`/api/projects/${params.id}/board`, { cookie: cookieHeader });
  return {
    boardResp: boardResult.ok ? boardResult.data ?? { configured: false } : { configured: false },
    boardError: boardResult.ok ? null : boardResult.error || null
  };
};
const actions = {
  setup: async ({ params, request, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
    const data = await request.formData();
    const columns = data.getAll("columns").map((v) => String(v)).filter((s) => s.length > 0);
    if (columns.length === 0) return fail(400, { error: "pick at least one label as a column" });
    const r = await apiFetch(
      `/api/projects/${params.id}/board`,
      { method: "POST", body: { columns }, cookie: cookieHeader }
    );
    if (!r.ok) return fail(r.status, { error: r.error || "setup failed" });
    return { ok: true, syncError: r.data?.sync_error };
  },
  sync: async ({ params, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
    const r = await apiFetch(`/api/projects/${params.id}/board/sync`, {
      method: "POST",
      cookie: cookieHeader
    });
    if (!r.ok) return fail(r.status, { error: r.error || "sync failed" });
    return { synced: true };
  }
};
export {
  actions,
  load
};
