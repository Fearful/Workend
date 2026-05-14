// @ts-nocheck
import { error, redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';
import { apiFetch } from '$lib/api';

const SESSION_COOKIE = 'workend_session';

interface Workspace {
  id: string;
  name: string;
  description: string;
  my_role: string;
}

interface FeedEvent {
  id: number;
  workspace_id: string;
  user_id: string | null;
  user_name: string;
  event_type: string;
  entity_type: string;
  entity_id: string;
  summary: string;
  metadata: Record<string, unknown> | null;
  created_at: string;
}

export const load = async ({ params, url, locals, cookies }: Parameters<PageServerLoad>[0]) => {
  if (!locals.user) throw redirect(303, '/login');
  const cookie = cookies.get(SESSION_COOKIE);
  const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : undefined;

  const before = url.searchParams.get('before') || '';
  const limit = 50;
  const feedUrl = before
    ? `/api/workspaces/${params.id}/feed?limit=${limit}&before=${before}`
    : `/api/workspaces/${params.id}/feed?limit=${limit}`;

  const [wsResp, feedResp] = await Promise.all([
    apiFetch<Workspace>(`/api/workspaces/${params.id}`, { cookie: cookieHeader }),
    apiFetch<FeedEvent[]>(feedUrl, { cookie: cookieHeader })
  ]);

  if (wsResp.status === 404) throw error(404, 'workspace not found');
  if (!wsResp.ok || !wsResp.data) throw error(500, wsResp.error || 'failed to load workspace');

  const events = feedResp.ok ? (feedResp.data ?? []) : [];
  const nextCursor = events.length >= limit ? String(events[events.length - 1].id) : null;

  return {
    workspace: wsResp.data,
    events,
    feedError: feedResp.ok ? null : (feedResp.error || 'failed to load activity feed'),
    nextCursor,
    currentBefore: before || null
  };
};
