import { error, fail, redirect } from '@sveltejs/kit';
import type { PageServerLoad, Actions } from './$types';
import { apiFetch } from '$lib/api';

const SESSION_COOKIE = 'workend_session';

interface Workspace {
  id: string;
  name: string;
  description: string;
  created_at: string;
  updated_at: string;
  my_role: string;
}

interface Project {
  id: string;
  workspace_id: string;
  name: string;
  git_url: string;
  default_branch: string | null;
  status: string;
  last_commit_sha: string | null;
  last_commit_message: string | null;
  last_commit_author: string | null;
  last_synced_at: string | null;
  created_at: string;
  updated_at: string;
}

interface Member {
  user_id: string;
  email: string;
  display_name: string;
  role: string;
  added_at: string;
}

export const load: PageServerLoad = async ({ params, locals, cookies }) => {
  if (!locals.user) throw redirect(303, '/login');
  const cookie = cookies.get(SESSION_COOKIE);
  const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : undefined;

  const [wsResp, projResp, membersResp] = await Promise.all([
    apiFetch<Workspace>(`/api/workspaces/${params.id}`, { cookie: cookieHeader }),
    apiFetch<Project[]>(`/api/workspaces/${params.id}/projects`, { cookie: cookieHeader }),
    apiFetch<Member[]>(`/api/workspaces/${params.id}/members`, { cookie: cookieHeader })
  ]);

  if (wsResp.status === 404) throw error(404, 'workspace not found');
  if (!wsResp.ok || !wsResp.data) throw error(500, wsResp.error || 'failed to load workspace');

  return {
    workspace: wsResp.data,
    projects: projResp.ok ? (projResp.data ?? []) : [],
    projectsError: projResp.ok ? null : (projResp.error || 'failed to load projects'),
    members: membersResp.ok ? (membersResp.data ?? []) : []
  };
};

export const actions: Actions = {
  delete: async ({ params, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const result = await apiFetch(`/api/workspaces/${params.id}`, {
      method: 'DELETE',
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : undefined
    });
    if (!result.ok) throw error(result.status, result.error || 'delete failed');
    throw redirect(303, '/');
  },
  addMember: async ({ params, request, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const data = await request.formData();
    const email = String(data.get('email') || '').trim();
    const role = String(data.get('role') || 'member');
    if (!email) return fail(400, { memberError: 'email required', email });
    const result = await apiFetch(`/api/workspaces/${params.id}/members`, {
      method: 'POST',
      body: { email, role },
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : undefined
    });
    if (!result.ok) return fail(result.status, { memberError: result.error || 'add failed', email });
    return { memberAdded: true };
  },
  removeMember: async ({ params, request, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const userID = String((await request.formData()).get('user_id') || '');
    if (!userID) return fail(400, { memberError: 'user_id required' });
    const result = await apiFetch(`/api/workspaces/${params.id}/members/${userID}`, {
      method: 'DELETE',
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : undefined
    });
    if (!result.ok) return fail(result.status, { memberError: result.error || 'remove failed' });
    return { memberRemoved: true };
  }
};
