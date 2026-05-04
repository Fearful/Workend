<script lang="ts">
  import { goto } from '$app/navigation';
  import { formatRelative } from '$lib/utils';
  import Breadcrumb from '$lib/components/Breadcrumb.svelte';
  import PageHeader from '$lib/components/PageHeader.svelte';
  import FlashMessage from '$lib/components/FlashMessage.svelte';

  let { data, form } = $props();

  let nameInput: HTMLInputElement | null = $state(null);
  let urlInput: HTMLInputElement | null = $state(null);
  let branchInput: HTMLInputElement | null = $state(null);

  function pickRepo(name: string, cloneURL: string, defaultBranch: string) {
    if (nameInput) nameInput.value = name;
    if (urlInput) urlInput.value = cloneURL;
    if (branchInput) branchInput.value = defaultBranch || '';
    nameInput?.scrollIntoView({ behavior: 'smooth', block: 'center' });
  }

  function providerLabel(kind: string): string {
    switch (kind) {
      case 'github': return 'GitHub';
      case 'gitlab': return 'GitLab';
      case 'gitea': return 'Gitea';
      default: return kind;
    }
  }

  function changeProvider(providerID: string) {
    const u = new URL(window.location.href);
    if (providerID) u.searchParams.set('from', providerID);
    else u.searchParams.delete('from');
    u.searchParams.delete('page');
    goto(u, { keepFocus: false });
  }

  function pageURL(p: number): string {
    const u = new URL(window.location.href);
    u.searchParams.set('page', String(p));
    return u.pathname + u.search;
  }

  let filter = $state('');
  let visibleRepos = $derived.by(() => {
    if (!filter.trim()) return data.repos;
    const q = filter.toLowerCase();
    return data.repos.filter(
      (r) => r.full_name.toLowerCase().includes(q) || r.description.toLowerCase().includes(q)
    );
  });
</script>

<style>
  h2 {
    font-size: 0.875rem;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: var(--text-muted);
    font-weight: 600;
    margin: 0 0 var(--space-4) 0;
  }

  .layout {
    display: grid;
    grid-template-columns: 480px 1fr;
    gap: var(--space-8);
    align-items: start;
    margin-top: var(--space-6);
  }
  @media (max-width: 900px) { .layout { grid-template-columns: 1fr; } }

  form { margin: 0; }

  .panel {
    background: var(--bg-panel);
    border: 1px solid var(--border);
    border-radius: var(--radius-lg);
    padding: var(--space-5) var(--space-6);
  }
  .panel-tight { padding: var(--space-2) 0.75rem; }

  .hint {
    color: var(--text-dim);
    font-size: 0.75rem;
    margin-top: 0.25rem;
  }
  .actions { display: flex; gap: var(--space-3); }

  .picker-tabs {
    display: flex;
    gap: var(--space-2);
    flex-wrap: wrap;
    margin-bottom: var(--space-4);
  }
  .pill {
    padding: 0.25rem 0.75rem;
    background: transparent;
    color: var(--text-muted);
    border: 1px solid var(--border-strong);
    border-radius: var(--radius-full);
    cursor: pointer;
    font-size: 0.8125rem;
    font-family: inherit;
  }
  .pill.active {
    background: var(--accent);
    color: white;
    border-color: var(--accent);
  }
  .pill:hover:not(.active) {
    background: var(--bg-hover);
    color: var(--text);
  }

  .repo-row {
    display: grid;
    grid-template-columns: 1fr auto auto;
    align-items: center;
    gap: var(--space-4);
    width: 100%;
    text-align: left;
    box-sizing: border-box;
    padding: 0.625rem 0.75rem;
    border: 0;
    border-bottom: 1px solid var(--border);
    background: transparent;
    color: inherit;
    font: inherit;
    font-size: 0.875rem;
    cursor: pointer;
  }
  .repo-row:hover { background: var(--bg-hover); }
  .repo-row:last-child { border-bottom: none; }
  .repo-name { font-family: var(--font-mono); }
  .repo-name .desc {
    color: var(--text-muted);
    font-size: 0.75rem;
    font-family: inherit;
    margin-top: 0.125rem;
  }
  .repo-meta {
    color: var(--text-dim);
    font-size: 0.75rem;
    font-family: var(--font-mono);
  }
  .lock { color: var(--warning); font-size: 0.75rem; }

  .panel-empty {
    color: var(--text-dim);
    text-align: center;
    padding: var(--space-6);
    font-size: 0.875rem;
  }
  .panel-empty.error { color: var(--danger-text); }

  .filter-input {
    margin-bottom: var(--space-2);
  }

  .pager {
    display: flex;
    gap: var(--space-2);
    margin-top: 0.75rem;
    justify-content: flex-end;
    font-size: 0.875rem;
  }
  .pager a {
    color: var(--link);
    padding: 0.25rem 0.5rem;
  }
  .pager-current { color: var(--text-dim); }
