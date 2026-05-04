<script lang="ts">
  import PageHeader from '$lib/components/PageHeader.svelte';
  import FlashMessage from '$lib/components/FlashMessage.svelte';

  let { data, form } = $props();

  let forceLocalLogin = $state(false);
  let showLocalLogin = $derived(!data.oidcConfigured || forceLocalLogin);
</script>

<style>
  form { max-width: 360px; }

  .alt {
    margin-top: var(--space-4);
    font-size: 0.875rem;
    color: var(--text-muted);
  }

  .oidc-section {
    max-width: 360px;
    margin-bottom: var(--space-6);
  }

  .oidc-btn {
    width: 100%;
    padding: 0.75rem;
    font-size: 1rem;
  }

  .divider {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    margin: var(--space-6) 0;
    max-width: 360px;
    color: var(--text-dim);
    font-size: 0.8125rem;
  }
  .divider::before, .divider::after {
    content: '';
    flex: 1;
    height: 1px;
    background: var(--border);
  }

  .local-toggle {
    background: none;
    border: none;
    color: var(--link);
    cursor: pointer;
    font-size: 0.8125rem;
    padding: 0;
  }
  .local-toggle:hover { text-decoration: underline; background: none; }
</style>

<PageHeader title="Log in" />

{#if data.oidcConfigured}
  <div class="oidc-section">
    <a href="/api/auth/oidc/start"><button type="button" class="oidc-btn">Sign in with {data.oidcProviderName}</button></a>
  </div>

  {#if !showLocalLogin}
    <button type="button" class="local-toggle" onclick={() => (forceLocalLogin = true)}>Use email and password instead</button>
  {:else}
    <div class="divider">or</div>
  {/if}
{/if}

{#if showLocalLogin}
  <form method="POST">
    <div class="field">
      <label for="email">Email</label>
      <input id="email" name="email" type="email" autocomplete="email" required value={form?.email || ''} />
    </div>
    <div class="field">
      <label for="password">Password</label>
      <input id="password" name="password" type="password" autocomplete="current-password" required />
    </div>

    {#if form?.error}<FlashMessage type="error">{form.error}</FlashMessage>{/if}

    <button type="submit">Log in</button>

    {#if !data.oidcConfigured}
      <p class="alt">No account? <a href="/signup">Sign up</a></p>
    {/if}
  </form>
{/if}
