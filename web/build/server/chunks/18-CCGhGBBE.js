import { r as redirect } from './index-BkmUvga9.js';
import { a as apiFetch } from './api-chcxdA-p.js';

const SESSION_COOKIE = "workend_session";
const load = async ({ params, url, locals, cookies }) => {
  if (!locals.user) throw redirect(303, "/login");
  const cookie = cookies.get(SESSION_COOKIE);
  const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
  const taskID = url.searchParams.get("task_id") || "";
  const daysRaw = url.searchParams.get("days") || "30";
  const days = Math.max(1, Math.min(365, parseInt(daysRaw, 10) || 30));
  const trendsPath = taskID ? `/api/projects/${params.id}/trends?days=${days}&task_id=${encodeURIComponent(taskID)}` : `/api/projects/${params.id}/trends?days=${days}`;
  const [tasksResult, trendsResult, activityResult] = await Promise.all([
    apiFetch(`/api/projects/${params.id}/tasks`, { cookie: cookieHeader }),
    apiFetch(trendsPath, { cookie: cookieHeader }),
    apiFetch(`/api/projects/${params.id}/activity?days=365`, { cookie: cookieHeader })
  ]);
  return {
    tasks: tasksResult.ok ? tasksResult.data ?? [] : [],
    trends: trendsResult.ok ? trendsResult.data ?? { days, points: [] } : { days, points: [] },
    activity: activityResult.ok ? activityResult.data ?? { days: 365, cells: [] } : { days: 365, cells: [] },
    filterTaskID: taskID,
    days
  };
};

var _page_server_ts = /*#__PURE__*/Object.freeze({
  __proto__: null,
  load: load
});

const index = 18;
let component_cache;
const component = async () => component_cache ??= (await import('./_page.svelte-BQpE39IM.js')).default;
const server_id = "src/routes/projects/[id]/trends/+page.server.ts";
const imports = ["_app/immutable/nodes/18.8OMNqkNq.js","_app/immutable/chunks/C4JmIpy3.js","_app/immutable/chunks/apxFJXnE.js","_app/immutable/chunks/CvynrFVy.js","_app/immutable/chunks/CvI2x4sI.js","_app/immutable/chunks/Bn0VCKPQ.js","_app/immutable/chunks/cLUGwKvc.js","_app/immutable/chunks/CYuiovvJ.js","_app/immutable/chunks/Dx6EV4QR.js","_app/immutable/chunks/BZxKEKJE.js","_app/immutable/chunks/ChBkQGhY.js","_app/immutable/chunks/BiQCKUjL.js","_app/immutable/chunks/Di6J7QjO.js"];
const stylesheets = ["_app/immutable/assets/Panel.Dd2v_-2B.css","_app/immutable/assets/EmptyState.ByauYn1J.css","_app/immutable/assets/18.wycGJu-N.css"];
const fonts = [];

export { component, fonts, imports, index, _page_server_ts as server, server_id, stylesheets };
//# sourceMappingURL=18-CCGhGBBE.js.map
