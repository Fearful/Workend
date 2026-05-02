<script lang="ts">
  let { data } = $props();

  function shortSha(sha: string | null): string {
    return sha ? sha.slice(0, 7) : '—';
  }

  function shortDigest(d: string | null): string {
    return d ? d.slice(0, 24) + '…' : '—';
  }

  function formatBytes(b: number | null): string {
    if (!b) return '—';
    const units = ['B', 'KB', 'MB', 'GB'];
    let v = b;
    let i = 0;
    while (v >= 1024 && i < units.length - 1) { v /= 1024; i++; }
    return `${v.toFixed(1)} ${units[i]}`;
  }
</script>

<style>
  h1 { font-size: 1.5rem; margin: 0 0 0.5rem 0; }
  .breadcrumb { color: #6b7280; font-size: 0.875rem; margin-bottom: 1.5rem; }

  .row {
    display: grid; grid-template-columns: 1fr auto auto auto auto;
    align-items: center; gap: 1rem;
    padding: 0.75rem 1rem; background: #14181d;
    border: 1px solid #1f2429; border-radius: 6px; margin-bottom: 0.375rem;
    font-size: 0.875rem;
  }
  .mono { font-family: ui-monospace, "SF Mono", Menlo, monospace; }
  .muted { color: #6b7280; font-size: 0.75rem; font-family: ui-monospace, "SF Mono", Menlo, monospace; }
  .empty { color: #9ca3af; text-align: center; padding: 3rem;
           background: #14181d; border: 1px dashed #1f2429; border-radius: 8px; }
</style>

<div class="breadcrumb">
  <a href="/">workspaces</a>
  {#if data.workspace}/ <a href={`/workspaces/${data.workspace.id}`}>{data.workspace.name}</a>{/if}
  / <a href={`/projects/${data.project.id}`}>{data.project.name}</a> / images
</div>

<h1>Images</h1>

{#if data.images.length === 0}
  <div class="empty">
    No images built yet. Add a <code>Dockerfile</code> to the repo and run the
    <code>dockerfile: build</code> task.
  </div>
{:else}
  {#each data.images as img (img.id)}
    <div class="row">
      <span class="mono">{shortDigest(img.digest)}</span>
      <span class="muted">{img.dockerfile_path}</span>
      <span class="muted">commit {shortSha(img.commit_sha)}</span>
      <span class="muted">{formatBytes(img.size_bytes)}</span>
      <span class="muted">{new Date(img.built_at).toLocaleString()}</span>
    </div>
  {/each}
{/if}
