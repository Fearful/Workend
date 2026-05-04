// @ts-nocheck
import { fail, redirect } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
import { apiFetch } from '$lib/api';

const SESSION_COOKIE = 'workend_session';

interface Label {
  name: string;
  color: string;
}
interface BoardIssue {
  id: string;
  provider_number: number;
  title: string;
  state: string;
  labels: string[];
  author_handle: string;
  html_url: string;
  upstream_updated_at: string | null;
}
interface Board {
  id: string;
  project_id: string;
  columns: { name: string; position: number }[];
  last_synced_at: string | null;
  issues: BoardIssue[];
}
interface BoardResp {
  configured: boolean;
  available_labels?: Label[];
  board?: Board;
}

export const load = async ({ params, locals, cookies }: Parameters<PageServerLoad>[0]) => {
  if (!locals.user) throw redirect(303, '/login');
  const cookie = cookies.get(SESSION_COOKIE);
  const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : undefined;

  const boardResult = await apiFetch<BoardResp>(`/api/projects/${params.id}/board`, { cookie: cookieHeader });

  return {
    boardResp: boardResult.ok ? (boardResult.data ?? { configured: false }) : { configured: false },
    boardError: boardResult.ok ? null : (boardResult.error || null)
  };
};

export const actions = {
  setup: async ({ params, request, cookies }: import('./$types').RequestEvent) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : undefined;
    const data = await request.formData();
    const columns = data.getAll('columns').map((v) => String(v)).filter((s) => s.length > 0);
    if (columns.length === 0) return fail(400, { error: 'pick at least one label as a column' });
    const r = await apiFetch<{ ok: boolean; sync_error?: string }>(
      `/api/projects/${params.id}/board`,
      { method: 'POST', body: { columns }, cookie: cookieHeader }
    );
    if (!r.ok) return fail(r.status, { error: r.error || 'setup failed' });
    return { ok: true, syncError: r.data?.sync_error };
  },
  sync: async ({ params, cookies }: import('./$types').RequestEvent) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : undefined;
    const r = await apiFetch(`/api/projects/${params.id}/board/sync`, {
      method: 'POST',
      cookie: cookieHeader
    });
    if (!r.ok) return fail(r.status, { error: r.error || 'sync failed' });
    return { synced: true };
  }
};
;null as any as Actions;