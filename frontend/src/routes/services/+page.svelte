<script>
	import { onMount } from 'svelte';

	const API_BASE = 'http://localhost:8000';

	// Services state using Svelte 5 reactive state
	let services = $state([]);
	let error = $state(null);
	
	// Form state
	let isDialogOpen = $state(false);
	let isEditMode = $state(false);
	let currentServiceId = $state(null);
	
	let formName = $state('');
	let formUrl = $state('');
	let formVerificationKeyword = $state('');
	let formExclusionKeyword = $state('');
	let formSkipTlsVerify = $state(false);
	let formIsActive = $state(true);

	// Fetch all services
	async function loadServices() {
		try {
			const res = await fetch(`${API_BASE}/api/services`);
			if (!res.ok) throw new Error('Failed to load services');
			services = await res.json();
		} catch (e) {
			error = e.message;
		}
	}

	// Open dialog for adding a service
	function openAddDialog() {
		isEditMode = false;
		currentServiceId = null;
		formName = '';
		formUrl = '';
		formVerificationKeyword = '';
		formExclusionKeyword = '';
		formSkipTlsVerify = false;
		formIsActive = true;
		isDialogOpen = true;
	}

	// Open dialog for editing a service
	function openEditDialog(service) {
		isEditMode = true;
		currentServiceId = service.id;
		formName = service.name;
		formUrl = service.url;
		formVerificationKeyword = service.verification_keyword || '';
		formExclusionKeyword = service.exclusion_keyword || '';
		formSkipTlsVerify = service.skip_tls_verify;
		formIsActive = service.is_active;
		isDialogOpen = true;
	}

	// Submit add/edit form
	async function handleSubmit(e) {
		e.preventDefault();
		const payload = {
			name: formName,
			url: formUrl,
			verification_keyword: formVerificationKeyword || null,
			exclusion_keyword: formExclusionKeyword || null,
			skip_tls_verify: formSkipTlsVerify,
			is_active: formIsActive
		};

		try {
			let res;
			if (isEditMode) {
				res = await fetch(`${API_BASE}/api/services/${currentServiceId}`, {
					method: 'PUT',
					headers: { 'Content-Type': 'application/json' },
					body: JSON.stringify(payload)
				});
			} else {
				res = await fetch(`${API_BASE}/api/services`, {
					method: 'POST',
					headers: { 'Content-Type': 'application/json' },
					body: JSON.stringify(payload)
				});
			}

			if (!res.ok) {
				const details = await res.json();
				throw new Error(details.detail || 'Save failed');
			}

			isDialogOpen = false;
			await loadServices();
		} catch (e) {
			alert(`Σφάλμα: ${e.message}`);
		}
	}

	// Delete service
	async function handleDelete(id, name) {
		if (!confirm(`Είστε σίγουροι ότι θέλετε να διαγράψετε την υπηρεσία "${name}";`)) return;

		try {
			const res = await fetch(`${API_BASE}/api/services/${id}`, {
				method: 'DELETE'
			});
			if (!res.ok) throw new Error('Delete failed');
			await loadServices();
		} catch (e) {
			alert(`Σφάλμα: ${e.message}`);
		}
	}

	onMount(loadServices);
</script>

