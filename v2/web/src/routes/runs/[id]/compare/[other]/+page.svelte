<script lang="ts">
  let { data } = $props();

  function statusColor(status: string): string {
    switch (status) {
      case 'succeeded': return '#22c55e';
      case 'queued': case 'running': return '#eab308';
      case 'failed': return '#ef4444';
      case 'cancelled': return '#6b7280';
      default: return '#6b7280';
    }
  }

  function shortSha(s: string | null): string { return s ? s.slice(0, 7) : '—'; }

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
  h1 { font-size: 1.5rem; margin: 0 0 0.5rem 0; }
  .breadcrumb { color: #6b7280; font-size: 0.875rem; margin-bottom: 1rem; }

  .meta-row {
    display: grid; grid-template-columns: 1fr 1fr; gap: 1rem; margin-bottom: 1rem;
  }
  .panel {
    background: #14181d; border: 1px solid #1f2429; border-radius: 8px;
    padding: 1rem 1.25rem;
  }
  .panel h2 {
    margin: 0 0 0.5rem 0;
    font-size: 0.75rem; text-transform: uppercase; letter-spacing: 0.05em;
    color: #9ca3af; font-weight: 600;
  }
  .meta-line { font-size: 0.875rem; padding: 0.25rem 0; color: #cbd5e1; }
  .meta-line .label { color: #6b7280; }
  .dot { width: 8px; height: 8px; border-radius: 50%; display: inline-block; margin-right: 0.5rem; }

  .summary {
    display: flex; gap: 1.5rem; padding: 0.75rem 1rem; background: #14181d;
    border: 1px solid #1f2429; border-radius: 8px; margin-bottom: 0.5rem;
    font-size: 0.875rem; font-family: ui-monospace, "SF Mono", Menlo, monospace;
  }
  .added   { color: #22c55e; }
  .removed { color: #ef4444; }
  .equal   { color: #6b7280; }

  pre.diff {
    margin: 0; padding: 0; background: #0d0f12;
    border: 1px solid #1f2429; border-radius: 8px;
    font-family: ui-monospace, "SF Mono", Menlo, monospace;
    font-size: 0.8125rem; line-height: 1.4;
    overflow-x: auto; max-height: 80vh;
  }
  .line { display: block; padding: 0 0.75rem; white-space: pre-wrap; word-break: break-word; }
  .line.eq { color: #94a3b8; }
  .line.add { color: #86efac; background: rgba(34, 197, 94, 0.1); }
  .line.del { color: #fca5a5; background: rgba(239, 68, 68, 0.1); }
  .line .marker { color: #475569; user-select: none; padding-right: 0.5rem; }
</style>

<div class="breadcrumb">
  <a href={`/runs/${data.left.id}`}>← back to run</a>
</div>

<h1>Run comparison</h1>

<div class="meta-row">
  <div class="panel">
    <h2>Left (older)</h2>
    <div class="meta-line"><span class="dot" style="background: {statusColor(data.left.status)}"></span><span class="label">status</span> {data.left.status} · exit {data.left.exit_code ?? '—'}</div>
    <div class="meta-line"><span class="label">commit</span> {shortSha(data.left.commit_sha)}</div>
    <div class="meta-line"><span class="label">started</span> {data.left.started_at ? new Date(data.left.started_at).toLocaleString() : '—'}</div>
  </div>
  <div class="panel">
    <h2>Right (newer)</h2>
    <div class="meta-line"><span class="dot" style="background: {statusColor(data.right.status)}"></span><span class="label">status</span> {data.right.status} · exit {data.right.exit_code ?? '—'}</div>
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
