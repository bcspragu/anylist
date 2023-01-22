import { PUBLIC_API_BASE_URL } from '$env/static/public';
import { browser } from '$app/environment';

/** @type {import('./$types').PageLoad} */
export async function load() {
	const baseURL = browser ? '' : PUBLIC_API_BASE_URL;
	const res = await fetch(`${baseURL}/api/list`);
	const item = await res.json();
	return {
		list: item
	};
}
