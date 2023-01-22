<script lang="ts">
	import type { PageData } from './$types';
	import type { Item } from '$lib/Checkbox.svelte';
	import Checkbox from '$lib/Checkbox.svelte';
	import { addItem, removeItem, checkItem } from '$lib/api';
	import { invalidateAll } from '$app/navigation';

	export let data: PageData;
	let newItemName = '';

	$: items = data.list.items.sort((a: Item, b: Item) => a.name.localeCompare(b.name));
	$: unchecked = items.filter((i: Item) => !i.checked);
	$: checked = items.filter((i: Item) => i.checked);

	const check = (idx: number) => {
		const targetID = unchecked[idx].id;
		const targetIdx = data.list.items.findIndex((i: Item) => i.id == targetID);
		data.list.items[targetIdx].checked = true;
		checkItem(targetID, true).then(invalidateAll);
	};
	const uncheck = (idx: number) => {
		const targetID = checked[idx].id;
		const targetIdx = data.list.items.findIndex((i: Item) => i.id == targetID);
		data.list.items[targetIdx].checked = false;
		checkItem(targetID, false).then(invalidateAll);
	};
	const addNewItem = () => {
		addItem(newItemName).then(() => {
			newItemName = '';
			return invalidateAll();
		});
	};
	const removeExistingItem = (id: string) => {
		removeItem(id).then(invalidateAll);
	};
</script>

<header>
	<div class="mx-auto px-4 pt-2">
		<h1 class="text-2xl font-bold">
			(Br)AnyList - {data.list.name}
		</h1>
	</div>
</header>
<form class="m-4" on:submit|preventDefault={addNewItem}>
	<input
		type="text"
		placeholder="Add item"
		class="block w-full px-3 py-1.5 text-base font-normal text-gray-700 bg-white border border-solid border-gray-300 rounded focus:text-gray-700 focus:bg-white focus:border-blue-600 focus:outline-none"
		bind:value={newItemName}
	/>
</form>
<div>
	{#each unchecked as item, index}
		<Checkbox
			{item}
			on:checked={() => check(index)}
			on:removed={() => removeExistingItem(item.id)}
		/>
	{/each}
</div>
<hr class="my-4 border-1 border-black w-1/3 mx-auto" />
<div>
	{#each checked as item, index}
		<Checkbox
			{item}
			on:checked={() => uncheck(index)}
			on:removed={() => removeExistingItem(item.id)}
		/>
	{/each}
</div>
