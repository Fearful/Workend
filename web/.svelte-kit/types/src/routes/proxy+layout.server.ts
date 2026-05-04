// @ts-nocheck
import type { LayoutServerLoad } from './$types';
import { apiFetch } from '$lib/api';

const SESSION_COOKIE = 'workend_session';

export const load = async ({ locals, cookies }: Parameters<LayoutServerLoad>[0]) => {
  let unreadMentions = 0;
  if (locals.user) {
    const cookie = cookies.get(SESSION_COOKIE);
    const r = await apiFetch<{ total: number; unread: number }>('/api/me/mentions/count', {
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : undefined
    });
    if (r.ok && r.data) unreadMentions = r.data.unread;
  }
  return { user: locals.user, unreadMentions };
};
