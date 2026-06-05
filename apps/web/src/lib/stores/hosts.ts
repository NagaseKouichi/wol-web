import { pb } from '$lib/pb';
import type { HostsRecord, RecordIdString } from '$lib/pocketbase-types';
import { toast } from 'svelte-sonner';
import { get, writable } from 'svelte/store';

export type HostStatus = {
	online: boolean;
	ip?: string;
	powerAvailable?: boolean;
};

export type HostFormData = {
	agentToken?: string;
	agentUrl?: string;
	ip: string;
	hostIp?: string;
	mac: string;
	name: string;
	port: number;
};

export function createHostsStore() {
	const hosts = writable<HostsRecord[]>([]);
	const statuses = writable<Record<string, HostStatus>>({});

	async function fetchHosts() {
		const records = await pb.collection('hosts').getFullList();
		hosts.set(records);
		fetchHostStatuses(records);
	}

	async function fetchHostStatuses(records = get(hosts)) {
		if (records.length === 0) {
			statuses.set({});
			return;
		}

		try {
			const result = await pb.send<Record<string, HostStatus>>('/api/host-statuses', {
				method: 'POST',
				headers: {
					Accept: 'application/json',
					'Content-Type': 'application/json'
				},
				body: JSON.stringify({ ids: records.map((host) => host.id) })
			});
			statuses.set(
				Object.fromEntries(records.map((host) => [host.id, result[host.id] ?? { online: false }]))
			);
		} catch (err) {
			console.error('Failed to fetch host statuses', err);
		}
	}

	async function createHost(host: HostFormData) {
		await pb.send('/api/host-config', {
			method: 'POST',
			headers: {
				Accept: 'application/json',
				'Content-Type': 'application/json'
			},
			body: JSON.stringify(host)
		});
		fetchHosts();
	}

	async function updateHost(id: RecordIdString, host: Partial<HostFormData>) {
		await pb.send('/api/host-config', {
			method: 'POST',
			headers: {
				Accept: 'application/json',
				'Content-Type': 'application/json'
			},
			body: JSON.stringify({ ...host, id })
		});
		fetchHosts();
	}

	async function deleteHost(host: HostsRecord) {
		await pb.collection('hosts').delete(host.id);
		fetchHosts();
	}

	async function wakeHost(host: HostsRecord) {
		pb.send('/api/wake', {
			method: 'POST',
			headers: {
				Accept: 'application/json',
				'Content-Type': 'application/json'
			},
			body: JSON.stringify({ id: host.id })
		})
			.then((res) => {
				toast.success('WakeOnLan Magic Packet Sent');
				setTimeout(() => fetchHostStatuses(), 5000);
			})
			.catch((err) => {
				toast.error('Failed to wake host', {
					description: err.message
				});
			});
	}

	async function powerHost(host: HostsRecord, action: 'shutdown' | 'sleep') {
		pb.send('/api/host-power', {
			method: 'POST',
			headers: {
				Accept: 'application/json',
				'Content-Type': 'application/json'
			},
			body: JSON.stringify({ id: host.id, action })
		})
			.then(() => {
				toast.success(action === 'shutdown' ? 'Shutdown requested' : 'Sleep requested');
				setTimeout(() => fetchHostStatuses(), 3000);
			})
			.catch((err) => {
				toast.error('Failed to request power action', {
					description: err.message
				});
			});
	}

	return {
		...hosts,
		statuses,
		fetchHosts,
		fetchHostStatuses,
		createHost,
		updateHost,
		deleteHost,
		wakeHost,
		powerHost
	};
}

export const hostsStore = createHostsStore();