</style>

<Breadcrumb segments={[
  { label: 'workspaces', href: '/' },
  { label: data.workspace.name, href: `/workspaces/${data.workspace.id}` },
  { label: 'new project' }
]} />

<PageHeader title="Add project" />

<div class="layout">
  <form method="POST">
    <div class="field">
      <label for="name">Name</label>
      <input bind:this={nameInput} id="name" name="name" type="text" required maxlength="100" value={form?.name || ''} />
      <p class="hint">Display name for this project. Must be unique within the workspace.</p>
    </div>

    <div class="field">
      <label for="git_url">Git URL</label>
      <input bind:this={urlInput} id="git_url" name="git_url" type="url" required placeholder="https://github.com/user/repo.git" value={form?.git_url || ''} />
      <p class="hint">Public HTTPS URL, or click a repo from a connected provider on the right →</p>
    </div>

    <div class="field">
      <label for="branch">Branch (optional)</label>
      <input bind:this={branchInput} id="branch" name="branch" type="text" placeholder="main" value={form?.branch || ''} />
      <p class="hint">Leave blank to use the repo's default branch.</p>
    </div>

    {#if form?.error}<FlashMessage type="error">{form.error}</FlashMessage>{/if}

    <div class="actions">
      <button type="submit">Add</button>
      <a href={`/workspaces/${data.workspace.id}`}><button type="button" class="ghost">Cancel</button></a>
    </div>
  </form>

  <div>
    <h2>Pick from a connected provider</h2>
    {#if data.connectedProviders.length === 0}
      <div class="panel panel-empty">
        No providers connected. <a href="/settings">Connect one in Settings →</a>
      </div>
    {:else}
      <div class="picker-tabs">
        <button class="pill {!data.activeProvider ? 'active' : ''}" onclick={() => changeProvider('')}>—</button>
        {#each data.connectedProviders as c (c.provider_id)}
          <button class="pill {data.activeProvider?.provider_id === c.provider_id ? 'active' : ''}"
                  onclick={() => changeProvider(c.provider_id)}>
            {providerLabel(c.provider)}: {c.instance_host}
          </button>
        {/each}
      </div>

      {#if !data.activeProvider}
        <div class="panel panel-empty">Select a provider to browse repos.</div>
      {:else if data.reposError}
        <div class="panel panel-empty error">{data.reposError}</div>
      {:else if data.repos.length === 0}
        <div class="panel panel-empty">No repos returned for page {data.page}.</div>
      {:else}
        <input type="text"
               placeholder="Filter this page…"
               bind:value={filter}
               class="filter-input" />
        <div class="panel panel-tight">
          {#if visibleRepos.length === 0}
            <div class="panel-empty">No repos match <code>{filter}</code>.</div>
          {:else}
            {#each visibleRepos as r (r.full_name)}
              <button type="button" class="repo-row" onclick={() => pickRepo(r.name, r.clone_url, r.default_branch)}>
                <div class="repo-name">
                  {r.full_name}
                  {#if r.private}<span class="lock">🔒</span>{/if}
                  {#if r.description}<div class="desc">{r.description}</div>{/if}
                </div>
                <div class="repo-meta">{r.default_branch}</div>
                <div class="repo-meta">{formatRelative(r.updated_at)}</div>
              </button>
            {/each}
          {/if}
        </div>
        <div class="pager">
          {#if data.page > 1}<a href={pageURL(data.page - 1)}>← prev</a>{/if}
          <span class="pager-current">page {data.page}</span>
          {#if data.repos.length >= 50}<a href={pageURL(data.page + 1)}>next →</a>{/if}
        </div>
      {/if}
    {/if}
  </div>
</div>
