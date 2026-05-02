import type { PageServerLoad } from './$types';

const API_URL = process.env.WORKEND_API_URL || 'http://api:8080';

export const load: PageServerLoad = async ({ fetch }) => {
  try {
    const res = await fetch(`${API_URL}/healthz`);
    const health = await res.json();
    return { health, apiReachable: true, error: null };
  } catch (err) {
    return {
      health: null,
      apiReachable: false,
      error: err instanceof Error ? err.message : String(err)
    };
  }
};
