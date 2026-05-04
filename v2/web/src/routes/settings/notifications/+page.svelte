<script lang="ts">
  import Breadcrumb from '$lib/components/Breadcrumb.svelte';
  import PageHeader from '$lib/components/PageHeader.svelte';
  import Panel from '$lib/components/Panel.svelte';
  import Badge from '$lib/components/Badge.svelte';
  import EmptyState from '$lib/components/EmptyState.svelte';
  import FlashMessage from '$lib/components/FlashMessage.svelte';

  let { data, form } = $props();

  let subConfig = $state('');
  $effect(() => {
    if (!subConfig && data.notifications.length > 0) subConfig = data.notifications[0].id;
  });
  let subScopeType = $state<'global' | 'workspace' | 'project'>('global');
  let subScopeID = $state('');
  let subSeverity = $state<'all' | 'failures' | 'off'>('failures');

  let projectsForWorkspace = $derived.by(() => {
    if (subScopeType !== 'project') return [];
    // flatten all projects across workspaces — UI lets user pick by name
    return data.workspaces.flatMap((w) => (data.projectsByWS[w.id] ?? []).map((p) => ({ ...p, ws_name: w.name })));
  });

  function kindVariant(kind: string): 'info' | 'warning' | 'success' | 'muted' {
    switch (kind) {
      case 'webhook': return 'info';
      case 'slack': return 'muted';
      case 'discord': return 'info';
      case 'teams': return 'info';
      case 'email': return 'success';
      default: return 'muted';
    }
  }
</script>

<style>
  .row {
    display: grid;
    grid-template-columns: auto 1fr 1fr auto auto;
    align-items: center;
    gap: var(--space-4);
    padding: 0.625rem 0;
    border-bottom: 1px solid var(--border);
    font-size: 0.875rem;
  }
  .row:last-child { border-bottom: none; }
  .mono { font-family: var(--font-mono); word-break: break-all; }
  .muted {
    color: var(--text-dim);
    font-size: 0.75rem;
    font-family: var(--font-mono);
  }

  .form-grid {
    display: grid;
    grid-template-columns: 140px 1fr 160px auto;
    gap: 0.75rem;
    align-items: end;
  }
  .form-grid select,
  .form-grid input {
    padding: 0.5rem 0.75rem;
    background: var(--bg-panel);
    color: var(--text);
    border: 1px solid var(--border-strong);
    border-radius: var(--radius-md);
    font-family: inherit;
    font-size: 0.875rem;
    width: 100%;
    box-sizing: border-box;
  }

  @media (max-width: 768px) {
    .row {
      grid-template-columns: 1fr;
      gap: var(--space-2);
      padding: var(--space-3) 0;
    }
    .form-grid { grid-template-columns: 1fr; }
  }

  .sub-row {
    grid-template-columns: auto 2fr auto auto auto;
  }
  .scope-tag {
    display: inline-block;
    padding: 0.0625rem 0.375rem;
    border-radius: var(--radius-sm);
    background: var(--status-info-bg);
    color: var(--status-info-fg);
    border: 1px solid var(--status-info-border);
    font-size: 0.6875rem;
    font-family: var(--font-mono);
    margin-right: 0.375rem;
    text-transform: uppercase;
  }
  .sub-form { margin-top: var(--space-3); }
  .sub-form h3 {
    margin: 0 0 var(--space-3);
    font-size: var(--fs-sm);
    color: var(--text-muted);
    text-transform: uppercase;
    letter-spacing: 0.04em;
    font-weight: var(--fw-semibold);
  }
  .sub-form-grid {
    grid-template-columns: 1.5fr 1fr 1fr 1fr auto;
  }
  @media (max-width: 768px) {
    .sub-row { grid-template-columns: 1fr; }
    .sub-form-grid { grid-template-columns: 1fr; }
  }
</style>

<Breadcrumb segments={[{ label: 'settings', href: '/settings' }, { label: 'notifications' }]} />

<PageHeader title="Notifications" />

