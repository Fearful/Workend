<script lang="ts">
  import { goto } from '$app/navigation';
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

  function formatRelative(iso: string | null): string {
    if (!iso) return '—';
    const ms = Date.now() - new Date(iso).getTime();
    const min = Math.floor(ms / 60000);
    if (min < 1) return 'just now';
    if (min < 60) return `${min}m ago`;
    const hr = Math.floor(min / 60);
    if (hr < 24) return `${hr}h ago`;
    const d = Math.floor(hr / 24);
    return `${d}d ago`;
  }
</script>

<style>
  h1 { font-size: 1.5rem; margin: 0 0 0.25rem 0; }
  h2 { font-size: 0.875rem; text-transform: uppercase; letter-spacing: 0.05em;
       color: #9ca3af; font-weight: 600; margin: 1.5rem 0 1rem 0; }
  .breadcrumb { color: #6b7280; font-size: 0.875rem; margin-bottom: 0.5rem; }

  .layout { display: grid; grid-template-columns: 480px 1fr; gap: 2rem; align-items: start; margin-top: 1.5rem; }
  @media (max-width: 900px) { .layout { grid-template-columns: 1fr; } }

  form { margin: 0; }
  .panel {
    background: #14181d; border: 1px solid #1f2429; border-radius: 8px;
    padding: 1.25rem 1.5rem;
  }

  .hint { color: #6b7280; font-size: 0.75rem; margin-top: 0.25rem; }
  .error { color: #ef4444; font-size: 0.875rem; margin: 0.5rem 0; }
  .actions { display: flex; gap: 0.75rem; }

  .picker-tabs { display: flex; gap: 0.5rem; flex-wrap: wrap; margin-bottom: 1rem; }
  .pill {
    padding: 0.25rem 0.75rem; background: transparent; color: #9ca3af;
    border: 1px solid #2d3540; border-radius: 999px; cursor: pointer;
    font-size: 0.8125rem; font-family: inherit;
  }
  .pill.active { background: #2563eb; color: white; border-color: #2563eb; }
  .pill:hover:not(.active) { background: #1a1f25; color: #e8eaed; }

  .repo-row {
    display: grid; grid-template-columns: 1fr auto auto;
    align-items: center; gap: 1rem;
    width: 100%; text-align: left; box-sizing: border-box;
    padding: 0.625rem 0.75rem; border: 0; border-bottom: 1px solid #1f2429;
    background: transparent; color: inherit; font: inherit;
    font-size: 0.875rem; cursor: pointer;
  }
  .repo-row:hover { background: #1a1f25; }
  .repo-row:last-child { border-bottom: none; }
  .repo-name { font-family: ui-monospace, "SF Mono", Menlo, monospace; }
  .repo-name .desc { color: #9ca3af; font-size: 0.75rem; font-family: inherit; margin-top: 0.125rem; }
  .repo-meta { color: #6b7280; font-size: 0.75rem; font-family: ui-monospace, "SF Mono", Menlo, monospace; }
  .lock { color: #eab308; font-size: 0.75rem; }

  .empty { color: #6b7280; text-align: center; padding: 1.5rem; font-size: 0.875rem; }

  .pager { display: flex; gap: 0.5rem; margin-top: 0.75rem; justify-content: flex-end; font-size: 0.875rem; }
  .pager a { color: #60a5fa; padding: 0.25rem 0.5rem; }
</style>

<div class="breadcrumb">
  <a href="/">workspaces</a> /
  <a href={`/workspaces/${data.workspace.id}`}>{data.workspace.name}</a> / new project
</div>

<h1>Add project</h1>

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

    {#if form?.error}<p class="error">{form.error}</p>{/if}

    <div class="actions">
      <button type="submit">Add</button>
      <a href={`/workspaces/${data.workspace.id}`}><button type="button" class="ghost">Cancel</button></a>
    </div>
  </form>

  <div>
    <h2 style="margin-top: 0;">Pick from a connected provider</h2>
    {#if data.connectedProviders.length === 0}
      <div class="panel empty">
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
        <div class="panel empty">Select a provider to browse repos.</div>
      {:else if data.reposError}
        <div class="panel empty" style="color:#ef4444;">{data.reposError}</div>
      {:else if data.repos.length === 0}
        <div class="panel empty">No repos returned for page {data.page}.</div>
      {:else}
        <div class="panel" style="padding: 0.5rem 0.75rem;">
          {#each data.repos as r (r.full_name)}
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
        </div>
        <div class="pager">
          {#if data.page > 1}<a href={pageURL(data.page - 1)}>← prev</a>{/if}
          <span style="color:#6b7280;">page {data.page}</span>
          {#if data.repos.length >= 50}<a href={pageURL(data.page + 1)}>next →</a>{/if}
        </div>
      {/if}
    {/if}
  </div>
</div>
