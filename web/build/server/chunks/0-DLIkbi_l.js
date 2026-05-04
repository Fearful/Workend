import { a as apiFetch } from './api-chcxdA-p.js';

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

var _layout_server_ts = /*#__PURE__*/Object.freeze({
  __proto__: null,
  load: load
});

const index = 0;
let component_cache;
const component = async () => component_cache ??= (await import('./_layout.svelte-CvE8EHNd.js')).default;
const server_id = "src/routes/+layout.server.ts";
const imports = ["_app/immutable/nodes/0.BgxKk1KP.js","_app/immutable/chunks/C4JmIpy3.js","_app/immutable/chunks/apxFJXnE.js","_app/immutable/chunks/Dd7pVJ5Y.js","_app/immutable/chunks/CvynrFVy.js","_app/immutable/chunks/CvI2x4sI.js","_app/immutable/chunks/Dx6EV4QR.js","_app/immutable/chunks/cLUGwKvc.js","_app/immutable/chunks/ChBkQGhY.js","_app/immutable/chunks/CgWWeIEW.js","_app/immutable/chunks/059uK4jn.js"];
const stylesheets = ["_app/immutable/assets/0.CwKSQQm9.css"];
const fonts = [];

export { component, fonts, imports, index, _layout_server_ts as server, server_id, stylesheets };
//# sourceMappingURL=0-DLIkbi_l.js.map
