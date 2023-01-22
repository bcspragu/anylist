<script lang="ts">
	import type { PageData } from './$types';
	import Checkbox from '$lib/Checkbox.svelte';

	export let data: PageData;
</script>

<header>
	<div class="mx-auto px-4 py-4">
		<h1 class="text-2xl font-bold">
			(Br)AnyList - {data.list.name}
		</h1>
	</div>
</header>
<div class="">
	{#each data.list.items as item}
		<Checkbox {item} />
	{/each}
</div>
