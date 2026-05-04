<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { invalidateAll } from '$app/navigation';
  import { formatRelative, shortSha as _shortSha, formatDuration } from '$lib/utils';
  import Breadcrumb from '$lib/components/Breadcrumb.svelte';
  import Panel from '$lib/components/Panel.svelte';
  import StatusPill from '$lib/components/StatusPill.svelte';
  import Badge from '$lib/components/Badge.svelte';
  import LogViewer from '$lib/components/LogViewer.svelte';
  import TimeAgo from '$lib/components/TimeAgo.svelte';
  import Tooltip from '$lib/components/Tooltip.svelte';

  let { data, form } = $props();
  let commentDraft = $state('');
  $effect(() => {
    if (form?.commentDraft) commentDraft = form.commentDraft;
  });

  function bodyParts(body: string): { text: string; mention: boolean }[] {
    const parts: { text: string; mention: boolean }[] = [];
    const re = /@([A-Za-z0-9._-]+)/g;
    let last = 0;
    let m: RegExpExecArray | null;
    while ((m = re.exec(body)) !== null) {
      if (m.index > last) parts.push({ text: body.slice(last, m.index), mention: false });
      parts.push({ text: m[0], mention: true });
      last = m.index + m[0].length;
    }
    if (last < body.length) parts.push({ text: body.slice(last), mention: false });
    return parts;
  }

  function shortSha(sha: string | null): string {
    return _shortSha(sha, 12);
  }

  function formatBytes(b: number): string {
    if (b < 1024) return `${b} B`;
    if (b < 1024 * 1024) return `${(b / 1024).toFixed(1)} KB`;
    if (b < 1024 * 1024 * 1024) return `${(b / 1024 / 1024).toFixed(1)} MB`;
    return `${(b / 1024 / 1024 / 1024).toFixed(2)} GB`;
  }

  let liveLog = $state('');
  let liveStatus = $state('queued');
  let liveExitCode = $state<number | null>(null);
  let cancelling = $state(false);
  let initialized = false;

  function isTerminal(status: string): boolean {
    return status === 'succeeded' || status === 'failed' || status === 'cancelled';
  }

  function isPendingApproval(status: string): boolean {
    return status === 'pending_approval';
  }

  let es: EventSource | null = null;
  let notifyOnComplete = $state(false);

  function maybeFireDesktopNotification() {
    if (!notifyOnComplete) return;
    if (typeof Notification === 'undefined') return;
    if (Notification.permission !== 'granted') return;
    if (typeof document !== 'undefined' && !document.hidden) return;
    const title = liveStatus === 'succeeded'
      ? `${data.run.task_name} succeeded`
      : `${data.run.task_name} ${liveStatus}`;
    try {
      new Notification(title, {
        body: `${data.run.task_source} · exit ${liveExitCode ?? '?'}`,
        tag: `workend-run-${data.run.id}`
      });
    } catch {
      // noop
    }
  }

  async function toggleNotify() {
    if (typeof Notification === 'undefined') {
      alert('This browser does not support desktop notifications.');
      return;
    }
    if (Notification.permission === 'denied') {
      alert('Notifications are blocked. Allow them in your browser settings to enable.');
      return;
    }
    if (notifyOnComplete) {
      notifyOnComplete = false;
      return;
    }
    if (Notification.permission === 'default') {
      const result = await Notification.requestPermission();
      if (result !== 'granted') return;
    }
    notifyOnComplete = true;
  }

  function startStream() {
    if (es || isTerminal(liveStatus)) return;
    es = new EventSource(`/runs/${data.run.id}/log-stream`);

    es.addEventListener('log', (ev) => {
      const e = ev as MessageEvent<string>;
      liveLog += e.data + '\n';
    });

    es.addEventListener('done', async (ev) => {
      const e = ev as MessageEvent<string>;
      try {
        const payload = JSON.parse(e.data) as { status: string; exit_code: number };
        liveStatus = payload.status;
        liveExitCode = payload.exit_code;
      } catch {
        // ignore
      }
      stopStream();
      maybeFireDesktopNotification();
      await invalidateAll();
    });

    es.onerror = () => {
      stopStream();
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

  async function cancel() {
    if (cancelling) return;
    cancelling = true;
    try {
      await fetch(`/runs/${data.run.id}/cancel`, { method: 'POST' });
    } finally {
      cancelling = false;
    }
  }

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
    font-size: var(--fs-xl);
    margin: 0;
    font-weight: var(--fw-semibold);
    letter-spacing: -0.02em;
    line-height: var(--lh-tight);
  }

  .hero {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: var(--space-4);
    margin-bottom: var(--space-5);
    flex-wrap: wrap;
  }
  .hero-main { display: flex; flex-direction: column; gap: var(--space-2); min-width: 0; flex: 1; }
  .hero-line {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: var(--space-2);
  }
  .hero-meta {
    display: flex;
    flex-wrap: wrap;
    align-items: baseline;
    gap: var(--space-2);
    color: var(--text-dim);
    font-size: var(--fs-sm);
  }
  .hero-meta-item { display: inline-flex; align-items: baseline; gap: 0.375rem; }
  .hero-meta-label {
    color: var(--text-dim);
    font-size: 0.6875rem;
    text-transform: uppercase;
    letter-spacing: 0.04em;
  }
  .hero-meta-item strong {
    color: var(--text);
    font-weight: var(--fw-medium);
  }
  .mono { font-family: var(--font-mono); }
  .hero-sep { color: var(--border-strong); user-select: none; }

  .actions {
    display: flex;
    gap: var(--space-2);
    flex-shrink: 0;
    flex-wrap: wrap;
  }
  .approve-form { display: flex; gap: var(--space-2); margin: 0; }

  .dim { color: var(--text-dim); }

  .layout {
    display: grid;
    grid-template-columns: 1fr;
    gap: var(--space-4);
  }

  @media (min-width: 1280px) {
    .layout {
      grid-template-columns: minmax(0, 1fr) 320px;
      align-items: start;
    }
    .layout > .main-col { min-width: 0; }
    .layout > .meta-col { position: sticky; top: var(--space-4); }
    .layout > .meta-col > :global(section) { margin-bottom: var(--space-4); }
  }

  .row {
    display: grid;
    grid-template-columns: 110px 1fr;
    gap: var(--space-3);
    padding: var(--space-2) 0;
    border-bottom: 1px solid var(--border);
    font-size: 0.875rem;
  }
  .row:last-child { border-bottom: none; }
  .row .label { color: var(--text-dim); }
  .row .value {
    color: var(--text);
    font-family: var(--font-mono);
    word-break: break-all;
  }
  .value-pre { white-space: pre-wrap; }

  .log-section {
    padding: var(--space-3) var(--space-4);
    margin-bottom: var(--space-4);
  }
  .log-section-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-2);
    margin-bottom: var(--space-3);
    flex-wrap: wrap;
  }
  .log-section-head h2 {
    margin: 0;
    font-size: var(--fs-md);
    font-weight: var(--fw-semibold);
    color: var(--text);
    display: flex;
    align-items: baseline;
    gap: var(--space-2);
  }
  .log-bytes {
    color: var(--text-dim);
    font-weight: var(--fw-regular);
    font-size: var(--fs-xs);
    font-family: var(--font-mono);
  }
  .live-badge {
    color: var(--status-info-fg);
    background: var(--status-info-bg);
    border: 1px solid var(--status-info-border);
    border-radius: var(--radius-full);
    padding: 0.125rem 0.5rem;
    font-size: var(--fs-xs);
    font-weight: var(--fw-medium);
    animation: pulse-glow 1.6s ease-in-out infinite;
  }

  .empty-log {
    color: var(--text-dim);
    font-style: italic;
    padding: var(--space-6) var(--space-5);
    text-align: center;
  }

  @keyframes pulse-glow {
    0%, 100% { opacity: 1; }
    50% { opacity: 0.55; }
  }

  .compare-select {
    background: var(--bg-panel);
    color: var(--text-muted);
    border: 1px solid var(--border-strong);
    padding: 0.5rem 0.75rem;
    border-radius: var(--radius-md);
    font: inherit;
    font-size: 0.875rem;
  }

  .empty-comments { color: var(--text-dim); font-size: 0.875rem; padding: var(--space-2) 0; }

  .artifact-list {
    list-style: none;
    padding: 0;
    margin: 0;
  }
  .artifact-list li {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    padding: 0.375rem 0;
    border-bottom: 1px solid var(--border);
    font-size: var(--fs-sm);
  }
  .artifact-list li:last-child { border-bottom: none; }
  .artifact-name {
    font-family: var(--font-mono);
    color: var(--link);
    flex: 1;
    word-break: break-all;
  }
  .artifact-size {
    color: var(--text-dim);
    font-family: var(--font-mono);
    font-size: var(--fs-xs);
    min-width: 60px;
    text-align: right;
  }
  .artifact-mime {
    color: var(--text-dim);
    font-family: var(--font-mono);
    font-size: var(--fs-xs);
  }

  .comment {
    padding: 0.625rem 0;
    border-bottom: 1px solid var(--border);
  }
  .comment:last-of-type { border-bottom: none; }
  .comment-head {
    display: flex;
    align-items: baseline;
    gap: var(--space-2);
    margin-bottom: 0.25rem;
    font-size: 0.8125rem;
  }
  .comment-author { color: var(--text); }
  .comment-time { color: var(--text-dim); }
  .comment-delete-form { margin: 0; margin-left: auto; }
  .comment-delete-btn { padding: 0.125rem 0.5rem; font-size: 0.6875rem; }

  .comment-body {
    white-space: pre-wrap;
    font-size: 0.875rem;
    color: var(--text);
  }
  .mention {
    color: var(--link);
    font-weight: 500;
  }

  .comment-form { margin-top: 0.875rem; }
  .comment-form-error {
    color: var(--danger-text);
    font-size: 0.8125rem;
    margin: 0.25rem 0 0 0;
  }
  .comment-form-actions {
    display: flex;
    justify-content: flex-end;
    margin-top: var(--space-2);
  }
