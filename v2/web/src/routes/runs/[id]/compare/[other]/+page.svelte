<script lang="ts">
  import { shortSha } from '$lib/utils';
  import Breadcrumb from '$lib/components/Breadcrumb.svelte';
  import PageHeader from '$lib/components/PageHeader.svelte';
  import StatusDot from '$lib/components/StatusDot.svelte';

  let { data } = $props();

  let stats = $derived.by(() => {
    let added = 0, removed = 0, equal = 0;
    for (const h of data.diff) {
      if (h.op === 0) equal++;
      else if (h.op === 1) added++;
      else removed++;
    }
    return { added, removed, equal };
  });

</script>

<style>
  .meta-row {
    display: grid;
    grid-template-columns: 1fr;
    gap: var(--space-4);
    margin-bottom: var(--space-4);
  }
  @media (min-width: 768px) {
    .meta-row { grid-template-columns: 1fr 1fr; }
  }

  .panel {
    background: var(--bg-panel);
    border: 1px solid var(--border);
    border-radius: var(--radius-lg);
    padding: var(--space-4) var(--space-5);
  }
  .panel h2 {
    margin: 0 0 var(--space-2) 0;
    font-size: 0.75rem;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: var(--text-muted);
    font-weight: 600;
  }

  .meta-line {
    font-size: 0.875rem;
    padding: 0.25rem 0;
    color: var(--text);
    display: flex;
    align-items: center;
    gap: var(--space-2);
    flex-wrap: wrap;
  }
  .meta-line .label { color: var(--text-dim); }

  .summary {
    display: flex;
    gap: var(--space-6);
    padding: var(--space-3) var(--space-4);
    background: var(--bg-panel);
    border: 1px solid var(--border);
    border-radius: var(--radius-lg);
    margin-bottom: var(--space-2);
    font-size: 0.875rem;
    font-family: var(--font-mono);
  }
  .added { color: var(--success); }
  .removed { color: var(--danger-text); }
  .equal { color: var(--text-dim); }

  pre.diff {
    margin: 0;
    padding: 0;
    background: var(--bg-page);
    border: 1px solid var(--border);
    border-radius: var(--radius-lg);
    font-family: var(--font-mono);
    font-size: 0.8125rem;
    line-height: 1.4;
    overflow-x: auto;
    max-height: 80vh;
  }
  .line {
    display: block;
    padding: 0 0.75rem;
    white-space: pre-wrap;
    word-break: break-word;
  }
  .line.eq { color: #94a3b8; }
  .line.add { color: #86efac; background: rgba(34, 197, 94, 0.1); }
  .line.del { color: #fca5a5; background: rgba(239, 68, 68, 0.1); }
  .line .marker { color: #475569; user-select: none; padding-right: var(--space-2); }
</style>

<Breadcrumb segments={[{ label: '← back to run', href: `/runs/${data.left.id}` }]} />

<PageHeader title="Run comparison" />

<div class="meta-row">
  <div class="panel">
    <h2>Left (older)</h2>
    <div class="meta-line">
      <StatusDot status={data.left.status} />
      <span class="label">status</span> {data.left.status} · exit {data.left.exit_code ?? '—'}
    </div>
    <div class="meta-line"><span class="label">commit</span> {shortSha(data.left.commit_sha)}</div>
    <div class="meta-line"><span class="label">started</span> {data.left.started_at ? new Date(data.left.started_at).toLocaleString() : '—'}</div>
  </div>
  <div class="panel">
    <h2>Right (newer)</h2>
    <div class="meta-line">
      <StatusDot status={data.right.status} />
      <span class="label">status</span> {data.right.status} · exit {data.right.exit_code ?? '—'}
    </div>
    <div class="meta-line"><span class="label">commit</span> {shortSha(data.right.commit_sha)}</div>
    <div class="meta-line"><span class="label">started</span> {data.right.started_at ? new Date(data.right.started_at).toLocaleString() : '—'}</div>
  </div>
</div>

<div class="summary">
  <span class="added">+{stats.added}</span>
  <span class="removed">-{stats.removed}</span>
  <span class="equal">{stats.equal} unchanged</span>
</div>

<pre class="diff">{#each data.diff as h, i (i)}<span class="line {h.op === 0 ? 'eq' : h.op === 1 ? 'add' : 'del'}"><span class="marker">{h.op === 0 ? ' ' : h.op === 1 ? '+' : '-'}</span>{h.text.replace(/\n$/, '')}
</span>{/each}</pre>
