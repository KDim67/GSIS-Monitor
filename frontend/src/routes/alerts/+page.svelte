<script>
	import { onMount } from 'svelte';

	const API_BASE = 'http://localhost:8000';

	// Alert logs and rules states
	let alerts = $state([]);
	let rules = $state([]);
	let services = $state([]);
	
	let isLoadingAlerts = $state(false);
	let isLoadingRules = $state(false);
	let error = $state(null);

	// Alert Rule Form State
	let formServiceId = $state('');
	let formMetric = $state('latency');
	let formOperator = $state('>');
	let formValue = $state(2500);

	// Fetch alerts history
	async function loadAlerts() {
		isLoadingAlerts = true;
		try {
			const res = await fetch(`${API_BASE}/api/alerts`);
			if (!res.ok) throw new Error('Failed to load alert logs');
			alerts = await res.json();
		} catch (e) {
			error = e.message;
		} finally {
			isLoadingAlerts = false;
		}
	}

	// Fetch rules list
	async function loadRules() {
		isLoadingRules = true;
		try {
			const res = await fetch(`${API_BASE}/api/rules`);
			if (!res.ok) throw new Error('Failed to load alert rules');
			rules = await res.json();
		} catch (e) {
			error = e.message;
		} finally {
			isLoadingRules = false;
		}
	}

	// Fetch services to populate dropdown
	async function loadServices() {
		try {
			const res = await fetch(`${API_BASE}/api/services`);
			if (!res.ok) throw new Error('Failed to load services');
			services = await res.json();
			if (services.length > 0) {
				formServiceId = services[0].id.toString();
			}
		} catch (e) {
			console.error(e);
		}
	}

	// Resolve alert manually
	async function resolveAlert(id) {
		try {
			const res = await fetch(`${API_BASE}/api/alerts/${id}/resolve`, {
				method: 'PUT'
			});
			if (!res.ok) throw new Error('Failed to resolve alert');
			await loadAlerts();
		} catch (e) {
			alert(`Σφάλμα: ${e.message}`);
		}
	}

	// Create new alert rule
	async function handleSubmitRule(e) {
		e.preventDefault();
		const payload = {
			service_id: parseInt(formServiceId),
			metric: formMetric,
			operator: formOperator,
			value: parseFloat(formValue)
		};

		try {
			const res = await fetch(`${API_BASE}/api/rules`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(payload)
			});
			if (!res.ok) throw new Error('Failed to create rule');
			
			// Reset form values defaults
			formValue = formMetric === 'ssl_expiry' ? 15 : formMetric === 'latency' ? 2500 : 200;
			await loadRules();
		} catch (e) {
			alert(`Σφάλμα: ${e.message}`);
		}
	}

	// Delete alert rule
	async function handleDeleteRule(id) {
		if (!confirm('Θέλετε σίγουρα να διαγράψετε αυτόν τον κανόνα ειδοποίησης;')) return;

		try {
			const res = await fetch(`${API_BASE}/api/rules/${id}`, {
				method: 'DELETE'
			});
			if (!res.ok) throw new Error('Failed to delete rule');
			await loadRules();
		} catch (e) {
			alert(`Σφάλμα: ${e.message}`);
		}
	}

	// Helper to find service name by ID
	function getServiceName(id) {
		const s = services.find(item => item.id === id);
		return s ? s.name : `Service #${id}`;
	}

	// Helper to format metric names
	function formatMetric(m) {
		switch (m) {
			case 'latency': return 'Καθυστέρηση (ms)';
			case 'ssl_expiry': return 'Λήξη SSL (ημέρες)';
			case 'status_code': return 'HTTP Status Code';
			case 'content_verified': return 'Keyword Verification';
			default: return m;
		}
	}

	onMount(() => {
		Promise.all([loadAlerts(), loadRules(), loadServices()]);
	});
</script>

