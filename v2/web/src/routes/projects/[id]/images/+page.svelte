<script lang="ts">
  import { shortSha } from '$lib/utils';
  import EmptyState from '$lib/components/EmptyState.svelte';
  import SectionHeader from '$lib/components/SectionHeader.svelte';
  import TimeAgo from '$lib/components/TimeAgo.svelte';
  import Tooltip from '$lib/components/Tooltip.svelte';

  let { data } = $props();

  function shortDigest(d: string | null): string {
    return d ? d.slice(0, 24) + '…' : '—';
  }

  function formatBytes(b: number | null): string {
    if (!b) return '—';
    const units = ['B', 'KB', 'MB', 'GB'];
    let v = b;
    let i = 0;
    while (v >= 1024 && i < units.length - 1) { v /= 1024; i++; }
    return `${v.toFixed(1)} ${units[i]}`;
  }

  type Severity = 'critical' | 'high' | 'medium' | 'low' | 'unknown';
  function sevColor(s: Severity): string {
    return ({ critical: '#ef4444', high: '#f97316', medium: '#eab308', low: '#60a5fa', unknown: 'var(--text-dim)' } as const)[s];
  }
</script>

<style>
  .row {
    display: grid;
    grid-template-columns: 1fr auto auto auto auto;
    align-items: center;
    gap: var(--space-4);
    padding: var(--space-3) var(--space-4);
    background: var(--bg-panel);
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    margin-bottom: 0.375rem;
    font-size: 0.875rem;
  }
  .mono { font-family: var(--font-mono); }
  .image-link { color: var(--link); text-decoration: none; }
  .image-link:hover { text-decoration: underline; }
  .muted {
    color: var(--text-dim);
    font-size: 0.75rem;
    font-family: var(--font-mono);
  }

  .scan-row {
    margin: -0.25rem 0 0.625rem var(--space-4);
    font-size: 0.75rem;
    display: flex;
    gap: var(--space-2);
    align-items: center;
    flex-wrap: wrap;
  }
  .scan-pending { color: var(--warning); }
  .scan-error { color: var(--danger-text); }
  .scan-ok { color: var(--success); }

  .sev-pill {
    padding: 0.0625rem 0.375rem;
    border-radius: var(--radius-sm);
  }
  .sev-pill.critical { background: rgba(239, 68, 68, 0.15); color: var(--danger-text); }
  .sev-pill.high { background: rgba(249, 115, 22, 0.15); color: #f97316; }
  .sev-pill.medium { background: rgba(234, 179, 8, 0.15); color: var(--warning); }
  .sev-pill.low { background: rgba(96, 165, 250, 0.15); color: var(--link); }

  .findings-summary {
    cursor: pointer;
    color: var(--text-muted);
  }
  .findings-list {
    margin: 0.375rem 0 0 var(--space-4);
    padding: 0;
  }
  .findings-list li {
    font-family: var(--font-mono);
    font-size: 0.6875rem;
    color: var(--text-muted);
  }

  @media (max-width: 768px) {
    .row { grid-template-columns: 1fr auto; }
    .row > :nth-child(3),
    .row > :nth-child(4) { display: none; }
  }
</style>

<SectionHeader title="Images" />

{#if data.images.length === 0}
  <EmptyState icon="◇">
    No images built yet. Add a <code>Dockerfile</code> to the repo and run the
    <code>dockerfile: build</code> task.
  </EmptyState>
{:else}
  {#each data.images as img (img.id)}
    <div class="row">
      {#if img.run_id}
        <a class="mono image-link" href={`/runs/${img.run_id}`} title="View the build run">{shortDigest(img.digest)}</a>
      {:else}
        <span class="mono">{shortDigest(img.digest)}</span>
      {/if}
      <span class="muted">{img.dockerfile_path}</span>
      <span class="muted">commit {shortSha(img.commit_sha)}</span>
      <span class="muted">{formatBytes(img.size_bytes)}</span>
      <span class="muted"><TimeAgo value={img.built_at} /></span>
    </div>
    {#if img.scan_status}
      <div class="scan-row">
        {#if img.scan_status === 'pending'}
          <span class="scan-pending">scan in progress…</span>
        {:else if img.scan_status === 'error'}
          <span class="scan-error">scan failed</span>
        {:else if img.scan_status === 'ok' && img.vuln_summary}
          {@const s = img.vuln_summary}
          {#if s.critical > 0}<span class="sev-pill critical">{s.critical} critical</span>{/if}
          {#if s.high > 0}<span class="sev-pill high">{s.high} high</span>{/if}
          {#if s.medium > 0}<span class="sev-pill medium">{s.medium} medium</span>{/if}
          {#if s.low > 0}<span class="sev-pill low">{s.low} low</span>{/if}
          {#if s.critical === 0 && s.high === 0 && s.medium === 0 && s.low === 0}
            <span class="scan-ok">no vulnerabilities</span>
          {/if}
          {#if s.top.length > 0}
            <details>
              <summary class="findings-summary">top findings</summary>
              <ul class="findings-list">
                {#each s.top as v (v.id + v.package)}
                  <li>
                    <span style="color: {sevColor((v.severity.toLowerCase() as Severity))};">{v.severity}</span>
                    · {v.id} · {v.package}{v.fixed_in ? ` → fix ${v.fixed_in}` : ''}
                  </li>
                {/each}
              </ul>
            </details>
          {/if}
        {/if}
      </div>
    {/if}
  {/each}
{/if}
