export { matchers } from './matchers.js';

export const nodes = [
	() => import('./nodes/0'),
	() => import('./nodes/1'),
	() => import('./nodes/2'),
	() => import('./nodes/3'),
	() => import('./nodes/4'),
	() => import('./nodes/5'),
	() => import('./nodes/6'),
	() => import('./nodes/7'),
	() => import('./nodes/8'),
	() => import('./nodes/9'),
	() => import('./nodes/10'),
	() => import('./nodes/11'),
	() => import('./nodes/12'),
	() => import('./nodes/13'),
	() => import('./nodes/14'),
	() => import('./nodes/15'),
	() => import('./nodes/16'),
	() => import('./nodes/17'),
	() => import('./nodes/18'),
	() => import('./nodes/19'),
	() => import('./nodes/20'),
	() => import('./nodes/21'),
	() => import('./nodes/22'),
	() => import('./nodes/23'),
	() => import('./nodes/24'),
	() => import('./nodes/25'),
	() => import('./nodes/26'),
	() => import('./nodes/27'),
	() => import('./nodes/28'),
	() => import('./nodes/29'),
	() => import('./nodes/30'),
	() => import('./nodes/31'),
	() => import('./nodes/32'),
	() => import('./nodes/33'),
	() => import('./nodes/34'),
	() => import('./nodes/35'),
	() => import('./nodes/36'),
	() => import('./nodes/37')
];

export const server_loads = [0,2];

export const dictionary = {
		"/": [~3],
		"/admin": [~4],
		"/dashboard": [~5],
		"/dockscope": [6],
		"/issues/[id]": [~7],
		"/login": [~8],
		"/logout": [9],
		"/mentions": [~10],
		"/otel": [11],
		"/pipeline-runs/[id]": [~12],
		"/projects/[id]": [~13,[2]],
		"/projects/[id]/blame": [~14,[2]],
		"/projects/[id]/board": [~15,[2]],
		"/projects/[id]/branches": [~16,[2]],
		"/projects/[id]/images": [~17,[2]],
		"/projects/[id]/monorepo": [~18,[2]],
		"/projects/[id]/pipelines": [~19,[2]],
		"/projects/[id]/previews": [~20,[2]],
		"/projects/[id]/runs": [~21,[2]],
		"/projects/[id]/schedules": [~22,[2]],
		"/projects/[id]/trends": [~23,[2]],
		"/runs/[id]": [~24],
		"/runs/[id]/compare/[other]": [~25],
		"/sandboxes": [~26],
		"/sandboxes/[id]": [~27],
		"/settings": [~28],
		"/settings/notifications": [~29],
		"/signup": [~30],
		"/workspaces/new": [~37],
		"/workspaces/[id]": [~31],
		"/workspaces/[id]/activity": [~32],
		"/workspaces/[id]/dashboard": [~33],
		"/workspaces/[id]/projects/new": [~34],
		"/workspaces/[id]/roles": [~35],
		"/workspaces/[id]/secrets": [~36]
	};

export const hooks = {
	handleError: (({ error }) => { console.error(error) }),
	
	reroute: (() => {}),
	transport: {}
};

export const decoders = Object.fromEntries(Object.entries(hooks.transport).map(([k, v]) => [k, v.decode]));
export const encoders = Object.fromEntries(Object.entries(hooks.transport).map(([k, v]) => [k, v.encode]));

export const hash = false;

export const decode = (type, value) => decoders[type](value);

export { default as root } from '../root.js';