<div class="container">
	<div class="alerts-header">
		<h2>Διαχείριση Ειδοποιήσεων & Κανόνων</h2>
		<p>Δείτε το ιστορικό σφαλμάτων και ρυθμίστε κανόνες ορίων (thresholds).</p>
	</div>

	<div class="alerts-layout">
		<!-- Left: Alert Logs History -->
		<div class="history-column">
			<div class="section-card">
				<div class="section-card-header">
					<h3>Ιστορικό Καταγραφών</h3>
					<button class="btn btn-secondary btn-sm" onclick={loadAlerts} disabled={isLoadingAlerts}>
						Ανανέωση
					</button>
				</div>

				<div class="table-wrapper">
					{#if alerts.length === 0}
						<div class="empty-state">
							Δεν υπάρχουν καταγραφές ειδοποιήσεων.
						</div>
					{:else}
						<table>
							<thead>
								<tr>
									<th>Υπηρεσία</th>
									<th>Κατάσταση</th>
									<th>Μήνυμα Σφάλματος</th>
									<th>Έναρξη</th>
									<th>Επίλυση</th>
									<th class="text-right">Ενέργεια</th>
								</tr>
							</thead>
							<tbody>
								{#each alerts as alert}
									<tr>
										<td class="font-bold">{alert.service_name}</td>
										<td>
											<span class="status-badge" class:healthy={alert.status === 'resolved'} class:unhealthy={alert.status === 'active'}>
												<span class="status-dot" class:healthy={alert.status === 'resolved'} class:unhealthy={alert.status === 'active'}></span>
												{alert.status === 'active' ? 'Ενεργό' : 'Επιλύθηκε'}
											</span>
										</td>
										<td class="text-break">{alert.message}</td>
										<td class="text-muted font-mono">{new Date(alert.triggered_at).toLocaleString()}</td>
										<td class="text-muted font-mono">
											{alert.resolved_at ? new Date(alert.resolved_at).toLocaleString() : '-'}
										</td>
										<td class="text-right">
											{#if alert.status === 'active'}
												<button class="btn btn-success btn-xs" onclick={() => resolveAlert(alert.id)}>
													Επίλυση
												</button>
											{:else}
												<span class="text-muted">-</span>
											{/if}
										</td>
									</tr>
								{/each}
							</tbody>
						</table>
					{/if}
				</div>
			</div>
		</div>

		<!-- Right: Alert Rules & Add Rule Form -->
		<div class="rules-column">
			<!-- Add Rule Form -->
			<div class="section-card margin-bottom-lg">
				<div class="section-card-header">
					<h3>Νέος Κανόνας Ειδοποίησης</h3>
				</div>
				<form onsubmit={handleSubmitRule}>
					<div class="form-group">
						<label for="rule-service">Υπηρεσία</label>
						<select id="rule-service" bind:value={formServiceId} required>
							{#each services as service}
								<option value={service.id.toString()}>{service.name}</option>
							{/each}
						</select>
					</div>

					<div class="form-row">
						<div class="form-group">
							<label for="rule-metric">Μετρικό</label>
							<select id="rule-metric" bind:value={formMetric} required>
								<option value="latency">Καθυστέρηση (ms)</option>
								<option value="ssl_expiry">Λήξη SSL (ημέρες)</option>
								<option value="status_code">HTTP Status Code</option>
								<option value="content_verified">Keyword Verification</option>
							</select>
						</div>
						<div class="form-group">
							<label for="rule-operator">Τελεστής</label>
							<select id="rule-operator" bind:value={formOperator} required>
								<option value=">">&gt;</option>
								<option value="<">&lt;</option>
								<option value="=">=</option>
								<option value="!=">!=</option>
							</select>
						</div>
					</div>

					<div class="form-group">
						<label for="rule-value">Τιμή Ορίου</label>
						<input type="number" id="rule-value" bind:value={formValue} required step="any" />
						<span class="input-hint">
							{#if formMetric === 'latency'}Τιμή σε milliseconds (π.χ. 2500){/if}
							{#if formMetric === 'ssl_expiry'}Τιμή σε ημέρες (π.χ. 15){/if}
							{#if formMetric === 'status_code'}HTTP Status (π.χ. 200){/if}
							{#if formMetric === 'content_verified'}1 για Επιτυχία, 0 για Αποτυχία{/if}
						</span>
					</div>

					<button type="submit" class="btn btn-primary w-full">
						Προσθήκη Κανόνα
					</button>
				</form>
			</div>

			<!-- Active Rules List -->
			<div class="section-card">
				<div class="section-card-header">
					<h3>Κανόνες σε Λειτουργία</h3>
				</div>
				<div class="table-wrapper">
					{#if rules.length === 0}
						<div class="empty-state">
							Δεν υπάρχουν κανόνες.
						</div>
					{:else}
						<table>
							<thead>
								<tr>
									<th>Υπηρεσία</th>
									<th>Κανόνας</th>
									<th class="text-right">Ενέργεια</th>
								</tr>
							</thead>
							<tbody>
								{#each rules as rule}
									<tr>
										<td class="font-bold">{getServiceName(rule.service_id)}</td>
										<td class="font-mono text-muted">
											{formatMetric(rule.metric)} {rule.operator} {rule.value}
										</td>
										<td class="text-right">
											<button class="btn-icon text-error-icon" onclick={() => handleDeleteRule(rule.id)} title="Διαγραφή">
												<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
													<polyline points="3 6 5 6 21 6"></polyline>
													<path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path>
													<line x1="10" y1="11" x2="10" y2="17"></line>
													<line x1="14" y1="11" x2="14" y2="17"></line>
												</svg>
											</button>
										</td>
									</tr>
								{/each}
							</tbody>
						</table>
					{/if}
				</div>
			</div>
		</div>
	</div>
</div>

<style>
	.alerts-header {
		margin-bottom: 2rem;
	}

	.alerts-header h2 {
		font-size: 1.75rem;
		color: var(--primary-color);
	}

	.alerts-header p {
		color: var(--text-secondary);
		font-size: 0.9375rem;
		margin-top: 0.25rem;
	}

	.alerts-layout {
		display: grid;
		grid-template-columns: 1fr 400px;
		gap: 2rem;
		align-items: start;
	}

	.section-card {
		background-color: var(--bg-card);
		border: 1px solid var(--border-color);
		border-radius: var(--radius-lg);
		box-shadow: var(--shadow-sm);
		overflow: hidden;
	}

	.section-card-header {
		padding: 1.25rem 1.5rem;
		border-bottom: 1px solid var(--border-color);
		display: flex;
		justify-content: space-between;
		align-items: center;
		background-color: var(--bg-card);
	}

	.section-card-header h3 {
		font-size: 1.125rem;
		color: var(--text-primary);
	}

	.table-wrapper {
		overflow-x: auto;
	}

	.empty-state {
		padding: 3rem;
		text-align: center;
		color: var(--text-muted);
	}

	.font-bold {
		font-weight: 600;
		font-family: 'Outfit', sans-serif;
		color: var(--text-primary);
	}

	.font-mono {
		font-family: SFMono-Regular, Consolas, Liberation Mono, Menlo, monospace;
	}

	.text-muted {
		color: var(--text-muted);
	}

	.text-center {
		text-align: center;
	}

	.text-right {
		text-align: right;
	}

	.text-break {
		word-break: break-all;
	}

	.margin-bottom-lg {
		margin-bottom: 2rem;
	}

	/* Buttons */
	.btn {
		padding: 0.5rem 1rem;
		font-size: 0.8125rem;
		font-weight: 600;
		border-radius: var(--radius-md);
		display: inline-flex;
		align-items: center;
		justify-content: center;
	}

	.btn-sm {
		padding: 0.375rem 0.75rem;
		font-size: 0.75rem;
	}

	.btn-xs {
		padding: 0.25rem 0.5rem;
		font-size: 0.6875rem;
		border-radius: var(--radius-sm);
	}

	.btn-primary {
		background-color: var(--primary-color);
		color: var(--bg-color);
		width: 100%;
	}

	.btn-primary:hover {
		background-color: var(--primary-hover);
	}

	.btn-secondary {
		background-color: var(--bg-darker);
		color: var(--text-secondary);
		border: 1px solid var(--border-color);
	}

	.btn-secondary:hover {
		background-color: var(--border-color);
	}

	.btn-success {
		background-color: var(--success-light);
		color: var(--success-color);
		border: 1px solid var(--success-color);
	}

	.btn-success:hover {
		background-color: var(--success-color);
		color: var(--bg-color);
	}

	/* Icon buttons */
	.btn-icon {
		background: none;
		border: none;
		padding: 0.375rem;
		border-radius: var(--radius-sm);
		color: var(--text-secondary);
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 2rem;
		height: 2rem;
		transition: var(--transition-all);
	}

	.btn-icon svg {
		width: 1.125rem;
		height: 1.125rem;
	}

	.btn-icon.text-error-icon:hover {
		background-color: var(--error-light);
		color: var(--error-color);
	}

	/* Form */
	form {
		padding: 1.5rem;
	}

	.form-group {
		margin-bottom: 1.25rem;
	}

	.form-row {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 1rem;
	}

	.input-hint {
		display: block;
		font-size: 0.6875rem;
		color: var(--text-muted);
		margin-top: 0.25rem;
	}

	.w-full {
		width: 100%;
	}

	@media (max-width: 1024px) {
		.alerts-layout {
			grid-template-columns: 1fr;
		}
	}
</style>
