import { r as redirect } from './index-BkmUvga9.js';
import { a as apiFetch } from './api-chcxdA-p.js';

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

var _page_server_ts = /*#__PURE__*/Object.freeze({
  __proto__: null,
  actions: actions
});

const index = 9;
const server_id = "src/routes/logout/+page.server.ts";
const imports = [];
const stylesheets = [];
const fonts = [];

export { fonts, imports, index, _page_server_ts as server, server_id, stylesheets };
//# sourceMappingURL=9-D5OsyEnf.js.map
