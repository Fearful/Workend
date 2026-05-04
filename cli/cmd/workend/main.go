// workend is a small companion CLI for the Workend self-hosted developer
// workstation. It wraps the HTTP API and stores a session cookie locally so
// you can list workspaces, trigger task runs, and tail logs from a terminal.
//
// Auth model: `workend login` posts your credentials to /api/auth/login and
// stores the returned cookie at $XDG_CONFIG_HOME/workend/session (or
// ~/.config/workend/session). All subsequent commands send that cookie.
//
// Usage:
//
//	export WORKEND_API_URL=http://localhost:8080   # default
//	workend login user@host
//	workend workspaces
//	workend projects <workspace-id>
//	workend tasks <project-id>
//	workend run <task-id> [--env KEY=VAL ...] [--arg STRING ...]
//	workend runs                       # recent runs across all projects
//	workend logs <run-id> [--follow]   # tail or full log
//	workend whoami
//	workend logout
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

const sessionCookieName = "workend_session"

type cli struct {
	apiURL string
	client *http.Client
	jar    *cookiejar.Jar
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	c, err := newCLI()
	if err != nil {
		fatal("init: %v", err)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "login":
		c.cmdLogin(args)
	case "logout":
		c.cmdLogout(args)
	case "whoami":
		c.cmdWhoami(args)
	case "workspaces", "ws":
		c.cmdWorkspaces(args)
	case "projects", "p":
		c.cmdProjects(args)
	case "tasks", "t":
		c.cmdTasks(args)
	case "run":
		c.cmdRun(args)
	case "runs":
		c.cmdRuns(args)
	case "logs":
		c.cmdLogs(args)
	case "-h", "--help", "help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", cmd)
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprint(os.Stderr, `workend — companion CLI for Workend (https://github.com/joanlascano/workend)

Usage:
  workend login <email>             prompt for password and create a session
  workend logout                    drop the local session
  workend whoami                    print the current user

  workend workspaces                list workspaces
  workend projects <workspace-id>   list projects in a workspace
  workend tasks <project-id>        list tasks discovered in a project
  workend run <task-id> [opts]      start a task run (and follow logs by default)
      --env KEY=VAL                 (repeatable) inject env var
      --arg VALUE                   (repeatable) append extra arg
      --no-follow                   return immediately with the run id
  workend runs                      list recent runs
  workend logs <run-id> [--follow]  print or tail the log

Environment:
  WORKEND_API_URL                   API base URL (default http://localhost:8080)
`)
}

// ---- HTTP plumbing ----

func newCLI() (*cli, error) {
	apiURL := os.Getenv("WORKEND_API_URL")
	if apiURL == "" {
		apiURL = "http://localhost:8080"
	}
	apiURL = strings.TrimRight(apiURL, "/")

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	c := &cli{
		apiURL: apiURL,
		jar:    jar,
		client: &http.Client{Timeout: 30 * time.Second, Jar: jar},
	}
	if token, err := loadSession(); err == nil && token != "" {
		c.applyCookie(token)
	}
	return c, nil
}

func (c *cli) applyCookie(value string) {
	u, _ := url.Parse(c.apiURL)
	c.jar.SetCookies(u, []*http.Cookie{{
		Name: sessionCookieName, Value: value, Path: "/",
	}})
}

func (c *cli) request(method, path string, body any) (*http.Response, error) {
	var reader io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(buf)
	}
	req, err := http.NewRequest(method, c.apiURL+path, reader)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return c.client.Do(req)
}

func (c *cli) requireOK(method, path string, body any, out any) {
	res, err := c.request(method, path, body)
	if err != nil {
		fatal("request failed: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusUnauthorized {
		fatal("not authenticated — run `workend login`")
	}
	if res.StatusCode >= 400 {
		buf, _ := io.ReadAll(res.Body)
		fatal("HTTP %d: %s", res.StatusCode, strings.TrimSpace(string(buf)))
	}
	if out == nil {
		return
	}
	if err := json.NewDecoder(res.Body).Decode(out); err != nil {
		fatal("decode: %v", err)
	}
}

// ---- session storage ----

func sessionPath() string {
	dir := os.Getenv("XDG_CONFIG_HOME")
	if dir == "" {
		home, _ := os.UserHomeDir()
		dir = filepath.Join(home, ".config")
	}
	return filepath.Join(dir, "workend", "session")
}

func loadSession() (string, error) {
	b, err := os.ReadFile(sessionPath())
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(b)), nil
}

func saveSession(token string) error {
	p := sessionPath()
	if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
		return err
	}
	return os.WriteFile(p, []byte(token), 0o600)
}

