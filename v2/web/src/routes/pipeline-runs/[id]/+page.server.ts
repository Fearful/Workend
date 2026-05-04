import { error, redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';
import { apiFetch } from '$lib/api';

const SESSION_COOKIE = 'workend_session';

interface Child {
  run_id: string;
  step: number;
  status: string;
  task_name: string;
}

interface PipelineRun {
  id: string;
  pipeline_id: string;
  status: string;
  started_at: string | null;
  finished_at: string | null;
  created_at: string;
  children?: Child[];
}

interface Pipeline {
  id: string;
  project_id: string;
  name: string;
}

export const load: PageServerLoad = async ({ params, locals, cookies }) => {
  if (!locals.user) throw redirect(303, '/login');
  const cookie = cookies.get(SESSION_COOKIE);
  const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : undefined;

  const prResult = await apiFetch<PipelineRun>(`/api/pipeline-runs/${params.id}`, { cookie: cookieHeader });
  if (prResult.status === 404) throw error(404, 'pipeline run not found');
  if (!prResult.ok || !prResult.data) throw error(500, prResult.error || 'failed to load');

  // Try to enrich with the pipeline name via project + pipelines list.
  // The pipeline-runs endpoint doesn't return the pipeline name directly;
  // we accept that and surface just the IDs if we can't resolve.
  let pipeline: Pipeline | null = null;
  // We don't know the project from the response, so skip enrichment for now;
  // a future endpoint could include pipeline_name + project_id.

  return { pipelineRun: prResult.data, pipeline };
};
