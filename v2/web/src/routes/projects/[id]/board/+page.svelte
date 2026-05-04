<script lang="ts">
  import { invalidateAll } from '$app/navigation';
  import Panel from '$lib/components/Panel.svelte';
  import FlashMessage from '$lib/components/FlashMessage.svelte';
  import SectionHeader from '$lib/components/SectionHeader.svelte';
  import TimeAgo from '$lib/components/TimeAgo.svelte';

  let { data, form } = $props();

  let dragIssueID = $state<string | null>(null);
  let dropColumnName = $state<string | null>(null);
  let pending = $state(false);
  let moveError = $state<string | null>(null);

  let columnsView = $derived.by(() => {
    if (!data.boardResp.configured || !data.boardResp.board) return [];
    const cols = data.boardResp.board.columns
      .slice()
      .sort((a, b) => a.position - b.position);
    const colNames = new Set(cols.map((c) => c.name));
    const buckets: Record<string, typeof data.boardResp.board.issues> = {};
    for (const c of cols) buckets[c.name] = [];
    const unassigned: typeof data.boardResp.board.issues = [];
    for (const issue of data.boardResp.board.issues) {
      if (issue.state === 'closed') continue;
      const matched = issue.labels.find((l) => colNames.has(l));
      if (matched) buckets[matched].push(issue);
      else unassigned.push(issue);
    }
    const out = cols.map((c) => ({ name: c.name, issues: buckets[c.name] }));
    if (unassigned.length > 0) out.push({ name: '__unassigned__', issues: unassigned });
    return out;
  });

  function onDragStart(issueID: string, e: DragEvent) {
    dragIssueID = issueID;
    if (e.dataTransfer) {
      e.dataTransfer.effectAllowed = 'move';
      e.dataTransfer.setData('text/plain', issueID);
    }
  }
  function onDragOver(col: string, e: DragEvent) {
    if (!dragIssueID || col === '__unassigned__') return;
    e.preventDefault();
    if (e.dataTransfer) e.dataTransfer.dropEffect = 'move';
    dropColumnName = col;
  }
  function onDragLeave(col: string) {
    if (dropColumnName === col) dropColumnName = null;
  }
  async function onDrop(col: string, e: DragEvent) {
    e.preventDefault();
    const issueID = dragIssueID;
    dragIssueID = null;
    dropColumnName = null;
    if (!issueID || col === '__unassigned__') return;
    if (pending) return;
    pending = true;
    try {
      const r = await fetch(`/issues/${issueID}/move`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
        body: new URLSearchParams({ to_column: col })
      });
      if (!r.ok) {
        moveError = `Move failed (HTTP ${r.status})`;
        setTimeout(() => { moveError = null; }, 4000);
        return;
      }
      moveError = null;
      await invalidateAll();
    } finally {
      pending = false;
    }
  }
  function onDragEnd() {
    dragIssueID = null;
    dropColumnName = null;
  }

</script>

<style>
  .sync-time { color: var(--text-dim); font-size: var(--fs-sm); }

  .label-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
    gap: 0.375rem;
    margin: var(--space-2) 0 var(--space-4) 0;
  }
  .label-pick {
    display: flex;
    gap: var(--space-2);
    align-items: center;
    padding: var(--space-2) var(--space-3);
    background: var(--bg-page);
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    cursor: pointer;
    font-size: 0.875rem;
  }
  .label-pick:hover { border-color: var(--accent); }
  .label-pick input { width: auto; }
  .swatch { width: 12px; height: 12px; border-radius: 3px; flex-shrink: 0; }

  .kanban {
    display: grid;
    grid-auto-flow: column;
    grid-auto-columns: minmax(280px, 1fr);
    gap: 0.875rem;
    overflow-x: auto;
    padding-bottom: var(--space-2);
    -webkit-overflow-scrolling: touch;
  }
  @media (min-width: 1280px) {
    .kanban {
      grid-auto-columns: minmax(280px, 1fr);
    }
  }
  @media (max-width: 767px) {
    .kanban {
      grid-auto-columns: 85vw;
    }
  }

  .col {
    background: var(--bg-panel);
    border: 1px solid var(--border);
    border-radius: var(--radius-lg);
    padding: 0.625rem 0.75rem;
    min-height: 200px;
    transition: border-color 80ms ease, background 80ms ease;
  }
  .col.drop-target {
    border-color: var(--accent);
    background: rgba(96, 165, 250, 0.05);
  }
  .col-head {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: var(--space-2);
    font-size: 0.8125rem;
    color: var(--text-muted);
    text-transform: uppercase;
    letter-spacing: 0.05em;
    font-weight: 600;
  }
  .col-count { color: var(--text-dim); font-weight: 400; }

  .card {
    background: var(--bg-page);
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    padding: var(--space-2) 0.625rem;
    margin-bottom: 0.375rem;
    cursor: grab;
    text-decoration: none;
    color: inherit;
    display: block;
    transition: border-color 80ms ease, box-shadow 80ms ease;
  }
  .card:hover {
    border-color: var(--border-strong);
    text-decoration: none;
    box-shadow: var(--shadow-card);
  }
  .card:active { cursor: grabbing; }
  .card.dragging { opacity: 0.5; }
  .card-num { color: var(--text-dim); font-family: var(--font-mono); font-size: 0.6875rem; }
  .card-title { font-size: 0.875rem; line-height: 1.3; margin: 0.125rem 0 0.25rem 0; }
  .card-tags { display: flex; flex-wrap: wrap; gap: 0.25rem; }
  .tag {
    font-size: 0.625rem;
    padding: 0.0625rem 0.375rem;
    border-radius: var(--radius-full);
    background: var(--bg-hover);
    color: var(--text-muted);
  }
  .col-empty {
    color: var(--text-dim);
    font-size: 0.75rem;
    text-align: center;
    padding: var(--space-4) 0;
  }
