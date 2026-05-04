import { ac as ensure_array_like, a9 as attr_class, a7 as attr, a8 as escape_html, ad as attr_style, ae as stringify, a2 as derived } from './renderer-mjPKoiGx.js';
import { o as onDestroy } from './index-server-BdEa2Uck.js';
import './root-gJ1T4-40.js';
import './state.svelte-CeeZin_s.js';
import { c as formatDuration, a as statusColor, s as shortSha } from './utils2-B5RqTmai.js';
import { P as Panel } from './Panel-BruuJ27z.js';
import { S as StatusDot } from './StatusDot-Cf5xUSPv.js';
import { B as Badge } from './Badge-BAVenLdx.js';
import { F as FlashMessage } from './FlashMessage-6DSDTh4l.js';
import { M as Modal } from './Modal-BJw4bo_R.js';

function _page($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    let { data, form } = $$props;
    function shortSha$1(sha) {
      return shortSha(sha, 12);
    }
    let collapsedPrefixes = /* @__PURE__ */ new Set();
    let taskFilter = "";
    let openMenuTaskID = null;
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
      "pipelines",
      "pipeline-configs",
      "compose"
    ];
    let panelOrder = [...PANEL_KEYS];
    let dragKey = null;
    let dropKey = null;
    let editLayoutMode = false;
    onDestroy(() => {
      if (typeof document !== "undefined") document.removeEventListener("click", closeMenuOnClickOutside);
    });
    function taskRow($$renderer3, t, isPinned) {
      $$renderer3.push(`<div class="task-row svelte-ffmenf"><form method="POST" action="?/togglePin" class="inline-form"><input type="hidden" name="task_id"${attr("value", t.id)}/> <input type="hidden" name="pinned"${attr("value", String(isPinned))}/> <button type="submit"${attr_class(`pin-btn ${stringify(isPinned ? "pinned" : "")}`, "svelte-ffmenf")}${attr("title", isPinned ? "Unpin from dashboard" : "Pin to dashboard")}>${escape_html(isPinned ? "★" : "☆")}</button></form> <div class="task-name-cell svelte-ffmenf"><div class="task-name svelte-ffmenf">${escape_html(t.name)} `);
      if (t.requires_approval) {
        $$renderer3.push("<!--[0-->");
        $$renderer3.push(`<span class="approval-tag svelte-ffmenf">approval</span>`);
      } else {
        $$renderer3.push("<!--[-1-->");
      }
      $$renderer3.push(`<!--]--></div> <div class="task-cmd svelte-ffmenf"${attr("title", t.raw_command)}>${escape_html(t.raw_command)}</div></div> <span class="task-meta svelte-ffmenf"><span title="Timeout">⏱ ${escape_html(formatTimeout(t.timeout_seconds))}</span> <span class="meta-sep svelte-ffmenf">·</span> <span title="Retry policy">↻ ${escape_html(formatRetry(t.retry_max, t.retry_backoff_sec))}</span> `);
      if (t.requires_approval) {
        $$renderer3.push("<!--[0-->");
        $$renderer3.push(`<span class="meta-sep svelte-ffmenf">·</span> <span class="gate-on svelte-ffmenf" title="Approval required">gate</span>`);
      } else {
        $$renderer3.push("<!--[-1-->");
      }
      $$renderer3.push(`<!--]--></span> <span class="task-actions svelte-ffmenf"><button type="button" class="menu-btn task-menu-trigger svelte-ffmenf" title="More actions" aria-haspopup="menu"${attr("aria-expanded", openMenuTaskID === t.id)}>⋯</button> <form method="POST" action="?/run" class="inline-form"><input type="hidden" name="task_id"${attr("value", t.id)}/> <button type="submit" class="run-btn svelte-ffmenf"${attr("disabled", data.project.status !== "ready", true)}>Run</button></form></span> `);
      if (openMenuTaskID === t.id) {
        $$renderer3.push("<!--[0-->");
        $$renderer3.push(`<div class="task-menu svelte-ffmenf" role="menu"><button class="menu-item svelte-ffmenf" type="button" role="menuitem"${attr("disabled", data.project.status !== "ready", true)}><span>Run with parameters…</span></button> <button class="menu-item svelte-ffmenf" type="button" role="menuitem"><span>Edit timeout</span> <span class="menu-item-value svelte-ffmenf">${escape_html(formatTimeout(t.timeout_seconds))}</span></button> <button class="menu-item svelte-ffmenf" type="button" role="menuitem"><span>Edit retry policy</span> <span class="menu-item-value svelte-ffmenf">${escape_html(formatRetry(t.retry_max, t.retry_backoff_sec))}</span></button> <form method="POST" action="?/toggleApproval" class="inline-form" style="display:contents;"><input type="hidden" name="task_id"${attr("value", t.id)}/> <input type="hidden" name="next"${attr("value", String(!t.requires_approval))}/> <button class="menu-item svelte-ffmenf" type="submit" role="menuitem"><span>${escape_html(t.requires_approval ? "Disable approval gate" : "Require approval")}</span> <span class="menu-item-value svelte-ffmenf">${escape_html(t.requires_approval ? "on" : "off")}</span></button></form></div>`);
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
                StatusDot($$renderer4, { status: r.status });
                $$renderer4.push(`<!----> <span class="run-name svelte-ffmenf">${escape_html(r.task_name)} <span class="dim svelte-ffmenf">(${escape_html(r.task_source)})</span></span> <span class="run-meta svelte-ffmenf">${escape_html(r.status)}</span> <span class="run-meta svelte-ffmenf">${escape_html(formatDuration(r.started_at, r.finished_at))}</span> <span class="run-meta svelte-ffmenf">${escape_html(new Date(r.created_at).toLocaleString())}</span></a>`);
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
            $$renderer4.push(`<div class="row svelte-ffmenf"><span class="label svelte-ffmenf">Webhook URL</span> <span class="value webhook-value svelte-ffmenf"><code class="webhook-code svelte-ffmenf">${escape_html(`${typeof window !== "undefined" ? window.location.origin : ""}/api/webhooks/projects/${data.project.webhook_token}`)}</code> <button type="button" class="ghost">Copy</button></span></div>`);
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
              StatusDot($$renderer4, { status: mappedStatus });
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
                $$renderer4.push(`<span class="pipeline-sha svelte-ffmenf">${escape_html(p.commit_sha.slice(0, 7))}</span>`);
              } else {
                $$renderer4.push("<!--[-1-->");
              }
              $$renderer4.push(`<!--]--></div> <span class="pipeline-status svelte-ffmenf"${attr_style(`color: ${stringify(statusColor(mappedStatus))};`)}>${escape_html(p.status)}</span> <span class="pipeline-time svelte-ffmenf">${escape_html(p.started_at ? formatDuration(p.started_at, p.finished_at) : "")}</span> `);
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
            StatusDot($$renderer4, {
              status: data.compose.status === "running" ? "ready" : data.compose.status === "starting" ? "cloning" : data.compose.status === "error" ? "error" : "cancelled"
            });
            $$renderer4.push(`<!----> ${escape_html(data.compose.status)}</span></div> `);
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
                $$renderer4.push(`<div class="field svelte-ffmenf"><label${attr("for", `env-${i}`)} class="svelte-ffmenf">${escape_html(envVar.key)}</label> <input${attr("id", `env-${i}`)} type="text"${attr("value", envVar.default_value)}${attr("data-env-key", envVar.key)} class="compose-env-input"/></div>`);
              }
              $$renderer4.push(`<!--]--> <button type="submit"${attr("disabled", data.project.status !== "ready", true)}>Start</button></form>`);
            } else if (data.compose.status === "stopped") {
              $$renderer4.push("<!--[1-->");
              $$renderer4.push(`<div class="compose-actions svelte-ffmenf"><button type="button"${attr("disabled", data.project.status !== "ready", true)}>Start</button></div>`);
            } else if (data.compose.status === "running") {
              $$renderer4.push("<!--[2-->");
              $$renderer4.push(`<div class="compose-actions svelte-ffmenf"><button type="button">Stop</button></div>`);
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
    $$renderer2.push(`<!--]--> <div class="layout-toolbar svelte-ffmenf"><span class="spacer svelte-ffmenf"></span> `);
    {
      $$renderer2.push("<!--[-1-->");
      $$renderer2.push(`<button type="button" class="ghost svelte-ffmenf">Customize layout</button>`);
    }
    $$renderer2.push(`<!--]--></div>     <!--[-->`);
    const each_array_9 = ensure_array_like(panelOrder);
    for (let $$index_9 = 0, $$length = each_array_9.length; $$index_9 < $$length; $$index_9++) {
      let key = each_array_9[$$index_9];
      $$renderer2.push(`<div${attr_class(`panel-shell ${stringify("")} ${stringify(dragKey === key ? "dragging" : "")} ${stringify(dropKey === key ? "drop-target" : "")}`, "svelte-ffmenf")}${attr("role", void 0)}${attr("aria-label", void 0)}${attr("draggable", editLayoutMode)}>`);
      {
        $$renderer2.push("<!--[-1-->");
      }
      $$renderer2.push(`<!--]--> `);
      if (key === "tasks") {
        $$renderer2.push("<!--[0-->");
        panelTasks($$renderer2);
      } else if (key === "recent-runs") {
        $$renderer2.push("<!--[1-->");
        panelRecentRuns($$renderer2);
      } else if (key === "repository") {
        $$renderer2.push("<!--[2-->");
        panelRepository($$renderer2);
      } else if (key === "latest-commit") {
        $$renderer2.push("<!--[3-->");
        panelLatestCommit($$renderer2);
      } else if (key === "code-stats") {
        $$renderer2.push("<!--[4-->");
        panelCodeStats($$renderer2);
      } else if (key === "pipelines") {
        $$renderer2.push("<!--[5-->");
        panelPipelines($$renderer2);
      } else if (key === "pipeline-configs") {
        $$renderer2.push("<!--[6-->");
        panelPipelineConfigs($$renderer2);
      } else if (key === "compose") {
        $$renderer2.push("<!--[7-->");
        panelCompose($$renderer2);
      } else {
        $$renderer2.push("<!--[-1-->");
      }
      $$renderer2.push(`<!--]--></div>`);
    }
    $$renderer2.push(`<!--]--> `);
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
    $$renderer2.push(`<!---->`);
  });
}

export { _page as default };
//# sourceMappingURL=_page.svelte-DCt_JGOp.js.map
