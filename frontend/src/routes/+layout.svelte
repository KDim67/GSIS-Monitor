<script>
	import '../global.css';
	import { page } from '$app/state';

	let { children } = $props();
</script>

<div class="app-layout">
	<header class="main-header">
		<div class="header-container">
			<div class="logo-area">
				<img src="/logo.png" alt="ΓΓΠΣΨΔ Logo" class="logo-img" />
				<h1 class="app-title">Σύστημα Παρακολούθησης Υπηρεσιών</h1>
			</div>
			<nav class="main-nav">
				<a href="/" class="nav-link" class:active={page.url.pathname === '/'}>
					Dashboard
				</a>
				<a href="/services" class="nav-link" class:active={page.url.pathname.startsWith('/services')}>
					Υπηρεσίες
				</a>
				<a href="/alerts" class="nav-link" class:active={page.url.pathname.startsWith('/alerts')}>
					Ειδοποιήσεις
				</a>
			</nav>
		</div>
	</header>

	<main class="main-content">
		{@render children()}
	</main>
</div>

<style>
	.app-layout {
		display: flex;
		flex-direction: column;
		min-height: 100vh;
	}

	.main-header {
		background-color: var(--bg-color);
		border-bottom: 2px solid var(--primary-color);
		box-shadow: var(--shadow-sm);
	}

	.header-container {
		max-width: 100%;
		padding: 1.25rem 3rem;
		display: flex;
		justify-content: space-between;
		align-items: center;
		flex-wrap: wrap;
		gap: 1.5rem;
	}
	.logo-area {
		display: flex;
		align-items: center;
		gap: 1.25rem;
	}

	.logo-img {
		height: 54px;
		width: auto;
	}

	.app-title {
		font-size: 1.5rem;
		font-weight: 700;
		color: var(--primary-color);
		letter-spacing: -0.02em;
		line-height: 1.2;
	}

	.main-nav {
		display: flex;
		gap: 0.5rem;
	}

	.nav-link {
		padding: 0.5rem 1rem;
		font-weight: 600;
		font-size: 0.875rem;
		color: var(--text-secondary);
		border-radius: var(--radius-md);
		border-bottom: 3px solid transparent;
		transition: var(--transition-all);
	}

	.nav-link:hover {
		color: var(--primary-color);
		background-color: var(--bg-darker);
	}

	.nav-link.active {
		color: var(--primary-color);
		background-color: var(--primary-light);
		border-bottom-color: var(--primary-color);
	}

	.main-content {
		flex: 1;
	}


	@media (max-width: 768px) {
		.header-container {
			flex-direction: column;
			align-items: flex-start;
		}

		.main-nav {
			width: 100%;
			justify-content: space-between;
		}
	}
</style>
