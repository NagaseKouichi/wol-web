<script lang="ts">
	import { superForm, defaults } from 'sveltekit-superforms';
	import { zodClient, zod } from 'sveltekit-superforms/adapters';
	import { z } from 'zod';
	import * as Form from '$lib/components/ui/form/index.js';
	import { Input } from './ui/input';
	import { toast } from 'svelte-sonner';
	import { cn } from '$lib/utils';
	import { pb } from '$lib/pb';
	import { hostsStore, type HostFormData } from '$lib/stores/hosts';
	import * as Popover from '$lib/components/ui/popover/index';
	import { Button } from './ui/button';
	import { InfoIcon } from 'lucide-svelte';
	import type { HostsRecord } from '$lib/pocketbase-types';

	const formSchema = z.object({
		name: z.string().min(1, 'Name is required'),
		mac: z
			.string()
			.min(1, 'MAC is required')
			.regex(/^([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})$/, 'Invalid MAC address format'),
		ip: z.string().min(1, 'IP is required').ip(),
		hostIp: z
			.string()
			.optional()
			.refine((value) => !value || z.string().ip().safeParse(value).success, 'Invalid host IP'),
		agentUrl: z
			.string()
			.optional()
			.refine((value) => !value || z.string().url().safeParse(value).success, 'Invalid agent URL'),
		agentToken: z.string().optional(),
		port: z.number().default(9)
	});

	function emptyFormData(): HostFormData {
		return {
			name: '',
			mac: '',
			ip: '',
			hostIp: '',
			port: 9,
			agentUrl: '',
			agentToken: ''
		};
	}

	function formDataFromHost(host: HostsRecord): HostFormData {
		return {
			name: host.name,
			mac: host.mac,
			ip: host.ip,
			hostIp: host.hostIp ?? '',
			port: host.port,
			agentUrl: host.agentUrl ?? '',
			agentToken: ''
		};
	}

	function updatePayload(data: HostFormData): Partial<HostFormData> {
		const payload: Partial<HostFormData> = {
			name: data.name,
			mac: data.mac,
			ip: data.ip,
			hostIp: data.hostIp || '',
			port: data.port,
			agentUrl: data.agentUrl || ''
		};

		if (data.agentToken) {
			payload.agentToken = data.agentToken;
		}

		return payload;
	}

	const form = superForm(defaults(zod(formSchema)), {
		validators: zodClient(formSchema),
		SPA: true,
		async onUpdate({ form, cancel }) {
			if (!form.valid) {
				toast.error('Invalid');
				return;
			}
			if (!pb.authStore.record) {
				toast.error('You must be logged in to create a host');
				return;
			}

			try {
				if (editingHost) {
					await hostsStore.updateHost(editingHost.id, updatePayload(form.data));
					toast.success('Host updated');
				} else {
					await hostsStore.createHost(form.data);
					toast.success('Host created');
				}
			} catch (err) {
				toast.error(editingHost ? 'Failed to update host' : 'Failed to create host', {
					description: err instanceof Error ? err.message : String(err)
				});
				cancel();
				return;
			}

			resetForm();
			onSaved?.();
			cancel();
		}
	});
	const { form: formData, enhance } = form;

	let {
		editingHost = $bindable<HostsRecord | null>(null),
		onCancel,
		onSaved,
		class: className
	}: {
		editingHost?: HostsRecord | null;
		onCancel?: () => void;
		onSaved?: () => void;
		class?: string;
	} = $props();

	let loadedHostId = $state<string | null>(null);

	function resetForm() {
		$formData = emptyFormData();
		editingHost = null;
		loadedHostId = null;
	}

	$effect(() => {
		if (editingHost && editingHost.id !== loadedHostId) {
			$formData = formDataFromHost(editingHost);
			loadedHostId = editingHost.id;
		}
	});
</script>

