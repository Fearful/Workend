<script lang="ts">
  let { data } = $props();

  async function connectGithub() {
    const r = await fetch('/auth/github/start', { method: 'POST' });
    if (!r.ok) {
      alert('failed to start GitHub OAuth');
      return;
    }
    const { url } = await r.json();
    window.location.href = url;
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
    display: flex; align-items: center; justify-content: space-between;
    padding: 0.5rem 0; gap: 1rem;
  }
  .label { color: #6b7280; font-size: 0.875rem; }
  .value { color: #e8eaed; font-family: ui-monospace, "SF Mono", Menlo, monospace; font-size: 0.875rem; }
  .badge-on { color: #22c55e; font-size: 0.75rem; font-weight: 600; }
  .badge-off { color: #6b7280; font-size: 0.75rem; }
  .hint { color: #6b7280; font-size: 0.75rem; margin-top: 0.5rem; }
</style>

<h1>Settings</h1>

<section class="panel">
  <h2>Account</h2>
  <div class="row"><span class="label">Display name</span><span class="value">{data.user?.display_name}</span></div>
  <div class="row"><span class="label">Email</span><span class="value">{data.user?.email}</span></div>
</section>

<section class="panel">
  <h2>GitHub</h2>
  {#if !data.ghEnabled}
    <p class="hint">GitHub OAuth is not configured on this Workend instance. Set <code>WORKEND_GITHUB_CLIENT_ID</code>, <code>WORKEND_GITHUB_CLIENT_SECRET</code>, and <code>WORKEND_TOKEN_KEY</code> in compose to enable.</p>
  {:else if data.ghStatus.connected}
    <div class="row">
      <div>
        <div class="value">@{data.ghStatus.handle} <span class="badge-on">connected</span></div>
        <p class="hint">Scopes: {data.ghStatus.scopes || '—'}</p>
      </div>
      <form method="POST" action="?/disconnectGithub" style="margin: 0;">
        <button type="submit" class="ghost">Disconnect</button>
      </form>
    </div>
  {:else}
    <div class="row">
      <div>
        <div class="value">Not connected <span class="badge-off">—</span></div>
        <p class="hint">Connect GitHub to clone private repositories you have access to.</p>
      </div>
      <button type="button" onclick={connectGithub}>Connect GitHub</button>
    </div>
  {/if}
</section>
