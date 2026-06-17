<script>
	import { onMount, onDestroy } from 'svelte';
	import Chart from 'chart.js/auto';

	const API_BASE = 'http://localhost:8000';

	// Dashboard state (Svelte 5 reactive state)
	let summary = $state({
		total_services: 0,
		active_services: 0,
		healthy_services: 0,
		unhealthy_services: 0,
		active_alerts: 0,
		average_response_time_ms: 0.0,
		services: []
	});

	let selectedService = $state(null);
	let selectedRange = $state('24h');
	let serviceLogs = $state([]);
	let isLoadingLogs = $state(false);
	let logsError = $state(null);

	let chartCanvas = $state(null);
	let chartInstance = null;
	let pollInterval = null;

	// Load dashboard summary
	async function loadSummary() {
		try {
			const res = await fetch(`${API_BASE}/api/dashboard/summary`);
			if (!res.ok) throw new Error('Failed to load summary');
			summary = await res.json();
			
			// Auto-select first service if none selected
			if (summary.services.length > 0 && !selectedService) {
				selectService(summary.services[0]);
			} else if (selectedService) {
				// Update selected service reference to get fresh metrics
				const updated = summary.services.find(s => s.id === selectedService.id);
				if (updated) selectedService = updated;
			}
		} catch (e) {
			console.error(e);
		}
	}

	// Select service to load logs and display chart
	function selectService(service) {
		selectedService = service;
		loadLogs();
	}

	// Change chart range
	function changeRange(range) {
		selectedRange = range;
		loadLogs();
	}

	// Load logs for the selected service and range
	async function loadLogs() {
		if (!selectedService) return;
		isLoadingLogs = true;
		logsError = null;
		try {
			const res = await fetch(`${API_BASE}/api/services/${selectedService.id}/logs?range=${selectedRange}`);
			if (!res.ok) throw new Error('Failed to load logs');
			serviceLogs = await res.json();
			renderChart();
		} catch (e) {
			logsError = e.message;
		} finally {
			isLoadingLogs = false;
		}
	}

	// Format timestamp labels based on time range and date
	function formatLabel(timeStr) {
		const d = new Date(timeStr);
		const now = new Date();
		const isToday = d.getDate() === now.getDate() && 
		                d.getMonth() === now.getMonth() && 
		                d.getFullYear() === now.getFullYear();
		
		const timePart = d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
		
		if (selectedRange === '1h' || selectedRange === '6h') {
			return timePart;
		}
		
		if (isToday) {
			return timePart;
		} else {
			return `${d.getDate()}/${d.getMonth() + 1} ${timePart}`;
		}
	}

	// Render the stacked-bar chart
	function renderChart() {
		if (!chartCanvas) return;

		// Clean up existing chart
		if (chartInstance) {
			chartInstance.destroy();
			chartInstance = null;
		}

		if (serviceLogs.length === 0) return;

		// Downsample logs if there are too many to keep the chart responsive and legible
		let sampledLogs = [...serviceLogs];
		const maxBars = 50;
		if (sampledLogs.length > maxBars) {
			const step = Math.ceil(sampledLogs.length / maxBars);
			sampledLogs = sampledLogs.filter((_, idx) => idx % step === 0);
		}

		const labels = sampledLogs.map(l => formatLabel(l.time));
		const dnsData = sampledLogs.map(l => l.dns_lookup_ms || 0);
		const tcpData = sampledLogs.map(l => l.tcp_connect_ms || 0);
		const tlsData = sampledLogs.map(l => l.tls_handshake_ms || 0);
		
		// TTFB represents the remainder of time up to the first byte
		// Total response represents total response time.
		const ttfbData = sampledLogs.map(l => {
			const dns = l.dns_lookup_ms || 0;
			const tcp = l.tcp_connect_ms || 0;
			const tls = l.tls_handshake_ms || 0;
			const ttfb = l.ttfb_ms || 0;
			// TTFB trace in Go might include DNS/TCP/TLS. Let's calculate network trace remainder:
			const remainder = ttfb - dns - tcp - tls;
			return remainder > 0 ? remainder : ttfb;
		});

		chartInstance = new Chart(chartCanvas, {
			type: 'bar',
			data: {
				labels,
				datasets: [
					{
						label: 'DNS Lookup (ms)',
						data: dnsData,
						backgroundColor: '#0188ca', // Primary blue
						borderWidth: 0
					},
					{
						label: 'TCP Connect (ms)',
						data: tcpData,
						backgroundColor: '#6b7280', // Slate gray
						borderWidth: 0
					},
					{
						label: 'TLS Handshake (ms)',
						data: tlsData,
						backgroundColor: '#ed5929', // Accent orange
						borderWidth: 0
					},
					{
						label: 'TTFB / Processing (ms)',
						data: ttfbData,
						backgroundColor: '#9ca3af', // Light gray
						borderWidth: 0
					}
				]
			},
			options: {
				responsive: true,
				maintainAspectRatio: false,
				scales: {
					x: {
						stacked: true,
						grid: {
							display: false
						},
						ticks: {
							font: {
								size: 10
							}
						}
					},
					y: {
						stacked: true,
						title: {
							display: true,
							text: 'Καθυστέρηση (ms)',
							font: {
								weight: 'bold'
							}
						},
						ticks: {
							font: {
								size: 10
							}
						}
					}
				},
				plugins: {
					legend: {
						position: 'bottom',
						labels: {
							boxWidth: 12,
							font: {
								size: 11
							}
						}
					},
					tooltip: {
						mode: 'index',
						intersect: false
					}
				}
			}
		});
	}

	onMount(() => {
		loadSummary();
		// Poll summary data every 10 seconds
		pollInterval = setInterval(loadSummary, 10000);
	});

	onDestroy(() => {
		if (pollInterval) clearInterval(pollInterval);
		if (chartInstance) chartInstance.destroy();
	});

	// Reactively re-run chart render if canvas references load or selectedRange changes
	$effect(() => {
		if (chartCanvas && serviceLogs.length > 0) {
			// Explicit dependency on selectedRange to trigger effect on change
			selectedRange;
			renderChart();
		}
	});
