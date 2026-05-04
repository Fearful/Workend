import { redirect } from "@sveltejs/kit";
import { a as apiFetch } from "../../../chunks/api.js";
const SESSION_COOKIE = "workend_session";
const actions = {
  default: async ({ cookies }) => {
    const token = cookies.get(SESSION_COOKIE);
    if (token) {
      await apiFetch("/api/auth/logout", {
        method: "POST",
        cookie: `${SESSION_COOKIE}=${token}`
      });
      cookies.delete(SESSION_COOKIE, { path: "/" });
    }
    throw redirect(303, "/login");
  }
};
export {
  actions
};
