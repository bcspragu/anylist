/** @type {import('./$types').PageLoad} */
export async function load() {
	const res = await fetch('http://localhost:8080/api/list');
	const item = await res.json();
	return {
		list: item
	};
}
