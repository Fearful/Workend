<script lang="ts">
  import { invalidateAll } from '$app/navigation';
  import Breadcrumb from '$lib/components/Breadcrumb.svelte';
  import Panel from '$lib/components/Panel.svelte';
  import StatusPill from '$lib/components/StatusPill.svelte';
  import Badge from '$lib/components/Badge.svelte';
  import Modal from '$lib/components/Modal.svelte';
  import TimeAgo from '$lib/components/TimeAgo.svelte';
  import FlashMessage from '$lib/components/FlashMessage.svelte';

  let { data, form } = $props();

  let showExtend = $state(false);
  let showDestroy = $state(false);
  let commandInput = $state('');
  let activeSessionId = $state<string | null>(null);

  interface TermEntry {
    command: string;
    output: string;
    exitCode: number;
  }

  let termHistory = $state<TermEntry[]>([]);

  function isActive(status: string): boolean {
    return status === 'running' || status === 'starting' || status === 'ready';
  }

  function timeRemaining(expiresAt: string): string {
    const diff = new Date(expiresAt).getTime() - Date.now();
    if (diff <= 0) return 'expired';
    const hours = Math.floor(diff / 3_600_000);
    const minutes = Math.floor((diff % 3_600_000) / 60_000);
    if (hours > 0) return `${hours}h ${minutes}m remaining`;
    return `${minutes}m remaining`;
  }

  // Pick the active shell session (open status)
  let openSessions = $derived(data.shellSessions.filter((s: { status: string }) => s.status === 'open' || s.status === 'active'));

  $effect(() => {
    if (form?.shellCreated && form.sessionId) {
      activeSessionId = form.sessionId;
    }
  });

  $effect(() => {
    if (form?.execResult && form.command) {
      termHistory = [
        ...termHistory,
        {
          command: form.command,
          output: form.execResult.output,
          exitCode: form.execResult.exit_code
        }
      ];
      commandInput = '';
    }
  });

  // Auto-select first open session if none is active
  $effect(() => {
    if (!activeSessionId && openSessions.length > 0) {
      activeSessionId = openSessions[0].id;
    }
  });

  function scrollTerminal() {
    requestAnimationFrame(() => {
      const el = document.querySelector('.term-output');
      if (el) el.scrollTop = el.scrollHeight;
    });
  }

  $effect(() => {
    if (termHistory.length > 0) scrollTerminal();
  });
</script>

