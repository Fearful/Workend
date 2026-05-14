<script lang="ts">
  import PageHeader from '$lib/components/PageHeader.svelte';
  import Panel from '$lib/components/Panel.svelte';
  import Badge from '$lib/components/Badge.svelte';
  import FlashMessage from '$lib/components/FlashMessage.svelte';

  let { data, form } = $props();

  let pushSupported = $state(false);
  let pushLoading = $state(false);
  let pushError = $state<string | null>(null);

  $effect(() => {
    pushSupported = typeof window !== 'undefined' && 'PushManager' in window && 'serviceWorker' in navigator;
  });

  async function connect(providerID: string) {
    const r = await fetch(`/auth/${providerID}/start`, { method: 'POST' });
    if (!r.ok) {
      alert('failed to start OAuth');
      return;
    }
    const { url } = await r.json();
    window.location.href = url;
  }

  async function enablePush() {
    if (!data.vapidPublicKey) {
      pushError = 'VAPID key not configured on server';
      return;
    }
    pushLoading = true;
    pushError = null;
    try {
      const permission = await Notification.requestPermission();
      if (permission !== 'granted') {
        pushError = 'Notification permission denied';
        return;
      }
      const reg = await navigator.serviceWorker.ready;
      const sub = await reg.pushManager.subscribe({
        userVisibleOnly: true,
        applicationServerKey: urlBase64ToUint8Array(data.vapidPublicKey).buffer as ArrayBuffer
      });
      const formData = new FormData();
      formData.set('subscription', JSON.stringify(sub.toJSON()));
      const r = await fetch('?/pushSubscribe', { method: 'POST', body: formData });
      if (!r.ok) {
        pushError = 'Failed to register subscription';
      } else {
        window.location.reload();
      }
    } catch (err) {
      pushError = err instanceof Error ? err.message : 'Push subscription failed';
    } finally {
      pushLoading = false;
    }
  }

  function urlBase64ToUint8Array(base64String: string): Uint8Array {
    const padding = '='.repeat((4 - (base64String.length % 4)) % 4);
    const base64 = (base64String + padding).replace(/-/g, '+').replace(/_/g, '/');
    const raw = atob(base64);
    const arr = new Uint8Array(raw.length);
    for (let i = 0; i < raw.length; i++) {
      arr[i] = raw.charCodeAt(i);
    }
    return arr;
  }

  function truncateEndpoint(endpoint: string): string {
    try {
      const u = new URL(endpoint);
      const path = u.pathname.length > 30 ? u.pathname.slice(0, 30) + '...' : u.pathname;
      return u.host + path;
    } catch {
      return endpoint.length > 50 ? endpoint.slice(0, 50) + '...' : endpoint;
    }
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

  const WEBHOOK_EVENTS = [
    'run_started',
    'run_completed',
    'run_failed',
    'project_created',
    'member_added'
  ] as const;
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

  .event-badges {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-1);
  }

  .webhook-url {
    font-family: var(--font-mono);
    font-size: 0.8125rem;
    word-break: break-all;
  }

  .webhook-actions {
    display: flex;
    gap: var(--space-2);
    align-items: center;
  }

  .checkbox-group {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-3);
    margin: var(--space-2) 0;
  }
  .checkbox-group label {
    display: inline-flex;
    align-items: center;
    gap: 0.375rem;
    font-size: 0.8125rem;
    color: var(--text);
    cursor: pointer;
    margin-bottom: 0;
  }
  .checkbox-group input[type="checkbox"] {
    width: auto;
    accent-color: var(--accent);
  }

  .push-info {
    background: var(--status-info-bg);
    color: var(--status-info-fg);
    border: 1px solid var(--status-info-border);
    padding: var(--space-3) var(--space-4);
    border-radius: var(--radius-md);
    font-size: 0.875rem;
  }

  .push-endpoint {
    font-family: var(--font-mono);
    font-size: 0.75rem;
    color: var(--text-muted);
  }

  @media (max-width: 640px) {
    .row { grid-template-columns: 1fr; gap: var(--space-2); }
    .row-three { grid-template-columns: 1fr; }
    .webhook-actions { flex-wrap: wrap; }
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

<Panel title="Outbound webhooks">
  <p class="hint">
    Receive HTTP POST callbacks when events happen in your workspaces.
    Optionally set a secret to verify webhook signatures.
  </p>

  {#if data.webhooks.length === 0}
    <p class="hint">No webhooks configured yet.</p>
  {:else}
    {#each data.webhooks as wh (wh.id)}
      <div class="row row-stretch">
        <div>
          <div class="webhook-url">{wh.url}</div>
          <div class="event-badges" style="margin-top: var(--space-1);">
            {#each wh.events as ev}
              <Badge variant="muted" size="sm">{ev}</Badge>
            {/each}
          </div>
          <p class="hint">added {new Date(wh.created_at).toLocaleDateString()}{wh.secret_hash ? ' · signed' : ''}</p>
        </div>
        <div class="webhook-actions">
          <form method="POST" action="?/testWebhook" class="inline-form">
            <input type="hidden" name="id" value={wh.id} />
            <button type="submit" class="ghost">Test</button>
          </form>
          <form method="POST" action="?/deleteWebhook" class="inline-form" onsubmit={(e) => !confirm('Delete this webhook?') && e.preventDefault()}>
            <input type="hidden" name="id" value={wh.id} />
            <button type="submit" class="ghost">Delete</button>
          </form>
        </div>
      </div>
    {/each}
  {/if}

  {#if form?.webhookTested}<p class="form-success">Test event sent.</p>{/if}
  {#if form?.webhookDeleted}<p class="form-success">Webhook deleted.</p>{/if}

  <form method="POST" action="?/addWebhook" style="margin-top: var(--space-4);">
    <div class="field">
      <label for="wh-url">Payload URL</label>
      <input id="wh-url" name="url" type="url" required
             placeholder="https://example.com/webhook"
             value={form?.webhookUrl || ''} />
    </div>
    <fieldset class="field" style="border: none; padding: 0; margin: 0 0 1rem 0;">
      <legend style="display: block; margin-bottom: 0.25rem; font-size: 0.875rem; color: var(--text-muted);">Events</legend>
      <div class="checkbox-group">
        {#each WEBHOOK_EVENTS as ev}
          <label>
            <input type="checkbox" name="events" value={ev} />
            {ev}
          </label>
        {/each}
      </div>
    </fieldset>
    <div class="field">
      <label for="wh-secret">Secret (optional)</label>
      <input id="wh-secret" name="secret" type="password" autocomplete="off"
             placeholder="Used for HMAC signature verification" />
    </div>
    {#if form?.webhookError}<p class="form-error">{form.webhookError}</p>{/if}
    {#if form?.webhookAdded}<p class="form-success">Webhook created.</p>{/if}
    <button type="submit">Add webhook</button>
  </form>
</Panel>

<Panel title="Push notifications">
  {#if !pushSupported}
    <div class="push-info">
      Push notifications are not supported in this browser. Use a modern browser with
      service worker support to enable push notifications.
    </div>
  {:else}
    <p class="hint">
      Receive browser push notifications for workspace events.
      You can manage subscriptions across all your devices.
    </p>

    <div style="margin: var(--space-3) 0;">
      <button type="button" onclick={enablePush} disabled={pushLoading}>
        {pushLoading ? 'Subscribing...' : 'Enable push notifications'}
      </button>
    </div>

    {#if pushError}
      <p class="form-error">{pushError}</p>
    {/if}
    {#if form?.pushSubscribed}<p class="form-success">Push notifications enabled.</p>{/if}
    {#if form?.pushUnsubscribed}<p class="form-success">Push notifications disabled.</p>{/if}
    {#if form?.pushError}<p class="form-error">{form.pushError}</p>{/if}
    {#if form?.pushSubDeleted}<p class="form-success">Subscription removed.</p>{/if}

    {#if data.pushSubscriptions.length > 0}
      <p class="hint" style="margin-top: var(--space-4);">Active subscriptions ({data.pushSubscriptions.length}):</p>
      {#each data.pushSubscriptions as sub (sub.id)}
        <div class="row row-stretch">
          <div>
            <span class="push-endpoint">{truncateEndpoint(sub.endpoint)}</span>
            <p class="hint">registered {new Date(sub.created_at).toLocaleDateString()}</p>
          </div>
          <form method="POST" action="?/deletePushSubscription" class="inline-form" onsubmit={(e) => !confirm('Remove this push subscription?') && e.preventDefault()}>
            <input type="hidden" name="id" value={sub.id} />
            <button type="submit" class="ghost">Remove</button>
          </form>
        </div>
      {/each}
    {:else}
      <p class="hint" style="margin-top: var(--space-3);">No active push subscriptions.</p>
    {/if}
  {/if}
</Panel>
