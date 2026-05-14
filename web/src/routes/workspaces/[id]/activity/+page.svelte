<script lang="ts">
  import Breadcrumb from '$lib/components/Breadcrumb.svelte';
  import EmptyState from '$lib/components/EmptyState.svelte';
  import FlashMessage from '$lib/components/FlashMessage.svelte';
  import Badge from '$lib/components/Badge.svelte';
  import TimeAgo from '$lib/components/TimeAgo.svelte';

  let { data } = $props();

  interface FeedEvent {
    id: number;
    workspace_id: string;
    user_id: string | null;
    user_name: string;
    event_type: string;
    entity_type: string;
    entity_id: string;
    summary: string;
    metadata: Record<string, unknown> | null;
    created_at: string;
  }

  let extraEvents = $state<FeedEvent[]>([]);
  let allEvents = $derived([...data.events, ...extraEvents]);
  let cursorOverride = $state<string | null | undefined>(undefined);
  let cursor = $derived(cursorOverride !== undefined ? cursorOverride : data.nextCursor);
  let loading = $state(false);

  function eventColor(eventType: string): 'success' | 'warning' | 'danger' | 'muted' | 'info' | 'accent' {
    if (eventType.includes('complete') || eventType.includes('passed') || eventType.includes('succeeded')) return 'success';
    if (eventType.includes('fail') || eventType.includes('error')) return 'danger';
    if (eventType.includes('member_added') || eventType.includes('joined') || eventType.includes('invited')) return 'info';
    if (eventType.includes('start') || eventType.includes('running') || eventType.includes('queued')) return 'warning';
    if (eventType.includes('create') || eventType.includes('deploy')) return 'accent';
    return 'muted';
  }

  function eventIcon(eventType: string): string {
    if (eventType.includes('run_complete') || eventType.includes('succeeded')) return '✓';
    if (eventType.includes('run_fail') || eventType.includes('failed')) return '✗';
    if (eventType.includes('run_start') || eventType.includes('running')) return '▶';
    if (eventType.includes('member')) return '○';
    if (eventType.includes('project')) return '◇';
    if (eventType.includes('secret')) return '⬡';
    if (eventType.includes('deploy')) return '↑';
    return '·';
  }

  function entityHref(entityType: string, entityId: string): string | null {
    if (!entityId) return null;
    if (entityType === 'run') return `/runs/${entityId}`;
    if (entityType === 'project') return `/projects/${entityId}`;
    if (entityType === 'pipeline') return `/pipelines/${entityId}`;
    return null;
  }

  function initials(name: string): string {
    if (!name) return '?';
    const parts = name.trim().split(/\s+/);
    if (parts.length >= 2) return (parts[0][0] + parts[1][0]).toUpperCase();
    return name.slice(0, 2).toUpperCase();
  }

  async function loadMore() {
    if (!cursor || loading) return;
    loading = true;
    try {
      const res = await fetch(`/workspaces/${data.workspace.id}/activity?before=${cursor}`, {
        headers: { 'Accept': 'application/json' }
      });
      if (!res.ok) return;
      const page = await res.json();
      const newEvents: FeedEvent[] = page.events ?? [];
      extraEvents = [...extraEvents, ...newEvents];
      cursorOverride = newEvents.length >= 50 ? String(newEvents[newEvents.length - 1].id) : null;
    } catch {
      // silently fail, user can retry
    } finally {
      loading = false;
    }
  }

  function dateGroup(iso: string): string {
    const d = new Date(iso);
    const now = new Date();
    const today = new Date(now.getFullYear(), now.getMonth(), now.getDate());
    const eventDate = new Date(d.getFullYear(), d.getMonth(), d.getDate());
    const diffDays = Math.floor((today.getTime() - eventDate.getTime()) / 86400000);
    if (diffDays === 0) return 'Today';
    if (diffDays === 1) return 'Yesterday';
    if (diffDays < 7) return d.toLocaleDateString(undefined, { weekday: 'long' });
    return d.toLocaleDateString(undefined, { month: 'short', day: 'numeric', year: d.getFullYear() !== now.getFullYear() ? 'numeric' : undefined });
  }

  let groupedEvents = $derived.by(() => {
    const groups: { label: string; events: FeedEvent[] }[] = [];
    let currentLabel = '';
    for (const e of allEvents) {
      const label = dateGroup(e.created_at);
      if (label !== currentLabel) {
        groups.push({ label, events: [] });
        currentLabel = label;
      }
      groups[groups.length - 1].events.push(e);
    }
    return groups;
  });
</script>

