<script lang="ts">
  import { invalidateAll } from '$app/navigation';
  import Panel from '$lib/components/Panel.svelte';
  import StatusPill from '$lib/components/StatusPill.svelte';
  import Modal from '$lib/components/Modal.svelte';
  import TimeAgo from '$lib/components/TimeAgo.svelte';
  import FlashMessage from '$lib/components/FlashMessage.svelte';
  import { formatDuration } from '$lib/utils';

  let { data, form } = $props();

  let syncPending = $state(false);
  let detailModal = $state<{
    sha: string;
    loading: boolean;
    commit: {
      sha: string; author: string; author_email: string; message: string;
      committed_at: string; files_changed: number; insertions: number; deletions: number;
    } | null;
    runs: { id: string; status: string; started_at: string | null; finished_at: string | null; duration_ms: number }[];
    error: string | null;
  } | null>(null);

  async function syncCommits() {
    if (syncPending) return;
    syncPending = true;
    try {
      const r = await fetch(`/projects/${data.project.id}/blame?/syncCommits`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
        body: ''
      });
      if (!r.ok) return;
      await invalidateAll();
    } finally {
      syncPending = false;
    }
  }

  async function openDetail(sha: string) {
    detailModal = { sha, loading: true, commit: null, runs: [], error: null };
    try {
      const r = await fetch(`/api/commits/${sha}`, { credentials: 'same-origin' });
      if (!r.ok) {
        if (detailModal) detailModal = { ...detailModal, loading: false, error: `HTTP ${r.status}` };
        return;
      }
      const json = await r.json() as { commit: typeof detailModal.commit; runs: typeof detailModal.runs };
      if (detailModal) {
        detailModal = { ...detailModal, loading: false, commit: json.commit, runs: json.runs ?? [] };
      }
    } catch (err) {
      if (detailModal) {
        detailModal = { ...detailModal, loading: false, error: err instanceof Error ? err.message : 'failed' };
      }
    }
  }

  function rowClass(c: { run_count: number; fail_count: number }): string {
    if (c.run_count === 0) return '';
    if (c.fail_count > 0) return 'row-fail';
    return 'row-pass';
  }
</script>

<style>
  .page-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: var(--space-4);
  }
  h2 { margin: 0; font-size: var(--fs-lg); font-weight: var(--fw-semibold); }

  .table-wrap {
    overflow-x: auto;
  }
  table {
    width: 100%;
    border-collapse: collapse;
    font-size: 0.8125rem;
  }
  th {
    text-align: left;
    color: var(--text-dim);
    font-size: 0.6875rem;
    text-transform: uppercase;
    letter-spacing: 0.04em;
    padding: 0.5rem 0.625rem;
    border-bottom: 2px solid var(--border);
    white-space: nowrap;
  }
  th.num { text-align: right; }
  td {
    padding: 0.625rem;
    border-bottom: 1px solid var(--border);
    white-space: nowrap;
  }
  td.num { text-align: right; }

  tr.clickable { cursor: pointer; }
  tr.clickable:hover { background: var(--bg-hover); }
  tr.row-pass { border-left: 3px solid var(--success); }
  tr.row-fail { border-left: 3px solid var(--danger-text); }

  .sha { font-family: var(--font-mono); }
  .author { max-width: 140px; overflow: hidden; text-overflow: ellipsis; }
  .msg { max-width: 300px; overflow: hidden; text-overflow: ellipsis; color: var(--text); }
  .date { color: var(--text-dim); font-size: 0.75rem; }
  .ins { color: var(--success); }
  .del { color: var(--danger-text); }
  .runs-cell { display: inline-flex; align-items: center; gap: 0.25rem; }
  .pass-count { color: var(--success); font-size: 0.6875rem; }
  .fail-count { color: var(--danger-text); font-size: 0.6875rem; }
  .dim { color: var(--text-dim); }

  .empty {
    text-align: center;
    padding: 3rem var(--space-4);
    color: var(--text-muted);
  }

  .detail-row {
    display: grid;
    grid-template-columns: 100px 1fr;
    gap: var(--space-3);
    padding: var(--space-2) 0;
    border-bottom: 1px solid var(--border);
    font-size: 0.875rem;
  }
  .detail-row:last-child { border-bottom: none; }
  .detail-label { color: var(--text-dim); }
  .detail-value { color: var(--text); font-family: var(--font-mono); word-break: break-all; }
  .detail-msg { font-family: inherit; white-space: pre-wrap; }

  .run-link {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    padding: var(--space-1) 0;
    color: inherit;
    text-decoration: none;
    font-size: 0.875rem;
  }
  .run-link:hover { background: var(--bg-hover); }
  .run-meta { color: var(--text-dim); font-size: 0.75rem; font-family: var(--font-mono); }

  .modal-actions {
    display: flex;
    gap: var(--space-2);
    justify-content: flex-end;
    margin-top: var(--space-4);
  }
  .modal-error { color: var(--danger-text); font-size: 0.875rem; }
