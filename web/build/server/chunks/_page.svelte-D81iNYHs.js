import { a7 as attr, ac as ensure_array_like, a8 as escape_html, a2 as derived } from './renderer-D0X3o35U.js';
import { B as Breadcrumb } from './Breadcrumb-Bl_x71bg.js';
import { E as EmptyState } from './EmptyState-5A_wG7lI.js';
import { F as FlashMessage } from './FlashMessage-CReJ7kuH.js';
import { B as Badge } from './Badge-imHJGUZA.js';
import { T as TimeAgo } from './TimeAgo-DaxliHBa.js';
import './Tooltip-Bg85MZ7d.js';
import './index-server-BFLhAcPs.js';

function _page($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    let { data } = $$props;
    let extraEvents = [];
    let allEvents = derived(() => [...data.events, ...extraEvents]);
    let cursor = derived(() => data.nextCursor);
    let loading = false;
    function eventColor(eventType) {
      if (eventType.includes("complete") || eventType.includes("passed") || eventType.includes("succeeded")) return "success";
      if (eventType.includes("fail") || eventType.includes("error")) return "danger";
      if (eventType.includes("member_added") || eventType.includes("joined") || eventType.includes("invited")) return "info";
      if (eventType.includes("start") || eventType.includes("running") || eventType.includes("queued")) return "warning";
      if (eventType.includes("create") || eventType.includes("deploy")) return "accent";
      return "muted";
    }
    function eventIcon(eventType) {
      if (eventType.includes("run_complete") || eventType.includes("succeeded")) return "✓";
      if (eventType.includes("run_fail") || eventType.includes("failed")) return "✗";
      if (eventType.includes("run_start") || eventType.includes("running")) return "▶";
      if (eventType.includes("member")) return "○";
      if (eventType.includes("project")) return "◇";
      if (eventType.includes("secret")) return "⬡";
      if (eventType.includes("deploy")) return "↑";
      return "·";
    }
    function entityHref(entityType, entityId) {
      if (!entityId) return null;
      if (entityType === "run") return `/runs/${entityId}`;
      if (entityType === "project") return `/projects/${entityId}`;
      if (entityType === "pipeline") return `/pipelines/${entityId}`;
      return null;
    }
    function initials(name) {
      if (!name) return "?";
      const parts = name.trim().split(/\s+/);
      if (parts.length >= 2) return (parts[0][0] + parts[1][0]).toUpperCase();
      return name.slice(0, 2).toUpperCase();
    }
    function dateGroup(iso) {
      const d = new Date(iso);
      const now = /* @__PURE__ */ new Date();
      const today = new Date(now.getFullYear(), now.getMonth(), now.getDate());
      const eventDate = new Date(d.getFullYear(), d.getMonth(), d.getDate());
      const diffDays = Math.floor((today.getTime() - eventDate.getTime()) / 864e5);
      if (diffDays === 0) return "Today";
      if (diffDays === 1) return "Yesterday";
      if (diffDays < 7) return d.toLocaleDateString(void 0, { weekday: "long" });
      return d.toLocaleDateString(void 0, {
        month: "short",
        day: "numeric",
        year: d.getFullYear() !== now.getFullYear() ? "numeric" : void 0
      });
    }
    let groupedEvents = derived(() => {
      const groups = [];
      let currentLabel = "";
      for (const e of allEvents()) {
        const label = dateGroup(e.created_at);
        if (label !== currentLabel) {
          groups.push({ label, events: [] });
          currentLabel = label;
        }
        groups[groups.length - 1].events.push(e);
      }
      return groups;
    });
    Breadcrumb($$renderer2, {
      segments: [
        { label: "workspaces", href: "/" },
        {
          label: data.workspace.name,
          href: `/workspaces/${data.workspace.id}`
        },
        { label: "activity" }
      ]
    });
    $$renderer2.push(`<!----> <div class="header-row svelte-1kxa1gx"><h1 class="svelte-1kxa1gx">Activity</h1></div> <nav class="ws-tabs svelte-1kxa1gx" aria-label="Workspace sections"><a class="ws-tab svelte-1kxa1gx"${attr("href", `/workspaces/${data.workspace.id}`)}>Overview</a> <a class="ws-tab svelte-1kxa1gx"${attr("href", `/workspaces/${data.workspace.id}/dashboard`)}>Dashboard</a> <a class="ws-tab active svelte-1kxa1gx"${attr("href", `/workspaces/${data.workspace.id}/activity`)}>Activity</a> <a class="ws-tab svelte-1kxa1gx"${attr("href", `/workspaces/${data.workspace.id}/secrets`)}>Secrets</a> <a class="ws-tab svelte-1kxa1gx"${attr("href", `/workspaces/${data.workspace.id}/roles`)}>Roles</a></nav> `);
    if (data.feedError) {
      $$renderer2.push("<!--[0-->");
      FlashMessage($$renderer2, {
        type: "error",
        children: ($$renderer3) => {
          $$renderer3.push(`<!---->${escape_html(data.feedError)}`);
        }
      });
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--> `);
    if (allEvents().length === 0) {
      $$renderer2.push("<!--[0-->");
      EmptyState($$renderer2, {
        icon: "○",
        message: "No activity yet. Events will appear here as your team works."
      });
    } else {
      $$renderer2.push("<!--[-1-->");
      $$renderer2.push(`<div class="timeline svelte-1kxa1gx"><!--[-->`);
      const each_array = ensure_array_like(groupedEvents());
      for (let $$index_1 = 0, $$length = each_array.length; $$index_1 < $$length; $$index_1++) {
        let group = each_array[$$index_1];
        $$renderer2.push(`<div class="date-label svelte-1kxa1gx">${escape_html(group.label)}</div> <!--[-->`);
        const each_array_1 = ensure_array_like(group.events);
        for (let $$index = 0, $$length2 = each_array_1.length; $$index < $$length2; $$index++) {
          let event = each_array_1[$$index];
          const href = entityHref(event.entity_type, event.entity_id);
          $$renderer2.push(`<div class="event-row svelte-1kxa1gx"><span class="event-marker svelte-1kxa1gx">${escape_html(eventIcon(event.event_type))}</span> <div class="event-body svelte-1kxa1gx"><div class="event-header svelte-1kxa1gx"><span class="avatar svelte-1kxa1gx">${escape_html(initials(event.user_name))}</span> <span class="event-user svelte-1kxa1gx">${escape_html(event.user_name || "System")}</span> `);
          Badge($$renderer2, {
            variant: eventColor(event.event_type),
            size: "sm",
            children: ($$renderer3) => {
              $$renderer3.push(`<!---->${escape_html(event.event_type.replace(/_/g, " "))}`);
            }
          });
          $$renderer2.push(`<!----></div> <div style="margin-top: 2px;"><span class="event-summary svelte-1kxa1gx">${escape_html(event.summary)}</span> `);
          if (href) {
            $$renderer2.push("<!--[0-->");
            $$renderer2.push(`<a${attr("href", href)} class="event-entity svelte-1kxa1gx">${escape_html(event.entity_id.slice(0, 8))}</a>`);
          } else {
            $$renderer2.push("<!--[-1-->");
          }
          $$renderer2.push(`<!--]--></div></div> <span class="event-time svelte-1kxa1gx">`);
          TimeAgo($$renderer2, { value: event.created_at });
          $$renderer2.push(`<!----></span></div>`);
        }
        $$renderer2.push(`<!--]-->`);
      }
      $$renderer2.push(`<!--]--></div> `);
      if (cursor()) {
        $$renderer2.push("<!--[0-->");
        $$renderer2.push(`<div class="load-more svelte-1kxa1gx"><button${attr("disabled", loading, true)}>${escape_html("Load more")}</button></div>`);
      } else {
        $$renderer2.push("<!--[-1-->");
      }
      $$renderer2.push(`<!--]-->`);
    }
    $$renderer2.push(`<!--]-->`);
  });
}

export { _page as default };
//# sourceMappingURL=_page.svelte-D81iNYHs.js.map
