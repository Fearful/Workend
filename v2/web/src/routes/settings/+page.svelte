<script lang="ts">
  import PageHeader from '$lib/components/PageHeader.svelte';
  import Panel from '$lib/components/Panel.svelte';
  import Badge from '$lib/components/Badge.svelte';
  import FlashMessage from '$lib/components/FlashMessage.svelte';

  let { data, form } = $props();

  async function connect(providerID: string) {
    const r = await fetch(`/auth/${providerID}/start`, { method: 'POST' });
    if (!r.ok) {
      alert('failed to start OAuth');
      return;
    }
    const { url } = await r.json();
    window.location.href = url;
  }

  function providerLabel(kind: string): string {
    switch (kind) {
      case 'github': return 'GitHub';
      case 'gitlab': return 'GitLab';
      case 'gitea': return 'Gitea';
      default: return kind;
    }
  }

  function providerVariant(kind: string): 'info' | 'warning' | 'success' | 'muted' {
    switch (kind) {
      case 'github': return 'info';
      case 'gitlab': return 'warning';
      case 'gitea': return 'success';
      default: return 'muted';
    }
  }
</script>

<style>
  .nav-link {
    margin: -1rem 0 1.5rem 0;
    color: var(--text-muted);
    font-size: 0.875rem;
  }

  .panels-grid {
    display: grid;
    grid-template-columns: 1fr;
    gap: 0;
  }

  @media (min-width: 1280px) {
    .panels-grid.two-col {
      grid-template-columns: 1fr 1fr;
      gap: var(--space-4);
    }
  }

  .row {
    display: grid;
    grid-template-columns: auto 1fr auto;
    align-items: center;
    gap: var(--space-4);
    padding: var(--space-3) 0;
    border-bottom: 1px solid var(--border);
    font-size: 0.875rem;
  }
  .row:last-child { border-bottom: none; }
  .row-stretch { grid-template-columns: 1fr auto; }
  .row-three { grid-template-columns: 1fr auto auto; }

  .label { color: var(--text-dim); font-size: 0.875rem; }
  .value { color: var(--text); font-family: var(--font-mono); font-size: 0.875rem; }
  .hint { color: var(--text-dim); font-size: 0.75rem; margin: 0.25rem 0 0 0; line-height: 1.5; }
  .ssh-key {
    word-break: break-all;
    font-family: var(--font-mono);
    font-size: 0.6875rem;
  }
  .form-error {
    color: var(--danger-text);
    font-size: 0.875rem;
    margin: 0 0 var(--space-2) 0;
  }
  .form-success {
    color: var(--success);
    font-size: 0.875rem;
    margin: 0 0 var(--space-2) 0;
  }

  @media (max-width: 640px) {
    .row { grid-template-columns: 1fr; gap: var(--space-2); }
    .row-three { grid-template-columns: 1fr; }
  }
</style>

<PageHeader title="Settings" />

<p class="nav-link">
  <a href="/settings/notifications">→ Notification targets</a>
</p>

