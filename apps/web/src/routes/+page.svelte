<script lang="ts">
	import { goto } from '$app/navigation';
	import CreateHostForm from '$lib/components/CreateHostForm.svelte';
	import HostCard from '$lib/components/HostCard.svelte';
	import { pb } from '$lib/pb';
	import { hostsStore } from '$lib/stores/hosts';
	import type { HostsRecord } from '$lib/pocketbase-types';
	import autoAnimate from '@formkit/auto-animate';
	import { Button } from '$lib/components/ui/button';
	import { Plus } from 'lucide-svelte';
	import { onDestroy, onMount } from 'svelte';

	let editingHost = $state<HostsRecord | null>(null);
	let isFormOpen = $state(false);
	let hasPushedFormState = false;

	function openForm(host: HostsRecord | null = null) {
		editingHost = host;
		isFormOpen = true;

		if (!hasPushedFormState) {
			window.history.pushState({ wolHostForm: true }, '', window.location.href);
			hasPushedFormState = true;
		}
	}

	function closeForm(useHistory = true) {
		editingHost = null;
		isFormOpen = false;

		if (useHistory && hasPushedFormState) {
			hasPushedFormState = false;
			window.history.back();
		}
	}

	$effect(() => {
		if (pb.authStore.isValid) {
			hostsStore.fetchHosts();
			hostsStore.startStatusRefresh();
		} else {
			hostsStore.stopStatusRefresh();
			goto('/auth');
		}
	});

	onMount(() => {
		function handlePopState() {
			if (isFormOpen) {
				hasPushedFormState = false;
				closeForm(false);
			}
		}

		window.addEventListener('popstate', handlePopState);
		return () => window.removeEventListener('popstate', handlePopState);
	});

	onDestroy(() => {
		hostsStore.stopStatusRefresh();
	});
</script>

<main class="pt-20 pb-20">
	{#if isFormOpen}
		<section class="mx-auto max-w-[40em] space-y-4 px-4">
			<h1 class="text-xl font-semibold">{editingHost ? 'Edit Host' : 'Create Host'}</h1>
			<CreateHostForm
				bind:editingHost
				onCancel={() => closeForm()}
				onSaved={() => closeForm()}
				class="rounded-lg border bg-card p-4"
			/>
		</section>
	{:else}
		<div class="space-y-2 px-4">
			<ul use:autoAnimate class="space-y-2">
				{#each $hostsStore as host (host.id)}
					<HostCard {host} onEdit={(selectedHost) => openForm(selectedHost)} class="mx-auto max-w-[40em]" />
				{/each}
			</ul>
		</div>

		<Button
			size="icon"
			class="fixed bottom-12 right-4 h-14 w-14 rounded-full shadow-lg"
			onclick={() => openForm()}
			title="Create host"
		>
			<Plus class="h-6 w-6" />
		</Button>
	{/if}
</main>
