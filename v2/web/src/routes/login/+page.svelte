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
</style>

<h1>Log in</h1>

<form method="POST">
  <div class="field">
    <label for="email">Email</label>
    <input id="email" name="email" type="email" autocomplete="email" required value={form?.email || ''} />
  </div>
  <div class="field">
    <label for="password">Password</label>
    <input id="password" name="password" type="password" autocomplete="current-password" required />
  </div>

  {#if form?.error}
    <p class="error">{form.error}</p>
  {/if}

  <button type="submit">Log in</button>

  <p class="alt">No account? <a href="/signup">Sign up</a></p>
</form>
