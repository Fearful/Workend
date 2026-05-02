<script lang="ts">
  let { data, form } = $props();
</script>

<style>
  h1 { font-size: 1.5rem; margin: 0 0 0.5rem 0; }
  .breadcrumb { color: #6b7280; font-size: 0.875rem; margin-bottom: 1.5rem; }
  h2 { font-size: 0.875rem; text-transform: uppercase; letter-spacing: 0.05em;
       color: #9ca3af; font-weight: 600; margin: 1.5rem 0 1rem 0; }

  .panel { background: #14181d; border: 1px solid #1f2429; border-radius: 8px;
           padding: 1.25rem 1.5rem; margin-bottom: 1rem; }
  .panel h2 { margin-top: 0; }

  .row {
    display: grid; grid-template-columns: auto 1fr 1fr auto auto;
    align-items: center; gap: 1rem;
    padding: 0.625rem 0; border-bottom: 1px solid #1f2429; font-size: 0.875rem;
  }
  .row:last-child { border-bottom: none; }
  .mono { font-family: ui-monospace, "SF Mono", Menlo, monospace; }
  .muted { color: #6b7280; font-size: 0.75rem; font-family: ui-monospace, "SF Mono", Menlo, monospace; }
  .kind-badge { padding: 0.125rem 0.5rem; border-radius: 4px; font-size: 0.6875rem;
                font-weight: 600; text-transform: uppercase; letter-spacing: 0.05em; }
  .kind-webhook { background: rgba(96, 165, 250, 0.15); color: #60a5fa; }
  .kind-slack { background: rgba(168, 85, 247, 0.15); color: #a855f7; }
  .kind-email { background: rgba(34, 197, 94, 0.15); color: #22c55e; }

  .form-grid { display: grid; grid-template-columns: 140px 1fr 160px auto;
               gap: 0.75rem; align-items: end; }
  select, input { padding: 0.5rem 0.75rem; background: #14181d; color: #e8eaed;
                  border: 1px solid #2d3540; border-radius: 6px; font-family: inherit;
                  font-size: 0.875rem; width: 100%; box-sizing: border-box; }
  .empty { color: #6b7280; text-align: center; padding: 1.5rem; font-size: 0.875rem; }
  .error { color: #ef4444; font-size: 0.875rem; margin-top: 0.5rem; }
  .success { color: #22c55e; font-size: 0.875rem; margin-top: 0.5rem; }
</style>

<div class="breadcrumb"><a href="/settings">settings</a> / notifications</div>

<h1>Notifications</h1>

<section class="panel">
  <h2>Configured targets</h2>
  {#if data.notifications.length === 0}
    <div class="empty">No notification targets yet.</div>
  {:else}
    {#each data.notifications as n (n.id)}
      <div class="row">
        <span class="kind-badge kind-{n.kind}">{n.kind}</span>
        <span class="mono">{n.target}</span>
        <span class="muted">trigger: {n.trigger}</span>
        <form method="POST" action="?/test" style="margin: 0;">
          <input type="hidden" name="id" value={n.id} />
          <button type="submit" class="ghost">Test</button>
        </form>
        <form method="POST" action="?/delete" style="margin: 0;" onsubmit={(e) => !confirm('Delete this notification target?') && e.preventDefault()}>
          <input type="hidden" name="id" value={n.id} />
          <button type="submit" class="ghost">Delete</button>
        </form>
      </div>
    {/each}
  {/if}
</section>

<section class="panel">
  <h2>Add notification target</h2>
  <form method="POST" action="?/create">
    <div class="form-grid">
      <div class="field">
        <label for="kind">Kind</label>
        <select id="kind" name="kind">
          <option value="webhook">Webhook</option>
          <option value="slack">Slack webhook</option>
          <option value="email">Email</option>
        </select>
      </div>
      <div class="field">
        <label for="target">Target</label>
        <input id="target" name="target" type="text" required placeholder="https://hooks.slack.com/... or alice@example.com" value={form?.target || ''} />
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
    {#if form?.error}<p class="error">{form.error}</p>{/if}
    {#if form?.tested}<p class="success">Test notification sent.</p>{/if}
  </form>
</section>
