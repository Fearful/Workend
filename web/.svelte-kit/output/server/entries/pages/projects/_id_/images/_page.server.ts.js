import { redirect } from "@sveltejs/kit";
import { a as apiFetch } from "../../../../../chunks/api.js";
const SESSION_COOKIE = "workend_session";
const load = async ({ params, locals, cookies }) => {
  if (!locals.user) throw redirect(303, "/login");
  const cookie = cookies.get(SESSION_COOKIE);
  const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
  const imagesResult = await apiFetch(`/api/projects/${params.id}/images`, { cookie: cookieHeader });
  return {
    images: imagesResult.ok ? imagesResult.data ?? [] : []
  };
};
export {
  load
};
