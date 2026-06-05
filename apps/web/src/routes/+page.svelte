<script lang="ts">
	import { goto } from '$app/navigation';
	import CreateHostForm from '$lib/components/CreateHostForm.svelte';
	import HostCard from '$lib/components/HostCard.svelte';
	import { pb } from '$lib/pb';
	import { hostsStore } from '$lib/stores/hosts';
	import type { HostsRecord } from '$lib/pocketbase-types';
	import autoAnimate from '@formkit/auto-animate';

	let editingHost = $state<HostsRecord | null>(null);

	$effect(() => {
		if (pb.authStore.isValid) {
			hostsStore.fetchHosts();
		} else {
			goto('/auth');
		}
	});
</script>

<main class="pt-20">
	<div class="space-y-2">
		<CreateHostForm bind:editingHost class="mx-auto max-w-[40em]" />
		<ul use:autoAnimate class="space-y-2">
			{#each $hostsStore as host (host.id)}
				<HostCard
					{host}
					onEdit={(selectedHost) => (editingHost = selectedHost)}
					class="mx-auto max-w-[40em]"
				/>
			{/each}
		</ul>
	</div>
</main>
