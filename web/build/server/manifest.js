const manifest = (() => {
function __memo(fn) {
	let value;
	return () => value ??= (value = fn());
}

return {
	appDir: "_app",
	appPath: "_app",
	assets: new Set([]),
	mimeTypes: {},
	_: {
		client: {start:"_app/immutable/entry/start.BNwKEhlV.js",app:"_app/immutable/entry/app.CBzzZ0mZ.js",imports:["_app/immutable/entry/start.BNwKEhlV.js","_app/immutable/chunks/DuSA90i8.js","_app/immutable/chunks/Di6_auBm.js","_app/immutable/chunks/BCGREYIq.js","_app/immutable/chunks/DdjgLl7P.js","_app/immutable/entry/app.CBzzZ0mZ.js","_app/immutable/chunks/BCGREYIq.js","_app/immutable/chunks/CWj6FrbW.js","_app/immutable/chunks/Di6_auBm.js","_app/immutable/chunks/glOr0qzD.js","_app/immutable/chunks/B2UVO_nM.js","_app/immutable/chunks/Bq1FU2Pv.js"],stylesheets:[],fonts:[],uses_env_dynamic_public:false},
		nodes: [
			__memo(() => import('./chunks/0-Ch1s4ZcT.js')),
			__memo(() => import('./chunks/1-L9ulOWMJ.js')),
			__memo(() => import('./chunks/2-CZ6qpr9O.js')),
			__memo(() => import('./chunks/3-bRwly-vz.js')),
			__memo(() => import('./chunks/4-J4qREDOa.js')),
			__memo(() => import('./chunks/5-DXvkK2li.js')),
			__memo(() => import('./chunks/6-BcHfcARt.js')),
			__memo(() => import('./chunks/7-Nvbl-Dn-.js')),
			__memo(() => import('./chunks/8-D8kdYSQ2.js')),
			__memo(() => import('./chunks/9-D5OsyEnf.js')),
			__memo(() => import('./chunks/10-BA4jblvm.js')),
			__memo(() => import('./chunks/11-CAE9rael.js')),
			__memo(() => import('./chunks/12-Y5VafT8c.js')),
			__memo(() => import('./chunks/13-zTsGl1Qz.js')),
			__memo(() => import('./chunks/14-CLP2_SfO.js')),
			__memo(() => import('./chunks/15-BNerK48_.js')),
			__memo(() => import('./chunks/16-MFky7oQO.js')),
			__memo(() => import('./chunks/17-DKoaDdN8.js')),
			__memo(() => import('./chunks/18-CmGY_SPU.js')),
			__memo(() => import('./chunks/19-278eXfBm.js')),
			__memo(() => import('./chunks/20-Dk4Qb6as.js')),
			__memo(() => import('./chunks/21-ef2cMoIL.js')),
			__memo(() => import('./chunks/22-CrEgDYz1.js')),
			__memo(() => import('./chunks/23-Ch_rqZYn.js')),
			__memo(() => import('./chunks/24-N43jGRf8.js')),
			__memo(() => import('./chunks/25-fy7zB2mu.js')),
			__memo(() => import('./chunks/26-CGlVJ-JJ.js')),
			__memo(() => import('./chunks/27-BdgfH1NX.js')),
			__memo(() => import('./chunks/28-Dopu6jW-.js')),
			__memo(() => import('./chunks/29-CYcsHvP2.js')),
			__memo(() => import('./chunks/30-Bv9_3z_o.js')),
			__memo(() => import('./chunks/31-CPvAAdKj.js')),
			__memo(() => import('./chunks/32-28IPNEKV.js')),
			__memo(() => import('./chunks/33-Cbck09EM.js')),
			__memo(() => import('./chunks/34-4XSUjw8h.js')),
			__memo(() => import('./chunks/35-D_rAatFa.js')),
			__memo(() => import('./chunks/36-CCqnCHad.js')),
			__memo(() => import('./chunks/37-B8k5x55x.js'))
		],
		remotes: {
			
		},
		routes: [
			{
				id: "/",
				pattern: /^\/$/,
				params: [],
				page: { layouts: [0,], errors: [1,], leaf: 3 },
				endpoint: null
			},
			{
				id: "/admin",
				pattern: /^\/admin\/?$/,
				params: [],
				page: { layouts: [0,], errors: [1,], leaf: 4 },
				endpoint: null
			},
			{
				id: "/auth/[provider]/callback",
				pattern: /^\/auth\/([^/]+?)\/callback\/?$/,
				params: [{"name":"provider","optional":false,"rest":false,"chained":false}],
				page: null,
				endpoint: __memo(() => import('./chunks/_server.ts-C3_o-U-b.js'))
			},
			{
				id: "/auth/[provider]/start",
				pattern: /^\/auth\/([^/]+?)\/start\/?$/,
				params: [{"name":"provider","optional":false,"rest":false,"chained":false}],
				page: null,
				endpoint: __memo(() => import('./chunks/_server.ts-BLbXDfJn.js'))
			},
			{
				id: "/dashboard",
				pattern: /^\/dashboard\/?$/,
				params: [],
				page: { layouts: [0,], errors: [1,], leaf: 5 },
				endpoint: null
			},
			{
				id: "/dockscope",
				pattern: /^\/dockscope\/?$/,
				params: [],
				page: { layouts: [0,], errors: [1,], leaf: 6 },
				endpoint: null
			},
			{
				id: "/issues/[id]",
				pattern: /^\/issues\/([^/]+?)\/?$/,
				params: [{"name":"id","optional":false,"rest":false,"chained":false}],
				page: { layouts: [0,], errors: [1,], leaf: 7 },
				endpoint: null
			},
			{
				id: "/issues/[id]/move",
				pattern: /^\/issues\/([^/]+?)\/move\/?$/,
				params: [{"name":"id","optional":false,"rest":false,"chained":false}],
				page: null,
				endpoint: __memo(() => import('./chunks/_server.ts-B93fzQ2U.js'))
			},
			{
				id: "/login",
				pattern: /^\/login\/?$/,
				params: [],
				page: { layouts: [0,], errors: [1,], leaf: 8 },
				endpoint: null
			},
			{
				id: "/logout",
				pattern: /^\/logout\/?$/,
				params: [],
				page: { layouts: [0,], errors: [1,], leaf: 9 },
				endpoint: null
			},
			{
				id: "/mentions",
				pattern: /^\/mentions\/?$/,
				params: [],
				page: { layouts: [0,], errors: [1,], leaf: 10 },
				endpoint: null
			},
			{
				id: "/otel",
				pattern: /^\/otel\/?$/,
				params: [],
				page: { layouts: [0,], errors: [1,], leaf: 11 },
				endpoint: null
			},
			{
				id: "/pipeline-runs/[id]",
				pattern: /^\/pipeline-runs\/([^/]+?)\/?$/,
				params: [{"name":"id","optional":false,"rest":false,"chained":false}],
				page: { layouts: [0,], errors: [1,], leaf: 12 },
				endpoint: null
			},
			{
				id: "/projects/[id]",
				pattern: /^\/projects\/([^/]+?)\/?$/,
				params: [{"name":"id","optional":false,"rest":false,"chained":false}],
				page: { layouts: [0,2,], errors: [1,,], leaf: 13 },
				endpoint: null
			},
			{
				id: "/projects/[id]/blame",
				pattern: /^\/projects\/([^/]+?)\/blame\/?$/,
				params: [{"name":"id","optional":false,"rest":false,"chained":false}],
				page: { layouts: [0,2,], errors: [1,,], leaf: 14 },
				endpoint: null
			},
			{
				id: "/projects/[id]/board",
				pattern: /^\/projects\/([^/]+?)\/board\/?$/,
				params: [{"name":"id","optional":false,"rest":false,"chained":false}],
				page: { layouts: [0,2,], errors: [1,,], leaf: 15 },
				endpoint: null
			},
			{
				id: "/projects/[id]/branches",
				pattern: /^\/projects\/([^/]+?)\/branches\/?$/,
				params: [{"name":"id","optional":false,"rest":false,"chained":false}],
				page: { layouts: [0,2,], errors: [1,,], leaf: 16 },
				endpoint: null
			},
			{
				id: "/projects/[id]/images",
				pattern: /^\/projects\/([^/]+?)\/images\/?$/,
				params: [{"name":"id","optional":false,"rest":false,"chained":false}],
				page: { layouts: [0,2,], errors: [1,,], leaf: 17 },
				endpoint: null
			},
			{
				id: "/projects/[id]/monorepo",
				pattern: /^\/projects\/([^/]+?)\/monorepo\/?$/,
				params: [{"name":"id","optional":false,"rest":false,"chained":false}],
				page: { layouts: [0,2,], errors: [1,,], leaf: 18 },
				endpoint: null
			},
			{
				id: "/projects/[id]/pipelines",
				pattern: /^\/projects\/([^/]+?)\/pipelines\/?$/,
				params: [{"name":"id","optional":false,"rest":false,"chained":false}],
				page: { layouts: [0,2,], errors: [1,,], leaf: 19 },
				endpoint: null
			},
			{
				id: "/projects/[id]/previews",
				pattern: /^\/projects\/([^/]+?)\/previews\/?$/,
				params: [{"name":"id","optional":false,"rest":false,"chained":false}],
				page: { layouts: [0,2,], errors: [1,,], leaf: 20 },
				endpoint: null
			},
			{
				id: "/projects/[id]/runs",
				pattern: /^\/projects\/([^/]+?)\/runs\/?$/,
				params: [{"name":"id","optional":false,"rest":false,"chained":false}],
				page: { layouts: [0,2,], errors: [1,,], leaf: 21 },
				endpoint: null
			},
			{
				id: "/projects/[id]/schedules",
				pattern: /^\/projects\/([^/]+?)\/schedules\/?$/,
				params: [{"name":"id","optional":false,"rest":false,"chained":false}],
				page: { layouts: [0,2,], errors: [1,,], leaf: 22 },
				endpoint: null
			},
			{
				id: "/projects/[id]/trends",
				pattern: /^\/projects\/([^/]+?)\/trends\/?$/,
				params: [{"name":"id","optional":false,"rest":false,"chained":false}],
				page: { layouts: [0,2,], errors: [1,,], leaf: 23 },
				endpoint: null
			},
			{
				id: "/runs/[id]",
				pattern: /^\/runs\/([^/]+?)\/?$/,
				params: [{"name":"id","optional":false,"rest":false,"chained":false}],
				page: { layouts: [0,], errors: [1,], leaf: 24 },
				endpoint: null
			},
			{
				id: "/runs/[id]/cancel",
				pattern: /^\/runs\/([^/]+?)\/cancel\/?$/,
				params: [{"name":"id","optional":false,"rest":false,"chained":false}],
				page: null,
				endpoint: __memo(() => import('./chunks/_server.ts-DVQY26Di.js'))
			},
			{
				id: "/runs/[id]/compare/[other]",
				pattern: /^\/runs\/([^/]+?)\/compare\/([^/]+?)\/?$/,
				params: [{"name":"id","optional":false,"rest":false,"chained":false},{"name":"other","optional":false,"rest":false,"chained":false}],
				page: { layouts: [0,], errors: [1,], leaf: 25 },
				endpoint: null
			},
			{
				id: "/runs/[id]/log-stream",
				pattern: /^\/runs\/([^/]+?)\/log-stream\/?$/,
				params: [{"name":"id","optional":false,"rest":false,"chained":false}],
				page: null,
				endpoint: __memo(() => import('./chunks/_server.ts-CY-KrSZM.js'))
			},
			{
				id: "/sandboxes",
				pattern: /^\/sandboxes\/?$/,
				params: [],
				page: { layouts: [0,], errors: [1,], leaf: 26 },
				endpoint: null
			},
			{
				id: "/sandboxes/[id]",
				pattern: /^\/sandboxes\/([^/]+?)\/?$/,
				params: [{"name":"id","optional":false,"rest":false,"chained":false}],
				page: { layouts: [0,], errors: [1,], leaf: 27 },
				endpoint: null
			},
			{
				id: "/settings",
				pattern: /^\/settings\/?$/,
				params: [],
				page: { layouts: [0,], errors: [1,], leaf: 28 },
				endpoint: null
			},
			{
				id: "/settings/notifications",
				pattern: /^\/settings\/notifications\/?$/,
				params: [],
				page: { layouts: [0,], errors: [1,], leaf: 29 },
				endpoint: null
			},
			{
				id: "/signup",
				pattern: /^\/signup\/?$/,
				params: [],
				page: { layouts: [0,], errors: [1,], leaf: 30 },
				endpoint: null
			},
			{
				id: "/workspaces/new",
				pattern: /^\/workspaces\/new\/?$/,
				params: [],
				page: { layouts: [0,], errors: [1,], leaf: 37 },
				endpoint: null
			},
			{
				id: "/workspaces/[id]",
				pattern: /^\/workspaces\/([^/]+?)\/?$/,
				params: [{"name":"id","optional":false,"rest":false,"chained":false}],
				page: { layouts: [0,], errors: [1,], leaf: 31 },
				endpoint: null
			},
			{
				id: "/workspaces/[id]/activity",
				pattern: /^\/workspaces\/([^/]+?)\/activity\/?$/,
				params: [{"name":"id","optional":false,"rest":false,"chained":false}],
				page: { layouts: [0,], errors: [1,], leaf: 32 },
				endpoint: null
			},
			{
				id: "/workspaces/[id]/dashboard",
				pattern: /^\/workspaces\/([^/]+?)\/dashboard\/?$/,
				params: [{"name":"id","optional":false,"rest":false,"chained":false}],
				page: { layouts: [0,], errors: [1,], leaf: 33 },
				endpoint: null
			},
			{
				id: "/workspaces/[id]/projects/new",
				pattern: /^\/workspaces\/([^/]+?)\/projects\/new\/?$/,
				params: [{"name":"id","optional":false,"rest":false,"chained":false}],
				page: { layouts: [0,], errors: [1,], leaf: 34 },
				endpoint: null
			},
			{
				id: "/workspaces/[id]/roles",
				pattern: /^\/workspaces\/([^/]+?)\/roles\/?$/,
				params: [{"name":"id","optional":false,"rest":false,"chained":false}],
				page: { layouts: [0,], errors: [1,], leaf: 35 },
				endpoint: null
			},
			{
				id: "/workspaces/[id]/secrets",
				pattern: /^\/workspaces\/([^/]+?)\/secrets\/?$/,
				params: [{"name":"id","optional":false,"rest":false,"chained":false}],
				page: { layouts: [0,], errors: [1,], leaf: 36 },
				endpoint: null
			}
		],
		prerendered_routes: new Set([]),
		matchers: async () => {
			
			return {  };
		},
		server_assets: {}
	}
}
})();

const prerendered = new Set([]);

const base = "";

export { base, manifest, prerendered };
//# sourceMappingURL=manifest.js.map
