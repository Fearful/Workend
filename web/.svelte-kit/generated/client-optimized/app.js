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
	() => import('./nodes/26')
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
		"/projects/[id]": [~12,[2]],
		"/projects/[id]/board": [~13,[2]],
		"/projects/[id]/branches": [~14,[2]],
		"/projects/[id]/images": [~15,[2]],
		"/projects/[id]/runs": [~16,[2]],
		"/projects/[id]/schedules": [~17,[2]],
		"/projects/[id]/trends": [~18,[2]],
		"/runs/[id]": [~19],
		"/runs/[id]/compare/[other]": [~20],
		"/settings": [~21],
		"/settings/notifications": [~22],
		"/signup": [~23],
		"/workspaces/new": [~26],
		"/workspaces/[id]": [~24],
		"/workspaces/[id]/projects/new": [~25]
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