<script lang="ts">
  import PageHeader from '$lib/components/PageHeader.svelte';
  import FlashMessage from '$lib/components/FlashMessage.svelte';

  let { form } = $props();
</script>

<style>
  form { max-width: 480px; }

  textarea {
    min-height: 4rem;
    resize: vertical;
  }

  .actions {
    display: flex;
    gap: var(--space-3);
  }
</style>

<PageHeader title="New workspace" />

<form method="POST">
  <div class="field">
    <label for="name">Name</label>
    <input id="name" name="name" type="text" required maxlength="100" value={form?.name || ''} />
  </div>
  <div class="field">
    <label for="description">Description (optional)</label>
    <textarea id="description" name="description" maxlength="500">{form?.description || ''}</textarea>
  </div>

  {#if form?.error}<FlashMessage type="error">{form.error}</FlashMessage>{/if}

  <div class="actions">
    <button type="submit">Create</button>
    <a href="/"><button type="button" class="ghost">Cancel</button></a>
  </div>
</form>
