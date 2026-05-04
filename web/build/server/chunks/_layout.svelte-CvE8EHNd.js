import { a7 as attr, a8 as escape_html, a9 as attr_class } from './renderer-mjPKoiGx.js';
import { o as onDestroy } from './index-server-BdEa2Uck.js';
import { g as goto } from './client-CqnwEXXv.js';
import { p as page } from './index2-E0jCU3nx.js';
import './state.svelte-CeeZin_s.js';
import './root-gJ1T4-40.js';
import './index-B8uUmbHp.js';

function _layout($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    let { children, data } = $$props;
    let theme = "dark";
    let helpOpen = false;
    let mobileNavOpen = false;
    let lastKey = null;
    let lastKeyAt = 0;
    function isActive(path) {
      if (path === "/") return page.url.pathname === "/";
      return page.url.pathname.startsWith(path);
    }
    onDestroy(() => {
      if (typeof window !== "undefined") window.removeEventListener("keydown", onKeyDown);
    });
    function toggleTheme() {
      theme = theme === "dark" ? "light" : "dark";
      document.documentElement.setAttribute("data-theme", theme);
      try {
        localStorage.setItem("workend_theme", theme);
      } catch {
      }
    }
    function onKeyDown(e) {
      const t = e.target;
      if (t && (t.tagName === "INPUT" || t.tagName === "TEXTAREA" || t.isContentEditable)) return;
      if (e.metaKey || e.ctrlKey || e.altKey) return;
      if (e.key === "?" && e.shiftKey) {
        e.preventDefault();
        helpOpen = !helpOpen;
        return;
      }
      if (e.key === "Escape") {
        helpOpen = false;
        mobileNavOpen = false;
        return;
      }
      const now = Date.now();
      if (lastKey === "g" && now - lastKeyAt < 1500) {
        lastKey = null;
        switch (e.key) {
          case "d":
            e.preventDefault();
            if (data.user) goto();
            return;
          case "h":
            e.preventDefault();
            goto();
            return;
          case "s":
            e.preventDefault();
            if (data.user) goto();
            return;
        }
      }
      if (e.key === "g") {
        lastKey = "g";
        lastKeyAt = now;
        return;
      }
      if (e.key === "t") {
        e.preventDefault();
        toggleTheme();
        return;
      }
      lastKey = null;
    }
    $$renderer2.push(`<header class="svelte-12qhfyh"><a href="/" class="brand svelte-12qhfyh">Workend <small class="svelte-12qhfyh">v0.0.1</small></a> <button type="button" class="nav-toggle svelte-12qhfyh" aria-label="Toggle menu"${attr("aria-expanded", mobileNavOpen)}>${escape_html(mobileNavOpen ? "✕" : "☰")}</button> <nav${attr_class("svelte-12qhfyh", void 0, { "open": mobileNavOpen })}>`);
    if (data.user) {
      $$renderer2.push("<!--[0-->");
      $$renderer2.push(`<a href="/dashboard"${attr_class("svelte-12qhfyh", void 0, { "active": isActive("/dashboard") })}>Dashboard</a> <a href="/mentions"${attr_class("nav-link svelte-12qhfyh", void 0, { "active": isActive("/mentions") })}>Mentions `);
      if (data.unreadMentions > 0) {
        $$renderer2.push("<!--[0-->");
        $$renderer2.push(`<span class="badge-count svelte-12qhfyh">${escape_html(data.unreadMentions)}</span>`);
      } else {
        $$renderer2.push("<!--[-1-->");
      }
      $$renderer2.push(`<!--]--></a> <a href="/dockscope"${attr_class("svelte-12qhfyh", void 0, { "active": isActive("/dockscope") })}>Dockscope</a> <a href="/otel"${attr_class("svelte-12qhfyh", void 0, { "active": isActive("/otel") })}>Telemetry</a> `);
      if (data.user.is_admin) {
        $$renderer2.push("<!--[0-->");
        $$renderer2.push(`<a href="/admin"${attr_class("svelte-12qhfyh", void 0, { "active": isActive("/admin") })}>Admin</a>`);
      } else {
        $$renderer2.push("<!--[-1-->");
      }
      $$renderer2.push(`<!--]--> <a href="/settings"${attr_class("svelte-12qhfyh", void 0, { "active": isActive("/settings") })}>${escape_html(data.user.display_name)}</a> <form method="POST" action="/logout" class="inline-form svelte-12qhfyh"><button type="submit" class="ghost svelte-12qhfyh">Log out</button></form>`);
    } else {
      $$renderer2.push("<!--[-1-->");
      $$renderer2.push(`<a href="/login"${attr_class("svelte-12qhfyh", void 0, { "active": isActive("/login") })}>Log in</a> <a href="/signup"${attr_class("svelte-12qhfyh", void 0, { "active": isActive("/signup") })}>Sign up</a>`);
    }
    $$renderer2.push(`<!--]--> <button type="button" class="theme-toggle svelte-12qhfyh"${attr("title", theme === "dark" ? "Switch to light mode" : "Switch to dark mode")} aria-label="Toggle theme">${escape_html(theme === "dark" ? "☀" : "☾")}</button></nav></header> `);
    if (mobileNavOpen) {
      $$renderer2.push("<!--[0-->");
      $$renderer2.push(`<div class="nav-overlay show svelte-12qhfyh" role="presentation"></div>`);
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--> <main class="svelte-12qhfyh">`);
    children($$renderer2);
    $$renderer2.push(`<!----></main> `);
    if (helpOpen) {
      $$renderer2.push("<!--[0-->");
      $$renderer2.push(`<div role="dialog" aria-modal="true" aria-label="Keyboard shortcuts" tabindex="-1" class="modal-backdrop svelte-12qhfyh"><div class="modal-content svelte-12qhfyh"><h2 style="margin:0 0 1rem 0; font-size:1.05rem;">Keyboard shortcuts</h2> <div class="shortcut-grid svelte-12qhfyh"><kbd>?</kbd><span>Toggle this help</span> <kbd>t</kbd><span>Toggle theme</span> <kbd>g h</kbd><span>Go to home (workspaces)</span> <kbd>g d</kbd><span>Go to dashboard</span> <kbd>g s</kbd><span>Go to settings</span> <kbd>Esc</kbd><span>Close dialogs</span></div> <div style="margin-top:1rem; text-align:right;"><button type="button" class="ghost">Close</button></div></div></div>`);
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]-->`);
  });
}

export { _layout as default };
//# sourceMappingURL=_layout.svelte-CvE8EHNd.js.map
