<script lang="ts">
	import { connectionStatus, type ConnectionStatus } from '$lib/services/connection';
	import { syncState } from '$lib/services/sync';

	let status: ConnectionStatus = 'online';
	let pending = $state(0);
	let syncing = $state(false);
	let lastSync: string | null = $state(null);
	let lastError: string | null = $state(null);

	$effect(() => {
		const unsub = connectionStatus.subscribe((v) => {
			status = v;
			pending = syncState.pending;
			syncing = syncState.syncing;
			lastSync = syncState.lastSync;
			lastError = syncState.lastError;
		});
		return unsub;
	});

	let showPanel = $state(false);

	function togglePanel() {
		showPanel = !showPanel;
	}

	function formatTime(iso: string | null): string {
		if (!iso) return 'Never';
		const d = new Date(iso);
		return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
	}
</script>

<div class="sync-container">
	{#if status === 'offline'}
		<button class="sync-badge offline" onclick={togglePanel} title="You are offline">
			<span class="dot offline-dot"></span>
			Offline
			{#if pending > 0}
				<span class="pending-count">{pending}</span>
			{/if}
		</button>
	{:else if syncing}
		<button class="sync-badge syncing" onclick={togglePanel} title="Syncing...">
			<span class="spinner"></span>
			Syncing {syncState.progress.current}/{syncState.progress.total}
		</button>
	{:else if pending > 0}
		<button class="sync-badge pending" onclick={togglePanel} title="Pending sync items">
			<span class="dot pending-dot"></span>
			{pending} pending
		</button>
	{:else}
		<button class="sync-badge online" onclick={togglePanel} title="All synced">
			<span class="dot online-dot"></span>
			Synced
		</button>
	{/if}

	{#if showPanel}
		<div class="sync-panel" onclick={(e) => e.stopPropagation()}>
			<div class="panel-header">
				<h3>Sync Status</h3>
				<button class="close-btn" onclick={() => (showPanel = false)}>&times;</button>
			</div>
			<div class="panel-body">
				<div class="row">
					<span>Status</span>
					<span class="capitalize">{status}</span>
				</div>
				<div class="row">
					<span>Pending items</span>
					<span>{pending}</span>
				</div>
				<div class="row">
					<span>Last sync</span>
					<span>{formatTime(lastSync)}</span>
				</div>
				{#if lastError}
					<div class="row error">
						<span>Last error</span>
						<span class="error-text">{lastError}</span>
					</div>
				{/if}
			</div>
		</div>
	{/if}
</div>

<style>
	.sync-container {
		position: relative;
		display: inline-flex;
	}

	.sync-badge {
		display: inline-flex;
		align-items: center;
		gap: 4px;
		padding: 4px 10px;
		border-radius: 12px;
		border: 1px solid;
		font-size: 12px;
		font-weight: 500;
		cursor: pointer;
		background: white;
		transition: all 0.2s;
	}

	.sync-badge:hover {
		filter: brightness(0.95);
	}

	.sync-badge.online {
		color: #16a34a;
		border-color: #bbf7d0;
	}

	.sync-badge.offline {
		color: #dc2626;
		border-color: #fecaca;
	}

	.sync-badge.syncing {
		color: #2563eb;
		border-color: #bfdbfe;
	}

	.sync-badge.pending {
		color: #d97706;
		border-color: #fde68a;
	}

	.dot {
		width: 8px;
		height: 8px;
		border-radius: 50%;
		display: inline-block;
	}

	.online-dot { background: #16a34a; }
	.offline-dot { background: #dc2626; }
	.pending-dot { background: #d97706; }

	.spinner {
		width: 8px;
		height: 8px;
		border: 2px solid #bfdbfe;
		border-top-color: #2563eb;
		border-radius: 50%;
		animation: spin 0.6s linear infinite;
	}

	@keyframes spin { to { transform: rotate(360deg); } }

	.pending-count {
		background: #dc2626;
		color: white;
		border-radius: 50%;
		padding: 0 5px;
		font-size: 10px;
		font-weight: 700;
		min-width: 16px;
		text-align: center;
	}

	.sync-panel {
		position: absolute;
		top: calc(100% + 8px);
		right: 0;
		width: 260px;
		background: white;
		border: 1px solid #e5e7eb;
		border-radius: 8px;
		box-shadow: 0 4px 12px rgba(0,0,0,0.1);
		z-index: 1000;
	}

	.panel-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 10px 14px;
		border-bottom: 1px solid #e5e7eb;
	}

	.panel-header h3 {
		margin: 0;
		font-size: 14px;
		font-weight: 600;
	}

	.close-btn {
		background: none;
		border: none;
		font-size: 18px;
		cursor: pointer;
		color: #6b7280;
		padding: 0;
		line-height: 1;
	}

	.panel-body {
		padding: 10px 14px;
	}

	.row {
		display: flex;
		justify-content: space-between;
		padding: 6px 0;
		font-size: 13px;
	}

	.row.error {
		flex-direction: column;
		gap: 4px;
	}

	.error-text {
		color: #dc2626;
		font-size: 12px;
		word-break: break-all;
	}

	.capitalize { text-transform: capitalize; }
</style>
