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

export type HostPowerAction = 'shutdown' | 'sleep' | 'reboot';

export function createHostsStore() {
	const hosts = writable<HostsRecord[]>([]);
	const statuses = writable<Record<string, HostStatus>>({});
	const statusPollTokens = new Map<string, symbol>();
	let statusRefreshInterval: ReturnType<typeof setInterval> | undefined;
	let statusRefreshInFlight = false;

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

	function refreshHostStatuses() {
		if (statusRefreshInFlight) {
			return;
		}

		statusRefreshInFlight = true;
		fetchHostStatuses().finally(() => {
			statusRefreshInFlight = false;
		});
	}

	function startStatusRefresh(intervalMs = 5000) {
		stopStatusRefresh();
		refreshHostStatuses();
		statusRefreshInterval = setInterval(refreshHostStatuses, intervalMs);
	}

	function stopStatusRefresh() {
		if (statusRefreshInterval !== undefined) {
			clearInterval(statusRefreshInterval);
			statusRefreshInterval = undefined;
		}
		statusRefreshInFlight = false;
	}

	async function fetchHostStatus(host: HostsRecord) {
		const result = await pb.send<Record<string, HostStatus>>('/api/host-statuses', {
			method: 'POST',
			headers: {
				Accept: 'application/json',
				'Content-Type': 'application/json'
			},
			body: JSON.stringify({ ids: [host.id] })
		});
		const status = result[host.id] ?? { online: false };
		statuses.update((current) => ({ ...current, [host.id]: status }));
		return status;
	}

	function pollHostStatus(host: HostsRecord, targetOnline: boolean) {
		const delays = [3000, 6000, 9000, 12000, 15000, 20000, 30000, 45000, 60000];
		const token = Symbol(host.id);
		statusPollTokens.set(host.id, token);

		for (const [index, delay] of delays.entries()) {
			setTimeout(async () => {
				if (statusPollTokens.get(host.id) !== token) {
					return;
				}

				try {
					const status = await fetchHostStatus(host);
					if (status.online === targetOnline) {
						statusPollTokens.delete(host.id);
						return;
					}
				} catch (err) {
					console.error('Failed to poll host status', err);
				}

				if (index === delays.length - 1 && statusPollTokens.get(host.id) === token) {
					statusPollTokens.delete(host.id);
				}
			}, delay);
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
				pollHostStatus(host, true);
			})
			.catch((err) => {
				toast.error('Failed to wake host', {
					description: err.message
				});
			});
	}

	async function powerHost(host: HostsRecord, action: HostPowerAction) {
		pb.send('/api/host-power', {
			method: 'POST',
			headers: {
				Accept: 'application/json',
				'Content-Type': 'application/json'
			},
			body: JSON.stringify({ id: host.id, action })
		})
			.then(() => {
				const message =
					action === 'shutdown'
						? 'Shutdown requested'
						: action === 'reboot'
							? 'Reboot requested'
							: 'Sleep requested';
				toast.success(message);
				pollHostStatus(host, false);
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
		fetchHostStatus,
		startStatusRefresh,
		stopStatusRefresh,
		createHost,
		updateHost,
		deleteHost,
		wakeHost,
		powerHost
	};
}

export const hostsStore = createHostsStore();
