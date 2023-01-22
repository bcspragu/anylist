<script context="module" lang="ts">
	export interface Item {
		id: string;
		name: string;
		details: string;
		checked: boolean;
	}
</script>

<script lang="ts">
	import { createEventDispatcher } from 'svelte';

	interface Emit {
		checked: {};
		removed: {};
	}
	const dispatch = createEventDispatcher<Emit>();

	export let item: Item;
</script>

<div
	class="rounded-md px-3 py-2 border-2 border-black mb-2 mx-3 flex items-center"
	class:opacity-25={item.checked}
>
	<div class="mr-3">
		<input
			type="checkbox"
			class="h-6 w-6 rounded-md border-gray-200 bg-white shadow-sm mt-2"
			bind:checked={item.checked}
			on:click|preventDefault={() => dispatch('checked')}
		/>
	</div>
	<div class="flex-1">
		<div class="text-gray-500 sm:pr-8">
			<h1 class="text-xl font-bold text-gray-900">{item.name}</h1>

			{#if item.details}
				<p class="mt-2 text-sm sm:block">{item.details}</p>
			{/if}
		</div>
	</div>
	<div>
		<img
			class="w-5 cursor-pointer"
			src="trash.svg"
			alt="Remove item"
			on:click={() => dispatch('removed')}
			on:keypress={() => dispatch('removed')}
		/>
	</div>
</div>
