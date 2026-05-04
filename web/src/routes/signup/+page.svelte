<script lang="ts">
  import PageHeader from '$lib/components/PageHeader.svelte';
  import FlashMessage from '$lib/components/FlashMessage.svelte';

  let { data, form } = $props();
</script>

<style>
  form { max-width: 360px; }

  .alt {
    margin-top: var(--space-4);
    font-size: 0.875rem;
    color: var(--text-muted);
  }

  .hint {
    color: var(--text-dim);
    font-size: 0.75rem;
    margin-top: 0.25rem;
  }

  .oidc-redirect { max-width: 360px; }
  .oidc-redirect p {
    color: var(--text-muted);
    font-size: 0.875rem;
    margin-bottom: var(--space-4);
  }
</style>

<PageHeader title="Sign up" />

{#if data.oidcConfigured}
  <div class="oidc-redirect">
    <p>This instance uses {data.oidcProviderName} for authentication. Your account will be created automatically when you sign in.</p>
    <a href="/api/auth/oidc/start"><button type="button">Sign in with {data.oidcProviderName}</button></a>
    <p class="alt">Already have an account? <a href="/login">Log in</a></p>
  </div>
{:else}
  <form method="POST">
    <div class="field">
      <label for="display_name">Display name</label>
      <input id="display_name" name="display_name" type="text" required value={form?.display_name || ''} />
    </div>
    <div class="field">
      <label for="email">Email</label>
      <input id="email" name="email" type="email" autocomplete="email" required value={form?.email || ''} />
    </div>
    <div class="field">
      <label for="password">Password</label>
      <input id="password" name="password" type="password" autocomplete="new-password" required minlength="8" />
      <p class="hint">At least 8 characters.</p>
    </div>

    {#if form?.error}<FlashMessage type="error">{form.error}</FlashMessage>{/if}

    <button type="submit">Create account</button>

    <p class="alt">Already have an account? <a href="/login">Log in</a></p>
  </form>
{/if}
