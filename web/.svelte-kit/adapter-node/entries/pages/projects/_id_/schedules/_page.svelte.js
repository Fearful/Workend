import { c as ensure_array_like, e as escape_html, a as attr } from "../../../../../chunks/renderer.js";
import { P as Panel } from "../../../../../chunks/Panel.js";
import { B as Badge } from "../../../../../chunks/Badge.js";
import { E as EmptyState } from "../../../../../chunks/EmptyState.js";
import { F as FlashMessage } from "../../../../../chunks/FlashMessage.js";
import { M as Modal } from "../../../../../chunks/Modal.js";
function _page($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    let { data, form } = $$props;
    let deleteModal = null;
    $$renderer2.push(`<h2 class="section-title svelte-1crrz5y">Schedules</h2> `);
    Panel($$renderer2, {
      title: "Existing schedules",
      children: ($$renderer3) => {
        if (data.schedules.length === 0) {
          $$renderer3.push("<!--[0-->");
          EmptyState($$renderer3, { message: "No schedules yet." });
        } else {
          $$renderer3.push("<!--[-1-->");
          $$renderer3.push(`<!--[-->`);
          const each_array = ensure_array_like(data.schedules);
          for (let $$index = 0, $$length = each_array.length; $$index < $$length; $$index++) {
            let s = each_array[$$index];
            $$renderer3.push(`<div class="row svelte-1crrz5y"><span class="mono svelte-1crrz5y">${escape_html(s.task_name)} <span class="dim svelte-1crrz5y">(${escape_html(s.task_source)})</span></span> <span class="mono svelte-1crrz5y">${escape_html(s.cron_expr)}</span> `);
            if (s.enabled) {
              $$renderer3.push("<!--[0-->");
              Badge($$renderer3, {
                variant: "success",
                size: "sm",
                children: ($$renderer4) => {
                  $$renderer4.push(`<!---->ON`);
                }
              });
            } else {
              $$renderer3.push("<!--[-1-->");
              Badge($$renderer3, {
                variant: "muted",
                size: "sm",
                children: ($$renderer4) => {
                  $$renderer4.push(`<!---->OFF`);
                }
              });
            }
            $$renderer3.push(`<!--]--> <span class="muted svelte-1crrz5y">next: ${escape_html(s.next_run_at ? new Date(s.next_run_at).toLocaleString() : "—")}</span> <form method="POST" action="?/toggle" class="inline-form"><input type="hidden" name="id"${attr("value", s.id)}/> <button type="submit" class="ghost">${escape_html(s.enabled ? "Pause" : "Resume")}</button></form> <button type="button" class="ghost">Delete</button></div>`);
          }
          $$renderer3.push(`<!--]-->`);
        }
        $$renderer3.push(`<!--]-->`);
      }
    });
    $$renderer2.push(`<!----> `);
    Panel($$renderer2, {
      title: "New schedule",
      children: ($$renderer3) => {
        if (data.tasks.length === 0) {
          $$renderer3.push("<!--[0-->");
          EmptyState($$renderer3, {
            message: "No tasks detected for this project. Add a package.json or justfile to its repo."
          });
        } else {
          $$renderer3.push("<!--[-1-->");
          $$renderer3.push(`<form method="POST" action="?/create"><div class="form-grid svelte-1crrz5y"><div class="field"><label for="task_id">Task</label> <select id="task_id" name="task_id" required="" class="schedule-select svelte-1crrz5y"><!--[-->`);
          const each_array_1 = ensure_array_like(data.tasks);
          for (let $$index_1 = 0, $$length = each_array_1.length; $$index_1 < $$length; $$index_1++) {
            let t = each_array_1[$$index_1];
            $$renderer3.option({ value: t.id }, ($$renderer4) => {
              $$renderer4.push(`${escape_html(t.source)}: ${escape_html(t.name)}`);
            });
          }
          $$renderer3.push(`<!--]--></select></div> <div class="field"><label for="cron_expr">Cron expression</label> <input id="cron_expr" name="cron_expr" type="text" required="" placeholder="0 9 * * 1"${attr("value", form?.cron_expr || "")}/> <small class="hint svelte-1crrz5y">Standard 5-field: m h dom mon dow. <code>0 9 * * 1</code> = 9am every Monday.</small></div> <label class="checkbox-label svelte-1crrz5y"><input type="checkbox" name="enabled" checked="" class="svelte-1crrz5y"/> <span class="svelte-1crrz5y">Enabled</span></label> <button type="submit">Create</button></div> `);
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
          $$renderer3.push(`<!--]--></form>`);
        }
        $$renderer3.push(`<!--]-->`);
      }
    });
    $$renderer2.push(`<!----> `);
    Modal($$renderer2, {
      open: deleteModal !== null,
      title: "Delete schedule?",
      width: 420,
      onClose: () => deleteModal = null,
      children: ($$renderer3) => {
        if (deleteModal) {
          $$renderer3.push("<!--[0-->");
          $$renderer3.push(`<p style="margin: 0 0 var(--space-3) 0;">Stop running <strong>${escape_html(deleteModal.taskName)}</strong> on <code>${escape_html(deleteModal.cron)}</code>?
      Past runs are kept; only the schedule is removed.</p> <div style="display:flex; gap: var(--space-2); justify-content:flex-end;"><button type="button" class="ghost">Cancel</button> <form method="POST" action="?/delete" class="inline-form"><input type="hidden" name="id"${attr("value", deleteModal.id)}/> <button type="submit" class="danger">Delete</button></form></div>`);
        } else {
          $$renderer3.push("<!--[-1-->");
        }
        $$renderer3.push(`<!--]-->`);
      }
    });
    $$renderer2.push(`<!---->`);
  });
}
export {
  _page as default
};
