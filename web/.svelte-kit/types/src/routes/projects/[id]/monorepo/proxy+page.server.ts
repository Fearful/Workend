// @ts-nocheck
import { fail, redirect } from '@sveltejs/kit';
import type { PageServerLoad, Actions } from './$types';
import { apiFetch } from '$lib/api';

const SESSION_COOKIE = 'workend_session';

interface MonorepoPackage {
  id: string;
  project_id: string;
  name: string;
  path: string;
  pkg_type: string;
  task_count: number;
}

interface Task {
  id: string;
  name: string;
  source: string;
}

export const load = async ({ params, locals, cookies }: Parameters<PageServerLoad>[0]) => {
  if (!locals.user) throw redirect(303, '/login');
  const cookie = cookies.get(SESSION_COOKIE);
  const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : undefined;

  const [packagesResult, tasksResult] = await Promise.all([
    apiFetch<MonorepoPackage[]>(
      `/api/projects/${params.id}/monorepo/packages`,
      { cookie: cookieHeader }
    ),
    apiFetch<Task[]>(
      `/api/projects/${params.id}/tasks`,
      { cookie: cookieHeader }
    )
  ]);

  return {
    packages: packagesResult.ok ? (packagesResult.data ?? []) : [],
    tasks: tasksResult.ok ? (tasksResult.data ?? []) : []
  };
};

export const actions = {
  detectPackages: async ({ params, cookies }: import('./$types').RequestEvent) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : undefined;
    const result = await apiFetch<{ detected: number; synced: number }>(
      `/api/projects/${params.id}/monorepo/detect`,
      { method: 'POST', cookie: cookieHeader }
    );
    if (!result.ok) return fail(result.status, { error: result.error || 'detection failed' });
    return { detected: result.data?.detected ?? 0 };
  },
  autoMap: async ({ params, cookies }: import('./$types').RequestEvent) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : undefined;
    const result = await apiFetch<{ mapped: number }>(
      `/api/projects/${params.id}/monorepo/auto-map`,
      { method: 'POST', cookie: cookieHeader }
    );
    if (!result.ok) return fail(result.status, { error: result.error || 'auto-map failed' });
    return { mapped: result.data?.mapped ?? 0 };
  }
};
;null as any as Actions;