func clearSession() error {
	p := sessionPath()
	if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// ---- commands ----

func (c *cli) cmdLogin(args []string) {
	if len(args) < 1 {
		fatal("usage: workend login <email>")
	}
	email := args[0]

	fmt.Fprint(os.Stderr, "Password: ")
	pw, err := readPassword()
	fmt.Fprintln(os.Stderr)
	if err != nil {
		fatal("read password: %v", err)
	}

	res, err := c.request("POST", "/api/auth/login",
		map[string]string{"email": email, "password": pw})
	if err != nil {
		fatal("login: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode >= 400 {
		buf, _ := io.ReadAll(res.Body)
		fatal("login failed (HTTP %d): %s", res.StatusCode, strings.TrimSpace(string(buf)))
	}

	for _, ck := range res.Cookies() {
		if ck.Name == sessionCookieName {
			if err := saveSession(ck.Value); err != nil {
				fatal("save session: %v", err)
			}
			fmt.Println("logged in")
			return
		}
	}
	fatal("server did not return a session cookie")
}

func (c *cli) cmdLogout(_ []string) {
	_, _ = c.request("POST", "/api/auth/logout", nil)
	if err := clearSession(); err != nil {
		fatal("clear session: %v", err)
	}
	fmt.Println("logged out")
}

func (c *cli) cmdWhoami(_ []string) {
	var u struct {
		Email       string `json:"email"`
		DisplayName string `json:"display_name"`
		IsAdmin     bool   `json:"is_admin"`
	}
	c.requireOK("GET", "/api/me", nil, &u)
	suffix := ""
	if u.IsAdmin {
		suffix = " (admin)"
	}
	fmt.Printf("%s <%s>%s\n", u.DisplayName, u.Email, suffix)
}

func (c *cli) cmdWorkspaces(_ []string) {
	var ws []struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	c.requireOK("GET", "/api/workspaces", nil, &ws)
	if len(ws) == 0 {
		fmt.Println("no workspaces")
		return
	}
	for _, w := range ws {
		fmt.Printf("%s  %s\n", w.ID, w.Name)
	}
}

func (c *cli) cmdProjects(args []string) {
	if len(args) < 1 {
		fatal("usage: workend projects <workspace-id>")
	}
	var projs []struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		Status string `json:"status"`
	}
	c.requireOK("GET", "/api/workspaces/"+args[0]+"/projects", nil, &projs)
	if len(projs) == 0 {
		fmt.Println("no projects")
		return
	}
	for _, p := range projs {
		fmt.Printf("%s  [%s]  %s\n", p.ID, p.Status, p.Name)
	}
}

func (c *cli) cmdTasks(args []string) {
	if len(args) < 1 {
		fatal("usage: workend tasks <project-id>")
	}
	var tasks []struct {
		ID         string `json:"id"`
		Source     string `json:"source"`
		Name       string `json:"name"`
		RawCommand string `json:"raw_command"`
	}
	c.requireOK("GET", "/api/projects/"+args[0]+"/tasks", nil, &tasks)
	if len(tasks) == 0 {
		fmt.Println("no tasks detected")
		return
	}
	for _, t := range tasks {
		fmt.Printf("%s  [%s]  %s\n    %s\n", t.ID, t.Source, t.Name, t.RawCommand)
	}
}

type runFlags struct {
	envs     stringSlice
	args     stringSlice
	noFollow bool
}

func (c *cli) cmdRun(argv []string) {
	if len(argv) < 1 {
		fatal("usage: workend run <task-id> [--env KEY=VAL] [--arg VALUE] [--no-follow]")
	}
	taskID := argv[0]
	rest := argv[1:]

	fs := flag.NewFlagSet("run", flag.ExitOnError)
	rf := runFlags{}
	fs.Var(&rf.envs, "env", "env var KEY=VAL (repeatable)")
	fs.Var(&rf.args, "arg", "extra arg (repeatable)")
	fs.BoolVar(&rf.noFollow, "no-follow", false, "do not tail logs after start")
	_ = fs.Parse(rest)

	envMap := map[string]string{}
	for _, kv := range rf.envs {
		eq := strings.Index(kv, "=")
		if eq < 1 {
			fatal("bad --env %q (expected KEY=VAL)", kv)
		}
		envMap[kv[:eq]] = kv[eq+1:]
	}

	body := map[string]any{}
	if len(envMap) > 0 {
		body["env"] = envMap
	}
	if len(rf.args) > 0 {
		body["args"] = []string(rf.args)
	}

	var run struct {
		ID string `json:"id"`
	}
	c.requireOK("POST", "/api/tasks/"+taskID+"/runs", body, &run)
	if run.ID == "" {
		fatal("server returned no run id")
	}
	fmt.Fprintf(os.Stderr, "started run %s\n", run.ID)
	if rf.noFollow {
		fmt.Println(run.ID)
		return
	}
	c.followLog(run.ID)
}

func (c *cli) cmdRuns(_ []string) {
	var runs []struct {
		ID            string  `json:"id"`
		Status        string  `json:"status"`
		ExitCode      *int    `json:"exit_code"`
		TaskName      string  `json:"task_name"`
		TaskSource    string  `json:"task_source"`
		ProjectName   string  `json:"project_name"`
		WorkspaceName string  `json:"workspace_name"`
		CreatedAt     string  `json:"created_at"`
		TimedOut      bool    `json:"timed_out"`
	}
	c.requireOK("GET", "/api/me/runs", nil, &runs)
	for _, r := range runs {
		exit := "—"
		if r.ExitCode != nil {
			exit = fmt.Sprintf("exit %d", *r.ExitCode)
		}
		extra := ""
		if r.TimedOut {
			extra = " (timeout)"
		}
		fmt.Printf("%s  %-9s%s  %s  %s/%s  %s [%s]  %s\n",
			r.ID, r.Status, extra, exit, r.WorkspaceName, r.ProjectName,
			r.TaskName, r.TaskSource, r.CreatedAt)
	}
}

func (c *cli) cmdLogs(argv []string) {
	if len(argv) < 1 {
		fatal("usage: workend logs <run-id> [--follow]")
	}
	runID := argv[0]
	follow := false
	for _, a := range argv[1:] {
		if a == "--follow" || a == "-f" {
			follow = true
		}
	}
	if follow {
		c.followLog(runID)
		return
	}
	res, err := c.request("GET", "/api/runs/"+runID+"/log", nil)
	if err != nil {
		fatal("logs: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode >= 400 {
		fatal("logs HTTP %d", res.StatusCode)
	}
	_, _ = io.Copy(os.Stdout, res.Body)
}

// followLog opens the SSE stream and prints `log` events as they arrive,
// then exits cleanly on the `done` event. Falls back to a static fetch if
// the stream errors out before completion.
func (c *cli) followLog(runID string) {
	req, err := http.NewRequest("GET", c.apiURL+"/api/runs/"+runID+"/log/stream", nil)
	if err != nil {
		fatal("stream: %v", err)
	}
	streamClient := &http.Client{Timeout: 0, Jar: c.jar}
	res, err := streamClient.Do(req)
	if err != nil {
		fatal("stream: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode >= 400 {
		fatal("stream HTTP %d", res.StatusCode)
	}

	scanner := bufio.NewScanner(res.Body)
	scanner.Buffer(make([]byte, 64*1024), 4*1024*1024)
	var (
		event string
		data  strings.Builder
	)
	flush := func() {
		switch event {
		case "log":
			fmt.Print(data.String())
			if !strings.HasSuffix(data.String(), "\n") {
				fmt.Println()
			}
		case "done":
			fmt.Fprintln(os.Stderr, "\n=== "+strings.TrimSpace(data.String()))
		}
		event = ""
		data.Reset()
	}
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case line == "":
			flush()
		case strings.HasPrefix(line, "event:"):
			event = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
		case strings.HasPrefix(line, "data:"):
			payload := strings.TrimPrefix(line, "data:")
			if strings.HasPrefix(payload, " ") {
				payload = payload[1:]
			}
			if data.Len() > 0 {
				data.WriteString("\n")
			}
			data.WriteString(payload)
		}
		if event == "done" && data.Len() > 0 {
			flush()
			return
		}
	}
	flush()
}

// ---- helpers ----

type stringSlice []string

func (s *stringSlice) String() string     { return strings.Join(*s, ",") }
func (s *stringSlice) Set(v string) error { *s = append(*s, v); return nil }

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

// readPassword reads a line from stdin without echoing. Falls back to
// echoed line-read if the terminal can't be put into raw mode (e.g. when
// stdin is a pipe — for scripts).
func readPassword() (string, error) {
	fd := int(syscall.Stdin)
	if !isTerminal(fd) {
		r := bufio.NewReader(os.Stdin)
		s, err := r.ReadString('\n')
		return strings.TrimRight(s, "\r\n"), err
	}
	pw, err := readSilent(fd)
	return string(pw), err
}
