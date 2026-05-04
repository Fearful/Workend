<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { goto } from '$app/navigation';
  import { page } from '$app/state';

  let { children, data } = $props();

  let theme = $state<'dark' | 'light'>('dark');
  let helpOpen = $state(false);
  let mobileNavOpen = $state(false);
  let lastKey = $state<string | null>(null);
  let lastKeyAt = 0;

  function isActive(path: string): boolean {
    if (path === '/') return page.url.pathname === '/';
    return page.url.pathname.startsWith(path);
  }

  onMount(() => {
    const current = document.documentElement.getAttribute('data-theme');
    theme = current === 'light' ? 'light' : 'dark';
    window.addEventListener('keydown', onKeyDown);
  });

  onDestroy(() => {
    if (typeof window !== 'undefined') window.removeEventListener('keydown', onKeyDown);
  });

  function toggleTheme() {
    theme = theme === 'dark' ? 'light' : 'dark';
    document.documentElement.setAttribute('data-theme', theme);
    try {
      localStorage.setItem('workend_theme', theme);
    } catch {
      // localStorage may be unavailable (private mode); the in-memory toggle
      // still works for this session.
    }
  }

  // Keyboard shortcuts. Skip when the user is typing into an input/textarea
  // or when a modifier key is held (so browser/OS shortcuts still work).
  function onKeyDown(e: KeyboardEvent) {
    const t = e.target as HTMLElement | null;
    if (t && (t.tagName === 'INPUT' || t.tagName === 'TEXTAREA' || t.isContentEditable)) return;
    if (e.metaKey || e.ctrlKey || e.altKey) return;

    if (e.key === '?' && e.shiftKey) {
      e.preventDefault();
      helpOpen = !helpOpen;
      return;
    }
    if (e.key === 'Escape') {
      helpOpen = false;
      mobileNavOpen = false;
      return;
    }

    // Two-key chords ("g" then "d", "g" then "h"). 1.5s window.
    const now = Date.now();
    if (lastKey === 'g' && now - lastKeyAt < 1500) {
      lastKey = null;
      switch (e.key) {
        case 'd':
          e.preventDefault();
          if (data.user) goto('/dashboard');
          return;
        case 'h':
          e.preventDefault();
          goto('/');
          return;
        case 's':
          e.preventDefault();
          if (data.user) goto('/settings');
          return;
      }
    }

    if (e.key === 'g') {
      lastKey = 'g';
      lastKeyAt = now;
      return;
    }
    if (e.key === 't') {
      e.preventDefault();
      toggleTheme();
      return;
    }
    lastKey = null;
  }
</script>

