<script lang="ts">
  import { formatRelative } from '$lib/utils';
  import Panel from '$lib/components/Panel.svelte';
  import Badge from '$lib/components/Badge.svelte';
  import FlashMessage from '$lib/components/FlashMessage.svelte';

  let { data, form } = $props();
  let draft = $state('');
  $effect(() => {
    if (form?.draft) draft = form.draft;
  });
</script>

<style>
  h1 { font-size: 1.5rem; margin: 0; line-height: 1.3; font-weight: 600; letter-spacing: -0.02em; }
  .num { color: var(--text-dim); font-family: var(--font-mono); font-size: 0.875rem; }

  .back-button {
    background: transparent;
    color: var(--link);
    border: none;
    padding: 0;
    font: inherit;
    cursor: pointer;
    margin-bottom: var(--space-2);
    font-size: 0.875rem;
  }
  .back-button:hover { text-decoration: underline; background: transparent; color: var(--link); }

  .header {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: var(--space-4);
    margin-bottom: var(--space-4);
  }
  .meta {
    display: flex;
    gap: var(--space-2);
    flex-wrap: wrap;
    align-items: center;
    color: var(--text-dim);
    font-size: 0.8125rem;
    margin-top: var(--space-2);
  }
  .author { color: var(--text); font-weight: 500; }
  .ext-link { margin-left: var(--space-2); }

  .body {
    color: var(--text);
    font-size: 0.9rem;
    line-height: 1.55;
    white-space: pre-wrap;
    word-wrap: break-word;
  }
  .comment {
    padding: var(--space-3) 0;
    border-bottom: 1px solid var(--border);
  }
  .comment:last-child { border-bottom: none; }
  .comment-head {
    display: flex;
    align-items: baseline;
    gap: var(--space-2);
    font-size: 0.8125rem;
    color: var(--text-muted);
    margin-bottom: var(--space-1);
  }
  .comment-actions { margin-left: auto; }

  .post-actions {
    display: flex;
    justify-content: flex-end;
    margin-top: var(--space-2);
  }

  @media (max-width: 640px) {
    .header { flex-direction: column; align-items: stretch; }
  }
</style>

<button type="button" onclick={() => history.back()} class="back-button">← back</button>

<div class="header">
  <div>
    <span class="num">#{data.issue.provider_number}</span>
    <h1>{data.issue.title}</h1>
    <div class="meta">
      <Badge variant={data.issue.state === 'open' ? 'success' : 'info'} size="sm">{data.issue.state}</Badge>
      {#if data.issue.author_handle}
        <span>opened by <span class="author">{data.issue.author_handle}</span></span>
      {/if}
      {#if data.issue.upstream_updated_at}
        <span>· updated {formatRelative(data.issue.upstream_updated_at)}</span>
      {/if}
      <a href={data.issue.html_url} target="_blank" rel="noopener" class="ext-link">view on provider →</a>
    </div>
    {#if data.issue.labels.length > 0}
      <div class="meta">
        {#each data.issue.labels as l (l)}
          <Badge variant="muted" size="sm">{l}</Badge>
        {/each}
      </div>
    {/if}
  </div>
  {#if data.issue.state === 'open'}
    <form method="POST" action="?/close" class="inline-form" onsubmit={(e) => !confirm('Close this issue upstream?') && e.preventDefault()}>
      <button type="submit" class="danger">Close issue</button>
    </form>
  {/if}
</div>

{#if form?.closeError}<FlashMessage type="error">{form.closeError}</FlashMessage>{/if}
{#if form?.closed}<FlashMessage type="success">Issue closed.</FlashMessage>{/if}

{#if data.issue.body}
  <Panel>
    <div class="body">{data.issue.body}</div>
  </Panel>
{/if}

<Panel title="Comments">
  {#if data.issue.comments.length === 0}
    <p style="color: var(--text-dim); font-size: 0.875rem; margin: 0;">No comments yet.</p>
  {:else}
    {#each data.issue.comments as c (c.id)}
      <div class="comment">
        <div class="comment-head">
          <span class="author">{c.author_handle || 'someone'}</span>
          <span>{formatRelative(c.created_at)}</span>
          {#if c.html_url}
            <a href={c.html_url} target="_blank" rel="noopener" class="comment-actions">view</a>
          {/if}
        </div>
        <div class="body">{c.body}</div>
      </div>
    {/each}
  {/if}

  <form method="POST" action="?/comment" style="margin-top: var(--space-4);">
    <textarea name="body" rows="4" placeholder="Add a comment…" bind:value={draft}></textarea>
    {#if form?.commentError}<FlashMessage type="error">{form.commentError}</FlashMessage>{/if}
    {#if form?.commented}<FlashMessage type="success">Comment posted.</FlashMessage>{/if}
    <div class="post-actions">
      <button type="submit">Post comment</button>
    </div>
  </form>
</Panel>
