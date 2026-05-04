// @ts-nocheck
import { redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';
import { apiFetch } from '$lib/api';

const SESSION_COOKIE = 'workend_session';

interface VulnSummary {
  critical: number;
  high: number;
  medium: number;
  low: number;
  unknown: number;
  top: { id: string; package: string; severity: string; fixed_in?: string }[];
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
  scan_status: 'pending' | 'ok' | 'error' | null;
  scan_completed_at: string | null;
  vuln_summary: VulnSummary | null;
}

export const load = async ({ params, locals, cookies }: Parameters<PageServerLoad>[0]) => {
  if (!locals.user) throw redirect(303, '/login');
  const cookie = cookies.get(SESSION_COOKIE);
  const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : undefined;

  const imagesResult = await apiFetch<Image[]>(`/api/projects/${params.id}/images`, { cookie: cookieHeader });

  return {
    images: imagesResult.ok ? (imagesResult.data ?? []) : []
  };
};
