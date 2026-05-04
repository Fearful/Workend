
// this file is generated — do not edit it


declare module "svelte/elements" {
	export interface HTMLAttributes<T> {
		'data-sveltekit-keepfocus'?: true | '' | 'off' | undefined | null;
		'data-sveltekit-noscroll'?: true | '' | 'off' | undefined | null;
		'data-sveltekit-preload-code'?:
			| true
			| ''
			| 'eager'
			| 'viewport'
			| 'hover'
			| 'tap'
			| 'off'
			| undefined
			| null;
		'data-sveltekit-preload-data'?: true | '' | 'hover' | 'tap' | 'off' | undefined | null;
		'data-sveltekit-reload'?: true | '' | 'off' | undefined | null;
		'data-sveltekit-replacestate'?: true | '' | 'off' | undefined | null;
	}
}

export {};


declare module "$app/types" {
	type MatcherParam<M> = M extends (param : string) => param is (infer U extends string) ? U : string;

	export interface AppTypes {
		RouteId(): "/" | "/admin" | "/auth" | "/auth/[provider]" | "/auth/[provider]/callback" | "/auth/[provider]/start" | "/dashboard" | "/dockscope" | "/issues" | "/issues/[id]" | "/issues/[id]/move" | "/login" | "/logout" | "/mentions" | "/otel" | "/pipeline-runs" | "/pipeline-runs/[id]" | "/projects" | "/projects/[id]" | "/projects/[id]/board" | "/projects/[id]/branches" | "/projects/[id]/images" | "/projects/[id]/pipelines" | "/projects/[id]/runs" | "/projects/[id]/schedules" | "/projects/[id]/trends" | "/runs" | "/runs/[id]" | "/runs/[id]/cancel" | "/runs/[id]/compare" | "/runs/[id]/compare/[other]" | "/runs/[id]/log-stream" | "/settings" | "/settings/notifications" | "/signup" | "/workspaces" | "/workspaces/new" | "/workspaces/[id]" | "/workspaces/[id]/dashboard" | "/workspaces/[id]/projects" | "/workspaces/[id]/projects/new";
		RouteParams(): {
			"/auth/[provider]": { provider: string };
			"/auth/[provider]/callback": { provider: string };
			"/auth/[provider]/start": { provider: string };
			"/issues/[id]": { id: string };
			"/issues/[id]/move": { id: string };
			"/pipeline-runs/[id]": { id: string };
			"/projects/[id]": { id: string };
			"/projects/[id]/board": { id: string };
			"/projects/[id]/branches": { id: string };
			"/projects/[id]/images": { id: string };
			"/projects/[id]/pipelines": { id: string };
			"/projects/[id]/runs": { id: string };
			"/projects/[id]/schedules": { id: string };
			"/projects/[id]/trends": { id: string };
			"/runs/[id]": { id: string };
			"/runs/[id]/cancel": { id: string };
			"/runs/[id]/compare": { id: string };
			"/runs/[id]/compare/[other]": { id: string; other: string };
			"/runs/[id]/log-stream": { id: string };
			"/workspaces/[id]": { id: string };
			"/workspaces/[id]/dashboard": { id: string };
			"/workspaces/[id]/projects": { id: string };
			"/workspaces/[id]/projects/new": { id: string }
		};
		LayoutParams(): {
			"/": { provider?: string; id?: string; other?: string };
			"/admin": Record<string, never>;
			"/auth": { provider?: string };
			"/auth/[provider]": { provider: string };
			"/auth/[provider]/callback": { provider: string };
			"/auth/[provider]/start": { provider: string };
			"/dashboard": Record<string, never>;
			"/dockscope": Record<string, never>;
			"/issues": { id?: string };
			"/issues/[id]": { id: string };
			"/issues/[id]/move": { id: string };
			"/login": Record<string, never>;
			"/logout": Record<string, never>;
			"/mentions": Record<string, never>;
			"/otel": Record<string, never>;
			"/pipeline-runs": { id?: string };
			"/pipeline-runs/[id]": { id: string };
			"/projects": { id?: string };
			"/projects/[id]": { id: string };
			"/projects/[id]/board": { id: string };
			"/projects/[id]/branches": { id: string };
			"/projects/[id]/images": { id: string };
			"/projects/[id]/pipelines": { id: string };
			"/projects/[id]/runs": { id: string };
			"/projects/[id]/schedules": { id: string };
			"/projects/[id]/trends": { id: string };
			"/runs": { id?: string; other?: string };
			"/runs/[id]": { id: string; other?: string };
			"/runs/[id]/cancel": { id: string };
			"/runs/[id]/compare": { id: string; other?: string };
			"/runs/[id]/compare/[other]": { id: string; other: string };
			"/runs/[id]/log-stream": { id: string };
			"/settings": Record<string, never>;
			"/settings/notifications": Record<string, never>;
			"/signup": Record<string, never>;
			"/workspaces": { id?: string };
			"/workspaces/new": Record<string, never>;
			"/workspaces/[id]": { id: string };
			"/workspaces/[id]/dashboard": { id: string };
			"/workspaces/[id]/projects": { id: string };
			"/workspaces/[id]/projects/new": { id: string }
		};
		Pathname(): "/" | "/admin" | `/auth/${string}/callback` & {} | `/auth/${string}/start` & {} | "/dashboard" | "/dockscope" | `/issues/${string}` & {} | `/issues/${string}/move` & {} | "/login" | "/logout" | "/mentions" | "/otel" | `/pipeline-runs/${string}` & {} | `/projects/${string}` & {} | `/projects/${string}/board` & {} | `/projects/${string}/branches` & {} | `/projects/${string}/images` & {} | `/projects/${string}/pipelines` & {} | `/projects/${string}/runs` & {} | `/projects/${string}/schedules` & {} | `/projects/${string}/trends` & {} | `/runs/${string}` & {} | `/runs/${string}/cancel` & {} | `/runs/${string}/compare/${string}` & {} | `/runs/${string}/log-stream` & {} | "/settings" | "/settings/notifications" | "/signup" | "/workspaces/new" | `/workspaces/${string}` & {} | `/workspaces/${string}/dashboard` & {} | `/workspaces/${string}/projects/new` & {};
		ResolvedPathname(): `${"" | `/${string}`}${ReturnType<AppTypes['Pathname']>}`;
		Asset(): string & {};
	}
}