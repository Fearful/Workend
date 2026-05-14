import { ac as ensure_array_like, a8 as escape_html, a7 as attr, a9 as attr_class, ad as stringify, ae as attr_style, a2 as derived } from './renderer-D0X3o35U.js';
import { f as freshnessClass, a as statusColor, b as formatRelative } from './utils2-BUPlP7zG.js';
import { P as PageHeader } from './PageHeader-C8D5NSCp.js';
import { P as Panel } from './Panel-B1HKSjmN.js';
import { S as StatusPill } from './StatusPill-j6_a2per.js';
import { B as Badge } from './Badge-imHJGUZA.js';
import { E as EmptyState } from './EmptyState-5A_wG7lI.js';
import { T as TimeAgo } from './TimeAgo-DaxliHBa.js';
import './Tooltip-Bg85MZ7d.js';
import './index-server-BFLhAcPs.js';

function _page($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    let { data } = $$props;
    let groupedByWorkspace = derived(() => {
      const groups = {};
      for (const c of data.cards) {
        if (!groups[c.workspace_name]) groups[c.workspace_name] = [];
        groups[c.workspace_name].push(c);
      }
      return groups;
    });
    let stats = derived(() => {
      const running = data.cards.filter((c) => c.last_run_status === "running" || c.last_run_status === "queued").length;
      const failing = data.cards.filter((c) => c.has_failing_recent_run).length;
      return { total: data.cards.length, running, failing };
    });
    const DAY_LABELS = ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"];
    const HOUR_LABELS = Array.from({ length: 24 }, (_, i) => {
      if (i === 0) return "12a";
      if (i < 12) return `${i}a`;
      if (i === 12) return "12p";
      return `${i - 12}p`;
    });
    let heatmapGrid = derived(() => {
      const grid = Array.from({ length: 7 }, () => Array(24).fill(0));
      let max = 0;
      for (const b of data.heatmap) {
        const row = b.day_of_week;
        const col = b.hour;
        if (row >= 0 && row < 7 && col >= 0 && col < 24) {
          grid[row][col] = b.failures;
          if (b.failures > max) max = b.failures;
        }
      }
      return { grid, max };
    });
    let queueItemCount = derived(() => data.myQueue.pending_approvals.length + data.myQueue.expiring_sandboxes.length + (data.myQueue.unread_mentions > 0 ? 1 : 0));
    let hasWidgets = derived(() => data.runPulse.length > 0 || data.heatmap.length > 0 || queueItemCount() > 0 || data.velocity.this_week.runs > 0 || data.velocity.last_week.runs > 0 || data.sandboxes.length > 0 || data.quota.length > 0);
    function formatMs(ms) {
      if (ms <= 0) return "0s";
      const sec = Math.round(ms / 1e3);
      if (sec < 60) return `${sec}s`;
      return `${Math.floor(sec / 60)}m ${sec % 60}s`;
    }
    function passRate(passed, total) {
      if (total === 0) return "—";
      return `${Math.round(passed / total * 100)}%`;
    }
    function deltaClass(delta, higherIsBetter) {
      if (delta === 0) return "";
      const positive = delta > 0;
      return positive === higherIsBetter ? "delta-good" : "delta-bad";
    }
    function deltaArrow(delta) {
      if (delta > 0) return "↑";
      if (delta < 0) return "↓";
      return "";
    }
    PageHeader($$renderer2, { title: "Dashboard" });
    $$renderer2.push(`<!----> `);
    if (hasWidgets()) {
      $$renderer2.push("<!--[0-->");
      $$renderer2.push(`<div class="widgets-section svelte-x1i5gj"><div class="widgets-grid svelte-x1i5gj">`);
      if (data.runPulse.length > 0) {
        $$renderer2.push("<!--[0-->");
        $$renderer2.push(`<div class="widget widget-run-pulse">`);
        Panel($$renderer2, {
          title: "Run Pulse",
          padding: "compact",
          children: ($$renderer3) => {
            $$renderer3.push(`<div class="pulse-list svelte-x1i5gj"><!--[-->`);
            const each_array = ensure_array_like(data.runPulse);
            for (let $$index = 0, $$length = each_array.length; $$index < $$length; $$index++) {
              let run = each_array[$$index];
              $$renderer3.push(`<a${attr("href", `/runs/${run.id}`)} class="pulse-row svelte-x1i5gj">`);
              StatusPill($$renderer3, { status: run.status, size: "sm" });
              $$renderer3.push(`<!----> <span class="pulse-task svelte-x1i5gj">${escape_html(run.task_name)}</span> <span class="pulse-meta svelte-x1i5gj">${escape_html(run.project_name)}</span> <span class="pulse-meta svelte-x1i5gj">${escape_html(run.workspace_name)}</span> `);
              TimeAgo($$renderer3, { value: run.started_at });
              $$renderer3.push(`<!----></a>`);
            }
            $$renderer3.push(`<!--]--></div>`);
          }
        });
        $$renderer2.push(`<!----></div>`);
      } else {
        $$renderer2.push("<!--[-1-->");
      }
      $$renderer2.push(`<!--]--> `);
      if (data.velocity.this_week.runs > 0 || data.velocity.last_week.runs > 0) {
        $$renderer2.push("<!--[0-->");
        $$renderer2.push(`<div class="widget widget-velocity">`);
        Panel($$renderer2, {
          title: "Sprint Velocity",
          padding: "compact",
          children: ($$renderer3) => {
            $$renderer3.push(`<div class="velocity-table svelte-x1i5gj"><div class="vel-header svelte-x1i5gj"><span class="vel-label svelte-x1i5gj"></span> <span class="vel-col svelte-x1i5gj">This week</span> <span class="vel-col svelte-x1i5gj">Last week</span> <span class="vel-col svelte-x1i5gj">Delta</span></div> <div class="vel-row svelte-x1i5gj"><span class="vel-label svelte-x1i5gj">Runs</span> <span class="vel-val svelte-x1i5gj">${escape_html(data.velocity.this_week.runs)}</span> <span class="vel-val svelte-x1i5gj">${escape_html(data.velocity.last_week.runs)}</span> <span${attr_class(`vel-val vel-delta ${stringify(deltaClass(data.velocity.deltas.runs, true))}`, "svelte-x1i5gj")}>${escape_html(deltaArrow(data.velocity.deltas.runs))} ${escape_html(Math.abs(data.velocity.deltas.runs))}</span></div> <div class="vel-row svelte-x1i5gj"><span class="vel-label svelte-x1i5gj">Pass rate</span> <span class="vel-val svelte-x1i5gj">${escape_html(passRate(data.velocity.this_week.passed, data.velocity.this_week.runs))}</span> <span class="vel-val svelte-x1i5gj">${escape_html(passRate(data.velocity.last_week.passed, data.velocity.last_week.runs))}</span> <span${attr_class(`vel-val vel-delta ${stringify(deltaClass(data.velocity.deltas.passed, true))}`, "svelte-x1i5gj")}>${escape_html(deltaArrow(data.velocity.deltas.passed))} ${escape_html(Math.abs(data.velocity.deltas.passed))}</span></div> <div class="vel-row svelte-x1i5gj"><span class="vel-label svelte-x1i5gj">Failed</span> <span class="vel-val svelte-x1i5gj">${escape_html(data.velocity.this_week.failed)}</span> <span class="vel-val svelte-x1i5gj">${escape_html(data.velocity.last_week.failed)}</span> <span${attr_class(`vel-val vel-delta ${stringify(deltaClass(data.velocity.deltas.failed, false))}`, "svelte-x1i5gj")}>${escape_html(deltaArrow(data.velocity.deltas.failed))} ${escape_html(Math.abs(data.velocity.deltas.failed))}</span></div> <div class="vel-row svelte-x1i5gj"><span class="vel-label svelte-x1i5gj">Avg duration</span> <span class="vel-val svelte-x1i5gj">${escape_html(formatMs(data.velocity.this_week.avg_duration_ms))}</span> <span class="vel-val svelte-x1i5gj">${escape_html(formatMs(data.velocity.last_week.avg_duration_ms))}</span> <span${attr_class(`vel-val vel-delta ${stringify(deltaClass(data.velocity.deltas.avg_duration_ms, false))}`, "svelte-x1i5gj")}>${escape_html(deltaArrow(data.velocity.deltas.avg_duration_ms))} ${escape_html(formatMs(Math.abs(data.velocity.deltas.avg_duration_ms)))}</span></div></div>`);
          }
        });
        $$renderer2.push(`<!----></div>`);
      } else {
        $$renderer2.push("<!--[-1-->");
      }
      $$renderer2.push(`<!--]--> `);
      if (queueItemCount() > 0) {
        $$renderer2.push("<!--[0-->");
        $$renderer2.push(`<div class="widget widget-queue">`);
        Panel($$renderer2, {
          title: "My Queue",
          padding: "compact",
          children: ($$renderer3) => {
            $$renderer3.push(`<div class="queue-list svelte-x1i5gj"><!--[-->`);
            const each_array_1 = ensure_array_like(data.myQueue.pending_approvals);
            for (let $$index_1 = 0, $$length = each_array_1.length; $$index_1 < $$length; $$index_1++) {
              let approval = each_array_1[$$index_1];
              $$renderer3.push(`<a${attr("href", `/runs/${approval.run_id}`)} class="queue-item svelte-x1i5gj">`);
              Badge($$renderer3, {
                variant: "accent",
                size: "sm",
                children: ($$renderer4) => {
                  $$renderer4.push(`<!---->approval`);
                }
              });
              $$renderer3.push(`<!----> <span class="queue-task svelte-x1i5gj">${escape_html(approval.task_name)}</span> <span class="queue-meta svelte-x1i5gj">${escape_html(approval.project_name)}</span> `);
              TimeAgo($$renderer3, { value: approval.created_at });
              $$renderer3.push(`<!----></a>`);
            }
            $$renderer3.push(`<!--]--> <!--[-->`);
            const each_array_2 = ensure_array_like(data.myQueue.expiring_sandboxes);
            for (let $$index_2 = 0, $$length = each_array_2.length; $$index_2 < $$length; $$index_2++) {
              let sb = each_array_2[$$index_2];
              $$renderer3.push(`<div class="queue-item svelte-x1i5gj">`);
              Badge($$renderer3, {
                variant: "warning",
                size: "sm",
                children: ($$renderer4) => {
                  $$renderer4.push(`<!---->expiring`);
                }
              });
              $$renderer3.push(`<!----> <span class="queue-task svelte-x1i5gj">${escape_html(sb.branch)}</span> <span class="queue-meta queue-urgent svelte-x1i5gj">${escape_html(sb.minutes_left)}m left</span></div>`);
            }
            $$renderer3.push(`<!--]--> `);
            if (data.myQueue.unread_mentions > 0) {
              $$renderer3.push("<!--[0-->");
              $$renderer3.push(`<div class="queue-item svelte-x1i5gj">`);
              Badge($$renderer3, {
                variant: "info",
                size: "sm",
                children: ($$renderer4) => {
                  $$renderer4.push(`<!---->mentions`);
                }
              });
              $$renderer3.push(`<!----> <span class="queue-task svelte-x1i5gj">${escape_html(data.myQueue.unread_mentions)} unread</span></div>`);
            } else {
              $$renderer3.push("<!--[-1-->");
            }
            $$renderer3.push(`<!--]--></div>`);
          }
        });
        $$renderer2.push(`<!----></div>`);
      } else {
        $$renderer2.push("<!--[-1-->");
      }
      $$renderer2.push(`<!--]--> `);
      if (data.heatmap.length > 0) {
        $$renderer2.push("<!--[0-->");
        $$renderer2.push(`<div class="widget widget-heatmap svelte-x1i5gj">`);
        Panel($$renderer2, {
          title: "Failure Heatmap",
          padding: "compact",
          children: ($$renderer3) => {
            $$renderer3.push(`<div class="heatmap-wrap svelte-x1i5gj"><div class="heatmap-grid svelte-x1i5gj"><div class="heatmap-corner svelte-x1i5gj"></div> <!--[-->`);
            const each_array_3 = ensure_array_like(HOUR_LABELS);
            for (let i = 0, $$length = each_array_3.length; i < $$length; i++) {
              let h = each_array_3[i];
              if (i % 3 === 0) {
                $$renderer3.push("<!--[0-->");
                $$renderer3.push(`<span class="heatmap-hour-label svelte-x1i5gj">${escape_html(h)}</span>`);
              } else {
                $$renderer3.push("<!--[-1-->");
                $$renderer3.push(`<span class="heatmap-hour-label svelte-x1i5gj"></span>`);
              }
              $$renderer3.push(`<!--]-->`);
            }
            $$renderer3.push(`<!--]--> <!--[-->`);
            const each_array_4 = ensure_array_like(heatmapGrid().grid);
            for (let dayIdx = 0, $$length = each_array_4.length; dayIdx < $$length; dayIdx++) {
              let dayRow = each_array_4[dayIdx];
              $$renderer3.push(`<span class="heatmap-day-label svelte-x1i5gj">${escape_html(DAY_LABELS[dayIdx])}</span> <!--[-->`);
              const each_array_5 = ensure_array_like(dayRow);
              for (let hourIdx = 0, $$length2 = each_array_5.length; hourIdx < $$length2; hourIdx++) {
                let count = each_array_5[hourIdx];
                $$renderer3.push(`<div class="heatmap-cell svelte-x1i5gj"${attr_style(`--intensity: ${stringify(heatmapGrid().max > 0 ? count / heatmapGrid().max : 0)}`)}${attr("title", `${stringify(DAY_LABELS[dayIdx])} ${stringify(HOUR_LABELS[hourIdx])}: ${stringify(count)} failure${stringify(count !== 1 ? "s" : "")}`)}></div>`);
              }
              $$renderer3.push(`<!--]-->`);
            }
            $$renderer3.push(`<!--]--></div> <div class="heatmap-legend svelte-x1i5gj"><span class="heatmap-legend-label svelte-x1i5gj">Less</span> <div class="heatmap-legend-cell svelte-x1i5gj" style="--intensity: 0"></div> <div class="heatmap-legend-cell svelte-x1i5gj" style="--intensity: 0.25"></div> <div class="heatmap-legend-cell svelte-x1i5gj" style="--intensity: 0.5"></div> <div class="heatmap-legend-cell svelte-x1i5gj" style="--intensity: 0.75"></div> <div class="heatmap-legend-cell svelte-x1i5gj" style="--intensity: 1"></div> <span class="heatmap-legend-label svelte-x1i5gj">More</span></div></div>`);
          }
        });
        $$renderer2.push(`<!----></div>`);
      } else {
        $$renderer2.push("<!--[-1-->");
      }
      $$renderer2.push(`<!--]--> `);
      if (data.sandboxes.length > 0) {
        $$renderer2.push("<!--[0-->");
        $$renderer2.push(`<div class="widget widget-sandboxes">`);
        Panel($$renderer2, {
          title: "Sandbox Status",
          padding: "compact",
          children: ($$renderer3) => {
            $$renderer3.push(`<div class="sandbox-list svelte-x1i5gj"><!--[-->`);
            const each_array_6 = ensure_array_like(data.sandboxes);
            for (let $$index_6 = 0, $$length = each_array_6.length; $$index_6 < $$length; $$index_6++) {
              let sb = each_array_6[$$index_6];
              $$renderer3.push(`<div class="sandbox-row svelte-x1i5gj">`);
              StatusPill($$renderer3, { status: sb.status, size: "sm" });
              $$renderer3.push(`<!----> <span class="sandbox-branch svelte-x1i5gj">${escape_html(sb.branch)}</span> <span class="sandbox-project svelte-x1i5gj">${escape_html(sb.project_name)}</span> `);
              if (sb.minutes_left <= 10) {
                $$renderer3.push("<!--[0-->");
                Badge($$renderer3, {
                  variant: "danger",
                  size: "sm",
                  children: ($$renderer4) => {
                    $$renderer4.push(`<!---->${escape_html(sb.minutes_left)}m left`);
                  }
                });
              } else if (sb.minutes_left <= 30) {
                $$renderer3.push("<!--[1-->");
                Badge($$renderer3, {
                  variant: "warning",
                  size: "sm",
                  children: ($$renderer4) => {
                    $$renderer4.push(`<!---->${escape_html(sb.minutes_left)}m left`);
                  }
                });
              } else {
                $$renderer3.push("<!--[-1-->");
                Badge($$renderer3, {
                  variant: "muted",
                  size: "sm",
                  children: ($$renderer4) => {
                    $$renderer4.push(`<!---->${escape_html(sb.minutes_left)}m left`);
                  }
                });
              }
              $$renderer3.push(`<!--]--></div>`);
            }
            $$renderer3.push(`<!--]--></div>`);
          }
        });
        $$renderer2.push(`<!----></div>`);
      } else {
        $$renderer2.push("<!--[-1-->");
      }
      $$renderer2.push(`<!--]--> `);
      if (data.quota.length > 0) {
        $$renderer2.push("<!--[0-->");
        $$renderer2.push(`<div class="widget widget-quota">`);
        Panel($$renderer2, {
          title: "Quota Meter",
          padding: "compact",
          children: ($$renderer3) => {
            $$renderer3.push(`<div class="quota-list svelte-x1i5gj"><!--[-->`);
            const each_array_7 = ensure_array_like(data.quota);
            for (let $$index_7 = 0, $$length = each_array_7.length; $$index_7 < $$length; $$index_7++) {
              let ws = each_array_7[$$index_7];
              $$renderer3.push(`<div class="quota-row svelte-x1i5gj"><div class="quota-head svelte-x1i5gj"><span class="quota-name svelte-x1i5gj">${escape_html(ws.workspace_name)}</span> <span class="quota-stats svelte-x1i5gj">${escape_html(ws.projects)} projects · ${escape_html(ws.runs_30d)} runs / 30d</span></div> <div class="quota-bar-track svelte-x1i5gj"><div${attr_class(
                `quota-bar-fill ${stringify(ws.usage_percent >= 90 ? "quota-danger" : ws.usage_percent >= 70 ? "quota-warn" : "")}`,
                "svelte-x1i5gj"
              )}${attr_style(`width: ${stringify(Math.min(ws.usage_percent, 100))}%`)}></div></div> <span class="quota-pct svelte-x1i5gj">${escape_html(Math.round(ws.usage_percent))}%</span></div>`);
            }
            $$renderer3.push(`<!--]--></div>`);
          }
        });
        $$renderer2.push(`<!----></div>`);
      } else {
        $$renderer2.push("<!--[-1-->");
      }
      $$renderer2.push(`<!--]--></div></div>`);
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--> <div class="layout svelte-x1i5gj"><div class="main-content svelte-x1i5gj">`);
    if (data.cards.length === 0) {
      $$renderer2.push("<!--[0-->");
      EmptyState($$renderer2, {
        icon: "◇",
        message: "No projects yet. Create a workspace and add some.",
        actionHref: "/",
        actionLabel: "Browse workspaces"
      });
    } else {
      $$renderer2.push("<!--[-1-->");
      $$renderer2.push(`<!--[-->`);
      const each_array_8 = ensure_array_like(Object.entries(groupedByWorkspace()));
      for (let $$index_9 = 0, $$length = each_array_8.length; $$index_9 < $$length; $$index_9++) {
        let [wsName, cards] = each_array_8[$$index_9];
        $$renderer2.push(`<h2 class="svelte-x1i5gj">${escape_html(wsName)}</h2> <div class="grid svelte-x1i5gj"><!--[-->`);
        const each_array_9 = ensure_array_like(cards);
        for (let $$index_8 = 0, $$length2 = each_array_9.length; $$index_8 < $$length2; $$index_8++) {
          let c = each_array_9[$$index_8];
          $$renderer2.push(`<a${attr("href", `/projects/${c.project_id}`)}${attr_class(`card ${stringify(c.has_failing_recent_run ? "alert" : "")}`, "svelte-x1i5gj")}><div class="card-head svelte-x1i5gj">`);
          StatusPill($$renderer2, { status: c.status, size: "sm" });
          $$renderer2.push(`<!----> <span class="card-name svelte-x1i5gj">${escape_html(c.project_name)}</span> `);
          if (c.has_failing_recent_run) {
            $$renderer2.push("<!--[0-->");
            Badge($$renderer2, {
              variant: "danger",
              size: "sm",
              children: ($$renderer3) => {
                $$renderer3.push(`<!---->recent failure`);
              }
            });
          } else {
            $$renderer2.push("<!--[-1-->");
          }
          $$renderer2.push(`<!--]--></div> `);
          if (c.last_commit_message) {
            $$renderer2.push("<!--[0-->");
            $$renderer2.push(`<div class="commit-line svelte-x1i5gj">${escape_html(c.last_commit_message)}</div>`);
          } else {
            $$renderer2.push("<!--[-1-->");
          }
          $$renderer2.push(`<!--]--> <div class="meta-grid svelte-x1i5gj"><span class="label svelte-x1i5gj">Status</span><span class="value svelte-x1i5gj">${escape_html(c.status)}</span> <span class="label svelte-x1i5gj">Tasks</span><span class="value svelte-x1i5gj">${escape_html(c.task_count)}</span> <span class="label svelte-x1i5gj">Lang</span><span class="value svelte-x1i5gj">${escape_html(c.top_language || "—")}</span> <span class="label svelte-x1i5gj">Lines</span><span class="value svelte-x1i5gj">${escape_html(c.total_lines ? c.total_lines.toLocaleString() : "—")}</span> <span class="label svelte-x1i5gj">Last sync</span> `);
          if (freshnessClass(c.last_synced_at) === "fresh-good") {
            $$renderer2.push("<!--[0-->");
            Badge($$renderer2, {
              variant: "success",
              size: "sm",
              children: ($$renderer3) => {
                $$renderer3.push(`<!---->${escape_html(formatRelative(c.last_synced_at))}`);
              }
            });
          } else if (freshnessClass(c.last_synced_at) === "fresh-warn") {
            $$renderer2.push("<!--[1-->");
            Badge($$renderer2, {
              variant: "warning",
              size: "sm",
              children: ($$renderer3) => {
                $$renderer3.push(`<!---->${escape_html(formatRelative(c.last_synced_at))}`);
              }
            });
          } else {
            $$renderer2.push("<!--[-1-->");
            Badge($$renderer2, {
              variant: "muted",
              size: "sm",
              children: ($$renderer3) => {
                $$renderer3.push(`<!---->${escape_html(formatRelative(c.last_synced_at))}`);
              }
            });
          }
          $$renderer2.push(`<!--]--> <span class="label svelte-x1i5gj">Last run</span> `);
          if (c.last_run_status) {
            $$renderer2.push("<!--[0-->");
            $$renderer2.push(`<span class="value svelte-x1i5gj"${attr_style(`color: ${stringify(statusColor(c.last_run_status))};`)}>${escape_html(c.last_run_task_name)} · ${escape_html(c.last_run_status)}</span>`);
          } else {
            $$renderer2.push("<!--[-1-->");
            $$renderer2.push(`<span class="value svelte-x1i5gj">never</span>`);
          }
          $$renderer2.push(`<!--]--></div></a>`);
        }
        $$renderer2.push(`<!--]--></div>`);
      }
      $$renderer2.push(`<!--]-->`);
    }
    $$renderer2.push(`<!--]--></div> <aside class="sidebar svelte-x1i5gj">`);
    if (data.cards.length > 0) {
      $$renderer2.push("<!--[0-->");
      $$renderer2.push(`<h2 class="svelte-x1i5gj">Overview</h2> <div class="stats-grid svelte-x1i5gj"><div class="stat-tile svelte-x1i5gj"><div class="stat-num svelte-x1i5gj">${escape_html(stats().total)}</div> <div class="stat-label svelte-x1i5gj">Projects</div></div> <div class="stat-tile svelte-x1i5gj"><div class="stat-num warning svelte-x1i5gj">${escape_html(stats().running)}</div> <div class="stat-label svelte-x1i5gj">Running</div></div> <div class="stat-tile svelte-x1i5gj"><div class="stat-num danger svelte-x1i5gj">${escape_html(stats().failing)}</div> <div class="stat-label svelte-x1i5gj">Failing</div></div></div>`);
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--> `);
    if (data.pinned.length > 0) {
      $$renderer2.push("<!--[0-->");
      $$renderer2.push(`<h2 class="svelte-x1i5gj">Pinned tasks</h2> <div class="pinned-grid svelte-x1i5gj"><!--[-->`);
      const each_array_10 = ensure_array_like(data.pinned);
      for (let $$index_10 = 0, $$length = each_array_10.length; $$index_10 < $$length; $$index_10++) {
        let p = each_array_10[$$index_10];
        $$renderer2.push(`<div class="pin-card svelte-x1i5gj"><div class="pin-head svelte-x1i5gj"><span class="pin-task svelte-x1i5gj">${escape_html(p.task_name)}</span> <span class="pin-source svelte-x1i5gj">${escape_html(p.task_source)}</span></div> <div class="pin-loc svelte-x1i5gj"><a${attr("href", `/projects/${p.project_id}`)} class="svelte-x1i5gj">${escape_html(p.workspace_name)} / ${escape_html(p.project_name)}</a></div> <div class="pin-actions svelte-x1i5gj"><form method="POST" action="?/runPinned" class="pin-run-form svelte-x1i5gj"><input type="hidden" name="task_id"${attr("value", p.task_id)}/> <button type="submit" class="svelte-x1i5gj">Run</button></form> <form method="POST" action="?/unpin" class="svelte-x1i5gj"><input type="hidden" name="task_id"${attr("value", p.task_id)}/> <button type="submit" class="ghost" title="Unpin" aria-label="Unpin">★</button></form></div></div>`);
      }
      $$renderer2.push(`<!--]--></div>`);
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--> `);
    if (data.flaky.length > 0) {
      $$renderer2.push("<!--[0-->");
      $$renderer2.push(`<h2 class="warn svelte-x1i5gj">Flaky tasks</h2> <div class="flaky-grid svelte-x1i5gj"><!--[-->`);
      const each_array_11 = ensure_array_like(data.flaky);
      for (let $$index_11 = 0, $$length = each_array_11.length; $$index_11 < $$length; $$index_11++) {
        let f = each_array_11[$$index_11];
        $$renderer2.push(`<a${attr("href", `/projects/${f.project_id}/trends?task_id=${f.task_id}`)} class="flaky-card svelte-x1i5gj"><div class="flaky-head svelte-x1i5gj"><span class="flaky-task svelte-x1i5gj">${escape_html(f.task_name)}</span> <span class="flaky-source svelte-x1i5gj">${escape_html(f.task_source)}</span></div> <div class="flaky-loc svelte-x1i5gj">${escape_html(f.workspace_name)} / ${escape_html(f.project_name)}</div> <div class="flaky-stats svelte-x1i5gj"><span><strong class="flaky-flips svelte-x1i5gj">${escape_html(f.flips)}</strong> flips / ${escape_html(f.total)} runs</span> <span>${escape_html(Math.round(f.success_rate * 100))}% success</span></div></a>`);
      }
      $$renderer2.push(`<!--]--></div>`);
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--></aside></div>`);
  });
}

export { _page as default };
//# sourceMappingURL=_page.svelte-BoCR3mDg.js.map
