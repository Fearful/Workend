// @ts-nocheck
import { error, redirect } from '@sveltejs/kit';
import type { LayoutServerLoad } from './$types';
import { apiFetch } from '$lib/api';

const SESSION_COOKIE = 'workend_session';

interface Project {
  id: string;
  workspace_id: string;
  name: string;
  git_url: string;
  default_branch: string | null;
  local_path: string | null;
  status: string;
  last_commit_sha: string | null;
  last_commit_message: string | null;
  last_commit_author: string | null;
  last_synced_at: string | null;
  created_at: string;
  updated_at: string;
  webhook_token?: string;
}

interface Workspace {
  id: string;
  name: string;
}

export const load = async ({ params, locals, cookies }: Parameters<LayoutServerLoad>[0]) => {
  if (!locals.user) throw redirect(303, '/login');
  const cookie = cookies.get(SESSION_COOKIE);
  const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : undefined;

  const projResult = await apiFetch<Project>(`/api/projects/${params.id}`, { cookie: cookieHeader });
  if (projResult.status === 404) throw error(404, 'project not found');
  if (!projResult.ok || !projResult.data) throw error(500, projResult.error || 'failed to load');

  const wsResult = await apiFetch<Workspace>(
    `/api/workspaces/${projResult.data.workspace_id}`,
    { cookie: cookieHeader }
  );

  return {
    project: projResult.data,
    workspace: wsResult.ok ? (wsResult.data ?? null) : null
  };
};