{#if data.flash.connected}
  <FlashMessage type="success">Connected to {data.flash.connected}.</FlashMessage>
{/if}
{#if data.flash.error}
  <FlashMessage type="error">Connection failed: {data.flash.error}</FlashMessage>
{/if}

<Panel title="Account">
  <div class="row"><span class="label">Display name</span><span class="value">{data.user?.display_name}</span><span></span></div>
  <div class="row"><span class="label">Email</span><span class="value">{data.user?.email}</span><span></span></div>
</Panel>

<div class="panels-grid two-col">
  <Panel title="Git provider connections">
    {#if !data.providersConfigured || data.connections.length === 0}
      <p class="hint">
        No OAuth providers configured on this Workend instance. Set
        <code>WORKEND_GITHUB_CLIENT_ID</code> /
        <code>WORKEND_GITLAB_CLIENT_ID</code> /
        <code>WORKEND_GITEA_CLIENT_ID</code> (with matching <code>_CLIENT_SECRET</code>
        and <code>WORKEND_TOKEN_KEY</code>) in compose to enable.
      </p>
    {:else}
      {#each data.connections as c (c.provider_id)}
        <div class="row">
          <Badge variant={providerVariant(c.provider)} size="sm">{providerLabel(c.provider)}</Badge>
          <div>
            <div class="value">{c.instance_host}</div>
            {#if c.connected}
              <p class="hint">@{c.handle} · scopes: {c.scopes || '—'}</p>
            {:else}
              <p class="hint">not connected</p>
            {/if}
          </div>
          {#if c.connected}
            <form method="POST" action="?/disconnect" class="inline-form">
              <input type="hidden" name="id" value={c.connection_id || ''} />
              <button type="submit" class="ghost">Disconnect</button>
            </form>
          {:else}
            <button type="button" onclick={() => connect(c.provider_id)}>Connect</button>
          {/if}
        </div>
      {/each}
    {/if}
  </Panel>

  <Panel title="HTTPS access tokens">
    <p class="hint">
      Used for cloning private repos via <code>https://host/user/repo.git</code>
      URLs. Paste a personal access token from your git host. The token is
      encrypted at rest and never shown again after saving.
    </p>
    {#if data.patCredentials.length === 0}
      <p class="hint">No access tokens yet.</p>
    {:else}
      {#each data.patCredentials as p (p.id)}
        <div class="row row-stretch">
          <div>
            <div class="value">{p.host}</div>
            <p class="hint">{p.label} · added {new Date(p.created_at).toLocaleDateString()}</p>
          </div>
          <form method="POST" action="?/deletePAT" class="inline-form" onsubmit={(e) => !confirm('Delete this access token?') && e.preventDefault()}>
            <input type="hidden" name="id" value={p.id} />
            <button type="submit" class="ghost">Delete</button>
          </form>
        </div>
      {/each}
    {/if}

    <form method="POST" action="?/addPAT" style="margin-top: var(--space-4);">
      <div class="field">
        <label for="pat-host">Host</label>
        <input id="pat-host" name="host" type="text" required
               placeholder="e.g. github.com or gitlab.example.com"
               value={form?.patHost || ''} />
      </div>
      <div class="field">
        <label for="pat-label">Label</label>
        <input id="pat-label" name="label" type="text" required
               placeholder="e.g. personal-pat (read:repo)"
               value={form?.patLabel || ''} />
      </div>
      <div class="field">
        <label for="pat-token">Token</label>
        <input id="pat-token" name="token" type="password" required
               autocomplete="off" placeholder="ghp_… / glpat-… / etc." />
      </div>
      {#if form?.patError}<p class="form-error">{form.patError}</p>{/if}
      {#if form?.patAdded}<p class="form-success">Access token saved.</p>{/if}
      <button type="submit">Save access token</button>
    </form>
  </Panel>
</div>

<Panel title="SSH keys">
  <p class="hint">
    Used for cloning repos via <code>git@host:user/repo.git</code> URLs. Paste an
    unencrypted PEM private key and add the matching public key to your git host.
  </p>
  {#if data.sshKeys.length === 0}
    <p class="hint">No SSH keys yet.</p>
  {:else}
    {#each data.sshKeys as k (k.id)}
      <div class="row row-three">
        <div>
          <div class="value">{k.name}</div>
          <p class="hint ssh-key">{k.public_key}</p>
        </div>
        <button type="button" class="ghost"
                onclick={() => navigator.clipboard?.writeText(k.public_key)}>Copy public key</button>
        <form method="POST" action="?/deleteSSHKey" class="inline-form" onsubmit={(e) => !confirm('Delete this SSH key?') && e.preventDefault()}>
          <input type="hidden" name="id" value={k.id} />
          <button type="submit" class="ghost">Delete</button>
        </form>
      </div>
    {/each}
  {/if}

  <form method="POST" action="?/addSSHKey" style="margin-top: var(--space-4);">
    <div class="field">
      <label for="ssh-name">Key name</label>
      <input id="ssh-name" name="name" type="text" required
             placeholder="e.g. laptop-ed25519"
             value={form?.sshName || ''} />
    </div>
    <div class="field">
      <label for="ssh-private">Private key (PEM)</label>
      <textarea id="ssh-private" name="private_key" rows="6" required
                placeholder="-----BEGIN OPENSSH PRIVATE KEY-----
..."></textarea>
    </div>
    {#if form?.sshError}<p class="form-error">{form.sshError}</p>{/if}
    {#if form?.sshAdded}<p class="form-success">SSH key added.</p>{/if}
    <button type="submit">Add SSH key</button>
  </form>
</Panel>