<div class="container">
	<div class="services-header">
		<div>
			<h2>Διαχείριση Ψηφιακών Πυλών</h2>
			<p>Προσθέστε, επεξεργαστείτε ή απενεργοποιήστε πύλες ελέγχου.</p>
		</div>
		<button class="btn btn-primary" onclick={openAddDialog}>
			Νέα Υπηρεσία
		</button>
	</div>

	{#if error}
		<div class="error-banner">
			Σφάλμα επικοινωνίας: {error}
		</div>
	{/if}

	<!-- Services Table -->
	<div class="table-wrapper">
		{#if services.length === 0}
			<div class="empty-state">
				Δεν υπάρχουν καταχωρημένες υπηρεσίες.
			</div>
		{:else}
			<table>
				<thead>
					<tr>
						<th>Όνομα</th>
						<th>Διεύθυνση (URL)</th>
						<th>Keyword Επαλήθευσης</th>
						<th>Keyword Αποκλεισμού</th>
						<th class="text-center">Παράκαμψη TLS</th>
						<th class="text-center">Κατάσταση</th>
						<th class="text-right">Ενέργειες</th>
					</tr>
				</thead>
				<tbody>
					{#each services as service}
						<tr>
							<td class="font-bold">{service.name}</td>
							<td class="text-muted text-break">{service.url}</td>
							<td>{service.verification_keyword || '-'}</td>
							<td>{service.exclusion_keyword || '-'}</td>
							<td class="text-center">
								<span class="badge" class:badge-warning={service.skip_tls_verify} class:badge-secondary={!service.skip_tls_verify}>
									{service.skip_tls_verify ? 'ΝΑΙ' : 'ΟΧΙ'}
								</span>
							</td>
							<td class="text-center">
								<span class="status-badge" class:healthy={service.is_active} class:inactive={!service.is_active}>
									<span class="status-dot" class:healthy={service.is_active} class:inactive={!service.is_active}></span>
									{service.is_active ? 'ΕΝΕΡΓΗ' : 'ΑΝΕΝΕΡΓΗ'}
								</span>
							</td>
							<td class="text-right actions-cell">
								<button class="btn-icon" onclick={() => openEditDialog(service)} title="Επεξεργασία">
									<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
										<path d="M11 5H6a2 2 0 0 0-2 2v11a2 2 0 0 0 2 2h11a2 2 0 0 0 2-2v-5"></path>
										<path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"></path>
									</svg>
								</button>
								<button class="btn-icon text-error-icon" onclick={() => handleDelete(service.id, service.name)} title="Διαγραφή">
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

	<!-- Modal Dialog (Overlay) -->
	{#if isDialogOpen}
		<div class="modal-overlay">
			<div class="modal-card">
				<div class="modal-header">
					<h3>{isEditMode ? 'Επεξεργασία Υπηρεσίας' : 'Προσθήκη Νέας Υπηρεσίας'}</h3>
					<button class="btn-close" onclick={() => isDialogOpen = false}>&times;</button>
				</div>
				<form onsubmit={handleSubmit}>
					<div class="form-group">
						<label for="name">Όνομα Υπηρεσίας</label>
						<input type="text" id="name" bind:value={formName} required placeholder="π.χ. e-Paravolo" />
					</div>
					<div class="form-group">
						<label for="url">Διεύθυνση (URL)</label>
						<input type="url" id="url" bind:value={formUrl} required placeholder="https://example.gov.gr" />
					</div>
					<div class="form-row">
						<div class="form-group">
							<label for="verification_keyword">Keyword Επαλήθευσης (Προαιρετικό)</label>
							<input type="text" id="verification_keyword" bind:value={formVerificationKeyword} placeholder="Έλεγχος ύπαρξης στο HTML" />
						</div>
						<div class="form-group">
							<label for="exclusion_keyword">Keyword Αποκλεισμού (Προαιρετικό)</label>
							<input type="text" id="exclusion_keyword" bind:value={formExclusionKeyword} placeholder="Σφάλμα αν βρεθεί (π.χ. maintenance)" />
						</div>
					</div>

					<div class="form-checkboxes">
						<label class="checkbox-label">
							<input type="checkbox" bind:checked={formSkipTlsVerify} />
							<span>Παράκαμψη ελέγχου TLS/SSL (InsecureSkipVerify)</span>
						</label>
						<label class="checkbox-label">
							<input type="checkbox" bind:checked={formIsActive} />
							<span>Ενεργή παρακολούθηση (Active)</span>
						</label>
					</div>

					<div class="modal-actions">
						<button type="button" class="btn btn-secondary" onclick={() => isDialogOpen = false}>Ακύρωση</button>
						<button type="submit" class="btn btn-primary">Αποθήκευση</button>
					</div>
				</form>
			</div>
		</div>
	{/if}
</div>

<style>
	.services-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 2rem;
	}

	.services-header h2 {
		font-size: 1.75rem;
		color: var(--primary-color);
	}

	.services-header p {
		color: var(--text-secondary);
		font-size: 0.9375rem;
		margin-top: 0.25rem;
	}

	.btn {
		padding: 0.625rem 1.25rem;
		font-size: 0.875rem;
		font-weight: 600;
		border-radius: var(--radius-md);
		display: inline-flex;
		align-items: center;
		justify-content: center;
	}

	.btn-primary {
		background-color: var(--primary-color);
		color: var(--bg-color);
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

	.error-banner {
		background-color: var(--error-light);
		color: var(--error-color);
		border: 1px solid var(--error-color);
		padding: 1rem;
		border-radius: var(--radius-md);
		margin-bottom: 1.5rem;
		font-weight: 500;
	}

	.table-wrapper {
		background-color: var(--bg-card);
		border: 1px solid var(--border-color);
		border-radius: var(--radius-lg);
		box-shadow: var(--shadow-sm);
		overflow: hidden;
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

	.text-center {
		text-align: center;
	}

	.text-right {
		text-align: right;
	}

	.text-break {
		word-break: break-all;
	}

	.badge {
		display: inline-block;
		padding: 0.125rem 0.5rem;
		font-size: 0.6875rem;
		font-weight: 700;
		border-radius: var(--radius-sm);
	}

	.badge-warning {
		background-color: var(--warning-light);
		color: var(--warning-color);
	}

	.badge-secondary {
		background-color: var(--bg-darker);
		color: var(--text-secondary);
		border: 1px solid var(--border-color);
	}

	/* Icon buttons */
	.actions-cell {
		display: flex;
		justify-content: flex-end;
		gap: 0.5rem;
	}

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

	.btn-icon:hover {
		background-color: var(--bg-darker);
		color: var(--primary-color);
	}

	.btn-icon.text-error-icon:hover {
		background-color: var(--error-light);
		color: var(--error-color);
	}

	/* Modal overlay & card */
	.modal-overlay {
		position: fixed;
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;
		background-color: rgba(15, 23, 42, 0.4);
		backdrop-filter: blur(4px);
		display: flex;
		justify-content: center;
		align-items: center;
		z-index: 50;
	}

	.modal-card {
		background-color: var(--bg-card);
		border-radius: var(--radius-lg);
		box-shadow: var(--shadow-lg);
		width: 100%;
		max-width: 580px;
		overflow: hidden;
		display: flex;
		flex-direction: column;
		border: 1px solid var(--border-color);
		animation: modalFadeIn 0.2s ease-out;
	}

	.modal-header {
		padding: 1.25rem 1.5rem;
		border-bottom: 1px solid var(--border-color);
		display: flex;
		justify-content: space-between;
		align-items: center;
	}

	.modal-header h3 {
		font-size: 1.125rem;
		color: var(--text-primary);
	}

	.btn-close {
		background: none;
		font-size: 1.5rem;
		color: var(--text-muted);
		line-height: 1;
		padding: 0.25rem;
		transition: var(--transition-all);
	}

	.btn-close:hover {
		color: var(--error-color);
	}

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

	.form-checkboxes {
		margin: 1.5rem 0;
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
	}

	.checkbox-label {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		cursor: pointer;
		font-size: 0.875rem;
		font-weight: 500;
		color: var(--text-secondary);
		margin-bottom: 0;
	}

	.checkbox-label input {
		width: auto;
		cursor: pointer;
	}

	.modal-actions {
		display: flex;
		justify-content: flex-end;
		gap: 0.75rem;
		border-top: 1px solid var(--border-color);
		padding-top: 1.25rem;
		margin-top: 1.5rem;
	}

	@keyframes modalFadeIn {
		from {
			opacity: 0;
			transform: scale(0.95);
		}
		to {
			opacity: 1;
			transform: scale(1);
		}
	}

	@media (max-width: 640px) {
		.form-row {
			grid-template-columns: 1fr;
			gap: 0;
		}
	}
</style>
