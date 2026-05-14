import { a8 as escape_html, a7 as attr, ac as ensure_array_like } from './renderer-D0X3o35U.js';
import { b as formatRelative } from './utils2-BUPlP7zG.js';
import { P as Panel } from './Panel-B1HKSjmN.js';
import { B as Badge } from './Badge-imHJGUZA.js';
import { F as FlashMessage } from './FlashMessage-CReJ7kuH.js';

function _page($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    let { data, form } = $$props;
    let draft = "";
    $$renderer2.push(`<button type="button" class="back-button svelte-2ebdn1">← back</button> <div class="header svelte-2ebdn1"><div><span class="num svelte-2ebdn1">#${escape_html(data.issue.provider_number)}</span> <h1 class="svelte-2ebdn1">${escape_html(data.issue.title)}</h1> <div class="meta svelte-2ebdn1">`);
    Badge($$renderer2, {
      variant: data.issue.state === "open" ? "success" : "info",
      size: "sm",
      children: ($$renderer3) => {
        $$renderer3.push(`<!---->${escape_html(data.issue.state)}`);
      }
    });
    $$renderer2.push(`<!----> `);
    if (data.issue.author_handle) {
      $$renderer2.push("<!--[0-->");
      $$renderer2.push(`<span>opened by <span class="author svelte-2ebdn1">${escape_html(data.issue.author_handle)}</span></span>`);
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--> `);
    if (data.issue.upstream_updated_at) {
      $$renderer2.push("<!--[0-->");
      $$renderer2.push(`<span>· updated ${escape_html(formatRelative(data.issue.upstream_updated_at))}</span>`);
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--> <a${attr("href", data.issue.html_url)} target="_blank" rel="noopener" class="ext-link svelte-2ebdn1">view on provider →</a></div> `);
    if (data.issue.labels.length > 0) {
      $$renderer2.push("<!--[0-->");
      $$renderer2.push(`<div class="meta svelte-2ebdn1"><!--[-->`);
      const each_array = ensure_array_like(data.issue.labels);
      for (let $$index = 0, $$length = each_array.length; $$index < $$length; $$index++) {
        let l = each_array[$$index];
        Badge($$renderer2, {
          variant: "muted",
          size: "sm",
          children: ($$renderer3) => {
            $$renderer3.push(`<!---->${escape_html(l)}`);
          }
        });
      }
      $$renderer2.push(`<!--]--></div>`);
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--></div> `);
    if (data.issue.state === "open") {
      $$renderer2.push("<!--[0-->");
      $$renderer2.push(`<form method="POST" action="?/close" class="inline-form"><button type="submit" class="danger">Close issue</button></form>`);
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--></div> `);
    if (form?.closeError) {
      $$renderer2.push("<!--[0-->");
      FlashMessage($$renderer2, {
        type: "error",
        children: ($$renderer3) => {
          $$renderer3.push(`<!---->${escape_html(form.closeError)}`);
        }
      });
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--> `);
    if (form?.closed) {
      $$renderer2.push("<!--[0-->");
      FlashMessage($$renderer2, {
        type: "success",
        children: ($$renderer3) => {
          $$renderer3.push(`<!---->Issue closed.`);
        }
      });
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--> `);
    if (data.issue.body) {
      $$renderer2.push("<!--[0-->");
      Panel($$renderer2, {
        children: ($$renderer3) => {
          $$renderer3.push(`<div class="body svelte-2ebdn1">${escape_html(data.issue.body)}</div>`);
        }
      });
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--> `);
    Panel($$renderer2, {
      title: "Comments",
      children: ($$renderer3) => {
        if (data.issue.comments.length === 0) {
          $$renderer3.push("<!--[0-->");
          $$renderer3.push(`<p style="color: var(--text-dim); font-size: 0.875rem; margin: 0;">No comments yet.</p>`);
        } else {
          $$renderer3.push("<!--[-1-->");
          $$renderer3.push(`<!--[-->`);
          const each_array_1 = ensure_array_like(data.issue.comments);
          for (let $$index_1 = 0, $$length = each_array_1.length; $$index_1 < $$length; $$index_1++) {
            let c = each_array_1[$$index_1];
            $$renderer3.push(`<div class="comment svelte-2ebdn1"><div class="comment-head svelte-2ebdn1"><span class="author svelte-2ebdn1">${escape_html(c.author_handle || "someone")}</span> <span>${escape_html(formatRelative(c.created_at))}</span> `);
            if (c.html_url) {
              $$renderer3.push("<!--[0-->");
              $$renderer3.push(`<a${attr("href", c.html_url)} target="_blank" rel="noopener" class="comment-actions svelte-2ebdn1">view</a>`);
            } else {
              $$renderer3.push("<!--[-1-->");
            }
            $$renderer3.push(`<!--]--></div> <div class="body svelte-2ebdn1">${escape_html(c.body)}</div></div>`);
          }
          $$renderer3.push(`<!--]-->`);
        }
        $$renderer3.push(`<!--]--> <form method="POST" action="?/comment" style="margin-top: var(--space-4);"><textarea name="body" rows="4" placeholder="Add a comment…">`);
        const $$body = escape_html(draft);
        if ($$body) {
          $$renderer3.push(`${$$body}`);
        }
        $$renderer3.push(`</textarea> `);
        if (form?.commentError) {
          $$renderer3.push("<!--[0-->");
          FlashMessage($$renderer3, {
            type: "error",
            children: ($$renderer4) => {
              $$renderer4.push(`<!---->${escape_html(form.commentError)}`);
            }
          });
        } else {
          $$renderer3.push("<!--[-1-->");
        }
        $$renderer3.push(`<!--]--> `);
        if (form?.commented) {
          $$renderer3.push("<!--[0-->");
          FlashMessage($$renderer3, {
            type: "success",
            children: ($$renderer4) => {
              $$renderer4.push(`<!---->Comment posted.`);
            }
          });
        } else {
          $$renderer3.push("<!--[-1-->");
        }
        $$renderer3.push(`<!--]--> <div class="post-actions svelte-2ebdn1"><button type="submit">Post comment</button></div></form>`);
      }
    });
    $$renderer2.push(`<!---->`);
  });
}

export { _page as default };
//# sourceMappingURL=_page.svelte-B2LTP9S_.js.map
