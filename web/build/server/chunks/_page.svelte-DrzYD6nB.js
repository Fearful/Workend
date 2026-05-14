import { a8 as escape_html, ac as ensure_array_like, a7 as attr } from './renderer-D0X3o35U.js';
import './root-BPP4VMgG.js';
import './state.svelte-DaEZ9E9Z.js';
import { P as Panel } from './Panel-B1HKSjmN.js';
import { E as EmptyState } from './EmptyState-5A_wG7lI.js';
import { F as FlashMessage } from './FlashMessage-CReJ7kuH.js';
import { S as SectionHeader } from './SectionHeader-BsGXUHV-.js';
import { T as TimeAgo } from './TimeAgo-DaxliHBa.js';
import { M as Modal } from './Modal-BnjRYgDm.js';
import './Tooltip-Bg85MZ7d.js';
import './index-server-BFLhAcPs.js';

function _page($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    let { data, form } = $$props;
    let creatorOpen = false;
    let pipelineName = "";
    let selectedTaskIDs = [];
    let deleteModal = null;
    function taskByID(id) {
      return data.tasks.find((t) => t.id === id);
    }
    {
      let actions = function($$renderer3) {
        $$renderer3.push(`<button type="button">New pipeline</button>`);
      };
      SectionHeader($$renderer2, { title: "Pipelines", actions });
    }
    $$renderer2.push(`<!----> <p style="color: var(--text-dim); font-size: var(--fs-sm); margin: 0 0 var(--space-3);">Chain tasks to run sequentially. Each step is a real run with its own log; the chain stops on first failure.</p> `);
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
    $$renderer2.push(`<!--]--> `);
    Panel($$renderer2, {
      title: "Existing pipelines",
      children: ($$renderer3) => {
        if (data.pipelines.length === 0) {
          $$renderer3.push("<!--[0-->");
          EmptyState($$renderer3, {
            message: "No pipelines yet. Click 'New pipeline' to chain tasks together."
          });
        } else {
          $$renderer3.push("<!--[-1-->");
          $$renderer3.push(`<!--[-->`);
          const each_array = ensure_array_like(data.pipelines);
          for (let $$index_1 = 0, $$length = each_array.length; $$index_1 < $$length; $$index_1++) {
            let p = each_array[$$index_1];
            $$renderer3.push(`<div class="pipeline-row svelte-1yl8vcr"><div><div class="pipeline-name svelte-1yl8vcr">${escape_html(p.name)}</div> <div class="pipeline-steps svelte-1yl8vcr"><!--[-->`);
            const each_array_1 = ensure_array_like(p.steps ?? []);
            for (let i = 0, $$length2 = each_array_1.length; i < $$length2; i++) {
              let s = each_array_1[i];
              $$renderer3.push(`<span class="step-pill svelte-1yl8vcr">${escape_html(s.task_source || "")}<span style="opacity:0.6">${escape_html(s.task_source ? ":" : "")}</span>${escape_html(s.task_name || s.task_id.slice(0, 8))}</span> `);
              if (i < (p.steps?.length ?? 0) - 1) {
                $$renderer3.push("<!--[0-->");
                $$renderer3.push(`<span class="step-arrow svelte-1yl8vcr">→</span>`);
              } else {
                $$renderer3.push("<!--[-1-->");
              }
              $$renderer3.push(`<!--]-->`);
            }
            $$renderer3.push(`<!--]--></div></div> <span style="color: var(--text-dim); font-size: var(--fs-xs);">created `);
            TimeAgo($$renderer3, { value: p.created_at });
            $$renderer3.push(`<!----></span> <div class="pipeline-actions svelte-1yl8vcr"><form method="POST" action="?/run" class="inline-form"><input type="hidden" name="id"${attr("value", p.id)}/> <button type="submit"${attr("disabled", (p.steps ?? []).length === 0, true)}>Run</button></form> <button type="button" class="ghost">Delete</button></div></div>`);
          }
          $$renderer3.push(`<!--]-->`);
        }
        $$renderer3.push(`<!--]-->`);
      }
    });
    $$renderer2.push(`<!----> `);
    Modal($$renderer2, {
      open: creatorOpen,
      title: "New pipeline",
      width: 720,
      onClose: () => creatorOpen = false,
      children: ($$renderer3) => {
        $$renderer3.push(`<form method="POST" action="?/create" class="builder svelte-1yl8vcr"><div class="field"><label for="pipeline-name">Name</label> <input id="pipeline-name" name="name" required=""${attr("value", pipelineName)} placeholder="e.g. ci, deploy, release"/></div> <div class="builder-grid svelte-1yl8vcr"><div class="field"><span class="field-label svelte-1yl8vcr">Available tasks</span> `);
        if (data.tasks.length === 0) {
          $$renderer3.push("<!--[0-->");
          $$renderer3.push(`<p style="color: var(--text-dim); font-size: var(--fs-sm);">No tasks detected for this project.</p>`);
        } else {
          $$renderer3.push("<!--[-1-->");
          $$renderer3.push(`<div class="task-list svelte-1yl8vcr"><!--[-->`);
          const each_array_2 = ensure_array_like(data.tasks);
          for (let $$index_2 = 0, $$length = each_array_2.length; $$index_2 < $$length; $$index_2++) {
            let t = each_array_2[$$index_2];
            $$renderer3.push(`<label class="task-item svelte-1yl8vcr"><input type="checkbox"${attr("checked", selectedTaskIDs.includes(t.id), true)} class="svelte-1yl8vcr"/> <span>${escape_html(t.name)}</span> <span class="task-source svelte-1yl8vcr">(${escape_html(t.source)})</span></label>`);
          }
          $$renderer3.push(`<!--]--></div>`);
        }
        $$renderer3.push(`<!--]--></div> <div class="field"><span class="field-label svelte-1yl8vcr">Pipeline steps (use arrows to reorder)</span> <div class="selected-list svelte-1yl8vcr">`);
        if (selectedTaskIDs.length === 0) {
          $$renderer3.push("<!--[0-->");
          $$renderer3.push(`<div class="selected-empty svelte-1yl8vcr">Select tasks on the left to add them here.</div>`);
        } else {
          $$renderer3.push("<!--[-1-->");
          $$renderer3.push(`<!--[-->`);
          const each_array_3 = ensure_array_like(selectedTaskIDs);
          for (let idx = 0, $$length = each_array_3.length; idx < $$length; idx++) {
            let id = each_array_3[idx];
            const t = taskByID(id);
            $$renderer3.push(`<div class="selected-row svelte-1yl8vcr"><span class="pos svelte-1yl8vcr">${escape_html(idx + 1)}</span> <span>${escape_html(t?.name ?? id)} <span class="task-source svelte-1yl8vcr">(${escape_html(t?.source ?? "?")})</span></span> <button type="button" class="arrow-btn svelte-1yl8vcr"${attr("disabled", idx === 0, true)} title="Move up">↑</button> <button type="button" class="arrow-btn svelte-1yl8vcr"${attr("disabled", idx === selectedTaskIDs.length - 1, true)} title="Move down">↓</button> <button type="button" class="remove-btn svelte-1yl8vcr" title="Remove">×</button> <input type="hidden" name="task_id"${attr("value", id)}/></div>`);
          }
          $$renderer3.push(`<!--]-->`);
        }
        $$renderer3.push(`<!--]--></div></div></div> <div class="builder-footer svelte-1yl8vcr"><span class="builder-hint svelte-1yl8vcr">${escape_html(selectedTaskIDs.length)} task${escape_html(selectedTaskIDs.length === 1 ? "" : "s")} selected</span> <div style="display:flex; gap: var(--space-2);"><button type="button" class="ghost">Cancel</button> <button type="submit"${attr("disabled", !pipelineName.trim() || selectedTaskIDs.length === 0, true)}>Create pipeline</button></div></div></form>`);
      }
    });
    $$renderer2.push(`<!----> `);
    Modal($$renderer2, {
      open: deleteModal !== null,
      title: "Delete pipeline?",
      width: 420,
      onClose: () => deleteModal = null,
      children: ($$renderer3) => {
        if (deleteModal) {
          $$renderer3.push("<!--[0-->");
          $$renderer3.push(`<p style="margin: 0 0 var(--space-3) 0;">Permanently delete pipeline <strong>${escape_html(deleteModal.name)}</strong>? Past pipeline runs are kept; only the chain definition is removed.</p> <div style="display:flex; gap: var(--space-2); justify-content:flex-end;"><button type="button" class="ghost">Cancel</button> <form method="POST" action="?/delete" class="inline-form"><input type="hidden" name="id"${attr("value", deleteModal.id)}/> <button type="submit" class="danger">Delete</button></form></div>`);
        } else {
          $$renderer3.push("<!--[-1-->");
        }
        $$renderer3.push(`<!--]-->`);
      }
    });
    $$renderer2.push(`<!---->`);
  });
}

export { _page as default };
//# sourceMappingURL=_page.svelte-DrzYD6nB.js.map
