// @ts-nocheck
import { error, fail, redirect } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
import { apiFetch } from '$lib/api';

const SESSION_COOKIE = 'workend_session';

interface IssueDetail {
  id: string;
  provider_number: number;
  title: string;
  state: string;
  labels: string[];
  body: string;
  author_handle: string;
  html_url: string;
  upstream_updated_at: string | null;
  comments: {
    id: string;
    provider_id: number;
    body: string;
    author_handle: string;
    html_url: string;
    created_at: string;
  }[];
}

export const load = async ({ params, locals, cookies }: Parameters<PageServerLoad>[0]) => {
  if (!locals.user) throw redirect(303, '/login');
  const cookie = cookies.get(SESSION_COOKIE);
  const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : undefined;
  const r = await apiFetch<IssueDetail>(`/api/issues/${params.id}`, { cookie: cookieHeader });
  if (r.status === 404) throw error(404, 'issue not found');
  if (!r.ok || !r.data) throw error(500, r.error || 'failed to load');
  return { issue: r.data };
};

export const actions = {
  comment: async ({ params, request, cookies }: import('./$types').RequestEvent) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : undefined;
    const data = await request.formData();
    const body = String(data.get('body') || '').trim();
    if (!body) return fail(400, { commentError: 'body required', draft: body });
    const r = await apiFetch(`/api/issues/${params.id}/comments`, {
      method: 'POST',
      body: { body },
      cookie: cookieHeader
    });
    if (!r.ok) return fail(r.status, { commentError: r.error || 'comment failed', draft: body });
    return { commented: true };
  },
  close: async ({ params, cookies }: import('./$types').RequestEvent) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : undefined;
    const r = await apiFetch(`/api/issues/${params.id}/close`, {
      method: 'POST',
      cookie: cookieHeader
    });
    if (!r.ok) return fail(r.status, { closeError: r.error || 'close failed' });
    return { closed: true };
  }
};
;null as any as Actions;