<style>
  h1 {
    font-size: var(--fs-2xl);
    margin: 0 0 0.25rem 0;
    font-weight: var(--fw-semibold);
    letter-spacing: -0.02em;
    line-height: var(--lh-tight);
  }
  .role-tag {
    color: var(--text-dim);
    font-size: var(--fs-xs);
    margin-left: var(--space-2);
    font-weight: var(--fw-regular);
  }
  .header-row {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: var(--space-3);
    margin-bottom: var(--space-4);
    flex-wrap: wrap;
  }
  .ws-tabs {
    border-bottom: 1px solid var(--border);
    display: flex;
    gap: 0;
    margin-bottom: var(--space-5);
  }
  .ws-tab {
    padding: 0.625rem 0.875rem;
    color: var(--text-dim);
    text-decoration: none;
    font-size: var(--fs-md);
    font-weight: var(--fw-medium);
    border-bottom: 2px solid transparent;
    margin-bottom: -1px;
  }
  .ws-tab:hover { color: var(--text); text-decoration: none; }
  .ws-tab.active { color: var(--text); border-bottom-color: var(--accent); }

  .date-label {
    font-size: var(--fs-xs);
    font-weight: var(--fw-semibold);
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: var(--text-dim);
    padding: var(--space-3) 0 var(--space-2);
  }
  .date-label:first-child { padding-top: 0; }

  .timeline {
    position: relative;
    padding-left: 2rem;
  }
  .timeline::before {
    content: '';
    position: absolute;
    left: 0.5rem;
    top: 0;
    bottom: 0;
    width: 2px;
    background: var(--border);
    border-radius: 1px;
  }

  .event-row {
    position: relative;
    display: grid;
    grid-template-columns: 1fr auto;
    gap: var(--space-3);
    align-items: start;
    padding: var(--space-3) 0;
    border-bottom: 1px solid var(--border);
    font-size: var(--fs-sm);
  }
  .event-row:last-child { border-bottom: none; }
  .event-row:hover { background: var(--bg-hover); }

  .event-marker {
    position: absolute;
    left: -1.75rem;
    top: 1rem;
    width: 20px;
    height: 20px;
    border-radius: 50%;
    background: var(--bg-panel);
    border: 2px solid var(--border);
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 0.625rem;
    z-index: 1;
  }

  .event-body { min-width: 0; }

  .event-header {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    flex-wrap: wrap;
  }

  .avatar {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 22px;
    height: 22px;
    border-radius: 50%;
    background: var(--bg-page);
    border: 1px solid var(--border);
    font-size: 0.5625rem;
    font-weight: 600;
    color: var(--text-dim);
    flex-shrink: 0;
  }

  .event-user {
    font-weight: var(--fw-medium);
    color: var(--text);
  }
  .event-summary {
    color: var(--text-muted);
  }
  .event-entity {
    font-family: var(--font-mono);
    font-size: var(--fs-xs);
    color: var(--link);
    margin-left: var(--space-1);
  }
  .event-time {
    color: var(--text-dim);
    font-size: var(--fs-xs);
    white-space: nowrap;
    padding-top: 2px;
  }

  .load-more {
    display: flex;
    justify-content: center;
    padding: var(--space-5) 0;
  }

  @media (max-width: 768px) {
    .timeline { padding-left: 1.5rem; }
    .event-marker { left: -1.25rem; width: 16px; height: 16px; font-size: 0.5rem; }
  }
</style>

<Breadcrumb segments={[
  { label: 'workspaces', href: '/' },
  { label: data.workspace.name, href: `/workspaces/${data.workspace.id}` },
  { label: 'activity' }
]} />

<div class="header-row">
  <h1>{data.workspace.name} <span class="role-tag">{data.workspace.my_role}</span></h1>
</div>

<nav class="ws-tabs" aria-label="Workspace sections">
  <a class="ws-tab" href={`/workspaces/${data.workspace.id}`}>Overview</a>
  <a class="ws-tab" href={`/workspaces/${data.workspace.id}/dashboard`}>Dashboard</a>
  <a class="ws-tab active" href={`/workspaces/${data.workspace.id}/activity`}>Activity</a>
  <a class="ws-tab" href={`/workspaces/${data.workspace.id}/secrets`}>Secrets</a>
  <a class="ws-tab" href={`/workspaces/${data.workspace.id}/roles`}>Roles</a>
</nav>

{#if data.feedError}
  <FlashMessage type="error">{data.feedError}</FlashMessage>
{/if}

{#if allEvents.length === 0}
  <EmptyState icon="○" message="No activity yet. Events will appear here as your team works." />
{:else}
  <div class="timeline">
    {#each groupedEvents as group (group.label)}
      <div class="date-label">{group.label}</div>
      {#each group.events as event (event.id)}
        {@const href = entityHref(event.entity_type, event.entity_id)}
        <div class="event-row">
          <span class="event-marker">{eventIcon(event.event_type)}</span>
          <div class="event-body">
            <div class="event-header">
              <span class="avatar">{initials(event.user_name)}</span>
              <span class="event-user">{event.user_name || 'System'}</span>
              <Badge variant={eventColor(event.event_type)} size="sm">{event.event_type.replace(/_/g, ' ')}</Badge>
            </div>
            <div style="margin-top: 2px;">
              <span class="event-summary">{event.summary}</span>
              {#if href}
                <a href={href} class="event-entity">{event.entity_id.slice(0, 8)}</a>
              {/if}
            </div>
          </div>
          <span class="event-time"><TimeAgo value={event.created_at} /></span>
        </div>
      {/each}
    {/each}
  </div>

  {#if cursor}
    <div class="load-more">
      <button onclick={loadMore} disabled={loading}>
        {loading ? 'Loading...' : 'Load more'}
      </button>
    </div>
  {/if}
{/if}
