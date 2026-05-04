<script lang="ts">
  import { page } from '$app/state';

  let { projectID }: { projectID: string } = $props();

  type Tab = { label: string; path: string; match: 'exact' | 'prefix' };

  let tabs: Tab[] = $derived([
    { label: 'Overview', path: `/projects/${projectID}`, match: 'exact' },
    { label: 'Runs', path: `/projects/${projectID}/runs`, match: 'prefix' },
    { label: 'Pipelines', path: `/projects/${projectID}/pipelines`, match: 'prefix' },
    { label: 'Board', path: `/projects/${projectID}/board`, match: 'prefix' },
    { label: 'Branches', path: `/projects/${projectID}/branches`, match: 'prefix' },
    { label: 'Trends', path: `/projects/${projectID}/trends`, match: 'prefix' },
    { label: 'Images', path: `/projects/${projectID}/images`, match: 'prefix' },
    { label: 'Schedules', path: `/projects/${projectID}/schedules`, match: 'prefix' }
  ]);

  function isActive(t: Tab): boolean {
    const p = page.url.pathname;
    return t.match === 'exact' ? p === t.path : p.startsWith(t.path);
  }
</script>

<nav class="project-tabs" aria-label="Project sections">
  <div class="tabs-inner">
    {#each tabs as t (t.path)}
      <a href={t.path} class="tab" class:active={isActive(t)} aria-current={isActive(t) ? 'page' : undefined}>
        {t.label}
      </a>
    {/each}
  </div>
</nav>

<style>
  .project-tabs {
    border-bottom: 1px solid var(--border);
    margin-bottom: var(--space-5);
    overflow-x: auto;
    -webkit-overflow-scrolling: touch;
  }
  .tabs-inner {
    display: flex;
    gap: 0;
    min-width: max-content;
  }
  .tab {
    padding: 0.625rem 0.875rem;
    color: var(--text-dim);
    text-decoration: none;
    font-size: 0.875rem;
    font-weight: 500;
    border-bottom: 2px solid transparent;
    margin-bottom: -1px;
    white-space: nowrap;
    transition: color 80ms ease, border-color 80ms ease;
  }
  .tab:hover {
    color: var(--text);
    text-decoration: none;
  }
  .tab.active {
    color: var(--text);
    border-bottom-color: var(--accent);
  }
</style>
