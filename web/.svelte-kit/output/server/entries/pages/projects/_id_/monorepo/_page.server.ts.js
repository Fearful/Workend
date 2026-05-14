import { fail, redirect } from "@sveltejs/kit";
import { a as apiFetch } from "../../../../../chunks/api.js";
const SESSION_COOKIE = "workend_session";
const load = async ({ params, locals, cookies }) => {
  if (!locals.user) throw redirect(303, "/login");
  const cookie = cookies.get(SESSION_COOKIE);
  const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
  const [packagesResult, tasksResult] = await Promise.all([
    apiFetch(
      `/api/projects/${params.id}/monorepo/packages`,
      { cookie: cookieHeader }
    ),
    apiFetch(
      `/api/projects/${params.id}/tasks`,
      { cookie: cookieHeader }
    )
  ]);
  return {
    packages: packagesResult.ok ? packagesResult.data ?? [] : [],
    tasks: tasksResult.ok ? tasksResult.data ?? [] : []
  };
};
const actions = {
  detectPackages: async ({ params, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
    const result = await apiFetch(
      `/api/projects/${params.id}/monorepo/detect`,
      { method: "POST", cookie: cookieHeader }
    );
    if (!result.ok) return fail(result.status, { error: result.error || "detection failed" });
    return { detected: result.data?.detected ?? 0 };
  },
  autoMap: async ({ params, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
    const result = await apiFetch(
      `/api/projects/${params.id}/monorepo/auto-map`,
      { method: "POST", cookie: cookieHeader }
    );
    if (!result.ok) return fail(result.status, { error: result.error || "auto-map failed" });
    return { mapped: result.data?.mapped ?? 0 };
  }
};
export {
  actions,
  load
};
