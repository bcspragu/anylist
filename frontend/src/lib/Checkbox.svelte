<script lang="ts">
	interface Item {
		id: string;
		name: string;
		details: string;
		checked: boolean;
	}
	export let item: Item;
</script>

<div
	class="rounded-xl p-3 border-2 border-black mb-2 mx-3 flex items-center"
	class:opacity-25={item.checked}
>
	<div class="mr-3">
		<input
			type="checkbox"
			class="h-6 w-6 rounded-md border-gray-200 bg-white shadow-sm mt-2"
			bind:checked={item.checked}
		/>
	</div>
	<div>
		<div class="text-gray-500 sm:pr-8">
			<h1 class="text-xl font-bold text-gray-900">{item.name}</h1>

			{#if item.details}
				<p class="mt-2 text-sm sm:block">{item.details}</p>
			{/if}
		</div>
	</div>
</div>
