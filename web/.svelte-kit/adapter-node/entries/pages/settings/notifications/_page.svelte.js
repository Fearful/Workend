import { c as ensure_array_like, e as escape_html, a as attr } from "../../../../chunks/renderer.js";
import { B as Breadcrumb } from "../../../../chunks/Breadcrumb.js";
import { P as PageHeader } from "../../../../chunks/PageHeader.js";
import { P as Panel } from "../../../../chunks/Panel.js";
import { B as Badge } from "../../../../chunks/Badge.js";
import { E as EmptyState } from "../../../../chunks/EmptyState.js";
import { F as FlashMessage } from "../../../../chunks/FlashMessage.js";
function _page($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    let { data, form } = $$props;
    function kindVariant(kind) {
      switch (kind) {
        case "webhook":
          return "info";
        case "slack":
          return "muted";
        case "discord":
          return "info";
        case "teams":
          return "info";
        case "email":
          return "success";
        default:
          return "muted";
      }
    }
    Breadcrumb($$renderer2, {
      segments: [
        { label: "settings", href: "/settings" },
        { label: "notifications" }
      ]
    });
    $$renderer2.push(`<!----> `);
    PageHeader($$renderer2, { title: "Notifications" });
    $$renderer2.push(`<!----> `);
    Panel($$renderer2, {
      title: "Configured targets",
      children: ($$renderer3) => {
        if (data.notifications.length === 0) {
          $$renderer3.push("<!--[0-->");
          EmptyState($$renderer3, { message: "No notification targets yet." });
        } else {
          $$renderer3.push("<!--[-1-->");
          $$renderer3.push(`<!--[-->`);
          const each_array = ensure_array_like(data.notifications);
          for (let $$index = 0, $$length = each_array.length; $$index < $$length; $$index++) {
            let n = each_array[$$index];
            $$renderer3.push(`<div class="row svelte-e87wh7">`);
            Badge($$renderer3, {
              variant: kindVariant(n.kind),
              size: "sm",
              children: ($$renderer4) => {
                $$renderer4.push(`<!---->${escape_html(n.kind)}`);
              }
            });
            $$renderer3.push(`<!----> <span class="mono svelte-e87wh7">${escape_html(n.target)}</span> <span class="muted svelte-e87wh7">trigger: ${escape_html(n.trigger)}</span> <form method="POST" action="?/test" class="inline-form"><input type="hidden" name="id"${attr("value", n.id)}/> <button type="submit" class="ghost">Test</button></form> <form method="POST" action="?/delete" class="inline-form"><input type="hidden" name="id"${attr("value", n.id)}/> <button type="submit" class="ghost">Delete</button></form></div>`);
          }
          $$renderer3.push(`<!--]-->`);
        }
        $$renderer3.push(`<!--]-->`);
      }
    });
    $$renderer2.push(`<!----> `);
    Panel($$renderer2, {
      title: "Add notification target",
      children: ($$renderer3) => {
        $$renderer3.push(`<form method="POST" action="?/create"><div class="form-grid svelte-e87wh7"><div class="field"><label for="kind">Kind</label> <select id="kind" name="kind" class="svelte-e87wh7">`);
        $$renderer3.option({ value: "webhook" }, ($$renderer4) => {
          $$renderer4.push(`Webhook`);
        });
        $$renderer3.option({ value: "slack" }, ($$renderer4) => {
          $$renderer4.push(`Slack webhook`);
        });
        $$renderer3.option({ value: "discord" }, ($$renderer4) => {
          $$renderer4.push(`Discord webhook`);
        });
        $$renderer3.option({ value: "teams" }, ($$renderer4) => {
          $$renderer4.push(`Teams webhook`);
        });
        $$renderer3.option({ value: "email" }, ($$renderer4) => {
          $$renderer4.push(`Email`);
        });
        $$renderer3.push(`</select></div> <div class="field"><label for="target">Target</label> <input id="target" name="target" type="text" required="" placeholder="https://hooks.slack.com/..., https://discord.com/api/webhooks/..., or alice@example.com"${attr("value", form?.target || "")} class="svelte-e87wh7"/></div> <div class="field"><label for="trigger">Trigger</label> <select id="trigger" name="trigger" class="svelte-e87wh7">`);
        $$renderer3.option({ value: "on_failure" }, ($$renderer4) => {
          $$renderer4.push(`On failure`);
        });
        $$renderer3.option({ value: "on_status_change" }, ($$renderer4) => {
          $$renderer4.push(`On status change`);
        });
        $$renderer3.option({ value: "always" }, ($$renderer4) => {
          $$renderer4.push(`Always`);
        });
        $$renderer3.push(`</select></div> <button type="submit">Add</button></div> `);
        if (form?.error) {
          $$renderer3.push("<!--[0-->");
          FlashMessage($$renderer3, {
            type: "error",
            children: ($$renderer4) => {
              $$renderer4.push(`<!---->${escape_html(form.error)}`);
            }
          });
        } else {
          $$renderer3.push("<!--[-1-->");
        }
        $$renderer3.push(`<!--]--> `);
        if (form?.tested) {
          $$renderer3.push("<!--[0-->");
          FlashMessage($$renderer3, {
            type: "success",
            children: ($$renderer4) => {
              $$renderer4.push(`<!---->Test notification sent.`);
            }
          });
        } else {
          $$renderer3.push("<!--[-1-->");
        }
        $$renderer3.push(`<!--]--></form>`);
      }
    });
    $$renderer2.push(`<!---->`);
  });
}
export {
  _page as default
};