</style>

<Breadcrumb segments={[
  { label: 'workspaces', href: '/' },
  { label: 'back to project', href: `/projects/${data.run.project_id}` },
  { label: `run ${data.run.id.slice(0, 8)}` }
]} />

<div class="hero">
  <div class="hero-main">
    <div class="hero-line">
      <StatusPill status={liveStatus} />
      <h1>{data.run.task_name} <span class="dim">({data.run.task_source})</span></h1>
      {#if data.run.timed_out}<Badge variant="warning" size="sm">timed out</Badge>{/if}
      {#if data.run.attempt > 1}<Badge variant="info" size="sm">attempt {data.run.attempt}</Badge>{/if}
      {#if isPendingApproval(liveStatus)}<Badge variant="warning" size="sm">awaiting approval</Badge>{/if}
    </div>
    <div class="hero-meta">
      {#if liveExitCode != null}
        <span class="hero-meta-item"><span class="hero-meta-label">Exit</span><strong>{liveExitCode}</strong></span>
        <span class="hero-sep">·</span>
      {/if}
      <span class="hero-meta-item"><span class="hero-meta-label">Duration</span><strong>{formatDuration(data.run.started_at, data.run.finished_at)}</strong></span>
      {#if data.run.commit_sha}
        <span class="hero-sep">·</span>
        <Tooltip text={data.run.commit_sha}>
          <span class="hero-meta-item"><span class="hero-meta-label">Commit</span><strong class="mono">{data.run.commit_sha.slice(0, 7)}</strong></span>
        </Tooltip>
      {/if}
      <span class="hero-sep">·</span>
      <span class="hero-meta-item"><span class="hero-meta-label">Started</span><strong><TimeAgo value={data.run.started_at} /></strong></span>
    </div>
  </div>
  <div class="actions">
    {#if isPendingApproval(liveStatus)}
      <form method="POST" action="?/approve" class="approve-form">
        <input type="hidden" name="approved" value="true" />
        <button type="submit">Approve</button>
      </form>
      <form method="POST" action="?/approve" class="inline-form">
        <input type="hidden" name="approved" value="false" />
        <button type="submit" class="danger">Reject</button>
      </form>
    {:else if !isTerminal(liveStatus)}
      <button type="button" class="ghost"
              title={notifyOnComplete ? 'Disable desktop notification on completion' : 'Notify me on this device when this run finishes'}
              aria-label="Toggle desktop notification"
              onclick={toggleNotify}>
        🔔 {notifyOnComplete ? 'On' : 'Off'}
      </button>
      <button type="button" class="danger" disabled={cancelling} onclick={cancel}>
        {cancelling ? 'Cancelling…' : 'Cancel run'}
      </button>
    {:else}
      <form method="POST" action="?/togglePin" class="inline-form">
        <input type="hidden" name="task_id" value={data.run.task_id} />
        <input type="hidden" name="pinned" value={String(data.isTaskPinned)} />
        <button type="submit" class="ghost"
                title={data.isTaskPinned ? 'Unpin this task from your dashboard' : 'Pin this task to your dashboard'}>
          {data.isTaskPinned ? '★ Pinned' : '☆ Pin task'}
        </button>
      </form>
      <form method="POST" action="?/rerun" class="inline-form">
        <button type="submit" title="Re-run against current repo HEAD">Re-run</button>
      </form>
      {#if data.run.commit_sha}
        <form method="POST" action="?/rerunPinned" class="inline-form">
          <button type="submit" class="ghost"
                  title={`Re-run pinned to commit ${data.run.commit_sha.slice(0, 12)}`}>
            Re-run @ {data.run.commit_sha.slice(0, 7)}
          </button>
        </form>
      {/if}
      {#if data.recentRuns.length > 0}
        <select class="compare-select"
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

<div class="layout">
  <div class="main-col">
    <section class="panel log-section">
      <div class="log-section-head">
        <h2>Log <span class="log-bytes">{liveLog.length.toLocaleString()} bytes</span></h2>
        {#if !isTerminal(liveStatus)}
          <span class="live-badge">live · {liveStatus}</span>
        {/if}
      </div>
      {#if liveLog}
        <LogViewer text={liveLog} autoScroll={!isTerminal(liveStatus)} />
      {:else if !isTerminal(liveStatus)}
        <div class="empty-log">Waiting for output…</div>
      {:else}
        <div class="empty-log">No output captured.</div>
      {/if}
    </section>

    {#if data.artifacts.length > 0}
      <Panel title="Artifacts">
        {#snippet actions()}
          <span style="color: var(--text-dim); font-size: var(--fs-xs); font-family: var(--font-mono);">{data.artifacts.length} file{data.artifacts.length === 1 ? '' : 's'}</span>
        {/snippet}
        <ul class="artifact-list">
          {#each data.artifacts as a (a.id)}
            <li>
              <a href={`/api/artifacts/${a.id}/download`} class="artifact-name">{a.relative_path}</a>
              <span class="artifact-size">{formatBytes(a.size_bytes)}</span>
              {#if a.mime_type}<span class="artifact-mime">{a.mime_type}</span>{/if}
            </li>
          {/each}
        </ul>
      </Panel>
    {/if}

    <Panel title="Comments">
      {#if data.comments.length === 0}
        <div class="empty-comments">No comments yet.</div>
      {:else}
        {#each data.comments as c (c.id)}
          <div class="comment">
            <div class="comment-head">
              <strong class="comment-author">{c.user_display_name || 'someone'}</strong>
              <span class="comment-time">{formatRelative(c.created_at)}</span>
              {#if data.user && c.user_id === data.user.id}
                <form method="POST" action="?/deleteComment" class="comment-delete-form">
                  <input type="hidden" name="id" value={c.id} />
                  <button type="submit" class="ghost comment-delete-btn" title="Delete">×</button>
                </form>
              {/if}
            </div>
            <div class="comment-body">
              {#each bodyParts(c.body) as part, i (i)}
                {#if part.mention}<span class="mention">{part.text}</span>
                {:else}<span>{part.text}</span>{/if}
              {/each}
            </div>
          </div>
        {/each}
      {/if}
      <form method="POST" action="?/comment" class="comment-form">
        <textarea name="body" rows="3"
                  placeholder="Add a comment. Use @display-name to mention a workspace member."
                  bind:value={commentDraft}></textarea>
        {#if form?.commentError}<p class="comment-form-error">{form.commentError}</p>{/if}
        <div class="comment-form-actions">
          <button type="submit">Post comment</button>
        </div>
      </form>
    </Panel>
  </div>

  <div class="meta-col">
    <Panel title="Run details">
      <div class="row"><span class="label">Status</span><span class="value"><StatusPill status={liveStatus} size="sm" /></span></div>
      <div class="row"><span class="label">Exit code</span><span class="value">{liveExitCode ?? '—'}</span></div>
      <div class="row"><span class="label">Commit</span><span class="value">
        {#if data.run.commit_sha}
          <Tooltip text={data.run.commit_sha}><span>{shortSha(data.run.commit_sha)}</span></Tooltip>
        {:else}—{/if}
      </span></div>
      <div class="row"><span class="label">Duration</span><span class="value">{formatDuration(data.run.started_at, data.run.finished_at)}</span></div>
      <div class="row"><span class="label">Started</span><span class="value"><TimeAgo value={data.run.started_at} /></span></div>
      <div class="row"><span class="label">Finished</span><span class="value"><TimeAgo value={data.run.finished_at} /></span></div>
      {#if data.run.params && (
        (data.run.params.env && Object.keys(data.run.params.env).length > 0) ||
        (data.run.params.args && data.run.params.args.length > 0)
      )}
        <div class="row">
          <span class="label">Inputs</span>
          <span class="value value-pre">{[
            ...(data.run.params.env ? Object.entries(data.run.params.env).map(([k, v]) => `${k}=${v}`) : []),
            ...(data.run.params.args ?? [])
          ].join('\n')}</span>
        </div>
      {/if}
    </Panel>
  </div>
</div>
