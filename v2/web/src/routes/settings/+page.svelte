<script lang="ts">
  let { data } = $props();

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
</script>

<style>
  h1 { font-size: 1.5rem; margin: 0 0 1.5rem 0; }
  h2 { font-size: 0.875rem; text-transform: uppercase; letter-spacing: 0.05em;
       color: #9ca3af; font-weight: 600; margin: 1.5rem 0 1rem 0; }
  .panel {
    background: #14181d; border: 1px solid #1f2429; border-radius: 8px;
    padding: 1.25rem 1.5rem; margin-bottom: 1rem;
  }
  .panel h2 { margin-top: 0; }
  .row {
    display: grid; grid-template-columns: auto 1fr auto;
    align-items: center; gap: 1rem;
    padding: 0.75rem 0; border-bottom: 1px solid #1f2429; font-size: 0.875rem;
  }
  .row:last-child { border-bottom: none; }
  .label { color: #6b7280; font-size: 0.875rem; }
  .value { color: #e8eaed; font-family: ui-monospace, "SF Mono", Menlo, monospace; font-size: 0.875rem; }
  .badge-on  { color: #22c55e; font-size: 0.6875rem; font-weight: 600; }
  .badge-off { color: #6b7280; font-size: 0.6875rem; }
  .hint { color: #6b7280; font-size: 0.75rem; margin: 0.25rem 0 0 0; }
  .kind-badge {
    padding: 0.125rem 0.5rem; border-radius: 4px; font-size: 0.6875rem;
    font-weight: 600; text-transform: uppercase; letter-spacing: 0.05em;
  }
  .kind-github { background: rgba(96, 165, 250, 0.15); color: #60a5fa; }
  .kind-gitlab { background: rgba(251, 146, 60, 0.15); color: #fb923c; }
  .kind-gitea  { background: rgba(34, 197, 94, 0.15);  color: #22c55e; }
  .flash-ok    { background: rgba(34, 197, 94, 0.1); border: 1px solid #166534;
                 color: #22c55e; padding: 0.5rem 0.75rem; border-radius: 6px;
                 margin-bottom: 1rem; font-size: 0.875rem; }
  .flash-err   { background: rgba(239, 68, 68, 0.1); border: 1px solid #5b1a1a;
                 color: #ef4444; padding: 0.5rem 0.75rem; border-radius: 6px;
                 margin-bottom: 1rem; font-size: 0.875rem; }
</style>

<h1>Settings</h1>

<p style="margin: -1rem 0 1.5rem 0; color: #9ca3af; font-size: 0.875rem;">
  <a href="/settings/notifications">→ Notification targets</a>
</p>

{#if data.flash.connected}
  <div class="flash-ok">Connected to {data.flash.connected}.</div>
{/if}
{#if data.flash.error}
  <div class="flash-err">Connection failed: {data.flash.error}</div>
{/if}

<section class="panel">
  <h2>Account</h2>
  <div class="row"><span class="label">Display name</span><span class="value">{data.user?.display_name}</span><span></span></div>
  <div class="row"><span class="label">Email</span><span class="value">{data.user?.email}</span><span></span></div>
</section>

<section class="panel">
  <h2>Git provider connections</h2>
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
        <span class="kind-badge kind-{c.provider}">{providerLabel(c.provider)}</span>
        <div>
          <div class="value">{c.instance_host}</div>
          {#if c.connected}
            <p class="hint">@{c.handle} <span class="badge-on">connected</span> · scopes: {c.scopes || '—'}</p>
          {:else}
            <p class="hint"><span class="badge-off">not connected</span></p>
          {/if}
        </div>
        {#if c.connected}
          <form method="POST" action="?/disconnect" style="margin: 0;">
            <input type="hidden" name="id" value={c.connection_id || ''} />
            <button type="submit" class="ghost">Disconnect</button>
          </form>
        {:else}
          <button type="button" onclick={() => connect(c.provider_id)}>Connect</button>
        {/if}
      </div>
    {/each}
  {/if}
</section>
