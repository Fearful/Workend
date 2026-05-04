import { a as apiFetch } from './api-chcxdA-p.js';

const SESSION_COOKIE = "workend_session";
const handle = async ({ event, resolve }) => {
  event.locals.user = null;
  const token = event.cookies.get(SESSION_COOKIE);
  if (token) {
    const result = await apiFetch("/api/me", {
      cookie: `${SESSION_COOKIE}=${token}`,
      userAgent: event.request.headers.get("user-agent") || ""
    });
    if (result.ok && result.data) {
      event.locals.user = result.data;
    } else if (result.status === 401) {
      event.cookies.delete(SESSION_COOKIE, { path: "/" });
    }
  }
  return resolve(event);
};

export { handle };
//# sourceMappingURL=hooks.server-DI4cVHPX.js.map
