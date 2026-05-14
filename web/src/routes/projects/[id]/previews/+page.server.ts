import { fail, redirect } from '@sveltejs/kit';
import type { PageServerLoad, Actions } from './$types';
import { apiFetch } from '$lib/api';

const SESSION_COOKIE = 'workend_session';

interface Preview {
  id: string;
  project_id: string;
  workspace_id: string;
  branch: string;
  status: string;
  url: string;
  auto_deploy: boolean;
  created_by: string;
  last_deployed_at: string | null;
  created_at: string;
}

export const load: PageServerLoad = async ({ params, locals, cookies }) => {
  if (!locals.user) throw redirect(303, '/login');
  const cookie = cookies.get(SESSION_COOKIE);
  const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : undefined;

  const previewsResult = await apiFetch<Preview[]>(
    `/api/projects/${params.id}/previews`,
    { cookie: cookieHeader }
  );

  return {
    previews: previewsResult.ok ? (previewsResult.data ?? []) : []
  };
};

export const actions: Actions = {
  createPreview: async ({ params, request, cookies }) => {
    const formData = await request.formData();
    const branch = String(formData.get('branch') || '').trim();
    if (!branch) return fail(400, { error: 'branch required' });
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : undefined;
    const result = await apiFetch(`/api/projects/${params.id}/previews`, {
      method: 'POST',
      body: { branch },
      cookie: cookieHeader
    });
    if (!result.ok) return fail(result.status, { error: result.error || 'preview creation failed' });
    return { previewCreated: true };
  },
  deletePreview: async ({ request, cookies }) => {
    const formData = await request.formData();
    const previewID = String(formData.get('preview_id') || '');
    if (!previewID) return fail(400, { error: 'preview_id required' });
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : undefined;
    const result = await apiFetch(`/api/previews/${previewID}`, {
      method: 'DELETE',
      cookie: cookieHeader
    });
    if (!result.ok) return fail(result.status, { error: result.error || 'delete failed' });
    return { deleted: true };
  }
};
