import { redirect } from "@sveltejs/kit";
import { a as apiFetch } from "../../../../../chunks/api.js";
const SESSION_COOKIE = "workend_session";
const allowedStatuses = /* @__PURE__ */ new Set(["succeeded", "failed", "cancelled", "running", "queued"]);
const load = async ({ params, url, locals, cookies }) => {
  if (!locals.user) throw redirect(303, "/login");
  const cookie = cookies.get(SESSION_COOKIE);
  const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
  const status = url.searchParams.get("status") || "";
  const filter = allowedStatuses.has(status) ? status : "";
  const path = filter ? `/api/projects/${params.id}/runs?status=${encodeURIComponent(filter)}` : `/api/projects/${params.id}/runs`;
  const runsResult = await apiFetch(path, { cookie: cookieHeader });
  return {
    runs: runsResult.ok ? runsResult.data ?? [] : [],
    filter
  };
};
export {
  load
};