<Panel title="Configured targets">
  {#if data.notifications.length === 0}
    <EmptyState message="No notification targets yet." />
  {:else}
    {#each data.notifications as n (n.id)}
      <div class="row">
        <Badge variant={kindVariant(n.kind)} size="sm">{n.kind}</Badge>
        <span class="mono">{n.target}</span>
        <span class="muted">trigger: {n.trigger}</span>
        <form method="POST" action="?/test" class="inline-form">
          <input type="hidden" name="id" value={n.id} />
          <button type="submit" class="ghost">Test</button>
        </form>
        <form method="POST" action="?/delete" class="inline-form" onsubmit={(e) => !confirm('Delete this notification target?') && e.preventDefault()}>
          <input type="hidden" name="id" value={n.id} />
          <button type="submit" class="ghost">Delete</button>
        </form>
      </div>
    {/each}
  {/if}
</Panel>

<Panel title="Add notification target">
  <form method="POST" action="?/create">
    <div class="form-grid">
      <div class="field">
        <label for="kind">Kind</label>
        <select id="kind" name="kind">
          <option value="webhook">Webhook</option>
          <option value="slack">Slack webhook</option>
          <option value="discord">Discord webhook</option>
          <option value="teams">Teams webhook</option>
          <option value="email">Email</option>
        </select>
      </div>
      <div class="field">
        <label for="target">Target</label>
        <input id="target" name="target" type="text" required placeholder="https://hooks.slack.com/..., https://discord.com/api/webhooks/..., or alice@example.com" value={form?.target || ''} />
      </div>
      <div class="field">
        <label for="trigger">Trigger</label>
        <select id="trigger" name="trigger">
          <option value="on_failure">On failure</option>
          <option value="on_status_change">On status change</option>
          <option value="always">Always</option>
        </select>
      </div>
      <button type="submit">Add</button>
    </div>
    {#if form?.error}<FlashMessage type="error">{form.error}</FlashMessage>{/if}
    {#if form?.tested}<FlashMessage type="success">Test notification sent.</FlashMessage>{/if}
  </form>
</Panel>

<Panel title="Subscriptions">
  {#snippet actions()}
    <span style="color: var(--text-dim); font-size: var(--fs-xs);">{data.subscriptions.length} active</span>
  {/snippet}
  <p style="color: var(--text-dim); font-size: var(--fs-sm); margin: 0 0 var(--space-3);">
    Each subscription routes runs to one notification target. A new target gets a default <code>global · failures</code> subscription so it isn't silent. Override per-workspace or per-project below — most-specific scope wins.
  </p>
  {#if data.subscriptions.length === 0}
    <EmptyState message="No subscriptions yet." />
  {:else}
    {#each data.subscriptions as s (s.id)}
      <div class="row sub-row">
        <Badge variant={kindVariant(s.config_kind ?? 'webhook')} size="sm">{s.config_kind ?? '?'}</Badge>
        <span>
          <span class="scope-tag">{s.scope_type}</span>
          <span class="mono">{s.scope_name || (s.scope_type === 'global' ? 'all runs' : '?')}</span>
        </span>
        <span class="muted">severity: {s.severity}</span>
        <span></span>
        <form method="POST" action="?/unsubscribe" class="inline-form">
          <input type="hidden" name="id" value={s.id} />
          <button type="submit" class="ghost">Remove</button>
        </form>
      </div>
    {/each}
  {/if}

  {#if data.notifications.length > 0}
    <form method="POST" action="?/subscribe" class="sub-form">
      <h3>Add subscription</h3>
      <div class="form-grid sub-form-grid">
        <div class="field">
          <label for="sub-config">Target</label>
          <select id="sub-config" name="config_id" bind:value={subConfig} required>
            {#each data.notifications as n (n.id)}
              <option value={n.id}>{n.kind} → {n.target.length > 40 ? n.target.slice(0, 40) + '…' : n.target}</option>
            {/each}
          </select>
        </div>
        <div class="field">
          <label for="sub-scope-type">Scope</label>
          <select id="sub-scope-type" name="scope_type" bind:value={subScopeType}>
            <option value="global">Global (all runs)</option>
            <option value="workspace">Workspace</option>
            <option value="project">Project</option>
          </select>
        </div>
        {#if subScopeType === 'workspace'}
          <div class="field">
            <label for="sub-scope-id">Workspace</label>
            <select id="sub-scope-id" name="scope_id" bind:value={subScopeID} required>
              <option value="">— pick one —</option>
              {#each data.workspaces as w (w.id)}
                <option value={w.id}>{w.name}</option>
              {/each}
            </select>
          </div>
        {:else if subScopeType === 'project'}
          <div class="field">
            <label for="sub-scope-id">Project</label>
            <select id="sub-scope-id" name="scope_id" bind:value={subScopeID} required>
              <option value="">— pick one —</option>
              {#each projectsForWorkspace as p (p.id)}
                <option value={p.id}>{p.ws_name} / {p.name}</option>
              {/each}
            </select>
          </div>
        {:else}
          <input type="hidden" name="scope_id" value="" />
          <span></span>
        {/if}
        <div class="field">
          <label for="sub-severity">Severity</label>
          <select id="sub-severity" name="severity" bind:value={subSeverity}>
            <option value="failures">Failures only</option>
            <option value="all">All runs</option>
            <option value="off">Mute</option>
          </select>
        </div>
        <button type="submit" disabled={subScopeType !== 'global' && !subScopeID}>Subscribe</button>
      </div>
    </form>
  {/if}
</Panel>