<style>
  .hero {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: var(--space-4);
    margin-bottom: var(--space-5);
    flex-wrap: wrap;
  }

  .hero-main {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
    min-width: 0;
    flex: 1;
  }

  .hero-line {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: var(--space-2);
  }

  h1 {
    font-size: var(--fs-xl);
    margin: 0;
    font-weight: var(--fw-semibold);
    letter-spacing: -0.02em;
    font-family: var(--font-mono);
  }

  .hero-meta {
    display: flex;
    flex-wrap: wrap;
    align-items: baseline;
    gap: var(--space-2);
    color: var(--text-dim);
    font-size: var(--fs-sm);
  }

  .hero-sep { color: var(--border-strong); user-select: none; }

  .actions {
    display: flex;
    gap: var(--space-2);
    flex-shrink: 0;
    flex-wrap: wrap;
  }

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

  .countdown {
    font-family: var(--font-mono);
    font-size: var(--fs-sm);
  }

  .shell-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-2);
    margin-bottom: var(--space-3);
    flex-wrap: wrap;
  }

  .shell-tabs {
    display: flex;
    gap: var(--space-1);
    align-items: center;
    flex-wrap: wrap;
  }

  .shell-tab {
    padding: 0.25rem 0.625rem;
    font-size: var(--fs-xs);
    font-family: var(--font-mono);
    border-radius: var(--radius-md);
    background: transparent;
    color: var(--text-muted);
    border: 1px solid var(--border-strong);
    cursor: pointer;
  }
  .shell-tab:hover { background: var(--bg-hover); color: var(--text); }
  .shell-tab.active {
    background: var(--accent);
    color: white;
    border-color: var(--accent);
  }

  .term-output {
    background: #0d0f12;
    border: 1px solid var(--border);
    border-radius: var(--radius-md) var(--radius-md) 0 0;
    padding: var(--space-3) var(--space-4);
    min-height: 240px;
    max-height: 480px;
    overflow-y: auto;
    font-family: var(--font-mono);
    font-size: var(--fs-sm);
    line-height: 1.6;
    color: var(--text);
  }

  .term-entry { margin-bottom: var(--space-2); }

  .term-prompt {
    color: var(--success);
    user-select: none;
  }

  .term-cmd { color: var(--text); }

  .term-result {
    white-space: pre-wrap;
    color: var(--text-muted);
  }

  .term-exit-error {
    color: var(--danger-text);
    font-size: var(--fs-xs);
  }

  .term-empty {
    color: var(--text-dim);
    font-style: italic;
  }

  .term-input-row {
    display: flex;
    border: 1px solid var(--border);
    border-top: none;
    border-radius: 0 0 var(--radius-md) var(--radius-md);
    background: var(--bg-input);
    overflow: hidden;
  }

  .term-input-row input {
    flex: 1;
    border: none;
    border-radius: 0;
    font-family: var(--font-mono);
    font-size: var(--fs-sm);
    padding: var(--space-2) var(--space-3);
    background: transparent;
  }
  .term-input-row input:focus { outline: none; }

  .term-input-row button {
    border-radius: 0;
    padding: var(--space-2) var(--space-4);
    font-size: var(--fs-sm);
    flex-shrink: 0;
  }

  .session-list {
    margin-top: var(--space-3);
  }

  .session-item {
    display: grid;
    grid-template-columns: 1fr auto auto;
    gap: var(--space-2);
    align-items: center;
    padding: var(--space-2) 0;
    border-bottom: 1px solid var(--border);
    font-size: var(--fs-sm);
  }
  .session-item:last-child { border-bottom: none; }

  .session-id {
    font-family: var(--font-mono);
    font-size: var(--fs-xs);
    color: var(--text);
  }

  .modal-actions {
    display: flex;
    gap: var(--space-2);
    justify-content: flex-end;
    margin-top: var(--space-4);
  }

  .modal-hint {
    color: var(--text-dim);
    font-size: 0.75rem;
    margin: 0.25rem 0 0 0;
  }

  .confirm-text {
    color: var(--text-muted);
    font-size: var(--fs-sm);
    margin: 0 0 var(--space-3) 0;
  }

  .no-shell {
    color: var(--text-dim);
    font-style: italic;
    font-size: var(--fs-sm);
    padding: var(--space-4) 0;
    text-align: center;
  }
</style>

