# Workend

A self-hosted, browser-based control surface for working across many code repositories: discover what each project can do, run those tasks reproducibly in isolated containers, and watch them happen live — without ever shelling into the host.

---

## Status: rebuild in progress

Workend has two lives in this repository.

### v2 — active development → see [`v2/`](v2/)

The current direction. SvelteKit + Go + Postgres + Dagger, orchestrated with Docker Compose.

Start with:
- [`v2/OBJECTIVE.md`](v2/OBJECTIVE.md) — what Workend is and isn't
- [`v2/PLAN.md`](v2/PLAN.md) — staged build plan, MVP through Stage 14+
- [`v2/ARCHITECTURE.md`](v2/ARCHITECTURE.md) — technical design (filled in per stage)

### v1 — archived (2015)

A Node.js + AngularJS prototype with the same concept but a fundamentally unsafe execution model (`child_process.exec` against host filesystem) and a stack that is now end-of-life across the board (AngularJS, Bower, Jade, Mongoose 4, Gulp 3).

The v1 source lives in `backend/`, `frontend/`, `server.js`, `bower.json`, etc. — preserved for historical reference. Tagged as `archive/v0.0.1-2015`.

```
git checkout archive/v0.0.1-2015
```

The v1 README pitch is preserved below for context.

---

## Original (2015) pitch

> The idea of Workend comes from many hours spent looking at command prompts and over functional software. Merging, testing, cloning, etc. I'm Joan Lascano, Front-end Developer at Making Sense, and I love the web in all its glory.
>
> The web 2.0 as we used to know it back in 2004 has changed, tools like nodejs, karma, bower and others have become a common base ground for many development projects. Working with new technologies can be a little pain, having not only to deal with the git tools but also node, and every npm that exposes methods to the command prompt.
>
> Workend aims to fill the gap between developers and the command prompt. It offers a simple UI where you can open/create/clone projects, you will be able to see your project statistics including test coverage, executable tasks available from gulpjs or grunt, news from github and manage your project status directly from the browser.
