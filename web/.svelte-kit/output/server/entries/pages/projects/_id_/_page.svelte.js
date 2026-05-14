import { c as ensure_array_like, f as attr_style, s as stringify, d as derived, a as attr, b as attr_class, e as escape_html } from "../../../../chunks/renderer.js";
import { o as onDestroy } from "../../../../chunks/index-server.js";
import "@sveltejs/kit/internal";
import "../../../../chunks/exports.js";
import "../../../../chunks/utils.js";
import "@sveltejs/kit/internal/server";
import "../../../../chunks/root.js";
import "../../../../chunks/state.svelte.js";
import { c as formatDuration, s as shortSha } from "../../../../chunks/utils2.js";
import { P as Panel } from "../../../../chunks/Panel.js";
import { S as StatusPill } from "../../../../chunks/StatusPill.js";
import { B as Badge } from "../../../../chunks/Badge.js";
import { F as FlashMessage } from "../../../../chunks/FlashMessage.js";
import { M as Modal } from "../../../../chunks/Modal.js";
import { T as TimeAgo } from "../../../../chunks/TimeAgo.js";
import { T as Tooltip } from "../../../../chunks/Tooltip.js";
function TaskSparkline($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    let { runs, max = 10 } = $$props;
    let display = derived(() => {
      const slice = runs.slice(0, max);
      return slice.reverse();
    });
    function color(status) {
      switch (status) {
        case "succeeded":
          return "var(--status-success-fg)";
        case "failed":
          return "var(--status-danger-fg)";
        case "cancelled":
          return "var(--text-dim)";
        case "running":
        case "queued":
          return "var(--status-warning-fg)";
        default:
          return "var(--text-dim)";
      }
    }
    function tipText() {
      if (display().length === 0) return "No runs yet";
      const latest = runs[0];
      const succeeded = runs.filter((r) => r.status === "succeeded").length;
      const failed = runs.filter((r) => r.status === "failed").length;
      return `${runs.length} run${runs.length === 1 ? "" : "s"} · ${succeeded} ok / ${failed} failed · last: ${latest.status}`;
    }
    if (display().length > 0) {
      $$renderer2.push("<!--[0-->");
      Tooltip($$renderer2, {
        text: tipText(),
        children: ($$renderer3) => {
          $$renderer3.push(`<span class="spark svelte-2wy1bs" aria-hidden="true"><!--[-->`);
          const each_array = ensure_array_like(display());
          for (let $$index = 0, $$length = each_array.length; $$index < $$length; $$index++) {
            let r = each_array[$$index];
            $$renderer3.push(`<span class="dot svelte-2wy1bs"${attr_style(`background: ${stringify(color(r.status))};`)}></span>`);
          }
          $$renderer3.push(`<!--]--></span>`);
        }
      });
    } else {
      $$renderer2.push("<!--[-1-->");
      $$renderer2.push(`<span class="spark empty svelte-2wy1bs"></span>`);
    }
    $$renderer2.push(`<!--]-->`);
  });
}
function _page($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    let { data, form } = $$props;
    function shortSha$1(sha) {
      return shortSha(sha, 12);
    }
    let collapsedPrefixes = /* @__PURE__ */ new Set();
    let taskFilter = "";
    let openMenuTaskID = null;
    let runsByTask = derived(() => {
      const out = {};
      for (const r of data.runs) {
        if (!out[r.task_id]) out[r.task_id] = [];
        out[r.task_id].push({ id: r.id, status: r.status, created_at: r.created_at });
      }
      return out;
    });
    let tasksBySource = derived(() => {
      const raw = {};
      for (const t of data.tasks) {
        if (!raw[t.source]) raw[t.source] = { all: [], byPrefix: {} };
        const sepIdx = Math.max(t.name.indexOf(":"), t.name.indexOf("."));
        if (sepIdx > 0) {
          const prefix = t.name.slice(0, sepIdx + 1);
          if (!raw[t.source].byPrefix[prefix]) raw[t.source].byPrefix[prefix] = [];
          raw[t.source].byPrefix[prefix].push(t);
        } else {
          raw[t.source].all.push(t);
        }
      }
      const result = {};
      for (const [source, g] of Object.entries(raw)) {
        result[source] = { ungrouped: [...g.all], prefixed: {} };
        for (const [prefix, tasks] of Object.entries(g.byPrefix)) {
          if (tasks.length >= 2) {
            result[source].prefixed[prefix] = tasks;
          } else {
            result[source].ungrouped.push(...tasks);
          }
        }
      }
      return result;
    });
    let topLangs = derived(() => {
      if (!data.stats) return [];
      return Object.entries(data.stats.languages).sort((a, b) => b[1].lines - a[1].lines).slice(0, 6);
    });
    let totalLangLines = derived(() => data.stats?.total_lines || 1);
    function formatTimeout(sec) {
      if (!sec || sec <= 0) return "default";
      if (sec < 60) return `${sec}s`;
      if (sec < 3600) return `${Math.round(sec / 60)}m`;
      return `${Math.round(sec / 360) / 10}h`;
    }
    let paramsModal = null;
    let timeoutModal = null;
    let retryModal = null;
    let concurrencyModal = null;
    let branchModal = null;
    let servicesModal = null;
    let artifactsModal = null;
    let syncingCommits = false;
    let blameDetailModal = null;
    function blameRowColor(c) {
      if (c.run_count === 0) return "";
      if (c.fail_count > 0) return "blame-fail";
      return "blame-pass";
    }
    let detectingPackages = false;
    let autoMapping = false;
    let packageDetailModal = null;
    function pkgTypeBadgeVariant(t) {
      switch (t) {
        case "npm":
          return "accent";
        case "go":
          return "info";
        case "cargo":
          return "warning";
        case "python":
          return "success";
        case "gradle":
          return "danger";
        default:
          return "muted";
      }
    }
    let newPreviewBranch = "";
    let previewActionPending = null;
    let previewDeleteConfirm = null;
    let metricsModal = null;
    function formatMs(ms) {
      if (ms < 1e3) return `${Math.round(ms)}ms`;
      if (ms < 6e4) return `${(ms / 1e3).toFixed(1)}s`;
      return `${(ms / 6e4).toFixed(1)}m`;
    }
    function formatConcurrency(max, policy) {
      if (max <= 0) return "single";
      return `${max}× (${policy})`;
    }
    function formatArtifacts(arr) {
      if (!arr || arr.length === 0) return "none";
      return `${arr.length} pattern${arr.length === 1 ? "" : "s"}`;
    }
    function formatServices(arr) {
      if (!arr || arr.length === 0) return "none";
      if (arr.length <= 2) return arr.join(", ");
      return `${arr.slice(0, 2).join(", ")} +${arr.length - 2}`;
    }
    function formatRetry(max, backoff) {
      if (max <= 0) return "no retry";
      return `${max}× / ${backoff}s`;
    }
    function closeMenuOnClickOutside(e) {
      const t = e.target;
      if (!t.closest(".task-menu") && !t.closest(".task-menu-trigger")) {
        openMenuTaskID = null;
      }
    }
    const PANEL_KEYS = [
      "tasks",
      "recent-runs",
      "repository",
      "latest-commit",
      "code-stats",
      "about",
      "contributors",
      "pipelines",
      "pipeline-configs",
      "compose",
      "blame-timeline",
      "monorepo",
      "previews",
      "alert-rules",
      "impact-radar",
      "dependency-tree",
      "live-run"
    ];
    let panelLayout = PANEL_KEYS.map((k) => [k]);
    let dragSource = null;
    let dropTarget = null;
    let editLayoutMode = false;
    function panelHasData(key) {
      switch (key) {
        case "tasks":
        case "repository":
        case "latest-commit":
          return true;
        case "recent-runs":
          return data.runs.length > 0;
        case "code-stats":
          return data.stats != null;
        case "pipelines":
          return data.remotePipelines.length > 0;
        case "pipeline-configs":
          return data.pipelineConfigs.length > 0;
        case "compose":
          return !!data.compose?.compose_file;
        case "about":
          return !!data.overview?.readme;
        case "contributors":
          return (data.overview?.contributors.length ?? 0) > 0;
        case "blame-timeline":
          return data.blameTimeline.length > 0;
        case "monorepo":
          return data.monorepoPackages.length > 0;
        case "previews":
          return data.previews.length > 0;
        case "alert-rules":
          return data.alertRules.length > 0 || data.widgetAlertRules.length > 0;
        case "impact-radar":
          return data.impactRadar != null && data.impactRadar.tasks.length > 0;
        case "dependency-tree":
          return data.dependencyTree != null && (data.dependencyTree.upstream.length > 0 || data.dependencyTree.downstream.length > 0);
        case "live-run":
          return data.liveRun != null;
      }
    }
    function visibleRow(row) {
      return row.filter(panelHasData);
    }
    onDestroy(() => {
      if (typeof document !== "undefined") document.removeEventListener("click", closeMenuOnClickOutside);
    });
    function taskRow($$renderer3, t, isPinned) {
      $$renderer3.push(`<div class="task-row svelte-ffmenf"><form method="POST" action="?/togglePin" class="inline-form"><input type="hidden" name="task_id"${attr("value", t.id)} class="svelte-ffmenf"/> <input type="hidden" name="pinned"${attr("value", String(isPinned))} class="svelte-ffmenf"/> <button type="submit"${attr_class(`pin-btn ${stringify(isPinned ? "pinned" : "")}`, "svelte-ffmenf")}${attr("title", isPinned ? "Unpin from dashboard" : "Pin to dashboard")}>${escape_html(isPinned ? "★" : "☆")}</button></form> <div class="task-name-cell svelte-ffmenf"><div class="task-name svelte-ffmenf">${escape_html(t.name)} `);
      if (t.requires_approval) {
        $$renderer3.push("<!--[0-->");
        $$renderer3.push(`<span class="approval-tag svelte-ffmenf">approval</span>`);
      } else {
        $$renderer3.push("<!--[-1-->");
      }
      $$renderer3.push(`<!--]--> <span class="task-spark svelte-ffmenf">`);
      TaskSparkline($$renderer3, { runs: runsByTask()[t.id] ?? [] });
      $$renderer3.push(`<!----></span></div> <div class="task-cmd svelte-ffmenf"${attr("title", t.raw_command)}>${escape_html(t.raw_command)}</div></div> <span class="task-meta svelte-ffmenf"><span title="Timeout">⏱ ${escape_html(formatTimeout(t.timeout_seconds))}</span> <span class="meta-sep svelte-ffmenf">·</span> <span title="Retry policy">↻ ${escape_html(formatRetry(t.retry_max, t.retry_backoff_sec))}</span> `);
      if (t.requires_approval) {
        $$renderer3.push("<!--[0-->");
        $$renderer3.push(`<span class="meta-sep svelte-ffmenf">·</span> <span class="gate-on svelte-ffmenf" title="Approval required">gate</span>`);
      } else {
        $$renderer3.push("<!--[-1-->");
      }
      $$renderer3.push(`<!--]--></span> <span class="task-actions svelte-ffmenf"><button type="button" class="menu-btn task-menu-trigger svelte-ffmenf" title="More actions" aria-haspopup="menu"${attr("aria-expanded", openMenuTaskID === t.id)}>⋯</button> <form method="POST" action="?/run" class="inline-form"><input type="hidden" name="task_id"${attr("value", t.id)} class="svelte-ffmenf"/> <button type="submit" class="run-btn svelte-ffmenf"${attr("disabled", data.project.status !== "ready", true)}>Run</button></form></span> `);
      if (openMenuTaskID === t.id) {
        $$renderer3.push("<!--[0-->");
        $$renderer3.push(`<div class="task-menu svelte-ffmenf" role="menu"><button class="menu-item svelte-ffmenf" type="button" role="menuitem"${attr("disabled", data.project.status !== "ready", true)}><span>Run with parameters…</span></button> <button class="menu-item svelte-ffmenf" type="button" role="menuitem"${attr("disabled", data.project.status !== "ready", true)}><span>Run on branch…</span></button> <button class="menu-item svelte-ffmenf" type="button" role="menuitem"><span>Edit timeout</span> <span class="menu-item-value svelte-ffmenf">${escape_html(formatTimeout(t.timeout_seconds))}</span></button> <button class="menu-item svelte-ffmenf" type="button" role="menuitem"><span>Edit retry policy</span> <span class="menu-item-value svelte-ffmenf">${escape_html(formatRetry(t.retry_max, t.retry_backoff_sec))}</span></button> <button class="menu-item svelte-ffmenf" type="button" role="menuitem"><span>Concurrency policy</span> <span class="menu-item-value svelte-ffmenf">${escape_html(formatConcurrency(t.max_concurrency, t.supersede_policy))}</span></button> <button class="menu-item svelte-ffmenf" type="button" role="menuitem"><span>Required services</span> <span class="menu-item-value svelte-ffmenf">${escape_html(formatServices(t.needs_services ?? []))}</span></button> <button class="menu-item svelte-ffmenf" type="button" role="menuitem"><span>Artifact patterns</span> <span class="menu-item-value svelte-ffmenf">${escape_html(formatArtifacts(t.artifact_patterns ?? []))}</span></button> <button class="menu-item svelte-ffmenf" type="button" role="menuitem"><span>View metrics</span></button> <form method="POST" action="?/toggleApproval" class="inline-form" style="display:contents;"><input type="hidden" name="task_id"${attr("value", t.id)} class="svelte-ffmenf"/> <input type="hidden" name="next"${attr("value", String(!t.requires_approval))} class="svelte-ffmenf"/> <button class="menu-item svelte-ffmenf" type="submit" role="menuitem"><span>${escape_html(t.requires_approval ? "Disable approval gate" : "Require approval")}</span> <span class="menu-item-value svelte-ffmenf">${escape_html(t.requires_approval ? "on" : "off")}</span></button></form></div>`);
      } else {
        $$renderer3.push("<!--[-1-->");
      }
      $$renderer3.push(`<!--]--></div>`);
    }
    function panelTasks($$renderer3) {
      {
        let actions = function($$renderer4) {
          if (data.tasks.length > 0) {
            $$renderer4.push("<!--[0-->");
            $$renderer4.push(`<span class="filter-count svelte-ffmenf">${escape_html(data.tasks.length)} total</span>`);
          } else {
            $$renderer4.push("<!--[-1-->");
          }
          $$renderer4.push(`<!--]-->`);
        };
        Panel($$renderer3, {
          title: "Tasks",
          actions,
          children: ($$renderer4) => {
            if (data.tasks.length === 0) {
              $$renderer4.push("<!--[0-->");
              $$renderer4.push(`<div class="empty svelte-ffmenf">`);
              if (data.project.status === "ready") {
                $$renderer4.push("<!--[0-->");
                $$renderer4.push(`No runnable tasks detected. Add a <code>package.json</code> with a <code>scripts</code> section, or a <code>justfile</code>.`);
              } else {
                $$renderer4.push("<!--[-1-->");
                $$renderer4.push(`Tasks will appear after the repo finishes cloning.`);
              }
              $$renderer4.push(`<!--]--></div>`);
            } else {
              $$renderer4.push("<!--[-1-->");
              if (data.tasks.length > 8) {
                $$renderer4.push("<!--[0-->");
                $$renderer4.push(`<div class="filter-row svelte-ffmenf"><input type="search" class="filter-input svelte-ffmenf" placeholder="Filter tasks…"${attr("value", taskFilter)}/></div>`);
              } else {
                $$renderer4.push("<!--[-1-->");
              }
              $$renderer4.push(`<!--]--> <!--[-->`);
              const each_array = ensure_array_like(Object.entries(tasksBySource()));
              for (let $$index_3 = 0, $$length = each_array.length; $$index_3 < $$length; $$index_3++) {
                let [source, group] = each_array[$$index_3];
                $$renderer4.push(`<div class="task-group svelte-ffmenf"><div class="task-source-header svelte-ffmenf"><span class="task-source-label svelte-ffmenf">${escape_html(source)}</span> <span class="task-source-count svelte-ffmenf">${escape_html(group.ungrouped.length + Object.values(group.prefixed).reduce((s, t) => s + t.length, 0))}</span></div> <!--[-->`);
                const each_array_1 = ensure_array_like(group.ungrouped);
                for (let $$index = 0, $$length2 = each_array_1.length; $$index < $$length2; $$index++) {
                  let t = each_array_1[$$index];
                  const isPinned = data.pinnedTaskIDs.includes(t.id);
                  taskRow($$renderer4, t, isPinned);
                }
                $$renderer4.push(`<!--]--> <!--[-->`);
                const each_array_2 = ensure_array_like(Object.entries(group.prefixed));
                for (let $$index_2 = 0, $$length2 = each_array_2.length; $$index_2 < $$length2; $$index_2++) {
                  let [prefix, prefixTasks] = each_array_2[$$index_2];
                  const prefixKey = source + "/" + prefix;
                  const isOpen = !collapsedPrefixes.has(prefixKey);
                  $$renderer4.push(`<button type="button" class="prefix-toggle svelte-ffmenf"><span class="prefix-chevron svelte-ffmenf">${escape_html(isOpen ? "▾" : "▸")}</span> <span class="prefix-label svelte-ffmenf">${escape_html(prefix)}</span> <span class="prefix-count svelte-ffmenf">(${escape_html(prefixTasks.length)})</span></button> `);
                  if (isOpen) {
                    $$renderer4.push("<!--[0-->");
                    $$renderer4.push(`<!--[-->`);
                    const each_array_3 = ensure_array_like(prefixTasks);
                    for (let $$index_1 = 0, $$length3 = each_array_3.length; $$index_1 < $$length3; $$index_1++) {
                      let t = each_array_3[$$index_1];
                      const isPinned = data.pinnedTaskIDs.includes(t.id);
                      taskRow($$renderer4, t, isPinned);
                    }
                    $$renderer4.push(`<!--]-->`);
                  } else {
                    $$renderer4.push("<!--[-1-->");
                  }
                  $$renderer4.push(`<!--]-->`);
                }
                $$renderer4.push(`<!--]--></div>`);
              }
              $$renderer4.push(`<!--]-->`);
            }
            $$renderer4.push(`<!--]-->`);
          }
        });
      }
    }
    function panelRecentRuns($$renderer3) {
      if (data.runs.length > 0) {
        $$renderer3.push("<!--[0-->");
        {
          let actions = function($$renderer4) {
            $$renderer4.push(`<a${attr("href", `/projects/${data.project.id}/runs`)} class="section-link svelte-ffmenf">View all →</a>`);
          };
          Panel($$renderer3, {
            title: "Recent runs",
            actions,
            children: ($$renderer4) => {
              $$renderer4.push(`<!--[-->`);
              const each_array_4 = ensure_array_like(data.runs.slice(0, 10));
              for (let $$index_4 = 0, $$length = each_array_4.length; $$index_4 < $$length; $$index_4++) {
                let r = each_array_4[$$index_4];
                $$renderer4.push(`<a${attr("href", `/runs/${r.id}`)} class="run-row svelte-ffmenf">`);
                StatusPill($$renderer4, { status: r.status, size: "sm" });
                $$renderer4.push(`<!----> <span class="run-name svelte-ffmenf">${escape_html(r.task_name)} <span class="dim svelte-ffmenf">(${escape_html(r.task_source)})</span></span> <span class="run-meta svelte-ffmenf">${escape_html(formatDuration(r.started_at, r.finished_at))}</span> <span class="run-meta svelte-ffmenf">`);
                TimeAgo($$renderer4, { value: r.created_at });
                $$renderer4.push(`<!----></span></a>`);
              }
              $$renderer4.push(`<!--]-->`);
            }
          });
        }
      } else {
        $$renderer3.push("<!--[-1-->");
      }
      $$renderer3.push(`<!--]-->`);
    }
    function panelRepository($$renderer3) {
      Panel($$renderer3, {
        title: "Repository",
        children: ($$renderer4) => {
          $$renderer4.push(`<div class="row svelte-ffmenf"><span class="label svelte-ffmenf">Git URL</span><span class="value svelte-ffmenf">${escape_html(data.project.git_url)}</span></div> <div class="row svelte-ffmenf"><span class="label svelte-ffmenf">Branch</span><span class="value svelte-ffmenf">${escape_html(data.project.default_branch || "—")}</span></div> <div class="row svelte-ffmenf"><span class="label svelte-ffmenf">Status</span><span class="value svelte-ffmenf">${escape_html(data.project.status)}</span></div> <div class="row svelte-ffmenf"><span class="label svelte-ffmenf">Local path</span><span class="value svelte-ffmenf">${escape_html(data.project.local_path || "—")}</span></div> `);
          if (data.project.webhook_token) {
            $$renderer4.push("<!--[0-->");
            $$renderer4.push(`<div class="row svelte-ffmenf"><span class="label svelte-ffmenf">Webhook URL</span> <span class="value webhook-value svelte-ffmenf"><code class="webhook-code svelte-ffmenf">${escape_html(`${typeof window !== "undefined" ? window.location.origin : ""}/api/webhooks/projects/${data.project.webhook_token}`)}</code> <button type="button" class="ghost svelte-ffmenf">Copy</button></span></div>`);
          } else {
            $$renderer4.push("<!--[-1-->");
          }
          $$renderer4.push(`<!--]-->`);
        }
      });
    }
    function panelLatestCommit($$renderer3) {
      Panel($$renderer3, {
        title: "Latest commit",
        children: ($$renderer4) => {
          $$renderer4.push(`<div class="row svelte-ffmenf"><span class="label svelte-ffmenf">SHA</span><span class="value svelte-ffmenf">${escape_html(shortSha$1(data.project.last_commit_sha))}</span></div> <div class="row svelte-ffmenf"><span class="label svelte-ffmenf">Author</span><span class="value commit-msg svelte-ffmenf">${escape_html(data.project.last_commit_author || "—")}</span></div> <div class="row svelte-ffmenf"><span class="label svelte-ffmenf">Message</span><span class="value commit-msg svelte-ffmenf">${escape_html(data.project.last_commit_message || "—")}</span></div> <div class="row svelte-ffmenf"><span class="label svelte-ffmenf">Synced</span><span class="value svelte-ffmenf">${escape_html(data.project.last_synced_at ? new Date(data.project.last_synced_at).toLocaleString() : "—")}</span></div>`);
        }
      });
    }
    function panelCodeStats($$renderer3) {
      if (data.stats) {
        $$renderer3.push("<!--[0-->");
        Panel($$renderer3, {
          title: "Code statistics",
          children: ($$renderer4) => {
            $$renderer4.push(`<div class="stats-summary svelte-ffmenf"><div class="stat-block svelte-ffmenf"><div class="stat-num svelte-ffmenf">${escape_html(data.stats.total_files.toLocaleString())}</div> <div class="stat-label svelte-ffmenf">Files</div></div> <div class="stat-block svelte-ffmenf"><div class="stat-num svelte-ffmenf">${escape_html(data.stats.total_lines.toLocaleString())}</div> <div class="stat-label svelte-ffmenf">Lines</div></div> <div class="stat-block svelte-ffmenf"><div class="stat-num svelte-ffmenf">${escape_html(data.stats.total_code.toLocaleString())}</div> <div class="stat-label svelte-ffmenf">Code</div></div></div> <!--[-->`);
            const each_array_5 = ensure_array_like(topLangs());
            for (let $$index_5 = 0, $$length = each_array_5.length; $$index_5 < $$length; $$index_5++) {
              let [name, l] = each_array_5[$$index_5];
              $$renderer4.push(`<div class="lang-row svelte-ffmenf"><span class="lang-name svelte-ffmenf">${escape_html(name)}</span> <div class="lang-bar svelte-ffmenf"><div class="lang-bar-fill svelte-ffmenf"${attr_style(`width: ${stringify(l.lines / totalLangLines() * 100)}%`)}></div></div> <span class="lang-num svelte-ffmenf">${escape_html(l.lines.toLocaleString())} lines</span> <span class="lang-num svelte-ffmenf">${escape_html(l.files)} files</span></div>`);
            }
            $$renderer4.push(`<!--]-->`);
          }
        });
      } else {
        $$renderer3.push("<!--[-1-->");
      }
      $$renderer3.push(`<!--]-->`);
    }
    function panelPipelines($$renderer3) {
      if (data.remotePipelines.length > 0) {
        $$renderer3.push("<!--[0-->");
        Panel($$renderer3, {
          title: "CI/CD pipelines",
          children: ($$renderer4) => {
            $$renderer4.push(`<!--[-->`);
            const each_array_6 = ensure_array_like(data.remotePipelines);
            for (let $$index_6 = 0, $$length = each_array_6.length; $$index_6 < $$length; $$index_6++) {
              let p = each_array_6[$$index_6];
              const mappedStatus = p.status === "success" ? "succeeded" : p.status === "failure" ? "failed" : p.status === "running" ? "running" : "pending";
              $$renderer4.push(`<div class="pipeline-row svelte-ffmenf">`);
              StatusPill($$renderer4, { status: mappedStatus, size: "sm" });
              $$renderer4.push(`<!----> <div class="pipeline-info svelte-ffmenf"><span class="pipeline-name svelte-ffmenf">${escape_html(p.workflow_name || "pipeline")}</span> `);
              if (p.branch) {
                $$renderer4.push("<!--[0-->");
                Badge($$renderer4, {
                  variant: "info",
                  size: "sm",
                  children: ($$renderer5) => {
                    $$renderer5.push(`<!---->${escape_html(p.branch)}`);
                  }
                });
              } else {
                $$renderer4.push("<!--[-1-->");
              }
              $$renderer4.push(`<!--]--> `);
              if (p.commit_sha) {
                $$renderer4.push("<!--[0-->");
                Tooltip($$renderer4, {
                  text: p.commit_sha,
                  children: ($$renderer5) => {
                    $$renderer5.push(`<span class="pipeline-sha svelte-ffmenf">${escape_html(p.commit_sha.slice(0, 7))}</span>`);
                  }
                });
              } else {
                $$renderer4.push("<!--[-1-->");
              }
              $$renderer4.push(`<!--]--></div> <span class="pipeline-time svelte-ffmenf">${escape_html(p.started_at ? formatDuration(p.started_at, p.finished_at) : "")}</span> `);
              if (p.html_url) {
                $$renderer4.push("<!--[0-->");
                $$renderer4.push(`<a${attr("href", p.html_url)} target="_blank" rel="noopener noreferrer" class="ghost pipeline-view svelte-ffmenf">View</a>`);
              } else {
                $$renderer4.push("<!--[-1-->");
              }
              $$renderer4.push(`<!--]--></div>`);
            }
            $$renderer4.push(`<!--]-->`);
          }
        });
      } else {
        $$renderer3.push("<!--[-1-->");
      }
      $$renderer3.push(`<!--]-->`);
    }
    function panelPipelineConfigs($$renderer3) {
      if (data.pipelineConfigs.length > 0) {
        $$renderer3.push("<!--[0-->");
        Panel($$renderer3, {
          title: "CI/CD configuration",
          children: ($$renderer4) => {
            $$renderer4.push(`<!--[-->`);
            const each_array_7 = ensure_array_like(data.pipelineConfigs);
            for (let $$index_7 = 0, $$length = each_array_7.length; $$index_7 < $$length; $$index_7++) {
              let cfg = each_array_7[$$index_7];
              $$renderer4.push(`<div class="ci-config-row svelte-ffmenf">`);
              Badge($$renderer4, {
                variant: "info",
                size: "sm",
                children: ($$renderer5) => {
                  $$renderer5.push(`<!---->${escape_html(cfg.ci_system)}`);
                }
              });
              $$renderer4.push(`<!----> <span class="ci-path svelte-ffmenf">${escape_html(cfg.file_path)}</span></div>`);
            }
            $$renderer4.push(`<!--]-->`);
          }
        });
      } else {
        $$renderer3.push("<!--[-1-->");
      }
      $$renderer3.push(`<!--]-->`);
    }
    function panelCompose($$renderer3) {
      if (data.compose?.compose_file) {
        $$renderer3.push("<!--[0-->");
        Panel($$renderer3, {
          title: "Docker Compose",
          children: ($$renderer4) => {
            $$renderer4.push(`<div class="row svelte-ffmenf"><span class="label svelte-ffmenf">File</span><span class="value svelte-ffmenf">${escape_html(data.compose.compose_file)}</span></div> <div class="row svelte-ffmenf"><span class="label svelte-ffmenf">Status</span> <span class="value compose-status-value svelte-ffmenf">`);
            StatusPill($$renderer4, {
              status: data.compose.status === "running" ? "ready" : data.compose.status === "starting" ? "cloning" : data.compose.status === "error" ? "failed" : "pending",
              size: "sm",
              label: data.compose.status
            });
            $$renderer4.push(`<!----></span></div> `);
            if (data.compose.services.length > 0) {
              $$renderer4.push("<!--[0-->");
              $$renderer4.push(`<div class="row svelte-ffmenf"><span class="label svelte-ffmenf">Services</span><span class="value svelte-ffmenf">${escape_html(data.compose.services.join(", "))}</span></div>`);
            } else {
              $$renderer4.push("<!--[-1-->");
            }
            $$renderer4.push(`<!--]--> `);
            if (data.compose.status === "stopped" && data.compose.env_vars.length > 0) {
              $$renderer4.push("<!--[0-->");
              $$renderer4.push(`<form class="compose-env-form svelte-ffmenf"><h3 class="svelte-ffmenf">Environment variables</h3> <!--[-->`);
              const each_array_8 = ensure_array_like(data.compose.env_vars);
              for (let i = 0, $$length = each_array_8.length; i < $$length; i++) {
                let envVar = each_array_8[i];
                $$renderer4.push(`<div class="field svelte-ffmenf"><label${attr("for", `env-${i}`)} class="svelte-ffmenf">${escape_html(envVar.key)}</label> <input${attr("id", `env-${i}`)} type="text"${attr("value", envVar.default_value)}${attr("data-env-key", envVar.key)} class="compose-env-input svelte-ffmenf"/></div>`);
              }
              $$renderer4.push(`<!--]--> <button type="submit"${attr("disabled", data.project.status !== "ready", true)} class="svelte-ffmenf">Start</button></form>`);
            } else if (data.compose.status === "stopped") {
              $$renderer4.push("<!--[1-->");
              $$renderer4.push(`<div class="compose-actions svelte-ffmenf"><button type="button"${attr("disabled", data.project.status !== "ready", true)} class="svelte-ffmenf">Start</button></div>`);
            } else if (data.compose.status === "running") {
              $$renderer4.push("<!--[2-->");
              $$renderer4.push(`<div class="compose-actions svelte-ffmenf"><button type="button" class="svelte-ffmenf">Stop</button></div>`);
            } else if (data.compose.status === "starting") {
              $$renderer4.push("<!--[3-->");
              $$renderer4.push(`<div class="compose-actions compose-info svelte-ffmenf">Starting services...</div>`);
            } else if (data.compose.status === "error") {
              $$renderer4.push("<!--[4-->");
              $$renderer4.push(`<div class="compose-actions compose-error svelte-ffmenf">Failed to start. Check logs and retry.</div> <button type="button" class="compose-actions svelte-ffmenf">Retry</button>`);
            } else {
              $$renderer4.push("<!--[-1-->");
            }
            $$renderer4.push(`<!--]-->`);
          }
        });
      } else {
        $$renderer3.push("<!--[-1-->");
      }
      $$renderer3.push(`<!--]-->`);
    }
    function panelAbout($$renderer3) {
      if (data.overview?.readme) {
        $$renderer3.push("<!--[0-->");
        const readme = data.overview.readme;
        {
          let actions = function($$renderer4) {
            $$renderer4.push(`<span style="color: var(--text-dim); font-family: var(--font-mono); font-size: var(--fs-xs);" class="svelte-ffmenf">${escape_html(readme.path)}</span>`);
          };
          Panel($$renderer3, {
            title: "About",
            actions,
            children: ($$renderer4) => {
              $$renderer4.push(`<div class="readme-body svelte-ffmenf">${escape_html(readme.content)}</div> `);
              if (readme.truncated) {
                $$renderer4.push("<!--[0-->");
                $$renderer4.push(`<div class="readme-truncated svelte-ffmenf">Truncated — open the full file in your repo for the rest.</div>`);
              } else {
                $$renderer4.push("<!--[-1-->");
              }
              $$renderer4.push(`<!--]--> <div class="readme-badges svelte-ffmenf">`);
              if (data.overview.license) {
                $$renderer4.push("<!--[0-->");
                Badge($$renderer4, {
                  variant: "info",
                  size: "sm",
                  children: ($$renderer5) => {
                    $$renderer5.push(`<!---->License: ${escape_html(data.overview.license.kind)}`);
                  }
                });
              } else {
                $$renderer4.push("<!--[-1-->");
              }
              $$renderer4.push(`<!--]--> `);
              if (data.overview.has_contributing) {
                $$renderer4.push("<!--[0-->");
                Badge($$renderer4, {
                  variant: "muted",
                  size: "sm",
                  children: ($$renderer5) => {
                    $$renderer5.push(`<!---->CONTRIBUTING ✓`);
                  }
                });
              } else {
                $$renderer4.push("<!--[-1-->");
              }
              $$renderer4.push(`<!--]--> `);
              if (data.overview.has_changelog) {
                $$renderer4.push("<!--[0-->");
                Badge($$renderer4, {
                  variant: "muted",
                  size: "sm",
                  children: ($$renderer5) => {
                    $$renderer5.push(`<!---->CHANGELOG ✓`);
                  }
                });
              } else {
                $$renderer4.push("<!--[-1-->");
              }
              $$renderer4.push(`<!--]--></div>`);
            }
          });
        }
      } else {
        $$renderer3.push("<!--[-1-->");
      }
      $$renderer3.push(`<!--]-->`);
    }
    function panelContributors($$renderer3) {
      if ((data.overview?.contributors.length ?? 0) > 0) {
        $$renderer3.push("<!--[0-->");
        {
          let actions = function($$renderer4) {
            $$renderer4.push(`<span style="color: var(--text-dim); font-family: var(--font-mono); font-size: var(--fs-xs);" class="svelte-ffmenf">top ${escape_html(data.overview?.contributors.length ?? 0)}</span>`);
          };
          Panel($$renderer3, {
            title: "Contributors",
            actions,
            children: ($$renderer4) => {
              $$renderer4.push(`<!--[-->`);
              const each_array_9 = ensure_array_like(data.overview?.contributors ?? []);
              for (let i = 0, $$length = each_array_9.length; i < $$length; i++) {
                let c = each_array_9[i];
                const max = data.overview?.contributors[0]?.commits ?? 1;
                $$renderer4.push(`<div class="contrib-row svelte-ffmenf"><span class="contrib-rank svelte-ffmenf">#${escape_html(i + 1)}</span> <span class="contrib-name svelte-ffmenf">${escape_html(c.name)}</span> <div class="contrib-bar svelte-ffmenf"><div class="contrib-bar-fill svelte-ffmenf"${attr_style(`width: ${stringify(c.commits / max * 100)}%`)}></div></div> <span class="contrib-count svelte-ffmenf">${escape_html(c.commits.toLocaleString())} commit${escape_html(c.commits === 1 ? "" : "s")}</span></div>`);
              }
              $$renderer4.push(`<!--]--> `);
              if (data.overview?.recent_files && data.overview.recent_files.length > 0) {
                $$renderer4.push("<!--[0-->");
                $$renderer4.push(`<div class="recent-files-head svelte-ffmenf">Recent files</div> <ul class="recent-files svelte-ffmenf"><!--[-->`);
                const each_array_10 = ensure_array_like(data.overview.recent_files.slice(0, 6));
                for (let $$index_10 = 0, $$length = each_array_10.length; $$index_10 < $$length; $$index_10++) {
                  let f = each_array_10[$$index_10];
                  $$renderer4.push(`<li><span class="recent-file-path svelte-ffmenf">${escape_html(f.path)}</span></li>`);
                }
                $$renderer4.push(`<!--]--></ul>`);
              } else {
                $$renderer4.push("<!--[-1-->");
              }
              $$renderer4.push(`<!--]-->`);
            }
          });
        }
      } else {
        $$renderer3.push("<!--[-1-->");
      }
      $$renderer3.push(`<!--]-->`);
    }
    function panelBlameTimeline($$renderer3) {
      {
        let actions = function($$renderer4) {
          $$renderer4.push(`<button type="button" class="ghost svelte-ffmenf" style="font-size: 0.75rem;"${attr("disabled", syncingCommits, true)}>${escape_html("Sync commits")}</button>`);
        };
        Panel($$renderer3, {
          title: "Blame timeline",
          actions,
          children: ($$renderer4) => {
            if (data.blameTimeline.length === 0) {
              $$renderer4.push("<!--[0-->");
              $$renderer4.push(`<div class="empty svelte-ffmenf">No commits synced yet. Click "Sync commits" to populate the timeline.</div>`);
            } else {
              $$renderer4.push("<!--[-1-->");
              $$renderer4.push(`<div class="blame-table-wrap svelte-ffmenf"><table class="blame-table svelte-ffmenf"><thead><tr><th class="svelte-ffmenf">SHA</th><th class="svelte-ffmenf">Author</th><th class="svelte-ffmenf">Message</th><th class="svelte-ffmenf">Date</th><th class="num svelte-ffmenf">Files</th><th class="num svelte-ffmenf">+/-</th><th class="num svelte-ffmenf">Runs</th></tr></thead><tbody><!--[-->`);
              const each_array_11 = ensure_array_like(data.blameTimeline);
              for (let $$index_11 = 0, $$length = each_array_11.length; $$index_11 < $$length; $$index_11++) {
                let c = each_array_11[$$index_11];
                $$renderer4.push(`<tr${attr_class(`blame-row ${stringify(blameRowColor(c))}`, "svelte-ffmenf")} role="button" tabindex="0"><td class="mono svelte-ffmenf">${escape_html(c.sha.slice(0, 8))}</td><td class="blame-author svelte-ffmenf">${escape_html(c.author)}</td><td class="blame-msg svelte-ffmenf">${escape_html(c.message.split("\n")[0].slice(0, 72))}</td><td class="blame-date svelte-ffmenf">`);
                TimeAgo($$renderer4, { value: c.committed_at });
                $$renderer4.push(`<!----></td><td class="num svelte-ffmenf">${escape_html(c.files_changed)}</td><td class="num svelte-ffmenf"><span class="ins svelte-ffmenf">+${escape_html(c.insertions)}</span> <span class="del svelte-ffmenf">-${escape_html(c.deletions)}</span></td><td class="num svelte-ffmenf">`);
                if (c.run_count > 0) {
                  $$renderer4.push("<!--[0-->");
                  $$renderer4.push(`<span class="blame-runs svelte-ffmenf">${escape_html(c.run_count)} `);
                  if (c.pass_count > 0) {
                    $$renderer4.push("<!--[0-->");
                    $$renderer4.push(`<span class="pass-count svelte-ffmenf">${escape_html(c.pass_count)}p</span>`);
                  } else {
                    $$renderer4.push("<!--[-1-->");
                  }
                  $$renderer4.push(`<!--]--> `);
                  if (c.fail_count > 0) {
                    $$renderer4.push("<!--[0-->");
                    $$renderer4.push(`<span class="fail-count svelte-ffmenf">${escape_html(c.fail_count)}f</span>`);
                  } else {
                    $$renderer4.push("<!--[-1-->");
                  }
                  $$renderer4.push(`<!--]--></span>`);
                } else {
                  $$renderer4.push("<!--[-1-->");
                  $$renderer4.push(`<span class="dim svelte-ffmenf">--</span>`);
                }
                $$renderer4.push(`<!--]--></td></tr>`);
              }
              $$renderer4.push(`<!--]--></tbody></table></div>`);
            }
            $$renderer4.push(`<!--]-->`);
          }
        });
      }
    }
    function panelMonorepo($$renderer3) {
      {
        let actions = function($$renderer4) {
          $$renderer4.push(`<button type="button" class="ghost svelte-ffmenf" style="font-size: 0.75rem;"${attr("disabled", detectingPackages, true)}>${escape_html("Detect packages")}</button> `);
          if (data.monorepoPackages.length > 0) {
            $$renderer4.push("<!--[0-->");
            $$renderer4.push(`<button type="button" class="ghost svelte-ffmenf" style="font-size: 0.75rem;"${attr("disabled", autoMapping, true)}>${escape_html("Auto-map tasks")}</button>`);
          } else {
            $$renderer4.push("<!--[-1-->");
          }
          $$renderer4.push(`<!--]-->`);
        };
        Panel($$renderer3, {
          title: "Monorepo packages",
          actions,
          children: ($$renderer4) => {
            if (data.monorepoPackages.length === 0) {
              $$renderer4.push("<!--[0-->");
              $$renderer4.push(`<div class="empty svelte-ffmenf">No packages detected. Click "Detect packages" to scan for monorepo packages.</div>`);
            } else {
              $$renderer4.push("<!--[-1-->");
              $$renderer4.push(`<!--[-->`);
              const each_array_12 = ensure_array_like(data.monorepoPackages);
              for (let $$index_12 = 0, $$length = each_array_12.length; $$index_12 < $$length; $$index_12++) {
                let pkg = each_array_12[$$index_12];
                $$renderer4.push(`<div class="pkg-row svelte-ffmenf" role="button" tabindex="0"><div class="pkg-info svelte-ffmenf"><span class="pkg-name svelte-ffmenf">${escape_html(pkg.name)}</span> <span class="pkg-path svelte-ffmenf">${escape_html(pkg.path)}</span></div> `);
                Badge($$renderer4, {
                  variant: pkgTypeBadgeVariant(pkg.pkg_type),
                  size: "sm",
                  children: ($$renderer5) => {
                    $$renderer5.push(`<!---->${escape_html(pkg.pkg_type)}`);
                  }
                });
                $$renderer4.push(`<!----> <span class="pkg-tasks svelte-ffmenf">${escape_html(pkg.task_count)} task${escape_html(pkg.task_count === 1 ? "" : "s")}</span></div>`);
              }
              $$renderer4.push(`<!--]-->`);
            }
            $$renderer4.push(`<!--]-->`);
          }
        });
      }
    }
    function panelPreviews($$renderer3) {
      {
        let actions = function($$renderer4) {
          $$renderer4.push(`<span class="filter-count svelte-ffmenf">${escape_html(data.previews.length)} preview${escape_html(data.previews.length === 1 ? "" : "s")}</span>`);
        };
        Panel($$renderer3, {
          title: "Deploy previews",
          actions,
          children: ($$renderer4) => {
            $$renderer4.push(`<form method="POST" action="?/createPreview" class="preview-create-form svelte-ffmenf"><input type="text" name="branch" placeholder="Branch name" class="preview-branch-input svelte-ffmenf"${attr("value", newPreviewBranch)}/> <button type="submit"${attr("disabled", !newPreviewBranch.trim(), true)} class="svelte-ffmenf">New preview</button></form> `);
            if (data.previews.length === 0) {
              $$renderer4.push("<!--[0-->");
              $$renderer4.push(`<div class="empty svelte-ffmenf" style="padding: var(--space-4) 0;">No deploy previews yet.</div>`);
            } else {
              $$renderer4.push("<!--[-1-->");
              $$renderer4.push(`<!--[-->`);
              const each_array_13 = ensure_array_like(data.previews);
              for (let $$index_13 = 0, $$length = each_array_13.length; $$index_13 < $$length; $$index_13++) {
                let p = each_array_13[$$index_13];
                $$renderer4.push(`<div class="preview-row svelte-ffmenf"><div class="preview-info svelte-ffmenf">`);
                StatusPill($$renderer4, {
                  status: p.status === "deployed" ? "ready" : p.status === "deploying" ? "cloning" : p.status === "stopped" ? "cancelled" : p.status === "failed" ? "failed" : "pending",
                  size: "sm"
                });
                $$renderer4.push(`<!----> `);
                Badge($$renderer4, {
                  variant: "info",
                  size: "sm",
                  children: ($$renderer5) => {
                    $$renderer5.push(`<!---->${escape_html(p.branch)}`);
                  }
                });
                $$renderer4.push(`<!----> `);
                if (p.url) {
                  $$renderer4.push("<!--[0-->");
                  $$renderer4.push(`<a${attr("href", p.url)} target="_blank" rel="noopener noreferrer" class="preview-url svelte-ffmenf">${escape_html(p.url)}</a>`);
                } else {
                  $$renderer4.push("<!--[-1-->");
                }
                $$renderer4.push(`<!--]--></div> <div class="preview-meta svelte-ffmenf">`);
                if (p.last_deployed_at) {
                  $$renderer4.push("<!--[0-->");
                  $$renderer4.push(`<span class="dim svelte-ffmenf">`);
                  TimeAgo($$renderer4, { value: p.last_deployed_at });
                  $$renderer4.push(`<!----></span>`);
                } else {
                  $$renderer4.push("<!--[-1-->");
                }
                $$renderer4.push(`<!--]--> <label class="auto-deploy-toggle svelte-ffmenf" title="Auto-deploy on push"><input type="checkbox"${attr("checked", p.auto_deploy, true)} class="svelte-ffmenf"/> <span class="auto-deploy-label svelte-ffmenf">Auto</span></label></div> <div class="preview-actions svelte-ffmenf"><button type="button" class="ghost svelte-ffmenf" style="font-size: 0.75rem;"${attr("disabled", previewActionPending === p.id, true)}>Redeploy</button> `);
                if (p.status === "deployed" || p.status === "deploying") {
                  $$renderer4.push("<!--[0-->");
                  $$renderer4.push(`<button type="button" class="ghost svelte-ffmenf" style="font-size: 0.75rem;"${attr("disabled", previewActionPending === p.id, true)}>Stop</button>`);
                } else {
                  $$renderer4.push("<!--[-1-->");
                }
                $$renderer4.push(`<!--]--> `);
                if (previewDeleteConfirm === p.id) {
                  $$renderer4.push("<!--[0-->");
                  $$renderer4.push(`<button type="button" class="danger svelte-ffmenf" style="font-size: 0.75rem;"${attr("disabled", previewActionPending === p.id, true)}>Confirm delete</button> <button type="button" class="ghost svelte-ffmenf" style="font-size: 0.75rem;">Cancel</button>`);
                } else {
                  $$renderer4.push("<!--[-1-->");
                  $$renderer4.push(`<button type="button" class="ghost danger svelte-ffmenf" style="font-size: 0.75rem;">Delete</button>`);
                }
                $$renderer4.push(`<!--]--></div></div>`);
              }
              $$renderer4.push(`<!--]-->`);
            }
            $$renderer4.push(`<!--]-->`);
          }
        });
      }
    }
    function panelAlertRules($$renderer3) {
      {
        let actions = function($$renderer4) {
          $$renderer4.push(`<button type="button" class="ghost svelte-ffmenf" style="font-size: 0.75rem;">${escape_html("New rule")}</button>`);
        };
        Panel($$renderer3, {
          title: "Alert rules",
          actions,
          children: ($$renderer4) => {
            {
              $$renderer4.push("<!--[-1-->");
            }
            $$renderer4.push(`<!--]--> `);
            if (data.alertRules.length === 0 && data.widgetAlertRules.length === 0) {
              $$renderer4.push("<!--[0-->");
              $$renderer4.push(`<div class="empty svelte-ffmenf" style="padding: var(--space-4) 0;">No alert rules configured for this project's tasks.</div>`);
            } else {
              $$renderer4.push("<!--[-1-->");
              $$renderer4.push(`<!--[-->`);
              const each_array_15 = ensure_array_like(data.alertRules);
              for (let $$index_15 = 0, $$length = each_array_15.length; $$index_15 < $$length; $$index_15++) {
                let rule = each_array_15[$$index_15];
                $$renderer4.push(`<div class="alert-row svelte-ffmenf"><div class="alert-info svelte-ffmenf"><span class="alert-task-name svelte-ffmenf">${escape_html(rule.task_name)}</span> <span class="alert-desc svelte-ffmenf">${escape_html(rule.type)} ${escape_html(rule.comparison)} ${escape_html(rule.threshold)}${escape_html(rule.type === "failure_rate" ? "%" : rule.type === "duration" ? "ms" : "")}</span></div> <div class="alert-actions svelte-ffmenf">`);
                if (rule.last_triggered_at) {
                  $$renderer4.push("<!--[0-->");
                  $$renderer4.push(`<span class="dim svelte-ffmenf" style="font-size: 0.75rem;">`);
                  TimeAgo($$renderer4, { value: rule.last_triggered_at });
                  $$renderer4.push(`<!----></span>`);
                } else {
                  $$renderer4.push("<!--[-1-->");
                }
                $$renderer4.push(`<!--]--> <form method="POST" action="?/toggleAlertRule" class="inline-form"><input type="hidden" name="rule_id"${attr("value", rule.id)} class="svelte-ffmenf"/> <button type="submit" class="ghost svelte-ffmenf" style="font-size: 0.75rem;"${attr("title", rule.enabled ? "Disable rule" : "Enable rule")}>${escape_html(rule.enabled ? "Enabled" : "Disabled")}</button></form> <form method="POST" action="?/deleteAlertRule" class="inline-form"><input type="hidden" name="rule_id"${attr("value", rule.id)} class="svelte-ffmenf"/> <button type="submit" class="ghost danger svelte-ffmenf" style="font-size: 0.75rem;">Delete</button></form></div></div>`);
              }
              $$renderer4.push(`<!--]-->`);
            }
            $$renderer4.push(`<!--]-->`);
          }
        });
      }
    }
    function panelImpactRadar($$renderer3) {
      if (data.impactRadar && data.impactRadar.tasks.length > 0) {
        $$renderer3.push("<!--[0-->");
        {
          let actions = function($$renderer4) {
            if (data.impactRadar && data.impactRadar.unmapped_tasks > 0) {
              $$renderer4.push("<!--[0-->");
              $$renderer4.push(`<span class="dim svelte-ffmenf" style="font-size: 0.75rem;">${escape_html(data.impactRadar.unmapped_tasks)} unmapped</span>`);
            } else {
              $$renderer4.push("<!--[-1-->");
            }
            $$renderer4.push(`<!--]-->`);
          };
          Panel($$renderer3, {
            title: "Impact radar",
            actions,
            children: ($$renderer4) => {
              $$renderer4.push(`<!--[-->`);
              const each_array_16 = ensure_array_like(data.impactRadar.tasks);
              for (let $$index_17 = 0, $$length = each_array_16.length; $$index_17 < $$length; $$index_17++) {
                let t = each_array_16[$$index_17];
                $$renderer4.push(`<div class="impact-row svelte-ffmenf">`);
                StatusPill($$renderer4, {
                  status: t.last_status === "succeeded" ? "succeeded" : t.last_status === "failed" ? "failed" : "pending",
                  size: "sm"
                });
                $$renderer4.push(`<!----> <span class="impact-name svelte-ffmenf">${escape_html(t.task_name)}</span> <div class="impact-patterns svelte-ffmenf"><!--[-->`);
                const each_array_17 = ensure_array_like(t.patterns.slice(0, 3));
                for (let $$index_16 = 0, $$length2 = each_array_17.length; $$index_16 < $$length2; $$index_16++) {
                  let p = each_array_17[$$index_16];
                  Badge($$renderer4, {
                    variant: "muted",
                    size: "sm",
                    children: ($$renderer5) => {
                      $$renderer5.push(`<!---->${escape_html(p)}`);
                    }
                  });
                }
                $$renderer4.push(`<!--]--> `);
                if (t.patterns.length > 3) {
                  $$renderer4.push("<!--[0-->");
                  $$renderer4.push(`<span class="dim svelte-ffmenf" style="font-size: 0.6875rem;">+${escape_html(t.patterns.length - 3)}</span>`);
                } else {
                  $$renderer4.push("<!--[-1-->");
                }
                $$renderer4.push(`<!--]--></div></div>`);
              }
              $$renderer4.push(`<!--]-->`);
            }
          });
        }
      } else {
        $$renderer3.push("<!--[-1-->");
      }
      $$renderer3.push(`<!--]-->`);
    }
    function panelDependencyTree($$renderer3) {
      if (data.dependencyTree && (data.dependencyTree.upstream.length > 0 || data.dependencyTree.downstream.length > 0)) {
        $$renderer3.push("<!--[0-->");
        Panel($$renderer3, {
          title: "Dependency tree",
          children: ($$renderer4) => {
            if (data.dependencyTree.upstream.length > 0) {
              $$renderer4.push("<!--[0-->");
              $$renderer4.push(`<div class="dep-section svelte-ffmenf"><span class="dep-heading svelte-ffmenf">Upstream (${escape_html(data.dependencyTree.upstream.length)})</span> <!--[-->`);
              const each_array_18 = ensure_array_like(data.dependencyTree.upstream);
              for (let $$index_18 = 0, $$length = each_array_18.length; $$index_18 < $$length; $$index_18++) {
                let d = each_array_18[$$index_18];
                $$renderer4.push(`<a${attr("href", `/projects/${d.project_id}`)} class="dep-row svelte-ffmenf"><span class="dep-name svelte-ffmenf">${escape_html(d.project_name)}</span> `);
                Badge($$renderer4, {
                  variant: "muted",
                  size: "sm",
                  children: ($$renderer5) => {
                    $$renderer5.push(`<!---->${escape_html(d.dep_type)}`);
                  }
                });
                $$renderer4.push(`<!----></a>`);
              }
              $$renderer4.push(`<!--]--></div>`);
            } else {
              $$renderer4.push("<!--[-1-->");
            }
            $$renderer4.push(`<!--]--> `);
            if (data.dependencyTree.downstream.length > 0) {
              $$renderer4.push("<!--[0-->");
              $$renderer4.push(`<div class="dep-section svelte-ffmenf"><span class="dep-heading svelte-ffmenf">Downstream (${escape_html(data.dependencyTree.downstream.length)})</span> <!--[-->`);
              const each_array_19 = ensure_array_like(data.dependencyTree.downstream);
              for (let $$index_19 = 0, $$length = each_array_19.length; $$index_19 < $$length; $$index_19++) {
                let d = each_array_19[$$index_19];
                $$renderer4.push(`<a${attr("href", `/projects/${d.project_id}`)} class="dep-row svelte-ffmenf"><span class="dep-name svelte-ffmenf">${escape_html(d.project_name)}</span> `);
                Badge($$renderer4, {
                  variant: "muted",
                  size: "sm",
                  children: ($$renderer5) => {
                    $$renderer5.push(`<!---->${escape_html(d.dep_type)}`);
                  }
                });
                $$renderer4.push(`<!----></a>`);
              }
              $$renderer4.push(`<!--]--></div>`);
            } else {
              $$renderer4.push("<!--[-1-->");
            }
            $$renderer4.push(`<!--]-->`);
          }
        });
      } else {
        $$renderer3.push("<!--[-1-->");
      }
      $$renderer3.push(`<!--]-->`);
    }
    function panelLiveRun($$renderer3) {
      if (data.liveRun) {
        $$renderer3.push("<!--[0-->");
        Panel($$renderer3, {
          title: "Live run",
          children: ($$renderer4) => {
            $$renderer4.push(`<div class="live-run-header svelte-ffmenf">`);
            StatusPill($$renderer4, {
              status: data.liveRun.run.status === "running" ? "running" : "queued",
              size: "sm"
            });
            $$renderer4.push(`<!----> <a${attr("href", `/runs/${data.liveRun.run.id}`)} class="live-run-name svelte-ffmenf">${escape_html(data.liveRun.run.task_name)}</a> <span class="dim svelte-ffmenf" style="font-size: 0.75rem;">`);
            TimeAgo($$renderer4, { value: data.liveRun.run.started_at });
            $$renderer4.push(`<!----></span></div> `);
            if (data.liveRun.log_tail) {
              $$renderer4.push("<!--[0-->");
              $$renderer4.push(`<pre class="live-log-tail svelte-ffmenf">${escape_html(data.liveRun.log_tail)}</pre>`);
            } else {
              $$renderer4.push("<!--[-1-->");
            }
            $$renderer4.push(`<!--]-->`);
          }
        });
      } else {
        $$renderer3.push("<!--[-1-->");
      }
      $$renderer3.push(`<!--]-->`);
    }
    function renderPanel($$renderer3, key) {
      if (key === "tasks") {
        $$renderer3.push("<!--[0-->");
        panelTasks($$renderer3);
      } else if (key === "recent-runs") {
        $$renderer3.push("<!--[1-->");
        panelRecentRuns($$renderer3);
      } else if (key === "repository") {
        $$renderer3.push("<!--[2-->");
        panelRepository($$renderer3);
      } else if (key === "latest-commit") {
        $$renderer3.push("<!--[3-->");
        panelLatestCommit($$renderer3);
      } else if (key === "code-stats") {
        $$renderer3.push("<!--[4-->");
        panelCodeStats($$renderer3);
      } else if (key === "about") {
        $$renderer3.push("<!--[5-->");
        panelAbout($$renderer3);
      } else if (key === "contributors") {
        $$renderer3.push("<!--[6-->");
        panelContributors($$renderer3);
      } else if (key === "pipelines") {
        $$renderer3.push("<!--[7-->");
        panelPipelines($$renderer3);
      } else if (key === "pipeline-configs") {
        $$renderer3.push("<!--[8-->");
        panelPipelineConfigs($$renderer3);
      } else if (key === "compose") {
        $$renderer3.push("<!--[9-->");
        panelCompose($$renderer3);
      } else if (key === "blame-timeline") {
        $$renderer3.push("<!--[10-->");
        panelBlameTimeline($$renderer3);
      } else if (key === "monorepo") {
        $$renderer3.push("<!--[11-->");
        panelMonorepo($$renderer3);
      } else if (key === "previews") {
        $$renderer3.push("<!--[12-->");
        panelPreviews($$renderer3);
      } else if (key === "alert-rules") {
        $$renderer3.push("<!--[13-->");
        panelAlertRules($$renderer3);
      } else if (key === "impact-radar") {
        $$renderer3.push("<!--[14-->");
        panelImpactRadar($$renderer3);
      } else if (key === "dependency-tree") {
        $$renderer3.push("<!--[15-->");
        panelDependencyTree($$renderer3);
      } else if (key === "live-run") {
        $$renderer3.push("<!--[16-->");
        panelLiveRun($$renderer3);
      } else {
        $$renderer3.push("<!--[-1-->");
      }
      $$renderer3.push(`<!--]-->`);
    }
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
    $$renderer2.push(`<!--]--> <div class="favorite-row svelte-ffmenf"><form method="POST" action="?/toggleFavorite" class="inline-form"><input type="hidden" name="is_favorited"${attr("value", String(data.isFavorited))}/> <button type="submit"${attr_class(`favorite-btn ${stringify(data.isFavorited ? "favorited" : "")}`, "svelte-ffmenf")}${attr("title", data.isFavorited ? "Remove from favorites" : "Add to favorites")}>${escape_html(data.isFavorited ? "★" : "☆")} <span class="favorite-label svelte-ffmenf">${escape_html(data.isFavorited ? "Favorited" : "Favorite")}</span></button></form></div> <div class="layout-toolbar svelte-ffmenf"><span class="spacer svelte-ffmenf"></span> `);
    {
      $$renderer2.push("<!--[-1-->");
      $$renderer2.push(`<button type="button" class="ghost svelte-ffmenf">Customize layout</button>`);
    }
    $$renderer2.push(`<!--]--></div>          <div${attr_class("panel-stack svelte-ffmenf", void 0, { "editing": editLayoutMode })}>`);
    {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--> <!--[-->`);
    const each_array_20 = ensure_array_like(panelLayout);
    for (let rowIdx = 0, $$length = each_array_20.length; rowIdx < $$length; rowIdx++) {
      let row = each_array_20[rowIdx];
      const visible = visibleRow(row);
      if (visible.length > 0) {
        $$renderer2.push("<!--[0-->");
        $$renderer2.push(`<div class="panel-row svelte-ffmenf"${attr_style(`grid-template-columns: repeat(${stringify(visible.length)}, minmax(0, 1fr));`)}><!--[-->`);
        const each_array_21 = ensure_array_like(row);
        for (let colIdx = 0, $$length2 = each_array_21.length; colIdx < $$length2; colIdx++) {
          let key = each_array_21[colIdx];
          if (panelHasData(key)) {
            $$renderer2.push("<!--[0-->");
            $$renderer2.push(`<div${attr_class("panel-shell svelte-ffmenf", void 0, {
              "editing": editLayoutMode,
              "dragging": dragSource,
              "drop-left": dropTarget,
              "drop-right": dropTarget
            })}${attr("role", void 0)}${attr("aria-label", void 0)}${attr("draggable", editLayoutMode)}>`);
            {
              $$renderer2.push("<!--[-1-->");
            }
            $$renderer2.push(`<!--]--> `);
            {
              $$renderer2.push("<!--[-1-->");
              renderPanel($$renderer2, key);
            }
            $$renderer2.push(`<!--]--></div>`);
          } else {
            $$renderer2.push("<!--[-1-->");
          }
          $$renderer2.push(`<!--]-->`);
        }
        $$renderer2.push(`<!--]--></div>`);
      } else {
        $$renderer2.push("<!--[-1-->");
      }
      $$renderer2.push(`<!--]--> `);
      {
        $$renderer2.push("<!--[-1-->");
      }
      $$renderer2.push(`<!--]-->`);
    }
    $$renderer2.push(`<!--]--></div> `);
    Modal($$renderer2, {
      open: paramsModal !== null,
      title: paramsModal ? `Run with parameters: ${paramsModal.taskName}` : "",
      width: 560,
      onClose: () => paramsModal = null,
      children: ($$renderer3) => {
        if (paramsModal) {
          $$renderer3.push("<!--[0-->");
          $$renderer3.push(`<div class="field"><label for="env-input">Env vars (one per line, KEY=VALUE)</label> <textarea id="env-input" rows="4" placeholder="NODE_ENV=production
