<script lang="ts">
  let { data, form } = $props();
</script>

<style>
  h1 {
    font-size: 1.5rem;
    margin: 0 0 0.25rem 0;
  }

  .breadcrumb {
    color: #6b7280;
    font-size: 0.875rem;
    margin-bottom: 0.5rem;
  }

  form {
    max-width: 480px;
    margin-top: 1.5rem;
  }

  .hint {
    color: #6b7280;
    font-size: 0.75rem;
    margin-top: 0.25rem;
  }

  .error {
    color: #ef4444;
    font-size: 0.875rem;
    margin: 0.5rem 0;
  }

  .actions {
    display: flex;
    gap: 0.75rem;
  }
</style>

<div class="breadcrumb"><a href="/">workspaces</a> / <a href={`/workspaces/${data.workspace.id}`}>{data.workspace.name}</a> / new project</div>

<h1>Add project</h1>

<form method="POST">
  <div class="field">
    <label for="name">Name</label>
    <input id="name" name="name" type="text" required maxlength="100" value={form?.name || ''} />
    <p class="hint">Display name for this project. Must be unique within the workspace.</p>
  </div>

  <div class="field">
    <label for="git_url">Git URL</label>
    <input id="git_url" name="git_url" type="url" required placeholder="https://github.com/user/repo.git" value={form?.git_url || ''} />
    <p class="hint">Public HTTPS URL only this stage. Private / SSH support comes in Stage 9.</p>
  </div>

  <div class="field">
    <label for="branch">Branch (optional)</label>
    <input id="branch" name="branch" type="text" placeholder="main" value={form?.branch || ''} />
    <p class="hint">Leave blank to use the repo's default branch.</p>
  </div>

  {#if form?.error}
    <p class="error">{form.error}</p>
  {/if}

  <div class="actions">
    <button type="submit">Add</button>
    <a href={`/workspaces/${data.workspace.id}`}><button type="button" class="ghost">Cancel</button></a>
  </div>
</form>
