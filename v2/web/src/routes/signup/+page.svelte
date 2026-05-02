<script lang="ts">
  let { form } = $props();
</script>

<style>
  h1 {
    font-size: 1.5rem;
    margin: 0 0 1.5rem 0;
  }

  form {
    max-width: 360px;
  }

  .error {
    color: #ef4444;
    font-size: 0.875rem;
    margin: 0.5rem 0;
  }

  .alt {
    margin-top: 1rem;
    font-size: 0.875rem;
    color: #9ca3af;
  }

  .hint {
    color: #6b7280;
    font-size: 0.75rem;
    margin-top: 0.25rem;
  }
</style>

<h1>Sign up</h1>

<form method="POST">
  <div class="field">
    <label for="display_name">Display name</label>
    <input id="display_name" name="display_name" type="text" required value={form?.display_name || ''} />
  </div>
  <div class="field">
    <label for="email">Email</label>
    <input id="email" name="email" type="email" autocomplete="email" required value={form?.email || ''} />
  </div>
  <div class="field">
    <label for="password">Password</label>
    <input id="password" name="password" type="password" autocomplete="new-password" required minlength="8" />
    <p class="hint">At least 8 characters.</p>
  </div>

  {#if form?.error}
    <p class="error">{form.error}</p>
  {/if}

  <button type="submit">Create account</button>

  <p class="alt">Already have an account? <a href="/login">Log in</a></p>
</form>
