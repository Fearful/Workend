// @ts-nocheck
import { error, fail, redirect } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
import { apiFetch } from '$lib/api';

const SESSION_COOKIE = 'workend_session';

interface Workspace {
  id: string;
  name: string;
}

interface ConnectionView {
  provider_id: string;
  provider: string;
  instance_url: string;
  instance_host: string;
  connected: boolean;
  handle?: string;
}

interface Repo {
  name: string;
  full_name: string;
  description: string;
  private: boolean;
  html_url: string;
  clone_url: string;
  default_branch: string;
  updated_at: string | null;
}

export const load = async ({ params, url, locals, cookies }: Parameters<PageServerLoad>[0]) => {
  if (!locals.user) throw redirect(303, '/login');
  const cookie = cookies.get(SESSION_COOKIE);
  const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : undefined;

  const wsResult = await apiFetch<Workspace>(`/api/workspaces/${params.id}`, { cookie: cookieHeader });
  if (wsResult.status === 404) throw error(404, 'workspace not found');
  if (!wsResult.ok || !wsResult.data) throw error(500, wsResult.error || 'failed to load workspace');

  const connectionsResult = await apiFetch<ConnectionView[]>('/api/me/connections', { cookie: cookieHeader });
  const connections = connectionsResult.ok ? (connectionsResult.data ?? []) : [];
  const connectedProviders = connections.filter((c) => c.connected);

  // If a `from` query param names a connected provider, fetch its repos.
  const from = url.searchParams.get('from') || '';
  const page = Math.max(1, parseInt(url.searchParams.get('page') || '1', 10));
  let repos: Repo[] = [];
  let reposError: string | null = null;
  let activeProvider: ConnectionView | null = null;

  if (from) {
    activeProvider = connectedProviders.find((c) => c.provider_id === from) ?? null;
    if (activeProvider) {
      const r = await apiFetch<Repo[]>(`/api/me/connections/${from}/repos?page=${page}`, { cookie: cookieHeader });
      if (r.ok) {
        repos = r.data ?? [];
      } else {
        reposError = r.error || 'failed to load repos';
      }
    }
  }

  return {
    workspace: wsResult.data,
    connectedProviders,
    activeProvider,
    repos,
    reposError,
    page
  };
};

export const actions = {
  default: async ({ params, request, cookies }: import('./$types').RequestEvent) => {
    const data = await request.formData();
    const name = String(data.get('name') || '').trim();
    const git_url = String(data.get('git_url') || '').trim();
    const branch = String(data.get('branch') || '').trim();

    if (!name || !git_url) {
      return fail(400, { name, git_url, branch, error: 'name and git_url required' });
    }

    const cookie = cookies.get(SESSION_COOKIE);
    const result = await apiFetch(`/api/workspaces/${params.id}/projects`, {
      method: 'POST',
      body: { name, git_url, branch },
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : undefined
    });

    if (!result.ok) {
      return fail(result.status, { name, git_url, branch, error: result.error || 'create failed' });
    }

    throw redirect(303, `/workspaces/${params.id}`);
  }
};
;null as any as Actions;