<style>
  :global(:root) {
    --bg-page: #0d0f12;
    --bg-panel: #14181d;
    --bg-hover: #1a1f25;
    --bg-input: #14181d;
    --border: #1f2429;
    --border-strong: #2d3540;
    --text: #e8eaed;
    --text-muted: #9ca3af;
    --text-dim: #6b7280;
    --accent: #2563eb;
    --accent-hover: #1d4ed8;
    --link: #60a5fa;
    --success: #22c55e;
    --warning: #eab308;
    --danger: #dc2626;
    --danger-text: #ef4444;
    --info: #60a5fa;

    --status-success-bg: rgba(34, 197, 94, 0.12);
    --status-success-fg: #4ade80;
    --status-success-border: rgba(34, 197, 94, 0.3);
    --status-warning-bg: rgba(234, 179, 8, 0.12);
    --status-warning-fg: #facc15;
    --status-warning-border: rgba(234, 179, 8, 0.3);
    --status-danger-bg: rgba(239, 68, 68, 0.12);
    --status-danger-fg: #f87171;
    --status-danger-border: rgba(239, 68, 68, 0.3);
    --status-info-bg: rgba(96, 165, 250, 0.12);
    --status-info-fg: #93c5fd;
    --status-info-border: rgba(96, 165, 250, 0.3);
    --status-neutral-bg: rgba(148, 163, 184, 0.12);
    --status-neutral-fg: #cbd5e1;
    --status-neutral-border: rgba(148, 163, 184, 0.3);

    --font-mono: ui-monospace, "SF Mono", Menlo, Consolas, monospace;

    --fs-xs: 0.75rem;
    --fs-sm: 0.8125rem;
    --fs-md: 0.875rem;
    --fs-lg: 1rem;
    --fs-xl: 1.25rem;
    --fs-2xl: 1.5rem;

    --fw-regular: 400;
    --fw-medium: 500;
    --fw-semibold: 600;

    --lh-tight: 1.2;
    --lh-normal: 1.45;
    --lh-loose: 1.6;

    --space-1: 0.25rem;
    --space-2: 0.5rem;
    --space-3: 0.75rem;
    --space-4: 1rem;
    --space-5: 1.25rem;
    --space-6: 1.5rem;
    --space-8: 2rem;

    --shadow-card: 0 1px 3px rgba(0, 0, 0, 0.08);
    --shadow-card-hover: 0 4px 12px rgba(0, 0, 0, 0.15);
    --shadow-popover: 0 8px 24px rgba(0, 0, 0, 0.35);

    --radius-sm: 4px;
    --radius-md: 6px;
    --radius-lg: 8px;
    --radius-full: 999px;
  }

  :global([data-theme="light"]) {
    --bg-page: #f8fafc;
    --bg-panel: #ffffff;
    --bg-hover: #f1f5f9;
    --bg-input: #ffffff;
    --border: #e2e8f0;
    --border-strong: #cbd5e1;
    --text: #0f172a;
    --text-muted: #475569;
    --text-dim: #64748b;
    --accent: #2563eb;
    --accent-hover: #1d4ed8;
    --link: #2563eb;
    --success: #16a34a;
    --warning: #ca8a04;
    --danger: #dc2626;
    --danger-text: #dc2626;
    --info: #2563eb;

    --status-success-bg: rgba(22, 163, 74, 0.10);
    --status-success-fg: #15803d;
    --status-success-border: rgba(22, 163, 74, 0.25);
    --status-warning-bg: rgba(202, 138, 4, 0.10);
    --status-warning-fg: #a16207;
    --status-warning-border: rgba(202, 138, 4, 0.25);
    --status-danger-bg: rgba(220, 38, 38, 0.10);
    --status-danger-fg: #b91c1c;
    --status-danger-border: rgba(220, 38, 38, 0.25);
    --status-info-bg: rgba(37, 99, 235, 0.10);
    --status-info-fg: #1d4ed8;
    --status-info-border: rgba(37, 99, 235, 0.25);
    --status-neutral-bg: rgba(100, 116, 139, 0.10);
    --status-neutral-fg: #475569;
    --status-neutral-border: rgba(100, 116, 139, 0.25);

    --shadow-popover: 0 8px 24px rgba(15, 23, 42, 0.15);
  }

  :global(body) {
    margin: 0;
    font-family: ui-sans-serif, system-ui, -apple-system, "Segoe UI", Roboto, sans-serif;
    background: var(--bg-page);
    color: var(--text);
    -webkit-font-smoothing: antialiased;
    -moz-osx-font-smoothing: grayscale;
    text-rendering: optimizeLegibility;
  }

  :global(*) {
    scrollbar-width: thin;
    scrollbar-color: var(--border-strong) transparent;
  }

  :global(*::-webkit-scrollbar) {
    width: 10px;
    height: 10px;
  }

  :global(*::-webkit-scrollbar-track) {
    background: transparent;
  }

  :global(*::-webkit-scrollbar-thumb) {
    background: var(--border-strong);
    border-radius: 6px;
    border: 2px solid var(--bg-page);
  }

  :global(*::-webkit-scrollbar-thumb:hover) {
    background: var(--text-dim);
  }

  :global(a) {
    color: var(--link);
    text-decoration: none;
  }

  :global(a:hover) {
    text-decoration: underline;
  }

  :global(button) {
    font-family: inherit;
    font-size: 0.875rem;
    background: var(--accent);
    color: white;
    border: none;
    padding: 0.5rem 1rem;
    border-radius: 6px;
    cursor: pointer;
  }

  :global(button:hover) {
    background: var(--accent-hover);
  }

  :global(button.ghost) {
    background: transparent;
    color: var(--text-muted);
    border: 1px solid var(--border-strong);
  }

  :global(button.ghost:hover) {
    background: var(--bg-hover);
    color: var(--text);
  }

  :global(button.danger) {
    background: var(--danger);
  }

  :global(button:disabled) {
    opacity: 0.5;
    cursor: not-allowed;
  }

  :global(button:disabled:hover) {
    background: var(--accent);
  }

  :global(button.ghost:disabled:hover) {
    background: transparent;
    color: var(--text-muted);
  }

  :global(select) {
    font-family: inherit;
    font-size: 0.875rem;
    background: var(--bg-input);
    color: var(--text);
    border: 1px solid var(--border-strong);
    padding: 0.5rem 0.75rem;
    border-radius: var(--radius-md);
    cursor: pointer;
  }

  :global(select:focus) {
    outline: none;
    border-color: var(--accent);
  }

  :global(code) {
    font-family: var(--font-mono);
    font-size: 0.875em;
    background: var(--bg-hover);
    padding: 0.125rem 0.375rem;
    border-radius: var(--radius-sm);
  }

  :global(input, textarea) {
    font-family: inherit;
    font-size: 0.875rem;
    background: var(--bg-input);
    color: var(--text);
    border: 1px solid var(--border-strong);
    padding: 0.5rem 0.75rem;
    border-radius: 6px;
    width: 100%;
    box-sizing: border-box;
  }

  :global(input:focus, textarea:focus) {
    outline: none;
    border-color: var(--accent);
  }

  :global(label) {
    display: block;
    margin-bottom: 0.25rem;
    font-size: 0.875rem;
    color: var(--text-muted);
  }

  :global(.field) {
    margin-bottom: 1rem;
  }

  :global(:focus-visible) {
    outline: 2px solid var(--accent);
    outline-offset: 2px;
  }

  :global(input:focus-visible, textarea:focus-visible) {
    outline: none;
    border-color: var(--accent);
  }

  :global(kbd) {
    font-family: var(--font-mono);
    background: var(--bg-hover);
    padding: 0.125rem 0.5rem;
    border-radius: var(--radius-sm);
    font-size: 0.8125rem;
  }

  :global(.panel) {
    background: var(--bg-panel);
    border: 1px solid var(--border);
    border-radius: var(--radius-lg);
    padding: var(--space-5) var(--space-6);
    margin-bottom: var(--space-4);
  }

  :global(.dot) {
    display: inline-block;
    width: 8px;
    height: 8px;
    border-radius: 50%;
    flex-shrink: 0;
  }

  :global(.empty) {
    text-align: center;
    padding: 3rem var(--space-4);
    color: var(--text-muted);
  }

  :global(.error-banner) {
    color: var(--danger-text);
    font-size: 0.875rem;
    background: rgba(239, 68, 68, 0.08);
    border: 1px solid rgba(239, 68, 68, 0.3);
    padding: var(--space-3) var(--space-4);
    border-radius: var(--radius-md);
    margin-bottom: var(--space-4);
  }

  :global(.success-banner) {
    color: var(--success);
    font-size: 0.875rem;
    background: rgba(34, 197, 94, 0.08);
    border: 1px solid rgba(34, 197, 94, 0.3);
    padding: var(--space-3) var(--space-4);
    border-radius: var(--radius-md);
    margin-bottom: var(--space-4);
  }

  main {
    max-width: 1440px;
    margin: 0 auto;
    padding: var(--space-6) var(--space-6);
    width: 100%;
    box-sizing: border-box;
  }

  @media (min-width: 1920px) {
    main { max-width: 1600px; }
  }

  /* Embed/iframe pages (telemetry, dockscope) fill the screen so the
     embedded tool gets every pixel. App pages always use the default. */
  main:global(.fullbleed) {
    max-width: none;
  }

  /* Forms-only pages (login/signup) opt into narrow for readability. */
  main:global(.narrow) {
    max-width: 560px;
  }

  @media (max-width: 767px) {
    main { padding: var(--space-4) var(--space-4); }
  }

  header {
    border-bottom: 1px solid var(--border);
    padding: 1rem 1.5rem;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-4);
  }

  .brand {
    font-size: 1.1rem;
    font-weight: 600;
    letter-spacing: -0.01em;
    color: var(--text);
    flex-shrink: 0;
  }

  .brand small {
    color: var(--text-dim);
    font-weight: 400;
    margin-left: 0.5rem;
  }

  nav {
    display: flex;
    align-items: center;
    gap: var(--space-4);
    font-size: 0.875rem;
  }

  nav a {
    position: relative;
    padding: 0.25rem 0;
    color: var(--text-muted);
    transition: color 120ms ease;
  }
  nav a:hover {
    color: var(--text);
    text-decoration: none;
  }
  nav a.active {
    color: var(--text);
  }
  nav a.active::after {
    content: '';
    position: absolute;
    left: 0;
    right: 0;
    bottom: -1.05rem;
    height: 2px;
    background: var(--accent);
    border-radius: 2px;
  }

  .theme-toggle {
    background: transparent;
    color: var(--text-muted);
    border: 1px solid var(--border-strong);
    padding: 0.375rem 0.625rem;
    border-radius: var(--radius-md);
    font-size: 0.875rem;
    cursor: pointer;
    line-height: 1;
    flex-shrink: 0;
  }

  .theme-toggle:hover {
    background: var(--bg-hover);
    color: var(--text);
  }

  .nav-toggle {
    display: none;
    background: transparent;
    color: var(--text);
    border: 1px solid var(--border-strong);
    padding: 0.5rem 0.625rem;
    border-radius: var(--radius-md);
    cursor: pointer;
    font-size: 1rem;
    line-height: 1;
  }
  .nav-toggle:hover { background: var(--bg-hover); }

  @media (max-width: 768px) {
    .nav-toggle { display: inline-flex; align-items: center; gap: 0.375rem; }

    nav {
      position: fixed;
      top: 60px;
      right: 0;
      bottom: 0;
      background: var(--bg-panel);
      border-left: 1px solid var(--border);
      flex-direction: column;
      align-items: stretch;
      gap: var(--space-2);
      padding: var(--space-4);
      width: min(280px, 80vw);
      transform: translateX(100%);
      transition: transform 200ms ease;
      z-index: 90;
      overflow-y: auto;
    }
    nav.open { transform: translateX(0); }

    nav a, nav button.theme-toggle, nav .inline-form button {
      width: 100%;
      text-align: left;
      padding: var(--space-3) var(--space-4);
      border-radius: var(--radius-md);
      box-sizing: border-box;
    }
    nav a:hover, nav button.theme-toggle:hover { background: var(--bg-hover); }
    nav a.active::after { display: none; }
    nav a.active { background: var(--bg-hover); }
    nav .inline-form { width: 100%; }
  }

  .nav-overlay {
    position: fixed;
    inset: 60px 0 0 0;
    background: rgba(0, 0, 0, 0.4);
    z-index: 80;
    display: none;
  }
  .nav-overlay.show { display: block; }
  @media (min-width: 769px) {
    .nav-overlay.show { display: none; }
  }

  .modal-backdrop {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.5);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 100;
  }

  .modal-content {
    background: var(--bg-panel);
    border: 1px solid var(--border);
    border-radius: var(--radius-lg);
    padding: var(--space-5) var(--space-6);
    width: min(420px, 90vw);
  }

  .shortcut-grid {
    display: grid;
    grid-template-columns: auto 1fr;
    gap: var(--space-2) var(--space-4);
    font-size: 0.875rem;
  }

  .nav-link {
    position: relative;
  }

  .badge-count {
    background: var(--warning);
    color: var(--bg-page);
    font-size: 0.6875rem;
    font-weight: 600;
    padding: 0.0625rem 0.375rem;
    border-radius: var(--radius-full);
    margin-left: var(--space-1);
  }

  :global(.inline-form) {
    margin: 0;
  }
