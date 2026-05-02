<script>
  let { children } = $props();
</script>

<style>
  :global(body) {
    margin: 0;
    font-family: ui-sans-serif, system-ui, -apple-system, "Segoe UI", Roboto, sans-serif;
    background: #0d0f12;
    color: #e8eaed;
  }

  main {
    max-width: 960px;
    margin: 0 auto;
    padding: 2rem 1.5rem;
  }

  header {
    border-bottom: 1px solid #1f2429;
    padding: 1rem 1.5rem;
  }

  header h1 {
    margin: 0;
    font-size: 1.1rem;
    font-weight: 600;
    letter-spacing: -0.01em;
  }

  header h1 :global(small) {
    color: #6b7280;
    font-weight: 400;
    margin-left: 0.5rem;
  }
</style>

<header>
  <h1>Workend <small>v0.0.1 — Stage 1</small></h1>
</header>

<main>
  {@render children()}
</main>
