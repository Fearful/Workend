import { a8 as escape_html, a7 as attr } from './renderer-D0X3o35U.js';
import { P as PageHeader } from './PageHeader-C8D5NSCp.js';
import { F as FlashMessage } from './FlashMessage-CReJ7kuH.js';

function _page($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    let { data, form } = $$props;
    PageHeader($$renderer2, { title: "Sign up" });
    $$renderer2.push(`<!----> `);
    if (data.oidcConfigured) {
      $$renderer2.push("<!--[0-->");
      $$renderer2.push(`<div class="oidc-redirect svelte-kmqcod"><p class="svelte-kmqcod">This instance uses ${escape_html(data.oidcProviderName)} for authentication. Your account will be created automatically when you sign in.</p> <a href="/api/auth/oidc/start"><button type="button">Sign in with ${escape_html(data.oidcProviderName)}</button></a> <p class="alt svelte-kmqcod">Already have an account? <a href="/login">Log in</a></p></div>`);
    } else {
      $$renderer2.push("<!--[-1-->");
      $$renderer2.push(`<form method="POST" class="svelte-kmqcod"><div class="field"><label for="display_name">Display name</label> <input id="display_name" name="display_name" type="text" required=""${attr("value", form?.display_name || "")}/></div> <div class="field"><label for="email">Email</label> <input id="email" name="email" type="email" autocomplete="email" required=""${attr("value", form?.email || "")}/></div> <div class="field"><label for="password">Password</label> <input id="password" name="password" type="password" autocomplete="new-password" required="" minlength="8"/> <p class="hint svelte-kmqcod">At least 8 characters.</p></div> `);
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
      $$renderer2.push(`<!--]--> <button type="submit">Create account</button> <p class="alt svelte-kmqcod">Already have an account? <a href="/login">Log in</a></p></form>`);
    }
    $$renderer2.push(`<!--]-->`);
  });
}

export { _page as default };
//# sourceMappingURL=_page.svelte-DccvH5ds.js.map
