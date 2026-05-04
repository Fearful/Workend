import { fail, redirect } from '@sveltejs/kit';
import type { PageServerLoad, Actions } from './$types';
import { apiFetch } from '$lib/api';

const SESSION_COOKIE = 'workend_session';

interface Branch {
  name: string;
  commit_sha: string;
  protected: boolean;
  default: boolean;
}

interface BranchesResp {
  source: 'provider' | 'ls-remote';
  branches: Branch[];
}

export const load: PageServerLoad = async ({ params, locals, cookies }) => {
  if (!locals.user) throw redirect(303, '/login');
  const cookie = cookies.get(SESSION_COOKIE);
  const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : undefined;

  const branchesResult = await apiFetch<BranchesResp>(`/api/projects/${params.id}/branches`, { cookie: cookieHeader });

  return {
    branchesSource: branchesResult.ok ? branchesResult.data?.source ?? 'ls-remote' : 'ls-remote',
    branches: branchesResult.ok ? (branchesResult.data?.branches ?? []) : [],
    branchesError: branchesResult.ok ? null : (branchesResult.error || 'failed to list branches'),
    canCreatePR: branchesResult.ok && branchesResult.data?.source === 'provider'
  };
};

export const actions: Actions = {
  switch: async ({ request, params, cookies }) => {
    const data = await request.formData();
    const name = String(data.get('name') || '').trim();
    if (!name) return fail(400, { error: 'branch name required' });
    const cookie = cookies.get(SESSION_COOKIE);
    const result = await apiFetch(`/api/projects/${params.id}/branch`, {
      method: 'POST',
      body: { name },
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : undefined
    });
    if (!result.ok) return fail(result.status, { error: result.error || 'switch failed' });
    throw redirect(303, `/projects/${params.id}`);
  },
  createPR: async ({ request, params, cookies }) => {
    const data = await request.formData();
    const source = String(data.get('source') || '').trim();
    const target = String(data.get('target') || '').trim();
    const title = String(data.get('title') || '').trim();
    const body = String(data.get('body') || '');
    if (!source || !target) return fail(400, { error: 'source and target required' });
    if (source === target) return fail(400, { error: 'source and target must differ' });

    const cookie = cookies.get(SESSION_COOKIE);
    const result = await apiFetch<{ url: string; number: number }>(`/api/projects/${params.id}/pull-requests`, {
      method: 'POST',
      body: { source, target, title, body },
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : undefined
    });
    if (!result.ok || !result.data) {
      return fail(result.status, { error: result.error || 'PR creation failed' });
    }
    return { prURL: result.data.url, prNumber: result.data.number };
  }
};