{#if form?.error}<FlashMessage type="error">{form.error}</FlashMessage>{/if}

<Breadcrumb segments={[
  { label: 'sandboxes', href: '/sandboxes' },
  { label: `${data.sandbox.branch || 'main'} (${data.sandbox.id.slice(0, 8)})` }
]} />

<div class="hero">
  <div class="hero-main">
    <div class="hero-line">
      <StatusPill status={data.sandbox.status} />
      <h1>{data.sandbox.branch || 'main'}</h1>
    </div>
    <div class="hero-meta">
      <span>Project <a href={`/projects/${data.sandbox.project_id}`}>{data.sandbox.project_id.slice(0, 8)}</a></span>
      <span class="hero-sep">*</span>
      <span class="countdown">{timeRemaining(data.sandbox.expires_at)}</span>
      <span class="hero-sep">*</span>
      <span>Created <TimeAgo value={data.sandbox.created_at} /></span>
    </div>
  </div>
  <div class="actions">
    {#if isActive(data.sandbox.status)}
      <button type="button" class="ghost" onclick={() => (showExtend = true)}>Extend</button>
      <button type="button" class="danger" onclick={() => (showDestroy = true)}>Destroy</button>
    {/if}
  </div>
</div>

<div class="layout">
  <div class="main-col">
    <!-- Shell Section -->
    <Panel title="Shell" padding="compact">
      {#if !isActive(data.sandbox.status)}
        <div class="no-shell">Shell is unavailable. Sandbox is {data.sandbox.status}.</div>
      {:else if openSessions.length === 0 && !activeSessionId}
        <div class="no-shell">
          No active shell session.
          <form method="POST" action="?/createShell" class="inline-form" style="display: inline;">
            <button type="submit">Create session</button>
          </form>
        </div>
      {:else}
        <div class="shell-header">
          <div class="shell-tabs">
            {#each openSessions as sess (sess.id)}
              <button type="button"
                      class="shell-tab"
                      class:active={activeSessionId === sess.id}
                      onclick={() => { activeSessionId = sess.id; }}>
                {sess.id.slice(0, 8)}
              </button>
            {/each}
            <form method="POST" action="?/createShell" class="inline-form" style="display: inline;">
              <button type="submit" class="shell-tab" title="New session">+</button>
            </form>
          </div>
          {#if activeSessionId}
            <form method="POST" action="?/closeShell" class="inline-form">
              <input type="hidden" name="session_id" value={activeSessionId} />
              <button type="submit" class="ghost" style="padding: 0.25rem 0.5rem; font-size: var(--fs-xs);">Close session</button>
            </form>
          {/if}
        </div>

        <div class="term-output">
          {#if termHistory.length === 0}
            <div class="term-empty">Ready. Type a command below.</div>
          {:else}
            {#each termHistory as entry, i (i)}
              <div class="term-entry">
                <div><span class="term-prompt">$ </span><span class="term-cmd">{entry.command}</span></div>
                {#if entry.output}<div class="term-result">{entry.output}</div>{/if}
                {#if entry.exitCode !== 0}
                  <div class="term-exit-error">exit {entry.exitCode}</div>
                {/if}
              </div>
            {/each}
          {/if}
        </div>

        {#if activeSessionId}
          <form method="POST" action="?/exec" class="term-input-row"
                onsubmit={() => { /* form will submit, $effect handles result */ }}>
            <input type="hidden" name="session_id" value={activeSessionId} />
            <input name="command" type="text" placeholder="Enter command..."
                   autocomplete="off" spellcheck="false"
                   bind:value={commandInput} />
            <button type="submit" disabled={!commandInput.trim()}>Run</button>
          </form>
        {/if}
      {/if}
    </Panel>

    <!-- Session List -->
    {#if data.shellSessions.length > 0}
      <Panel title="Shell Sessions">
        <div class="session-list">
          {#each data.shellSessions as sess (sess.id)}
            <div class="session-item">
              <span class="session-id">{sess.id.slice(0, 12)}</span>
              <StatusPill status={sess.status === 'open' || sess.status === 'active' ? 'running' : 'succeeded'} size="sm" label={sess.status} />
              <TimeAgo value={sess.created_at} />
            </div>
          {/each}
        </div>
      </Panel>
    {/if}
  </div>

  <div class="meta-col">
    <Panel title="Sandbox details">
      <div class="row"><span class="label">Status</span><span class="value"><StatusPill status={data.sandbox.status} size="sm" /></span></div>
      <div class="row"><span class="label">URL</span><span class="value">
        {#if data.sandbox.url && isActive(data.sandbox.status)}
          <a href={data.sandbox.url} target="_blank" rel="noopener noreferrer">{data.sandbox.url}</a>
        {:else}
          {data.sandbox.url || '---'}
        {/if}
      </span></div>
      <div class="row"><span class="label">Port</span><span class="value">{data.sandbox.port || '---'}</span></div>
      <div class="row"><span class="label">Container</span><span class="value">{data.sandbox.container_id ? data.sandbox.container_id.slice(0, 12) : '---'}</span></div>
      <div class="row"><span class="label">Branch</span><span class="value">{data.sandbox.branch || 'main'}</span></div>
      <div class="row"><span class="label">Created</span><span class="value"><TimeAgo value={data.sandbox.created_at} /></span></div>
      <div class="row"><span class="label">Expires</span><span class="value"><TimeAgo value={data.sandbox.expires_at} /></span></div>
      {#if data.sandbox.destroyed_at}
        <div class="row"><span class="label">Destroyed</span><span class="value"><TimeAgo value={data.sandbox.destroyed_at} /></span></div>
      {/if}
    </Panel>
  </div>
</div>

<!-- Extend Modal -->
<Modal open={showExtend} title="Extend Sandbox" width={400} onClose={() => (showExtend = false)}>
  <form method="POST" action="?/extend" onsubmit={() => { showExtend = false; }}>
    <div class="field">
      <label for="ext-hours">Additional hours</label>
      <input id="ext-hours" name="hours" type="number" min="1" max="72" value="2" />
      <p class="modal-hint">1-72 hours to add to the current expiry.</p>
    </div>
    <div class="modal-actions">
      <button type="button" class="ghost" onclick={() => (showExtend = false)}>Cancel</button>
      <button type="submit">Extend</button>
    </div>
  </form>
</Modal>

<!-- Destroy Modal -->
<Modal open={showDestroy} title="Destroy Sandbox" width={400} onClose={() => (showDestroy = false)}>
  <form method="POST" action="?/destroy">
    <p class="confirm-text">
      This will permanently destroy this sandbox. All data inside the container will be lost.
    </p>
    <div class="modal-actions">
      <button type="button" class="ghost" onclick={() => (showDestroy = false)}>Cancel</button>
      <button type="submit" class="danger">Destroy</button>
    </div>
  </form>
</Modal>
