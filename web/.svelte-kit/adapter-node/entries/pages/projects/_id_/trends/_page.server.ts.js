import { redirect } from "@sveltejs/kit";
import { a as apiFetch } from "../../../../../chunks/api.js";
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
export {
  load
};
