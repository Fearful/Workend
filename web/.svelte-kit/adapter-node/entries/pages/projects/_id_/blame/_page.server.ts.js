import { fail, redirect } from "@sveltejs/kit";
import { a as apiFetch } from "../../../../../chunks/api.js";
const SESSION_COOKIE = "workend_session";
const load = async ({ params, locals, cookies }) => {
  if (!locals.user) throw redirect(303, "/login");
  const cookie = cookies.get(SESSION_COOKIE);
  const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
  const blameResult = await apiFetch(
    `/api/projects/${params.id}/blame-timeline?limit=50`,
    { cookie: cookieHeader }
  );
  return {
    blameTimeline: blameResult.ok ? blameResult.data ?? [] : []
  };
};
const actions = {
  syncCommits: async ({ params, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
    const result = await apiFetch(
      `/api/projects/${params.id}/sync-commits`,
      { method: "POST", cookie: cookieHeader }
    );
    if (!result.ok) return fail(result.status, { error: result.error || "commit sync failed" });
    return { synced: result.data?.synced ?? 0, inserted: result.data?.inserted ?? 0 };
  }
};
export {
  actions,
  load
};
