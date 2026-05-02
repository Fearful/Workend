<script lang="ts">
  import { onMount, onDestroy, tick } from 'svelte';
  import { invalidateAll } from '$app/navigation';

  let { data } = $props();

  let liveLog = $state('');
  let liveStatus = $state('queued');
  let liveExitCode = $state<number | null>(null);
  let autoScroll = $state(true);
  let logEl: HTMLPreElement | null = $state(null);
  let cancelling = $state(false);
  let initialized = false;

  function statusColor(status: string): string {
    switch (status) {
      case 'succeeded':
        return '#22c55e';
      case 'queued':
      case 'running':
        return '#eab308';
      case 'failed':
        return '#ef4444';
      case 'cancelled':
        return '#6b7280';
      default:
        return '#6b7280';
    }
  }

  function shortSha(sha: string | null): string {
    return sha ? sha.slice(0, 12) : '—';
  }

  function formatDuration(start: string | null, end: string | null): string {
    if (!start) return '—';
    const startMs = new Date(start).getTime();
    const endMs = end ? new Date(end).getTime() : Date.now();
    const sec = Math.round((endMs - startMs) / 1000);
    if (sec < 60) return `${sec}s`;
    return `${Math.floor(sec / 60)}m ${sec % 60}s`;
  }

  function isTerminal(status: string): boolean {
    return status === 'succeeded' || status === 'failed' || status === 'cancelled';
  }

  let es: EventSource | null = null;

  function startStream() {
    if (es || isTerminal(liveStatus)) return;
    es = new EventSource(`/runs/${data.run.id}/log-stream`);

    es.addEventListener('log', async (ev) => {
      const e = ev as MessageEvent<string>;
      liveLog += e.data + '\n';
      if (autoScroll) {
        await tick();
        scrollToBottom();
      }
    });

    es.addEventListener('done', async (ev) => {
      const e = ev as MessageEvent<string>;
      try {
        const payload = JSON.parse(e.data) as { status: string; exit_code: number };
        liveStatus = payload.status;
        liveExitCode = payload.exit_code;
      } catch {
        // ignore parse errors
      }
      stopStream();
      // Refresh server-side data so timestamps + final log are authoritative.
      await invalidateAll();
    });

    es.onerror = () => {
      stopStream();
      // The connection may have closed naturally; if the run is still
      // non-terminal in our local state, retry after a short delay.
      if (!isTerminal(liveStatus)) {
        setTimeout(startStream, 2000);
      }
    };
  }

  function stopStream() {
    if (es) {
      es.close();
      es = null;
    }
  }

  function scrollToBottom() {
    if (logEl) logEl.scrollTop = logEl.scrollHeight;
  }

  function handleScroll() {
    if (!logEl) return;
    const atBottom = logEl.scrollHeight - logEl.scrollTop - logEl.clientHeight < 16;
    autoScroll = atBottom;
  }

  async function cancel() {
    if (cancelling) return;
    cancelling = true;
    try {
      await fetch(`/runs/${data.run.id}/cancel`, { method: 'POST' });
    } finally {
      cancelling = false;
    }
  }

  // Sync server data into reactive local state. On first mount we seed from
  // whatever the server load returned; afterwards we only overwrite when the
  // run reaches a terminal state and a fresh full log is available (so SSE
  // appends mid-run aren't clobbered by a re-fetch).
  $effect(() => {
    if (!initialized) {
      liveLog = data.log;
      liveStatus = data.run.status;
      liveExitCode = data.run.exit_code;
      initialized = true;
      return;
    }
    if (isTerminal(data.run.status)) {
      if (data.log) liveLog = data.log;
      liveStatus = data.run.status;
      liveExitCode = data.run.exit_code;
    }
  });

  onMount(startStream);
  onDestroy(stopStream);
</script>

