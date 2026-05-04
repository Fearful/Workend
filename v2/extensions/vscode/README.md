# Workend for VS Code

A sidebar extension that lists your Workend workspaces and recent runs and
lets you trigger task runs without leaving the editor.

This is a **scaffold** — it's wired to the existing API and the same session
cookie used by the `workend` CLI, but distribution to the marketplace is left
to the user's release pipeline.

## Setup (dev)

1. `cd v2/extensions/vscode && npm install && npm run compile`
2. Open this folder in VS Code and press `F5` to launch a new Extension Host.
3. In the host window, log in with the [Workend CLI](../../cli) so a session
   cookie exists at `~/.config/workend/session`.
4. Set `workend.apiUrl` in settings if your API isn't on `localhost:8080`.

## What's wired

- **Workspaces tree**: lists your workspaces; each expands to its projects;
  each project expands to its detected tasks.
- **Run task**: from the tree or via the `Workend: Run task…` palette
  command. Output streams into a dedicated VS Code output channel.
- **Recent runs view**: most-recent 25 runs across all your projects.
- **`Workend: Open recent run…`** quick-pick to jump to a specific run's
  log in the output channel.

## What it doesn't do (yet)

- Workspace creation / membership management — use the web UI.
- Issue board, branches, PR creation, schedules, notifications — same.
- Authentication — relies on the CLI to bootstrap the session cookie. A
  proper "Sign in to Workend" flow would store an API key per VS Code
  session via `vscode.SecretStorage`.
