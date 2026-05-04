<script lang="ts">
  import { onMount, tick } from 'svelte';

  let {
    text,
    height = 480,
    autoScroll = false
  }: {
    text: string;
    height?: number;
    autoScroll?: boolean;
  } = $props();

  let lines = $derived(text.split('\n'));
  let query = $state('');
  let wrap = $state(true);
  let follow = $state(false);
  let bottomEl: HTMLDivElement | undefined = $state();
  let copied = $state(false);

  onMount(() => {
    follow = autoScroll;
  });

  let matches = $derived.by(() => {
    if (!query) return new Set<number>();
    const q = query.toLowerCase();
    const out = new Set<number>();
    lines.forEach((l, i) => { if (l.toLowerCase().includes(q)) out.add(i); });
    return out;
  });

  let firstMatch = $derived.by(() => {
    if (matches.size === 0) return -1;
    return Math.min(...Array.from(matches));
  });

  $effect(() => {
    if (follow && bottomEl) {
      tick().then(() => bottomEl?.scrollIntoView({ block: 'end' }));
    }
  });

  $effect(() => {
    if (firstMatch >= 0) {
      tick().then(() => {
        const el = document.querySelector(`[data-line="${firstMatch}"]`);
        el?.scrollIntoView({ block: 'center', behavior: 'smooth' });
      });
    }
  });

  async function copyAll() {
    try {
      await navigator.clipboard.writeText(text);
      copied = true;
      setTimeout(() => { copied = false; }, 1500);
    } catch { /* ignore */ }
  }

  function highlight(line: string): string {
    if (!query) return escape(line);
    const q = query;
    const lc = line.toLowerCase();
    const qlc = q.toLowerCase();
    let out = '';
    let i = 0;
    while (i < line.length) {
      const idx = lc.indexOf(qlc, i);
      if (idx < 0) {
        out += escape(line.slice(i));
        break;
      }
      out += escape(line.slice(i, idx));
      out += `<mark>${escape(line.slice(idx, idx + q.length))}</mark>`;
      i = idx + q.length;
    }
    return out;
  }

  function escape(s: string): string {
    return s.replace(/[&<>"']/g, (c) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c]!));
  }

  onMount(() => {
    if (autoScroll && bottomEl) bottomEl.scrollIntoView({ block: 'end' });
  });
</script>

<div class="log-viewer">
  <div class="toolbar">
    <input type="search" placeholder="Search lines…" bind:value={query} class="search" />
    {#if matches.size > 0}<span class="hits">{matches.size} match{matches.size === 1 ? '' : 'es'}</span>{/if}
    <span class="spacer"></span>
    <label class="toggle">
      <input type="checkbox" bind:checked={wrap} />
      Wrap
    </label>
    <label class="toggle">
      <input type="checkbox" bind:checked={follow} />
      Follow
    </label>
    <button type="button" class="ghost copy-btn" onclick={copyAll}>{copied ? 'Copied' : 'Copy'}</button>
  </div>
  <div class="body" style="max-height:{height}px" class:wrap>
    {#each lines as line, i (i)}
      <div class="line" class:match={matches.has(i)} data-line={i}>
        <span class="ln">{i + 1}</span>
        <pre class="content">{@html highlight(line || ' ')}</pre>
      </div>
    {/each}
    <div bind:this={bottomEl}></div>
  </div>
</div>

<style>
  .log-viewer {
    background: var(--bg-page);
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    overflow: hidden;
    font-family: var(--font-mono);
    font-size: var(--fs-xs);
  }
  .toolbar {
    display: flex;
    gap: var(--space-2);
    align-items: center;
    padding: var(--space-2) var(--space-3);
    background: var(--bg-panel);
    border-bottom: 1px solid var(--border);
    flex-wrap: wrap;
  }
  .search {
    flex: 1;
    min-width: 160px;
    max-width: 280px;
    padding: 0.25rem 0.5rem;
    font-size: var(--fs-xs);
  }
  .hits {
    color: var(--text-dim);
    font-size: var(--fs-xs);
  }
  .spacer { flex: 1; }
  .toggle {
    display: inline-flex;
    align-items: center;
    gap: 0.25rem;
    color: var(--text-dim);
    font-size: var(--fs-xs);
    margin: 0;
    cursor: pointer;
  }
  .toggle input { width: auto; margin: 0; }
  .copy-btn {
    padding: 0.25rem 0.625rem;
    font-size: var(--fs-xs);
  }
  .body {
    overflow: auto;
    padding: var(--space-2) 0;
  }
  .body.wrap .content {
    white-space: pre-wrap;
    word-break: break-word;
  }
  .body:not(.wrap) .content {
    white-space: pre;
  }
  .line {
    display: grid;
    grid-template-columns: 56px 1fr;
    gap: 0;
    padding: 0 var(--space-3);
  }
  .line:hover { background: var(--bg-hover); }
  .line.match { background: rgba(234, 179, 8, 0.08); }
  .ln {
    color: var(--text-dim);
    user-select: none;
    text-align: right;
    padding-right: var(--space-3);
    border-right: 1px solid var(--border);
  }
  .content {
    margin: 0;
    padding-left: var(--space-3);
    color: var(--text);
    font-family: inherit;
    font-size: inherit;
    line-height: var(--lh-normal);
  }
  .content :global(mark) {
    background: rgba(234, 179, 8, 0.45);
    color: var(--text);
    padding: 0 1px;
    border-radius: 2px;
  }
</style>
