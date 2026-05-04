import { redirect, error } from "@sveltejs/kit";
import { a as apiFetch } from "../../../../../../chunks/api.js";
const SESSION_COOKIE = "workend_session";
const load = async ({ params, locals, cookies }) => {
  if (!locals.user) throw redirect(303, "/login");
  const cookie = cookies.get(SESSION_COOKIE);
  const result = await apiFetch(
    `/api/runs/${params.id}/compare?to=${encodeURIComponent(params.other)}`,
    { cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : void 0 }
  );
  if (!result.ok || !result.data) throw error(result.status, result.error || "compare failed");
  return result.data;
};
export {
  load
};
