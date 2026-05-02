<script lang="ts">
  let { children, data } = $props();
</script>

<style>
  :global(body) {
    margin: 0;
    font-family: ui-sans-serif, system-ui, -apple-system, "Segoe UI", Roboto, sans-serif;
    background: #0d0f12;
    color: #e8eaed;
  }

  :global(a) {
    color: #60a5fa;
    text-decoration: none;
  }

  :global(a:hover) {
    text-decoration: underline;
  }

  :global(button) {
    font-family: inherit;
    font-size: 0.875rem;
    background: #2563eb;
    color: white;
    border: none;
    padding: 0.5rem 1rem;
    border-radius: 6px;
    cursor: pointer;
  }

  :global(button:hover) {
    background: #1d4ed8;
  }

  :global(button.ghost) {
    background: transparent;
    color: #9ca3af;
    border: 1px solid #2d3540;
  }

  :global(button.ghost:hover) {
    background: #1a1f25;
    color: #e8eaed;
  }

  :global(button.danger) {
    background: #dc2626;
  }

  :global(input, textarea) {
    font-family: inherit;
    font-size: 0.875rem;
    background: #14181d;
    color: #e8eaed;
    border: 1px solid #2d3540;
    padding: 0.5rem 0.75rem;
    border-radius: 6px;
    width: 100%;
    box-sizing: border-box;
  }

  :global(input:focus, textarea:focus) {
    outline: none;
    border-color: #2563eb;
  }

  :global(label) {
    display: block;
    margin-bottom: 0.25rem;
    font-size: 0.875rem;
    color: #9ca3af;
  }

  :global(.field) {
    margin-bottom: 1rem;
  }

  main {
    max-width: 960px;
    margin: 0 auto;
    padding: 2rem 1.5rem;
  }

  header {
    border-bottom: 1px solid #1f2429;
    padding: 1rem 1.5rem;
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  .brand {
    font-size: 1.1rem;
    font-weight: 600;
    letter-spacing: -0.01em;
    color: #e8eaed;
  }

  .brand small {
    color: #6b7280;
    font-weight: 400;
    margin-left: 0.5rem;
  }

  nav {
    display: flex;
    align-items: center;
    gap: 1rem;
    font-size: 0.875rem;
  }

  .user-name {
    color: #9ca3af;
  }
</style>

<header>
  <a href="/" class="brand">Workend <small>v0.0.1</small></a>
  <nav>
    {#if data.user}
      <span class="user-name">{data.user.display_name}</span>
      <form method="POST" action="/logout" style="margin: 0;">
        <button type="submit" class="ghost">Log out</button>
      </form>
    {:else}
      <a href="/login">Log in</a>
      <a href="/signup">Sign up</a>
    {/if}
  </nav>
</header>

<main>
  {@render children()}
</main>
