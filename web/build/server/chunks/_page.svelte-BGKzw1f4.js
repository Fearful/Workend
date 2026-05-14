import { ac as ensure_array_like, a8 as escape_html, a7 as attr } from './renderer-D0X3o35U.js';
import { B as Breadcrumb } from './Breadcrumb-Bl_x71bg.js';
import { P as PageHeader } from './PageHeader-C8D5NSCp.js';
import { P as Panel } from './Panel-B1HKSjmN.js';
import { B as Badge } from './Badge-imHJGUZA.js';
import { E as EmptyState } from './EmptyState-5A_wG7lI.js';
import { F as FlashMessage } from './FlashMessage-CReJ7kuH.js';

function _page($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    let { data, form } = $$props;
    let subConfig = "";
    let subScopeType = "global";
    let subSeverity = "failures";
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
    $$renderer2.push(`<!----> `);
    {
      let actions = function($$renderer3) {
        $$renderer3.push(`<span style="color: var(--text-dim); font-size: var(--fs-xs);">${escape_html(data.subscriptions.length)} active</span>`);
      };
      Panel($$renderer2, {
        title: "Subscriptions",
        actions,
        children: ($$renderer3) => {
          $$renderer3.push(`<p style="color: var(--text-dim); font-size: var(--fs-sm); margin: 0 0 var(--space-3);">Each subscription routes runs to one notification target. A new target gets a default <code>global · failures</code> subscription so it isn't silent. Override per-workspace or per-project below — most-specific scope wins.</p> `);
          if (data.subscriptions.length === 0) {
            $$renderer3.push("<!--[0-->");
            EmptyState($$renderer3, { message: "No subscriptions yet." });
          } else {
            $$renderer3.push("<!--[-1-->");
            $$renderer3.push(`<!--[-->`);
            const each_array_1 = ensure_array_like(data.subscriptions);
            for (let $$index_1 = 0, $$length = each_array_1.length; $$index_1 < $$length; $$index_1++) {
              let s = each_array_1[$$index_1];
              $$renderer3.push(`<div class="row sub-row svelte-e87wh7">`);
              Badge($$renderer3, {
                variant: kindVariant(s.config_kind ?? "webhook"),
                size: "sm",
                children: ($$renderer4) => {
                  $$renderer4.push(`<!---->${escape_html(s.config_kind ?? "?")}`);
                }
              });
              $$renderer3.push(`<!----> <span><span class="scope-tag svelte-e87wh7">${escape_html(s.scope_type)}</span> <span class="mono svelte-e87wh7">${escape_html(s.scope_name || (s.scope_type === "global" ? "all runs" : "?"))}</span></span> <span class="muted svelte-e87wh7">severity: ${escape_html(s.severity)}</span> <span></span> <form method="POST" action="?/unsubscribe" class="inline-form"><input type="hidden" name="id"${attr("value", s.id)}/> <button type="submit" class="ghost">Remove</button></form></div>`);
            }
            $$renderer3.push(`<!--]-->`);
          }
          $$renderer3.push(`<!--]--> `);
          if (data.notifications.length > 0) {
            $$renderer3.push("<!--[0-->");
            $$renderer3.push(`<form method="POST" action="?/subscribe" class="sub-form svelte-e87wh7"><h3 class="svelte-e87wh7">Add subscription</h3> <div class="form-grid sub-form-grid svelte-e87wh7"><div class="field"><label for="sub-config">Target</label> `);
            $$renderer3.select(
              {
                id: "sub-config",
                name: "config_id",
                value: subConfig,
                required: true,
                class: ""
              },
              ($$renderer4) => {
                $$renderer4.push(`<!--[-->`);
                const each_array_2 = ensure_array_like(data.notifications);
                for (let $$index_2 = 0, $$length = each_array_2.length; $$index_2 < $$length; $$index_2++) {
                  let n = each_array_2[$$index_2];
                  $$renderer4.option({ value: n.id }, ($$renderer5) => {
                    $$renderer5.push(`${escape_html(n.kind)} → ${escape_html(n.target.length > 40 ? n.target.slice(0, 40) + "…" : n.target)}`);
                  });
                }
                $$renderer4.push(`<!--]-->`);
              },
              "svelte-e87wh7"
            );
            $$renderer3.push(`</div> <div class="field"><label for="sub-scope-type">Scope</label> `);
            $$renderer3.select(
              {
                id: "sub-scope-type",
                name: "scope_type",
                value: subScopeType,
                class: ""
              },
              ($$renderer4) => {
                $$renderer4.option({ value: "global" }, ($$renderer5) => {
                  $$renderer5.push(`Global (all runs)`);
                });
                $$renderer4.option({ value: "workspace" }, ($$renderer5) => {
                  $$renderer5.push(`Workspace`);
                });
                $$renderer4.option({ value: "project" }, ($$renderer5) => {
                  $$renderer5.push(`Project`);
                });
              },
              "svelte-e87wh7"
            );
            $$renderer3.push(`</div> `);
            {
              $$renderer3.push("<!--[-1-->");
              $$renderer3.push(`<input type="hidden" name="scope_id" value="" class="svelte-e87wh7"/> <span></span>`);
            }
            $$renderer3.push(`<!--]--> <div class="field"><label for="sub-severity">Severity</label> `);
            $$renderer3.select(
              {
                id: "sub-severity",
                name: "severity",
                value: subSeverity,
                class: ""
              },
              ($$renderer4) => {
                $$renderer4.option({ value: "failures" }, ($$renderer5) => {
                  $$renderer5.push(`Failures only`);
                });
                $$renderer4.option({ value: "all" }, ($$renderer5) => {
                  $$renderer5.push(`All runs`);
                });
                $$renderer4.option({ value: "off" }, ($$renderer5) => {
                  $$renderer5.push(`Mute`);
                });
              },
              "svelte-e87wh7"
            );
            $$renderer3.push(`</div> <button type="submit"${attr("disabled", subScopeType !== "global", true)}>Subscribe</button></div></form>`);
          } else {
            $$renderer3.push("<!--[-1-->");
          }
          $$renderer3.push(`<!--]-->`);
        }
      });
    }
    $$renderer2.push(`<!---->`);
  });
}

export { _page as default };
//# sourceMappingURL=_page.svelte-BGKzw1f4.js.map