LOG_LEVEL=debug">`);
          const $$body = escape_html(paramsModal.env);
          if ($$body) {
            $$renderer3.push(`${$$body}`);
          }
          $$renderer3.push(`</textarea></div> <div class="field"><label for="args-input">Extra args (one per line)</label> <textarea id="args-input" rows="4" placeholder="--watch
--reporter=verbose">`);
          const $$body_1 = escape_html(paramsModal.args);
          if ($$body_1) {
            $$renderer3.push(`${$$body_1}`);
          }
          $$renderer3.push(`</textarea></div> `);
          {
            $$renderer3.push("<!--[-1-->");
          }
          $$renderer3.push(`<!--]--> <div class="modal-actions svelte-ffmenf"><button type="button" class="ghost">Cancel</button> <button type="button">Run</button></div>`);
        } else {
          $$renderer3.push("<!--[-1-->");
        }
        $$renderer3.push(`<!--]-->`);
      }
    });
    $$renderer2.push(`<!----> `);
    Modal($$renderer2, {
      open: timeoutModal !== null,
      title: timeoutModal ? `Timeout: ${timeoutModal.taskName}` : "",
      width: 420,
      onClose: () => timeoutModal = null,
      children: ($$renderer3) => {
        if (timeoutModal) {
          $$renderer3.push("<!--[0-->");
          $$renderer3.push(`<div class="field"><label for="timeout-input">Timeout in seconds</label> <input id="timeout-input" type="number" min="1" max="86400" placeholder="e.g. 1800 = 30m, 7200 = 2h"${attr("value", timeoutModal.value)}/> <p class="modal-hint svelte-ffmenf">Leave blank for the server default. 1–86400 seconds (24h).</p></div> `);
          if (timeoutModal.error) {
            $$renderer3.push("<!--[0-->");
            $$renderer3.push(`<p class="modal-error svelte-ffmenf">${escape_html(timeoutModal.error)}</p>`);
          } else {
            $$renderer3.push("<!--[-1-->");
          }
          $$renderer3.push(`<!--]--> <div class="modal-actions svelte-ffmenf"><button type="button" class="ghost">Cancel</button> <button type="button">Save</button></div>`);
        } else {
          $$renderer3.push("<!--[-1-->");
        }
        $$renderer3.push(`<!--]-->`);
      }
    });
    $$renderer2.push(`<!----> `);
    Modal($$renderer2, {
      open: retryModal !== null,
      title: retryModal ? `Retry policy: ${retryModal.taskName}` : "",
      width: 460,
      onClose: () => retryModal = null,
      children: ($$renderer3) => {
        if (retryModal) {
          $$renderer3.push("<!--[0-->");
          $$renderer3.push(`<div class="modal-fields-grid svelte-ffmenf"><div class="field"><label for="retry-max-input">Max retries</label> <input id="retry-max-input" type="number" min="0" max="10"${attr("value", retryModal.max)}/> <p class="modal-hint svelte-ffmenf">0 = no retry, up to 10.</p></div> <div class="field"><label for="retry-backoff-input">Backoff (seconds)</label> <input id="retry-backoff-input" type="number" min="1" max="3600"${attr("disabled", parseInt(retryModal.max, 10) <= 0, true)}${attr("value", retryModal.backoff)}/> <p class="modal-hint svelte-ffmenf">Wait between retries.</p></div></div> `);
          if (retryModal.error) {
            $$renderer3.push("<!--[0-->");
            $$renderer3.push(`<p class="modal-error svelte-ffmenf">${escape_html(retryModal.error)}</p>`);
          } else {
            $$renderer3.push("<!--[-1-->");
          }
          $$renderer3.push(`<!--]--> <div class="modal-actions svelte-ffmenf"><button type="button" class="ghost">Cancel</button> <button type="button">Save</button></div>`);
        } else {
          $$renderer3.push("<!--[-1-->");
        }
        $$renderer3.push(`<!--]-->`);
      }
    });
    $$renderer2.push(`<!----> `);
    Modal($$renderer2, {
      open: artifactsModal !== null,
      title: artifactsModal ? `Artifact patterns: ${artifactsModal.taskName}` : "",
      width: 560,
      onClose: () => artifactsModal = null,
      children: ($$renderer3) => {
        if (artifactsModal) {
          $$renderer3.push("<!--[0-->");
          $$renderer3.push(`<p style="color: var(--text-dim); font-size: var(--fs-sm); margin: 0 0 var(--space-3);">One glob pattern per line. After each run finishes, matching files are captured and listed on the run detail page.</p> <div class="field"><label for="artifacts-input">Patterns</label> <textarea id="artifacts-input" rows="6" placeholder="**/junit.xml
coverage/*.xml
build/output.tar.gz">`);
          const $$body_2 = escape_html(artifactsModal.raw);
          if ($$body_2) {
            $$renderer3.push(`${$$body_2}`);
          }
          $$renderer3.push(`</textarea> <p class="modal-hint svelte-ffmenf">Supported: literal paths, basename globs (<code>*.xml</code>), and <code>**/</code> for any-depth.</p></div> `);
          if (artifactsModal.error) {
            $$renderer3.push("<!--[0-->");
            $$renderer3.push(`<p class="modal-error svelte-ffmenf">${escape_html(artifactsModal.error)}</p>`);
          } else {
            $$renderer3.push("<!--[-1-->");
          }
          $$renderer3.push(`<!--]--> <div class="modal-actions svelte-ffmenf"><button type="button" class="ghost">Cancel</button> <button type="button">Save</button></div>`);
        } else {
          $$renderer3.push("<!--[-1-->");
        }
        $$renderer3.push(`<!--]-->`);
      }
    });
    $$renderer2.push(`<!----> `);
    Modal($$renderer2, {
      open: servicesModal !== null,
      title: servicesModal ? `Required services: ${servicesModal.taskName}` : "",
      width: 520,
      onClose: () => servicesModal = null,
      children: ($$renderer3) => {
        if (servicesModal) {
          $$renderer3.push("<!--[0-->");
          $$renderer3.push(`<p style="color: var(--text-dim); font-size: var(--fs-sm); margin: 0 0 var(--space-3);">Comma-separated compose service names. Workend will refuse to run this task unless the project's compose stack is up.</p> <div class="field"><label for="services-input">Services</label> <input id="services-input" type="text" placeholder="postgres, redis"${attr("value", servicesModal.raw)}/> <p class="modal-hint svelte-ffmenf">Match service keys from your <code>compose.yaml</code>. Leave blank to remove the requirement.</p></div> `);
          if (servicesModal.error) {
            $$renderer3.push("<!--[0-->");
            $$renderer3.push(`<p class="modal-error svelte-ffmenf">${escape_html(servicesModal.error)}</p>`);
          } else {
            $$renderer3.push("<!--[-1-->");
          }
          $$renderer3.push(`<!--]--> <div class="modal-actions svelte-ffmenf"><button type="button" class="ghost">Cancel</button> <button type="button">Save</button></div>`);
        } else {
          $$renderer3.push("<!--[-1-->");
        }
        $$renderer3.push(`<!--]-->`);
      }
    });
    $$renderer2.push(`<!----> `);
    Modal($$renderer2, {
      open: branchModal !== null,
      title: branchModal ? `Run on branch: ${branchModal.taskName}` : "",
      width: 480,
      onClose: () => branchModal = null,
      children: ($$renderer3) => {
        if (branchModal) {
          $$renderer3.push("<!--[0-->");
          $$renderer3.push(`<p style="color: var(--text-dim); font-size: var(--fs-sm); margin: 0 0 var(--space-3);">Workend will fetch this branch's HEAD into an ephemeral checkout — your project's main checkout is not switched.</p> `);
          if (branchModal.loading) {
            $$renderer3.push("<!--[0-->");
            $$renderer3.push(`<div style="color: var(--text-dim); font-style: italic; padding: var(--space-3) 0;">Loading branches…</div>`);
          } else if (branchModal.error) {
            $$renderer3.push("<!--[1-->");
            $$renderer3.push(`<p class="modal-error svelte-ffmenf">${escape_html(branchModal.error)}</p>`);
          } else if (branchModal.branches && branchModal.branches.length > 0) {
            $$renderer3.push("<!--[2-->");
            $$renderer3.push(`<form method="POST" action="?/runOnBranch" class="inline-form"><input type="hidden" name="task_id"${attr("value", branchModal.taskID)}/> <div class="field"><label for="branch-pick">Branch</label> `);
            $$renderer3.select(
              {
                id: "branch-pick",
                name: "branch",
                value: branchModal.selected,
                required: true
              },
              ($$renderer4) => {
                $$renderer4.push(`<!--[-->`);
                const each_array_22 = ensure_array_like(branchModal.branches);
                for (let $$index_22 = 0, $$length = each_array_22.length; $$index_22 < $$length; $$index_22++) {
                  let b = each_array_22[$$index_22];
                  $$renderer4.option({ value: b.name }, ($$renderer5) => {
                    $$renderer5.push(`${escape_html(b.name)}`);
                  });
                }
                $$renderer4.push(`<!--]-->`);
              }
            );
            $$renderer3.push(` <p class="modal-hint svelte-ffmenf">${escape_html(branchModal.branches.length)} branches available.</p></div> <div class="modal-actions svelte-ffmenf"><button type="button" class="ghost">Cancel</button> <button type="submit">Run</button></div></form>`);
          } else {
            $$renderer3.push("<!--[-1-->");
            $$renderer3.push(`<p style="color: var(--text-dim);">No branches available.</p>`);
          }
          $$renderer3.push(`<!--]-->`);
        } else {
          $$renderer3.push("<!--[-1-->");
        }
        $$renderer3.push(`<!--]-->`);
      }
    });
    $$renderer2.push(`<!----> `);
    Modal($$renderer2, {
      open: concurrencyModal !== null,
      title: concurrencyModal ? `Concurrency: ${concurrencyModal.taskName}` : "",
      width: 520,
      onClose: () => concurrencyModal = null,
      children: ($$renderer3) => {
        if (concurrencyModal) {
          $$renderer3.push("<!--[0-->");
          $$renderer3.push(`<div class="modal-fields-grid svelte-ffmenf"><div class="field"><label for="conc-max">Max concurrent runs</label> <input id="conc-max" type="number" min="0" max="100"${attr("value", concurrencyModal.max)}/> <p class="modal-hint svelte-ffmenf">0 = use the legacy single-run guard. 1+ allows that many in flight.</p></div> <div class="field"><label for="conc-policy">When at limit</label> `);
          $$renderer3.select(
            {
              id: "conc-policy",
              value: concurrencyModal.policy,
              disabled: parseInt(concurrencyModal.max, 10) <= 0
            },
            ($$renderer4) => {
              $$renderer4.option({ value: "queue" }, ($$renderer5) => {
                $$renderer5.push(`queue (reject new for now)`);
              });
              $$renderer4.option({ value: "cancel-old" }, ($$renderer5) => {
                $$renderer5.push(`cancel-old (supersede the oldest)`);
              });
              $$renderer4.option({ value: "reject" }, ($$renderer5) => {
                $$renderer5.push(`reject (return 429)`);
              });
            }
          );
          $$renderer3.push(` <p class="modal-hint svelte-ffmenf">cancel-old is the right choice for fast-pushing branches and \`deploy\` tasks.</p></div></div> `);
          if (concurrencyModal.error) {
            $$renderer3.push("<!--[0-->");
            $$renderer3.push(`<p class="modal-error svelte-ffmenf">${escape_html(concurrencyModal.error)}</p>`);
          } else {
            $$renderer3.push("<!--[-1-->");
          }
          $$renderer3.push(`<!--]--> <div class="modal-actions svelte-ffmenf"><button type="button" class="ghost">Cancel</button> <button type="button">Save</button></div>`);
        } else {
          $$renderer3.push("<!--[-1-->");
        }
        $$renderer3.push(`<!--]-->`);
      }
    });
    $$renderer2.push(`<!----> `);
    Modal($$renderer2, {
      open: blameDetailModal !== null,
      title: blameDetailModal ? `Commit ${blameDetailModal.sha.slice(0, 8)}` : "",
      width: 640,
      onClose: () => blameDetailModal = null,
      children: ($$renderer3) => {
        if (blameDetailModal) {
          $$renderer3.push("<!--[0-->");
          if (blameDetailModal.loading) {
            $$renderer3.push("<!--[0-->");
            $$renderer3.push(`<div style="color: var(--text-dim); font-style: italic; padding: var(--space-3) 0;">Loading commit details...</div>`);
          } else if (blameDetailModal.error) {
            $$renderer3.push("<!--[1-->");
            $$renderer3.push(`<p class="modal-error svelte-ffmenf">${escape_html(blameDetailModal.error)}</p>`);
          } else if (blameDetailModal.commit) {
            $$renderer3.push("<!--[2-->");
            $$renderer3.push(`<div class="row svelte-ffmenf"><span class="label svelte-ffmenf">SHA</span><span class="value svelte-ffmenf">${escape_html(blameDetailModal.commit.sha)}</span></div> <div class="row svelte-ffmenf"><span class="label svelte-ffmenf">Author</span><span class="value commit-msg svelte-ffmenf">${escape_html(blameDetailModal.commit.author)} &lt;${escape_html(blameDetailModal.commit.author_email)}></span></div> <div class="row svelte-ffmenf"><span class="label svelte-ffmenf">Date</span><span class="value svelte-ffmenf">${escape_html(new Date(blameDetailModal.commit.committed_at).toLocaleString())}</span></div> <div class="row svelte-ffmenf"><span class="label svelte-ffmenf">Message</span><span class="value commit-msg svelte-ffmenf">${escape_html(blameDetailModal.commit.message)}</span></div> <div class="row svelte-ffmenf"><span class="label svelte-ffmenf">Changes</span><span class="value svelte-ffmenf">${escape_html(blameDetailModal.commit.files_changed)} files, <span class="ins svelte-ffmenf">+${escape_html(blameDetailModal.commit.insertions)}</span> <span class="del svelte-ffmenf">-${escape_html(blameDetailModal.commit.deletions)}</span></span></div> `);
            if (blameDetailModal.runs.length > 0) {
              $$renderer3.push("<!--[0-->");
              $$renderer3.push(`<h3 style="margin: var(--space-4) 0 var(--space-2) 0; font-size: 0.875rem; color: var(--text-muted);">Associated runs (${escape_html(blameDetailModal.runs.length)})</h3> <!--[-->`);
              const each_array_23 = ensure_array_like(blameDetailModal.runs);
              for (let $$index_23 = 0, $$length = each_array_23.length; $$index_23 < $$length; $$index_23++) {
                let r = each_array_23[$$index_23];
                $$renderer3.push(`<a${attr("href", `/runs/${r.id}`)} class="run-row svelte-ffmenf" style="padding: var(--space-1) 0;">`);
                StatusPill($$renderer3, { status: r.status, size: "sm" });
                $$renderer3.push(`<!----> <span class="run-meta svelte-ffmenf">${escape_html(formatDuration(r.started_at, r.finished_at))}</span> <span class="run-meta svelte-ffmenf">${escape_html(r.started_at ? new Date(r.started_at).toLocaleString() : "--")}</span></a>`);
              }
              $$renderer3.push(`<!--]-->`);
            } else {
              $$renderer3.push("<!--[-1-->");
              $$renderer3.push(`<p style="color: var(--text-dim); font-size: 0.875rem; margin-top: var(--space-3);">No runs associated with this commit.</p>`);
            }
            $$renderer3.push(`<!--]-->`);
          } else {
            $$renderer3.push("<!--[-1-->");
          }
          $$renderer3.push(`<!--]--> <div class="modal-actions svelte-ffmenf" style="margin-top: var(--space-4);"><button type="button" class="ghost">Close</button></div>`);
        } else {
          $$renderer3.push("<!--[-1-->");
        }
        $$renderer3.push(`<!--]-->`);
      }
    });
    $$renderer2.push(`<!----> `);
    Modal($$renderer2, {
      open: packageDetailModal !== null,
      title: packageDetailModal?.pkg ? packageDetailModal.pkg.name : "Package detail",
      width: 560,
      onClose: () => packageDetailModal = null,
      children: ($$renderer3) => {
        if (packageDetailModal) {
          $$renderer3.push("<!--[0-->");
          if (packageDetailModal.loading) {
            $$renderer3.push("<!--[0-->");
            $$renderer3.push(`<div style="color: var(--text-dim); font-style: italic; padding: var(--space-3) 0;">Loading package...</div>`);
          } else if (packageDetailModal.error) {
            $$renderer3.push("<!--[1-->");
            $$renderer3.push(`<p class="modal-error svelte-ffmenf">${escape_html(packageDetailModal.error)}</p>`);
          } else if (packageDetailModal.pkg) {
            $$renderer3.push("<!--[2-->");
            $$renderer3.push(`<div class="row svelte-ffmenf"><span class="label svelte-ffmenf">Path</span><span class="value svelte-ffmenf">${escape_html(packageDetailModal.pkg.path)}</span></div> <div class="row svelte-ffmenf"><span class="label svelte-ffmenf">Type</span><span class="value svelte-ffmenf">`);
            Badge($$renderer3, {
              variant: pkgTypeBadgeVariant(packageDetailModal.pkg.pkg_type),
              size: "sm",
              children: ($$renderer4) => {
                $$renderer4.push(`<!---->${escape_html(packageDetailModal.pkg.pkg_type)}`);
              }
            });
            $$renderer3.push(`<!----></span></div> <h3 style="margin: var(--space-4) 0 var(--space-2) 0; font-size: 0.875rem; color: var(--text-muted);">Scoped tasks (${escape_html(packageDetailModal.scopes.length)})</h3> `);
            if (packageDetailModal.scopes.length > 0) {
              $$renderer3.push("<!--[0-->");
              $$renderer3.push(`<!--[-->`);
              const each_array_24 = ensure_array_like(packageDetailModal.scopes);
              for (let $$index_24 = 0, $$length = each_array_24.length; $$index_24 < $$length; $$index_24++) {
                let scope = each_array_24[$$index_24];
                $$renderer3.push(`<div class="scope-row svelte-ffmenf"><span class="scope-name svelte-ffmenf">${escape_html(scope.task_name)}</span> <button type="button" class="ghost danger" style="font-size: 0.75rem;">Remove</button></div>`);
              }
              $$renderer3.push(`<!--]-->`);
            } else {
              $$renderer3.push("<!--[-1-->");
              $$renderer3.push(`<p style="color: var(--text-dim); font-size: 0.875rem;">No tasks scoped to this package.</p>`);
            }
            $$renderer3.push(`<!--]--> `);
            if (data.tasks.length > 0) {
              $$renderer3.push("<!--[0-->");
              const scopedIDs = new Set(packageDetailModal.scopes.map((s) => s.task_id));
              const availableTasks = data.tasks.filter((t) => !scopedIDs.has(t.id));
              if (availableTasks.length > 0) {
                $$renderer3.push("<!--[0-->");
                $$renderer3.push(`<div style="margin-top: var(--space-3); border-top: 1px solid var(--border); padding-top: var(--space-3);"><span style="font-size: 0.8125rem; color: var(--text-muted);">Add task scope:</span> <div class="scope-add-list svelte-ffmenf"><!--[-->`);
                const each_array_25 = ensure_array_like(availableTasks.slice(0, 10));
                for (let $$index_25 = 0, $$length = each_array_25.length; $$index_25 < $$length; $$index_25++) {
                  let t = each_array_25[$$index_25];
                  $$renderer3.push(`<button type="button" class="ghost" style="font-size: 0.75rem;">+ ${escape_html(t.name)}</button>`);
                }
                $$renderer3.push(`<!--]--></div></div>`);
              } else {
                $$renderer3.push("<!--[-1-->");
              }
              $$renderer3.push(`<!--]-->`);
            } else {
              $$renderer3.push("<!--[-1-->");
            }
            $$renderer3.push(`<!--]-->`);
          } else {
            $$renderer3.push("<!--[-1-->");
          }
          $$renderer3.push(`<!--]--> <div class="modal-actions svelte-ffmenf" style="margin-top: var(--space-4);"><button type="button" class="ghost">Close</button></div>`);
        } else {
          $$renderer3.push("<!--[-1-->");
        }
        $$renderer3.push(`<!--]-->`);
      }
    });
    $$renderer2.push(`<!----> `);
    Modal($$renderer2, {
      open: metricsModal !== null,
      title: metricsModal ? `Metrics: ${metricsModal.taskName}` : "",
      width: 640,
      onClose: () => metricsModal = null,
      children: ($$renderer3) => {
        if (metricsModal) {
          $$renderer3.push("<!--[0-->");
          if (metricsModal.loading) {
            $$renderer3.push("<!--[0-->");
            $$renderer3.push(`<div style="color: var(--text-dim); font-style: italic; padding: var(--space-3) 0;">Loading metrics...</div>`);
          } else if (metricsModal.error) {
            $$renderer3.push("<!--[1-->");
            $$renderer3.push(`<p class="modal-error svelte-ffmenf">${escape_html(metricsModal.error)}</p>`);
          } else if (metricsModal.metrics) {
            $$renderer3.push("<!--[2-->");
            $$renderer3.push(`<div class="metrics-grid svelte-ffmenf"><div class="metric-card svelte-ffmenf"><div class="metric-num svelte-ffmenf">${escape_html(formatMs(metricsModal.metrics.p50_ms))}</div> <div class="metric-label svelte-ffmenf">p50</div></div> <div class="metric-card svelte-ffmenf"><div class="metric-num svelte-ffmenf">${escape_html(formatMs(metricsModal.metrics.p95_ms))}</div> <div class="metric-label svelte-ffmenf">p95</div></div> <div class="metric-card svelte-ffmenf"><div class="metric-num svelte-ffmenf">${escape_html(formatMs(metricsModal.metrics.p99_ms))}</div> <div class="metric-label svelte-ffmenf">p99</div></div> <div class="metric-card svelte-ffmenf"><div class="metric-num svelte-ffmenf">${escape_html(metricsModal.metrics.total_runs.toLocaleString())}</div> <div class="metric-label svelte-ffmenf">Total runs</div></div></div> <div class="success-rate-bar svelte-ffmenf" style="margin-top: var(--space-4);"><div class="success-rate-label svelte-ffmenf"><span>Success rate</span> <span class="success-rate-pct svelte-ffmenf">${escape_html((metricsModal.metrics.success_rate * 100).toFixed(1))}%</span></div> <div class="rate-track svelte-ffmenf"><div class="rate-fill svelte-ffmenf"${attr_style(`width: ${stringify(metricsModal.metrics.success_rate * 100)}%;`)}></div></div> <span style="font-size: 0.75rem; color: var(--text-dim);">${escape_html(metricsModal.metrics.last_30d_runs)} runs in last 30 days</span></div> `);
            if (metricsModal.trends.length > 0) {
              $$renderer3.push("<!--[0-->");
              const trendMax = Math.max(...metricsModal.trends.map((t) => t.count), 1);
              $$renderer3.push(`<h3 style="margin: var(--space-4) 0 var(--space-2) 0; font-size: 0.875rem; color: var(--text-muted);">Daily trend (last 30 days)</h3> <div class="trend-chart svelte-ffmenf"><!--[-->`);
              const each_array_26 = ensure_array_like(metricsModal.trends);
              for (let $$index_26 = 0, $$length = each_array_26.length; $$index_26 < $$length; $$index_26++) {
                let day = each_array_26[$$index_26];
                $$renderer3.push(`<div class="trend-bar-group svelte-ffmenf"${attr("title", `${stringify(day.date)}: ${stringify(day.count)} runs (${stringify(day.passed)}p/${stringify(day.failed)}f) avg ${stringify(formatMs(day.avg_ms))}`)}><div class="trend-bar svelte-ffmenf">`);
                if (day.passed > 0) {
                  $$renderer3.push("<!--[0-->");
                  $$renderer3.push(`<div class="trend-pass svelte-ffmenf"${attr_style(`height: ${stringify(day.passed / trendMax * 100)}%;`)}></div>`);
                } else {
                  $$renderer3.push("<!--[-1-->");
                }
                $$renderer3.push(`<!--]--> `);
                if (day.failed > 0) {
                  $$renderer3.push("<!--[0-->");
                  $$renderer3.push(`<div class="trend-fail svelte-ffmenf"${attr_style(`height: ${stringify(day.failed / trendMax * 100)}%;`)}></div>`);
                } else {
                  $$renderer3.push("<!--[-1-->");
                }
                $$renderer3.push(`<!--]--></div></div>`);
              }
              $$renderer3.push(`<!--]--></div>`);
            } else {
              $$renderer3.push("<!--[-1-->");
            }
            $$renderer3.push(`<!--]-->`);
          } else {
            $$renderer3.push("<!--[-1-->");
          }
          $$renderer3.push(`<!--]--> <div class="modal-actions svelte-ffmenf" style="margin-top: var(--space-4);"><button type="button" class="ghost">Close</button></div>`);
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