<form method="POST" use:enhance class={cn('grid grid-cols-2 gap-2', className)}>
	<Form.Field {form} name="name">
		<Form.Control>
			{#snippet children({ props })}
				<div class="flex flex-col gap-2">
					<Form.Label>Name</Form.Label>
					<Input {...props} name="name" bind:value={$formData.name} placeholder="TrueNAS" />
				</div>
			{/snippet}
		</Form.Control>
		<Form.FieldErrors />
	</Form.Field>
	<Form.Field {form} name="mac">
		<Form.Control>
			{#snippet children({ props })}
				<div class="flex flex-col gap-2">
					<Form.Label>MAC</Form.Label>
					<Input {...props} name="mac" bind:value={$formData.mac} placeholder="86:2f:57:c1:df:65" />
				</div>
			{/snippet}
		</Form.Control>
		<Form.FieldErrors />
	</Form.Field>
	<Form.Field {form} name="ip">
		<Form.Control>
			{#snippet children({ props })}
				<div class="flex flex-col gap-2">
					<Form.Label>Broadcast IP</Form.Label>
					<div class="flex space-x-1">
						<Input
							{...props}
							name="ip"
							bind:value={$formData.ip}
							placeholder="e.g. 255.255.255.255"
						/>

						<Popover.Root>
							<Popover.Trigger>
								<Button variant="secondary" size="icon"><InfoIcon /></Button>
							</Popover.Trigger>
							<Popover.Content>
								<p>Use a broadcast IP address. Default should be <code>255.255.255.255</code></p>
								<p>
									If your computer is connected to multiple networks, then use a more specific
									subnet broadcast ip
								</p>
								<p>
									If the target host's ip is <code>192.168.1.123</code>, then use
									<code>192.168.1.255</code>
								</p>
							</Popover.Content>
						</Popover.Root>
					</div>
				</div>
			{/snippet}
		</Form.Control>
		<Form.FieldErrors />
	</Form.Field>
	<Form.Field {form} name="hostIp">
		<Form.Control>
			{#snippet children({ props })}
				<div class="flex flex-col gap-2">
					<Form.Label>Host IP</Form.Label>
					<Input
						{...props}
						name="hostIp"
						bind:value={$formData.hostIp}
						placeholder="e.g. 192.168.1.54"
					/>
				</div>
			{/snippet}
		</Form.Control>
		<Form.FieldErrors />
	</Form.Field>
	<Form.Field {form} name="port">
		<Form.Control>
			{#snippet children({ props })}
				<div class="flex flex-col gap-2">
					<Form.Label>Port</Form.Label>
					<Input {...props} name="port" bind:value={$formData.port} placeholder="Port" />
				</div>
			{/snippet}
		</Form.Control>
		<Form.FieldErrors />
	</Form.Field>
	<Form.Field {form} name="agentUrl">
		<Form.Control>
			{#snippet children({ props })}
				<div class="flex flex-col gap-2">
					<Form.Label>Agent URL</Form.Label>
					<Input
						{...props}
						name="agentUrl"
						bind:value={$formData.agentUrl}
						placeholder="e.g. http://192.168.1.54:8765"
					/>
				</div>
			{/snippet}
		</Form.Control>
		<Form.FieldErrors />
	</Form.Field>
	<Form.Field {form} name="agentToken">
		<Form.Control>
			{#snippet children({ props })}
				<div class="flex flex-col gap-2">
					<Form.Label>Agent Token</Form.Label>
					<Input
						{...props}
						name="agentToken"
						type="password"
						bind:value={$formData.agentToken}
						placeholder="Long random token"
					/>
				</div>
			{/snippet}
		</Form.Control>
		<Form.FieldErrors />
	</Form.Field>
	<div class="col-span-2 mt-4 flex gap-2">
		<Button
			type="button"
			variant="outline"
			class="w-full"
			onclick={() => {
				resetForm();
				onCancel?.();
			}}
		>
			Cancel
		</Button>
		<Form.Button class="w-full">{editingHost ? 'Edit' : 'Create'}</Form.Button>
	</div>
</form>