</script>

<div class="container">
	<div class="dashboard-header">
		<h2>Επισκόπηση Διαθεσιμότητας Υπηρεσιών</h2>
		<p>Σύστημα ελέγχου και καταγραφής τηλεμετρίας σε πραγματικό χρόνο.</p>
	</div>

	<!-- Summary Cards -->
	<div class="metrics-grid">
		<div class="metric-card">
			<span class="metric-title">Σύνολο Υπηρεσιών</span>
			<span class="metric-value">{summary.total_services}</span>
			<div class="metric-sub">
				Ενεργές: <strong>{summary.active_services}</strong>
			</div>
		</div>
		<div class="metric-card">
			<span class="metric-title">Υγιείς Πύλες</span>
			<span class="metric-value text-success">{summary.healthy_services}</span>
			<div class="metric-sub">
				Σε λειτουργία
			</div>
		</div>
		<div class="metric-card">
			<span class="metric-title">Εκτός Λειτουργίας</span>
			<span class="metric-value" class:text-error={summary.unhealthy_services > 0} class:text-muted={summary.unhealthy_services === 0}>
				{summary.unhealthy_services}
			</span>
			<div class="metric-sub">
				Απαιτούν προσοχή
			</div>
		</div>
		<div class="metric-card">
			<span class="metric-title">Ενεργές Ειδοποιήσεις</span>
			<span class="metric-value" class:text-warning={summary.active_alerts > 0} class:text-muted={summary.active_alerts === 0}>
				{summary.active_alerts}
			</span>
			<div class="metric-sub">
				Alert logs
			</div>
		</div>
		<div class="metric-card">
			<span class="metric-title">Μέση Απόκριση (24h)</span>
			<span class="metric-value text-primary">{summary.average_response_time_ms} ms</span>
			<div class="metric-sub">
				Απόκριση δικτύου
			</div>
		</div>
	</div>

	<div class="dashboard-layout">
		<!-- Services Cards Grid -->
		<div class="services-column">
			<h3 class="column-title">Κατάσταση Υπηρεσιών</h3>
			<div class="services-list">
				{#each summary.services as service}
					<button 
						class="service-card" 
						class:active={selectedService && selectedService.id === service.id}
						class:unhealthy={!service.is_healthy}
						onclick={() => selectService(service)}
					>
						<div class="service-card-header">
							<span class="service-name">{service.name}</span>
							<span class="status-badge" class:healthy={service.is_healthy} class:unhealthy={!service.is_healthy}>
								<span class="status-dot" class:healthy={service.is_healthy} class:unhealthy={!service.is_healthy}></span>
								{service.is_healthy ? 'ONLINE' : 'OFFLINE'}
							</span>
						</div>
						
						<div class="service-url">{service.url}</div>
						
						<div class="service-metrics">
							<div class="metric-item">
								<span class="label">Απόκριση:</span>
								<span class="val">
									{service.last_response_time_ms ? `${service.last_response_time_ms.toFixed(1)} ms` : '-'}
								</span>
							</div>
							<div class="metric-item">
								<span class="label">SSL Expiry:</span>
								<span class="val" class:text-error={service.ssl_expiry_days !== null && service.ssl_expiry_days < 15}>
									{service.ssl_expiry_days !== null ? `${service.ssl_expiry_days} ημ.` : '-'}
								</span>
							</div>
						</div>

						{#if service.active_alerts_count > 0}
							<div class="service-alert-pill">
								Ενεργές Ειδοποιήσεις: {service.active_alerts_count}
							</div>
						{/if}
					</button>
				{/each}
			</div>
		</div>

		<!-- Latency Chart Area -->
		<div class="chart-column">
			<div class="chart-card">
				{#if selectedService}
					<div class="chart-header">
						<div>
							<h3 class="chart-title">Ανάλυση Καθυστέρησης: {selectedService.name}</h3>
							<p class="chart-subtitle">Ανάλυση DNS, TCP, TLS, TTFB (ms)</p>
						</div>
						<div class="range-selector">
							<button class:active={selectedRange === '1h'} onclick={() => changeRange('1h')}>1h</button>
							<button class:active={selectedRange === '6h'} onclick={() => changeRange('6h')}>6h</button>
							<button class:active={selectedRange === '24h'} onclick={() => changeRange('24h')}>24h</button>
							<button class:active={selectedRange === '7d'} onclick={() => changeRange('7d')}>7d</button>
						</div>
					</div>

					<div class="chart-container-wrapper">
						{#if isLoadingLogs}
							<div class="chart-overlay">
								<span class="spinner"></span>
								<span>Φόρτωση δεδομένων...</span>
							</div>
						{/if}

						{#if logsError}
							<div class="chart-overlay error-text">
								<span>Σφάλμα: {logsError}</span>
							</div>
						{/if}

						{#if !isLoadingLogs && !logsError && serviceLogs.length === 0}
							<div class="chart-overlay">
								<span>Δεν βρέθηκαν telemetry logs για αυτή την περίοδο.</span>
							</div>
						{/if}

						<canvas bind:this={chartCanvas}></canvas>
					</div>
				{:else}
					<div class="no-selection">
						<p>Παρακαλώ επιλέξτε μια υπηρεσία από τη λίστα για να δείτε τα γραφήματα τηλεμετρίας.</p>
					</div>
				{/if}
			</div>
		</div>
	</div>
</div>

<style>
	.dashboard-header {
		margin-bottom: 2rem;
	}

	.dashboard-header h2 {
		font-size: 1.75rem;
		color: var(--primary-color);
	}

	.dashboard-header p {
		color: var(--text-secondary);
		font-size: 0.9375rem;
		margin-top: 0.25rem;
	}

	/* Metrics summary grid */
	.metrics-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
		gap: 1rem;
		margin-bottom: 2.5rem;
	}

	.metric-card {
		background-color: var(--bg-card);
		border: 1px solid var(--border-color);
		border-radius: var(--radius-lg);
		padding: 1.25rem;
		box-shadow: var(--shadow-sm);
		display: flex;
		flex-direction: column;
	}

	.metric-title {
		font-size: 0.8125rem;
		font-weight: 600;
		color: var(--text-secondary);
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.metric-value {
		font-family: 'Outfit', sans-serif;
		font-size: 1.875rem;
		font-weight: 700;
		margin: 0.375rem 0;
	}

	.metric-sub {
		font-size: 0.75rem;
		color: var(--text-muted);
	}

	.text-success {
		color: var(--success-color);
	}

	.text-error {
		color: var(--error-color);
	}

	.text-warning {
		color: var(--accent-color);
	}

	.text-primary {
		color: var(--primary-color);
	}

	.text-muted {
		color: var(--text-muted);
	}

	/* Layout structure */
	.dashboard-layout {
		display: grid;
		grid-template-columns: 380px 1fr;
		gap: 2rem;
		align-items: start;
	}

	.column-title {
		font-size: 1.125rem;
		color: var(--text-primary);
		margin-bottom: 1rem;
		padding-left: 0.25rem;
	}

	.services-list {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.service-card {
		background-color: var(--bg-card);
		border: 1px solid var(--border-color);
		border-radius: var(--radius-lg);
		padding: 1.25rem;
		text-align: left;
		width: 100%;
		display: flex;
		flex-direction: column;
		box-shadow: var(--shadow-sm);
		transition: var(--transition-all);
	}

	.service-card:hover {
		border-color: var(--primary-color);
		transform: translateY(-2px);
		box-shadow: var(--shadow-md);
	}

	.service-card.active {
		border-color: var(--primary-color);
		box-shadow: 0 0 0 3px rgba(1, 136, 202, 0.15);
		background-color: var(--primary-light);
	}

	.service-card.unhealthy {
		border-left: 4px solid var(--error-color);
	}

	.service-card-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 0.5rem;
	}

	.service-name {
		font-family: 'Outfit', sans-serif;
		font-weight: 600;
		font-size: 1rem;
		color: var(--text-primary);
	}

	.service-url {
		font-size: 0.75rem;
		color: var(--text-muted);
		word-break: break-all;
		margin-bottom: 0.875rem;
	}

	.service-metrics {
		display: flex;
		justify-content: space-between;
		font-size: 0.8125rem;
		color: var(--text-secondary);
	}

	.metric-item {
		display: flex;
		gap: 0.25rem;
	}

	.metric-item .label {
		color: var(--text-muted);
	}

	.metric-item .val {
		font-weight: 600;
	}

	.service-alert-pill {
		margin-top: 0.875rem;
		background-color: var(--error-light);
		color: var(--error-color);
		font-size: 0.75rem;
		font-weight: 700;
		padding: 0.375rem 0.625rem;
		border-radius: var(--radius-sm);
		text-align: center;
	}

	/* Chart container styling */
	.chart-card {
		background-color: var(--bg-card);
		border: 1px solid var(--border-color);
		border-radius: var(--radius-lg);
		padding: 1.5rem;
		box-shadow: var(--shadow-md);
		min-height: 480px;
		display: flex;
		flex-direction: column;
	}

	.chart-header {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		margin-bottom: 1.5rem;
		border-bottom: 1px solid var(--border-color);
		padding-bottom: 1rem;
	}

	.chart-title {
		font-size: 1.25rem;
		color: var(--text-primary);
	}

	.chart-subtitle {
		font-size: 0.8125rem;
		color: var(--text-muted);
		margin-top: 0.125rem;
	}

	.range-selector {
		display: flex;
		background-color: var(--bg-darker);
		border: 1px solid var(--border-color);
		border-radius: var(--radius-md);
		padding: 0.25rem;
	}

	.range-selector button {
		background: none;
		padding: 0.375rem 0.75rem;
		font-size: 0.75rem;
		font-weight: 600;
		color: var(--text-secondary);
		border-radius: var(--radius-sm);
	}

	.range-selector button.active {
		background-color: var(--primary-color);
		color: var(--bg-color);
	}

	.chart-container-wrapper {
		position: relative;
		flex: 1;
		min-height: 350px;
	}

	.chart-overlay {
		position: absolute;
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;
		background-color: rgba(255, 255, 255, 0.8);
		display: flex;
		flex-direction: column;
		justify-content: center;
		align-items: center;
		gap: 1rem;
		font-size: 0.875rem;
		color: var(--text-secondary);
		z-index: 10;
		border-radius: var(--radius-md);
	}

	.error-text {
		color: var(--error-color);
		font-weight: 500;
	}

	.spinner {
		width: 2rem;
		height: 2rem;
		border: 3px solid var(--border-color);
		border-top-color: var(--primary-color);
		border-radius: 50%;
		animation: spin 1s linear infinite;
	}

	.no-selection {
		display: flex;
		flex: 1;
		justify-content: center;
		align-items: center;
		color: var(--text-muted);
		text-align: center;
		padding: 3rem;
	}

	@keyframes spin {
		to {
			transform: rotate(360deg);
		}
	}

	@media (max-width: 1024px) {
		.dashboard-layout {
			grid-template-columns: 1fr;
		}

		.chart-card {
			min-height: 400px;
		}
	}
</style>
