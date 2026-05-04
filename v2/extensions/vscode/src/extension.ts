import * as vscode from 'vscode';
import * as fs from 'node:fs';
import * as path from 'node:path';
import * as os from 'node:os';

// --- API client ---------------------------------------------------------

interface Workspace {
  id: string;
  name: string;
}
interface Project {
  id: string;
  name: string;
  status: string;
}
interface Task {
  id: string;
  source: string;
  name: string;
  raw_command: string;
}
interface Run {
  id: string;
  status: string;
  exit_code: number | null;
  task_name: string;
  task_source: string;
  project_name: string;
  workspace_name: string;
  created_at: string;
}

class Api {
  private baseURL: string;
  private cookie: string | null = null;

  constructor() {
    this.baseURL = (vscode.workspace.getConfiguration('workend').get<string>('apiUrl') ||
      'http://localhost:8080').replace(/\/+$/, '');
    this.loadSession();
  }

  // The session token is shared with the CLI: a single line file at
  // $XDG_CONFIG_HOME/workend/session (or ~/.config/workend/session).
  private loadSession() {
    const override = vscode.workspace.getConfiguration('workend').get<string>('sessionFile');
    let p = override && override.length > 0 ? override :
      path.join(process.env.XDG_CONFIG_HOME || path.join(os.homedir(), '.config'), 'workend', 'session');
    try {
      this.cookie = fs.readFileSync(p, 'utf8').trim();
    } catch {
      this.cookie = null;
    }
  }

  hasSession(): boolean {
    return !!this.cookie;
  }

  private async req<T>(path: string, init: RequestInit = {}): Promise<T> {
    if (!this.cookie) throw new Error('no Workend session — run `workend login` in your terminal');
    const headers = new Headers(init.headers);
    headers.set('Cookie', `workend_session=${this.cookie}`);
    if (init.body && !headers.has('Content-Type')) headers.set('Content-Type', 'application/json');
    const res = await fetch(this.baseURL + path, { ...init, headers });
    if (!res.ok) throw new Error(`HTTP ${res.status}: ${(await res.text()).trim()}`);
    if (res.status === 204) return undefined as T;
    return (await res.json()) as T;
  }

  workspaces() { return this.req<Workspace[]>('/api/workspaces'); }
  projects(wsID: string) { return this.req<Project[]>(`/api/workspaces/${wsID}/projects`); }
  tasks(projectID: string) { return this.req<Task[]>(`/api/projects/${projectID}/tasks`); }
  recentRuns() { return this.req<Run[]>('/api/me/runs'); }
  startRun(taskID: string) { return this.req<{ id: string }>(`/api/tasks/${taskID}/runs`, { method: 'POST' }); }
  runLog(runID: string) {
    return fetch(`${this.baseURL}/api/runs/${runID}/log`, {
      headers: { Cookie: `workend_session=${this.cookie}` }
    }).then((r) => r.text());
  }
  // SSE log tail using fetch + ReadableStream (Node 22 supports it).
  async streamLog(runID: string, onChunk: (s: string) => void, onDone: () => void): Promise<void> {
    const res = await fetch(`${this.baseURL}/api/runs/${runID}/log/stream`, {
      headers: { Cookie: `workend_session=${this.cookie}`, Accept: 'text/event-stream' }
    });
    if (!res.ok || !res.body) {
      onChunk(`stream error: HTTP ${res.status}\n`);
      onDone();
      return;
    }
    const reader = res.body.getReader();
    const decoder = new TextDecoder();
    let buf = '';
    while (true) {
      const { done, value } = await reader.read();
      if (done) break;
      buf += decoder.decode(value, { stream: true });
      // Each SSE message is a block separated by a blank line.
      let idx;
      while ((idx = buf.indexOf('\n\n')) >= 0) {
        const block = buf.slice(0, idx);
        buf = buf.slice(idx + 2);
        let event = '';
        const dataLines: string[] = [];
        for (const line of block.split('\n')) {
          if (line.startsWith('event:')) event = line.slice(6).trim();
          else if (line.startsWith('data:')) dataLines.push(line.slice(5).replace(/^ /, ''));
        }
        const data = dataLines.join('\n');
        if (event === 'log') onChunk(data + '\n');
        else if (event === 'done') {
          onDone();
          return;
        }
      }
    }
    onDone();
  }
}

// --- Tree provider: workspaces → projects → tasks ----------------------

type Node =
  | { kind: 'workspace'; ws: Workspace }
  | { kind: 'project'; p: Project; wsName: string }
  | { kind: 'task'; t: Task; projectID: string };

class WorkspacesProvider implements vscode.TreeDataProvider<Node> {
  private _onDidChange = new vscode.EventEmitter<Node | undefined | void>();
  onDidChangeTreeData = this._onDidChange.event;

  constructor(private api: Api) {}
  refresh() { this._onDidChange.fire(); }

  getTreeItem(n: Node): vscode.TreeItem {
    if (n.kind === 'workspace') {
      const item = new vscode.TreeItem(n.ws.name, vscode.TreeItemCollapsibleState.Collapsed);
      item.iconPath = new vscode.ThemeIcon('folder-library');
      return item;
    }
    if (n.kind === 'project') {
      const item = new vscode.TreeItem(n.p.name, vscode.TreeItemCollapsibleState.Collapsed);
      item.description = n.p.status;
      item.iconPath = new vscode.ThemeIcon('git-branch');
      return item;
    }
    const item = new vscode.TreeItem(`${n.t.source}: ${n.t.name}`, vscode.TreeItemCollapsibleState.None);
    item.description = n.t.raw_command;
    item.tooltip = `Run ${n.t.name} (${n.t.source})`;
    item.iconPath = new vscode.ThemeIcon('play');
    item.command = { command: 'workend._runTaskFromTree', title: 'Run', arguments: [n.t] };
    return item;
  }

