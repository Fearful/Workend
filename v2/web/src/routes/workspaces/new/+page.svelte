<script lang="ts">
  let { form } = $props();
</script>

<style>
  h1 {
    font-size: 1.5rem;
    margin: 0 0 1.5rem 0;
  }

  form {
    max-width: 480px;
  }

  textarea {
    min-height: 4rem;
    resize: vertical;
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

<h1>New workspace</h1>

<form method="POST">
  <div class="field">
    <label for="name">Name</label>
    <input id="name" name="name" type="text" required maxlength="100" value={form?.name || ''} />
  </div>
  <div class="field">
    <label for="description">Description (optional)</label>
    <textarea id="description" name="description" maxlength="500">{form?.description || ''}</textarea>
  </div>

  {#if form?.error}
    <p class="error">{form.error}</p>
  {/if}

  <div class="actions">
    <button type="submit">Create</button>
    <a href="/"><button type="button" class="ghost">Cancel</button></a>
  </div>
</form>
