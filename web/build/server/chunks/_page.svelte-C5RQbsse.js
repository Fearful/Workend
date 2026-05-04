import { a8 as escape_html, a7 as attr, a2 as derived } from './renderer-mjPKoiGx.js';
import { P as PageHeader } from './PageHeader-CaLG0rbe.js';
import { F as FlashMessage } from './FlashMessage-6DSDTh4l.js';

function _page($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    let { data, form } = $$props;
    let forceLocalLogin = false;
    let showLocalLogin = derived(() => !data.oidcConfigured || forceLocalLogin);
    PageHeader($$renderer2, { title: "Log in" });
    $$renderer2.push(`<!----> `);
    if (data.oidcConfigured) {
      $$renderer2.push("<!--[0-->");
      $$renderer2.push(`<div class="oidc-section svelte-1x05zx6"><a href="/api/auth/oidc/start"><button type="button" class="oidc-btn svelte-1x05zx6">Sign in with ${escape_html(data.oidcProviderName)}</button></a></div> `);
      if (!showLocalLogin()) {
        $$renderer2.push("<!--[0-->");
        $$renderer2.push(`<button type="button" class="local-toggle svelte-1x05zx6">Use email and password instead</button>`);
      } else {
        $$renderer2.push("<!--[-1-->");
        $$renderer2.push(`<div class="divider svelte-1x05zx6">or</div>`);
      }
      $$renderer2.push(`<!--]-->`);
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--> `);
    if (showLocalLogin()) {
      $$renderer2.push("<!--[0-->");
      $$renderer2.push(`<form method="POST" class="svelte-1x05zx6"><div class="field"><label for="email">Email</label> <input id="email" name="email" type="email" autocomplete="email" required=""${attr("value", form?.email || "")}/></div> <div class="field"><label for="password">Password</label> <input id="password" name="password" type="password" autocomplete="current-password" required=""/></div> `);
      if (form?.error) {
        $$renderer2.push("<!--[0-->");
        FlashMessage($$renderer2, {
          type: "error",
          children: ($$renderer3) => {
            $$renderer3.push(`<!---->${escape_html(form.error)}`);
          }
        });
      } else {
        $$renderer2.push("<!--[-1-->");
      }
      $$renderer2.push(`<!--]--> <button type="submit">Log in</button> `);
      if (!data.oidcConfigured) {
        $$renderer2.push("<!--[0-->");
        $$renderer2.push(`<p class="alt svelte-1x05zx6">No account? <a href="/signup">Sign up</a></p>`);
      } else {
        $$renderer2.push("<!--[-1-->");
      }
      $$renderer2.push(`<!--]--></form>`);
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]-->`);
  });
}

export { _page as default };
//# sourceMappingURL=_page.svelte-C5RQbsse.js.map