</style>

<header>
  <a href="/" class="brand">Workend <small>v0.0.1</small></a>
  <button type="button"
          class="nav-toggle"
          onclick={() => (mobileNavOpen = !mobileNavOpen)}
          aria-label="Toggle menu"
          aria-expanded={mobileNavOpen}>
    {mobileNavOpen ? '✕' : '☰'}
  </button>
  <nav class:open={mobileNavOpen}>
    {#if data.user}
      <a href="/dashboard" class:active={isActive('/dashboard')} onclick={() => (mobileNavOpen = false)}>Dashboard</a>
      <a href="/mentions" class="nav-link" class:active={isActive('/mentions')} onclick={() => (mobileNavOpen = false)}>
        Mentions
        {#if data.unreadMentions > 0}
          <span class="badge-count">{data.unreadMentions}</span>
        {/if}
      </a>
      <a href="/dockscope" class:active={isActive('/dockscope')} onclick={() => (mobileNavOpen = false)}>Dockscope</a>
      <a href="/otel" class:active={isActive('/otel')} onclick={() => (mobileNavOpen = false)}>Telemetry</a>
      {#if data.user.is_admin}
        <a href="/admin" class:active={isActive('/admin')} onclick={() => (mobileNavOpen = false)}>Admin</a>
      {/if}
      <a href="/settings" class:active={isActive('/settings')} onclick={() => (mobileNavOpen = false)}>{data.user.display_name}</a>
      <form method="POST" action="/logout" class="inline-form">
        <button type="submit" class="ghost">Log out</button>
      </form>
    {:else}
      <a href="/login" class:active={isActive('/login')} onclick={() => (mobileNavOpen = false)}>Log in</a>
      <a href="/signup" class:active={isActive('/signup')} onclick={() => (mobileNavOpen = false)}>Sign up</a>
    {/if}
    <button type="button"
            class="theme-toggle"
            onclick={toggleTheme}
            title={theme === 'dark' ? 'Switch to light mode' : 'Switch to dark mode'}
            aria-label="Toggle theme">
      {theme === 'dark' ? '☀' : '☾'}
    </button>
  </nav>
</header>

{#if mobileNavOpen}
  <div class="nav-overlay show" onclick={() => (mobileNavOpen = false)} role="presentation"></div>
{/if}

<main>
  {@render children()}
</main>

{#if helpOpen}
  <div role="dialog"
       aria-modal="true"
       aria-label="Keyboard shortcuts"
       tabindex="-1"
       class="modal-backdrop"
       onclick={(e) => { if (e.target === e.currentTarget) helpOpen = false; }}
       onkeydown={(e) => { if (e.key === 'Escape') helpOpen = false; }}>
    <div class="modal-content">
      <h2 style="margin:0 0 1rem 0; font-size:1.05rem;">Keyboard shortcuts</h2>
      <div class="shortcut-grid">
        <kbd>?</kbd><span>Toggle this help</span>
        <kbd>t</kbd><span>Toggle theme</span>
        <kbd>g h</kbd><span>Go to home (workspaces)</span>
        <kbd>g d</kbd><span>Go to dashboard</span>
        <kbd>g s</kbd><span>Go to settings</span>
        <kbd>Esc</kbd><span>Close dialogs</span>
      </div>
      <div style="margin-top:1rem; text-align:right;">
        <button type="button" class="ghost" onclick={() => (helpOpen = false)}>Close</button>
      </div>
    </div>
  </div>
{/if}
