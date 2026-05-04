import { a7 as attr, a8 as escape_html } from './renderer-mjPKoiGx.js';
import { P as PageHeader } from './PageHeader-CaLG0rbe.js';
import { F as FlashMessage } from './FlashMessage-6DSDTh4l.js';

function _page($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    let { form } = $$props;
    PageHeader($$renderer2, { title: "New workspace" });
    $$renderer2.push(`<!----> <form method="POST" class="svelte-1bzkdak"><div class="field"><label for="name">Name</label> <input id="name" name="name" type="text" required="" maxlength="100"${attr("value", form?.name || "")}/></div> <div class="field"><label for="description">Description (optional)</label> <textarea id="description" name="description" maxlength="500" class="svelte-1bzkdak">`);
    const $$body = escape_html(form?.description || "");
    if ($$body) {
      $$renderer2.push(`${$$body}`);
    }
    $$renderer2.push(`</textarea></div> `);
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
    $$renderer2.push(`<!--]--> <div class="actions svelte-1bzkdak"><button type="submit">Create</button> <a href="/"><button type="button" class="ghost">Cancel</button></a></div></form>`);
  });
}

export { _page as default };
//# sourceMappingURL=_page.svelte-CuQBZxeY.js.map