</style>

<SectionHeader title="Issue board">
  {#snippet actions()}
    {#if data.boardResp.configured && data.boardResp.board}
      <span class="sync-time">
        {#if data.boardResp.board.last_synced_at}
          synced <TimeAgo value={data.boardResp.board.last_synced_at} />
        {:else}
          never synced
        {/if}
      </span>
      <form method="POST" action="?/sync" class="inline-form">
        <button type="submit" class="ghost">Sync</button>
      </form>
    {/if}
  {/snippet}
</SectionHeader>

{#if data.boardError}<FlashMessage type="error">{data.boardError}</FlashMessage>{/if}
{#if form?.error}<FlashMessage type="error">{form.error}</FlashMessage>{/if}
{#if moveError}<FlashMessage type="error">{moveError}</FlashMessage>{/if}
{#if form?.syncError}<FlashMessage type="error">Setup saved, but the initial sync failed: {form.syncError}</FlashMessage>{/if}
{#if form?.synced}<FlashMessage type="success">Board synced.</FlashMessage>{/if}

{#if !data.boardResp.configured}
  <Panel title="Choose your columns">
    <p style="color: var(--text-muted); font-size: 0.875rem; margin: 0 0 var(--space-2) 0;">
      Pick the upstream labels that should become board columns. Workflow-style
      labels like <em>todo</em>, <em>in progress</em>, <em>done</em> work best.
      You can change this later by re-setting up.
    </p>
    {#if (data.boardResp.available_labels ?? []).length === 0}
      <p style="color: var(--text-dim); font-size: 0.875rem;">
        No upstream labels found. Connect this project's git provider in
        <a href="/settings">Settings</a>, then come back.
      </p>
    {:else}
      <form method="POST" action="?/setup">
        <div class="label-grid">
          {#each data.boardResp.available_labels ?? [] as l (l.name)}
            <label class="label-pick">
              <input type="checkbox" name="columns" value={l.name} />
              <span class="swatch" style="background: #{l.color || '6b7280'};"></span>
              <span>{l.name}</span>
            </label>
          {/each}
        </div>
        <button type="submit">Save & sync</button>
      </form>
    {/if}
  </Panel>
{:else if data.boardResp.board}
  <div class="kanban">
    {#each columnsView as col (col.name)}
      <div class="col {dropColumnName === col.name ? 'drop-target' : ''}"
           role="region"
           aria-label={col.name === '__unassigned__' ? 'Unassigned column' : `Column ${col.name}`}
           ondragover={(e) => onDragOver(col.name, e)}
           ondragleave={() => onDragLeave(col.name)}
           ondrop={(e) => onDrop(col.name, e)}>
        <div class="col-head">
          <span>{col.name === '__unassigned__' ? 'Unassigned' : col.name}</span>
          <span class="col-count">{col.issues.length}</span>
        </div>
        {#each col.issues as i (i.id)}
          <a href={`/issues/${i.id}`}
             class="card {dragIssueID === i.id ? 'dragging' : ''}"
             draggable={col.name !== '__unassigned__'}
             ondragstart={(e) => onDragStart(i.id, e)}
             ondragend={onDragEnd}>
            <span class="card-num">#{i.provider_number}</span>
            <p class="card-title">{i.title}</p>
            <div class="card-tags">
              {#each i.labels as l (l)}
                {#if l !== col.name}
                  <span class="tag">{l}</span>
                {/if}
              {/each}
            </div>
          </a>
        {/each}
        {#if col.issues.length === 0}
          <p class="col-empty">{col.name === '__unassigned__' ? '' : 'Drop issues here'}</p>
        {/if}
      </div>
    {/each}
  </div>
{/if}
