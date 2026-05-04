import { a as apiFetch } from "../../chunks/api.js";
const SESSION_COOKIE = "workend_session";
const load = async ({ locals, cookies }) => {
  let unreadMentions = 0;
  if (locals.user) {
    const cookie = cookies.get(SESSION_COOKIE);
    const r = await apiFetch("/api/me/mentions/count", {
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : void 0
    });
    if (r.ok && r.data) unreadMentions = r.data.unread;
  }
  return { user: locals.user, unreadMentions };
};
export {
  load
};