</style>

{#if form?.error}<FlashMessage type="error">{form.error}</FlashMessage>{/if}

<div class="page-header">
  <h2>Blame timeline</h2>
  <button type="button" onclick={syncCommits} disabled={syncPending}>
    {syncPending ? 'Syncing...' : 'Sync commits'}
  </button>
</div>

{#if data.blameTimeline.length === 0}
  <div class="empty">
    <p>No commits synced yet. Click "Sync commits" to pull commit history and correlate with runs.</p>
  </div>
{:else}
  <div class="table-wrap">
    <table>
      <thead>
        <tr>
          <th>SHA</th>
          <th>Author</th>
          <th>Message</th>
          <th>Date</th>
          <th class="num">Files</th>
          <th class="num">+/-</th>
          <th class="num">Runs</th>
          <th class="num">Avg</th>
        </tr>
      </thead>
      <tbody>
        {#each data.blameTimeline as c (c.sha)}
          <tr class="clickable {rowClass(c)}" onclick={() => openDetail(c.sha)} role="button" tabindex="0">
            <td class="sha">{c.sha.slice(0, 8)}</td>
            <td class="author">{c.author}</td>
            <td class="msg">{c.message.split('\n')[0].slice(0, 80)}</td>
            <td class="date"><TimeAgo value={c.committed_at} /></td>
            <td class="num">{c.files_changed}</td>
            <td class="num"><span class="ins">+{c.insertions}</span> <span class="del">-{c.deletions}</span></td>
            <td class="num">
              {#if c.run_count > 0}
                <span class="runs-cell">
                  {c.run_count}
                  {#if c.pass_count > 0}<span class="pass-count">{c.pass_count}p</span>{/if}
                  {#if c.fail_count > 0}<span class="fail-count">{c.fail_count}f</span>{/if}
                </span>
              {:else}
                <span class="dim">--</span>
              {/if}
            </td>
            <td class="num">
              {#if c.avg_duration_ms > 0}
                {c.avg_duration_ms < 1000 ? `${Math.round(c.avg_duration_ms)}ms` : `${(c.avg_duration_ms / 1000).toFixed(1)}s`}
              {:else}
                <span class="dim">--</span>
              {/if}
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
{/if}

<Modal open={detailModal !== null} title={detailModal ? `Commit ${detailModal.sha.slice(0, 8)}` : ''} width={640} onClose={() => detailModal = null}>
  {#if detailModal}
    {#if detailModal.loading}
      <div style="color: var(--text-dim); font-style: italic; padding: var(--space-3) 0;">Loading...</div>
    {:else if detailModal.error}
      <p class="modal-error">{detailModal.error}</p>
    {:else if detailModal.commit}
      <div class="detail-row"><span class="detail-label">SHA</span><span class="detail-value">{detailModal.commit.sha}</span></div>
      <div class="detail-row"><span class="detail-label">Author</span><span class="detail-value detail-msg">{detailModal.commit.author} &lt;{detailModal.commit.author_email}&gt;</span></div>
      <div class="detail-row"><span class="detail-label">Date</span><span class="detail-value">{new Date(detailModal.commit.committed_at).toLocaleString()}</span></div>
      <div class="detail-row"><span class="detail-label">Message</span><span class="detail-value detail-msg">{detailModal.commit.message}</span></div>
      <div class="detail-row"><span class="detail-label">Changes</span><span class="detail-value">{detailModal.commit.files_changed} files, <span class="ins">+{detailModal.commit.insertions}</span> <span class="del">-{detailModal.commit.deletions}</span></span></div>
      {#if detailModal.runs.length > 0}
        <h3 style="margin: var(--space-4) 0 var(--space-2) 0; font-size: 0.875rem; color: var(--text-muted);">Associated runs ({detailModal.runs.length})</h3>
        {#each detailModal.runs as r (r.id)}
          <a href={`/runs/${r.id}`} class="run-link">
            <StatusPill status={r.status} size="sm" />
            <span class="run-meta">{formatDuration(r.started_at, r.finished_at)}</span>
            <span class="run-meta">{r.started_at ? new Date(r.started_at).toLocaleString() : '--'}</span>
          </a>
        {/each}
      {:else}
        <p style="color: var(--text-dim); font-size: 0.875rem; margin-top: var(--space-3);">No runs associated with this commit.</p>
      {/if}
    {/if}
    <div class="modal-actions">
      <button type="button" class="ghost" onclick={() => (detailModal = null)}>Close</button>
    </div>
  {/if}
</Modal>
