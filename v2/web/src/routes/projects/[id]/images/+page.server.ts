import { error, redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';
import { apiFetch } from '$lib/api';

const SESSION_COOKIE = 'workend_session';

interface Project {
  id: string;
  workspace_id: string;
  name: string;
}

interface Workspace {
  id: string;
  name: string;
}

interface Image {
  id: string;
  project_id: string;
  run_id: string | null;
  dockerfile_path: string;
  digest: string | null;
  size_bytes: number | null;
  commit_sha: string | null;
  built_at: string;
}

export const load: PageServerLoad = async ({ params, locals, cookies }) => {
  if (!locals.user) throw redirect(303, '/login');
  const cookie = cookies.get(SESSION_COOKIE);
  const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : undefined;

  const projResult = await apiFetch<Project>(`/api/projects/${params.id}`, { cookie: cookieHeader });
  if (projResult.status === 404) throw error(404, 'project not found');
  if (!projResult.ok || !projResult.data) throw error(500, projResult.error || 'failed to load');

  const [wsResult, imagesResult] = await Promise.all([
    apiFetch<Workspace>(`/api/workspaces/${projResult.data.workspace_id}`, { cookie: cookieHeader }),
    apiFetch<Image[]>(`/api/projects/${params.id}/images`, { cookie: cookieHeader })
  ]);

  return {
    project: projResult.data,
    workspace: wsResult.ok ? (wsResult.data ?? null) : null,
    images: imagesResult.ok ? (imagesResult.data ?? []) : []
  };
};
