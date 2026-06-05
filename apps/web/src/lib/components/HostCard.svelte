<script lang="ts">
	import * as Card from '$lib/components/ui/card/index.js';
	import type { HostsRecord } from '$lib/pocketbase-types';
	import { cn } from '$lib/utils';
	import {
		Trash2,
		BellRing,
		CircleCheck,
		CircleX,
		LoaderCircle,
		Moon,
		Power
	} from 'lucide-svelte';
	import { Button } from './ui/button';
	import { hostsStore } from '$lib/stores/hosts';

	let { host, class: className }: { host: HostsRecord; class?: string } = $props();
	const statuses = hostsStore.statuses;
	let status = $derived($statuses[host.id]);
	let powerAvailable = $derived(Boolean(status?.powerAvailable));

	function wake() {
		hostsStore.wakeHost(host);
	}

	function shutdown() {
		if (window.confirm(`Shutdown ${host.name}?`)) {
			hostsStore.powerHost(host, 'shutdown');
		}
	}

	function sleep() {
		hostsStore.powerHost(host, 'sleep');
	}

	function deleteHost() {
		if (window.confirm('Are you sure you want to delete this host?')) {
			hostsStore.deleteHost(host);
		}
	}
</script>

<Card.Root class={cn('relative', className)}>
	<Card.Content>
		<Button
			size="icon"
			class="absolute right-2 top-2 h-6 w-6 rounded-full"
			variant="destructive"
			onclick={deleteHost}
		>
			<Trash2 class="h-3 w-3" />
		</Button>
		<div class={cn('grid grid-cols-2 gap-2', className)}>
			<p class="text-md font-bold">
				Name: <span class="font-mono font-medium">{host.name}</span>
			</p>
			<p class="text-md font-bold">
				Broadcast IP: <span class="font-mono font-medium">{host.ip}</span>
			</p>
			<p class="text-md font-bold">
				Mac Address: <span class="font-mono font-medium">{host.mac}</span>
			</p>
			<p class="text-md font-bold">
				Port: <span class="font-mono font-medium">{host.port}</span>
			</p>
			<p class="text-md flex items-center gap-2 font-bold">
				Status:
				{#if status === undefined}
					<span class="inline-flex items-center gap-1 font-medium text-muted-foreground">
						<LoaderCircle class="h-4 w-4 animate-spin" />
						Checking
					</span>
				{:else if status.online}
					<span class="inline-flex items-center gap-1 font-medium text-green-600">
						<CircleCheck class="h-4 w-4" />
						Online
					</span>
				{:else}
					<span class="inline-flex items-center gap-1 font-medium text-muted-foreground">
						<CircleX class="h-4 w-4" />
						Offline
					</span>
				{/if}
			</p>
			{#if status?.online && status.ip}
				<p class="text-md font-bold">
					Host IP: <span class="font-mono font-medium">{status.ip}</span>
				</p>
			{/if}
		</div>
	</Card.Content>
	<Card.Footer class="flex justify-between gap-2">
		{#if status?.online}
			<Button
				size="sm"
				variant="destructive"
				class="w-full"
				onclick={shutdown}
				disabled={!powerAvailable}
				title={powerAvailable ? 'Shutdown' : 'Host agent is not configured'}
			>
				<Power class="h-4 w-4" />
				Shutdown
			</Button>
			<Button
				size="sm"
				variant="outline"
				class="w-full"
				onclick={sleep}
				disabled={!powerAvailable}
				title={powerAvailable ? 'Sleep' : 'Host agent is not configured'}
			>
				<Moon class="h-4 w-4" />
				Sleep
			</Button>
		{:else}
			<Button
				size="sm"
				variant="outline"
				class="w-full bg-green-400/60 hover:bg-green-400/40"
				onclick={wake}
			>
				<BellRing class="h-4 w-4" />
			</Button>
		{/if}
	</Card.Footer>
</Card.Root>
