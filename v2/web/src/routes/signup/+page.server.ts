import { fail, redirect } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
import { apiFetch, parseSetCookie } from '$lib/api';

interface AuthConfig {
  oidc_configured: boolean;
  oidc_provider_name?: string;
}

export const load: PageServerLoad = async ({ locals }) => {
  if (locals.user) throw redirect(303, '/');
  const authCfg = await apiFetch<AuthConfig>('/api/auth/config');
  return {
    oidcConfigured: authCfg.ok ? (authCfg.data?.oidc_configured ?? false) : false,
    oidcProviderName: authCfg.ok ? (authCfg.data?.oidc_provider_name ?? 'SSO') : 'SSO'
  };
};

export const actions: Actions = {
  default: async ({ request, cookies }) => {
    const data = await request.formData();
    const email = String(data.get('email') || '');
    const password = String(data.get('password') || '');
    const display_name = String(data.get('display_name') || '');

    if (!email || !password || !display_name) {
      return fail(400, { email, display_name, error: 'all fields required' });
    }
    if (password.length < 8) {
      return fail(400, { email, display_name, error: 'password must be at least 8 characters' });
    }

    const result = await apiFetch('/api/auth/signup', {
      method: 'POST',
      body: { email, password, display_name }
    });

    if (!result.ok) {
      return fail(result.status, { email, display_name, error: result.error || 'signup failed' });
    }

    if (result.setCookie) {
      const parsed = parseSetCookie(result.setCookie);
      if (parsed) cookies.set(parsed.name, parsed.value, parsed.options);
    }

    throw redirect(303, '/');
  }
};
