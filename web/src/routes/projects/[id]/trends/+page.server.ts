import { redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';
import { apiFetch } from '$lib/api';

const SESSION_COOKIE = 'workend_session';

interface Task {
  id: string;
  source: string;
  name: string;
}

interface TrendPoint {
  run_id: string;
  task_id: string;
  task_name: string;
  task_source: string;
  status: string;
  timed_out: boolean;
  duration_sec: number;
  finished_at: string;
}

interface TrendsResponse {
  days: number;
  points: TrendPoint[];
}

interface ActivityCell {
  date: string;
  total: number;
  succeeded: number;
  failed: number;
  cancelled: number;
}

interface ActivityResponse {
  days: number;
  cells: ActivityCell[];
}

export const load: PageServerLoad = async ({ params, url, locals, cookies }) => {
  if (!locals.user) throw redirect(303, '/login');
  const cookie = cookies.get(SESSION_COOKIE);
  const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : undefined;

  const taskID = url.searchParams.get('task_id') || '';
  const daysRaw = url.searchParams.get('days') || '30';
  const days = Math.max(1, Math.min(365, parseInt(daysRaw, 10) || 30));

  const trendsPath = taskID
    ? `/api/projects/${params.id}/trends?days=${days}&task_id=${encodeURIComponent(taskID)}`
    : `/api/projects/${params.id}/trends?days=${days}`;

  const [tasksResult, trendsResult, activityResult] = await Promise.all([
    apiFetch<Task[]>(`/api/projects/${params.id}/tasks`, { cookie: cookieHeader }),
    apiFetch<TrendsResponse>(trendsPath, { cookie: cookieHeader }),
    apiFetch<ActivityResponse>(`/api/projects/${params.id}/activity?days=365`, { cookie: cookieHeader })
  ]);

  return {
    tasks: tasksResult.ok ? (tasksResult.data ?? []) : [],
    trends: trendsResult.ok ? (trendsResult.data ?? { days, points: [] }) : { days, points: [] },
    activity: activityResult.ok ? (activityResult.data ?? { days: 365, cells: [] }) : { days: 365, cells: [] },
    filterTaskID: taskID,
    days
  };
};