  async getChildren(parent?: Node): Promise<Node[]> {
    if (!parent) {
      try {
        const ws = await this.api.workspaces();
        return ws.map((w) => ({ kind: 'workspace', ws: w } as Node));
      } catch (err) {
        vscode.window.showErrorMessage(`Workend: ${(err as Error).message}`);
        return [];
      }
    }
    if (parent.kind === 'workspace') {
      const projects = await this.api.projects(parent.ws.id);
      return projects.map((p) => ({ kind: 'project', p, wsName: parent.ws.name } as Node));
    }
    if (parent.kind === 'project') {
      const tasks = await this.api.tasks(parent.p.id);
      return tasks.map((t) => ({ kind: 'task', t, projectID: parent.p.id } as Node));
    }
    return [];
  }
}

class RunsProvider implements vscode.TreeDataProvider<Run> {
  private _onDidChange = new vscode.EventEmitter<Run | undefined | void>();
  onDidChangeTreeData = this._onDidChange.event;
  constructor(private api: Api) {}
  refresh() { this._onDidChange.fire(); }
  getTreeItem(r: Run): vscode.TreeItem {
    const item = new vscode.TreeItem(`${r.task_name} · ${r.status}`, vscode.TreeItemCollapsibleState.None);
    item.description = `${r.workspace_name} / ${r.project_name}`;
    item.tooltip = `${r.task_source} · created ${r.created_at}${r.exit_code !== null ? ` · exit ${r.exit_code}` : ''}`;
    item.iconPath = new vscode.ThemeIcon(
      r.status === 'succeeded' ? 'pass'
        : r.status === 'failed' ? 'error'
        : r.status === 'cancelled' ? 'circle-slash'
        : 'sync'
    );
    item.command = { command: 'workend._openRunFromTree', title: 'Open', arguments: [r] };
    return item;
  }
  async getChildren(): Promise<Run[]> {
    try {
      return await this.api.recentRuns();
    } catch {
      return [];
    }
  }
}

// --- Activation --------------------------------------------------------

export function activate(ctx: vscode.ExtensionContext) {
  const api = new Api();
  const wsProvider = new WorkspacesProvider(api);
  const runsProvider = new RunsProvider(api);
  const channel = vscode.window.createOutputChannel('Workend', 'log');

  ctx.subscriptions.push(
    vscode.window.registerTreeDataProvider('workend.workspaces', wsProvider),
    vscode.window.registerTreeDataProvider('workend.recentRuns', runsProvider),
    channel
  );

  if (!api.hasSession()) {
    vscode.window.showWarningMessage(
      'Workend: no session found. Run `workend login` in your terminal first.'
    );
  }

  ctx.subscriptions.push(
    vscode.commands.registerCommand('workend.refresh', () => {
      wsProvider.refresh();
      runsProvider.refresh();
    }),

    vscode.commands.registerCommand('workend._runTaskFromTree', (t: Task) => runAndStream(api, channel, t.id, `${t.name} (${t.source})`)),
    vscode.commands.registerCommand('workend._openRunFromTree', (r: Run) => openRun(api, channel, r.id, `${r.task_name} (${r.task_source})`)),

    vscode.commands.registerCommand('workend.runTask', async () => {
      const ws = await pick(api.workspaces(), (w) => w.name, 'Select workspace');
      if (!ws) return;
      const proj = await pick(api.projects(ws.id), (p) => p.name, 'Select project');
      if (!proj) return;
      const task = await pick(api.tasks(proj.id), (t) => `${t.source}: ${t.name}`, 'Select task');
      if (!task) return;
      await runAndStream(api, channel, task.id, `${task.name} (${task.source})`);
    }),

    vscode.commands.registerCommand('workend.openRun', async () => {
      const r = await pick(api.recentRuns(), (r) => `${r.task_name} · ${r.status} · ${r.workspace_name}/${r.project_name}`, 'Open run');
      if (r) await openRun(api, channel, r.id, `${r.task_name} (${r.task_source})`);
    })
  );
}

export function deactivate() {}

async function pick<T>(items: Promise<T[]>, label: (t: T) => string, placeHolder: string): Promise<T | undefined> {
  const list = await items;
  const choice = await vscode.window.showQuickPick(
    list.map((it) => ({ label: label(it), it })),
    { placeHolder }
  );
  return choice?.it;
}

async function runAndStream(api: Api, channel: vscode.OutputChannel, taskID: string, label: string) {
  channel.show(true);
  channel.appendLine(`\n=== Starting ${label} ===`);
  try {
    const { id } = await api.startRun(taskID);
    channel.appendLine(`run id: ${id}`);
    await api.streamLog(
      id,
      (chunk) => channel.append(chunk),
      () => channel.appendLine(`\n=== run ${id} finished ===`)
    );
  } catch (err) {
    channel.appendLine(`error: ${(err as Error).message}`);
  }
}

async function openRun(api: Api, channel: vscode.OutputChannel, runID: string, label: string) {
  channel.show(true);
  channel.appendLine(`\n=== Run ${label} (${runID}) ===`);
  try {
    const log = await api.runLog(runID);
    channel.append(log);
    channel.appendLine(`\n=== end of log ===`);
  } catch (err) {
    channel.appendLine(`error: ${(err as Error).message}`);
  }
}