<style>
  h1 {
    font-size: 1.5rem;
    margin: 0 0 0.25rem 0;
    display: flex;
    align-items: center;
    gap: 0.75rem;
  }

  .dot {
    width: 10px;
    height: 10px;
    border-radius: 50%;
    flex-shrink: 0;
  }

  .breadcrumb {
    color: #6b7280;
    font-size: 0.875rem;
    margin-bottom: 0.5rem;
  }

  .header-row {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    margin-bottom: 1.5rem;
    gap: 1rem;
  }

  .actions {
    display: flex;
    gap: 0.5rem;
    flex-shrink: 0;
  }

  .panel {
    background: #14181d;
    border: 1px solid #1f2429;
    border-radius: 8px;
    padding: 1.25rem 1.5rem;
    margin-bottom: 1rem;
  }

  .panel h2 {
    margin-top: 0;
    font-size: 0.875rem;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: #9ca3af;
    font-weight: 600;
  }

  .row {
    display: grid;
    grid-template-columns: 140px 1fr;
    gap: 0.75rem;
    padding: 0.5rem 0;
    border-bottom: 1px solid #1f2429;
    font-size: 0.875rem;
  }

  .row:last-child {
    border-bottom: none;
  }

  .row .label {
    color: #6b7280;
  }

  .row .value {
    color: #e8eaed;
    font-family: ui-monospace, "SF Mono", Menlo, monospace;
    word-break: break-all;
  }

  .log-panel {
    padding: 0;
  }

  .log-header {
    padding: 0.875rem 1.25rem;
    border-bottom: 1px solid #1f2429;
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  .log-header h2 {
    margin: 0;
  }

  .log-controls {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    color: #6b7280;
    font-size: 0.75rem;
    font-family: ui-monospace, "SF Mono", Menlo, monospace;
  }

  pre {
    margin: 0;
    padding: 1rem 1.25rem;
    background: #0d0f12;
    color: #cbd5e1;
    font-family: ui-monospace, "SF Mono", Menlo, monospace;
    font-size: 0.8125rem;
    line-height: 1.4;
    overflow-y: auto;
    overflow-x: auto;
    white-space: pre-wrap;
    word-break: break-word;
    min-height: 8rem;
    max-height: 70vh;
  }

  .empty-log {
    color: #6b7280;
    font-style: italic;
    padding: 1.5rem 1.25rem;
    text-align: center;
  }

  .spinner {
    display: inline-block;
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: #eab308;
    margin-right: 0.25rem;
    animation: pulse 1.5s ease-in-out infinite;
  }

  @keyframes pulse {
    0%, 100% { opacity: 0.4; }
    50% { opacity: 1; }
  }
</style>

<div class="breadcrumb">
  <a href="/">workspaces</a> / <a href={`/projects/${data.run.project_id}`}>back to project</a> / run {data.run.id.slice(0, 8)}
</div>

<div class="header-row">
  <h1>
    <span class="dot" style="background: {statusColor(liveStatus)}"></span>
    {data.run.task_name} <span style="color:#6b7280">({data.run.task_source})</span>
  </h1>
  <div class="actions">
    {#if !isTerminal(liveStatus)}
      <button type="button" class="danger" disabled={cancelling} onclick={cancel}>
        {cancelling ? 'Cancelling…' : 'Cancel'}
      </button>
    {:else}
      <form method="POST" action="?/rerun" style="margin: 0;">
        <button type="submit">Re-run</button>
      </form>
      {#if data.recentRuns.length > 0}
        <select class="ghost"
                style="background:#14181d; color:#9ca3af; border:1px solid #2d3540; padding:0.5rem 0.75rem; border-radius:6px; font: inherit; font-size:0.875rem;"
                onchange={(e) => {
                  const id = (e.currentTarget as HTMLSelectElement).value;
                  if (id) window.location.href = `/runs/${data.run.id}/compare/${id}`;
                }}>
          <option value="">Compare with…</option>
          {#each data.recentRuns as r (r.id)}
            <option value={r.id}>{r.status} · {r.created_at.slice(0, 16).replace('T', ' ')}</option>
          {/each}
        </select>
      {/if}
    {/if}
  </div>
</div>

<section class="panel">
  <h2>Run</h2>
  <div class="row"><span class="label">Status</span><span class="value">{liveStatus}</span></div>
  <div class="row"><span class="label">Exit code</span><span class="value">{liveExitCode ?? '—'}</span></div>
  <div class="row"><span class="label">Commit</span><span class="value">{shortSha(data.run.commit_sha)}</span></div>
  <div class="row"><span class="label">Duration</span><span class="value">{formatDuration(data.run.started_at, data.run.finished_at)}</span></div>
  <div class="row"><span class="label">Started</span><span class="value">{data.run.started_at ? new Date(data.run.started_at).toLocaleString() : '—'}</span></div>
  <div class="row"><span class="label">Finished</span><span class="value">{data.run.finished_at ? new Date(data.run.finished_at).toLocaleString() : '—'}</span></div>
</section>

<section class="panel log-panel">
  <div class="log-header">
    <h2>Log</h2>
    <div class="log-controls">
      {#if !isTerminal(liveStatus)}
        <span><span class="spinner"></span>{liveStatus}</span>
      {/if}
      <span>{liveLog.length} bytes</span>
      {#if !autoScroll && !isTerminal(liveStatus)}
        <button type="button" class="ghost" onclick={() => { autoScroll = true; scrollToBottom(); }}>Resume tail</button>
      {/if}
    </div>
  </div>
  {#if liveLog}
    <pre bind:this={logEl} onscroll={handleScroll}>{liveLog}</pre>
  {:else if !isTerminal(liveStatus)}
    <div class="empty-log">Waiting for output…</div>
  {:else}
    <div class="empty-log">No output captured.</div>
  {/if}
</section>
