<script lang="ts">
	import { goto } from '$app/navigation';

	let username = $state('');
	let password = $state('');
	let showPassword = $state(false);

	// Tambahan state untuk API
	let isLoading = $state(false);
	let errorMessage = $state('');

	const API_BASE_URL = import.meta.env.VITE_API_URL;

	// Logika Submit ke API Backend
	const handleSubmit = async (e: Event) => {
		e.preventDefault();
		isLoading = true;
		errorMessage = '';

		try {
			const res = await fetch(`${API_BASE_URL}/login`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ username, password })
			});

			if (!res.ok) {
				const errorText = await res.text();
				try {
					const errJson = JSON.parse(errorText);
					errorMessage = errJson.error || 'Login gagal.';
				} catch {
					errorMessage = errorText || 'Username atau password salah.';
				}
				isLoading = false;
				return;
			}

			const data = await res.json();

			// Simpan token ke LocalStorage
			localStorage.setItem('jwt_token', data.token);
			localStorage.setItem('user_role', data.role);

			// Arahkan ke Dashboard Admin
			goto('/admin/dashboard');
		} catch (err) {
			console.error('Network Error:', err);
			errorMessage = 'Tidak dapat terhubung ke server.';
		} finally {
			isLoading = false;
		}
	};
</script>

<div class="flex min-h-screen w-full items-center justify-center bg-slate-900 p-4">
	<a
		href="/"
		class="absolute top-6 left-6 flex items-center gap-2 text-sm font-semibold text-slate-400 transition-all hover:text-white"
	>
		<svg
			xmlns="http://www.w3.org/2000/svg"
			class="h-4 w-4"
			viewBox="0 0 24 24"
			fill="none"
			stroke="currentColor"
			stroke-width="3"><path d="m15 18-6-6 6-6" /></svg
		>
		Kembali
	</a>

	<div class="w-full max-w-md rounded-[2rem] bg-white p-8 shadow-2xl md:p-10">
		<div class="mb-10 text-center">
			<div class="mb-4 inline-flex h-16 w-16 items-center justify-center rounded-2xl bg-slate-100">
				<svg
					xmlns="http://www.w3.org/2000/svg"
					class="h-8 w-8 text-slate-800"
					viewBox="0 0 24 24"
					fill="none"
					stroke="currentColor"
					stroke-width="2"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" /></svg
				>
			</div>
			<h1 class="text-2xl font-black tracking-tight text-slate-800 uppercase">Admin Access</h1>
			<p class="text-sm font-medium text-slate-400">Sistem Manajemen Absensi</p>
		</div>

		{#if errorMessage}
			<div
				class="mb-6 rounded-2xl border border-red-100 bg-red-50 p-4 text-center text-xs font-black tracking-widest text-red-600 uppercase"
			>
				{errorMessage}
			</div>
		{/if}

		<form onsubmit={handleSubmit} class="space-y-6">
			<div class="flex flex-col gap-2">
				<label
					for="username"
					class="ml-1 text-xs font-black tracking-wider text-slate-500 uppercase"
					>Admin Username</label
				>
				<input
					type="text"
					id="username"
					bind:value={username}
					placeholder="Username Administrator"
					class="w-full rounded-2xl border-2 border-slate-100 bg-slate-50 px-5 py-4 font-medium transition-all outline-none focus:border-slate-800 focus:bg-white"
					required
				/>
			</div>

			<div class="flex flex-col gap-2">
				<label
					for="password"
					class="ml-1 text-xs font-black tracking-wider text-slate-500 uppercase"
					>Secure Password</label
				>
				<div class="relative">
					<input
						type={showPassword ? 'text' : 'password'}
						id="password"
						bind:value={password}
						placeholder="••••••••"
						class="w-full rounded-2xl border-2 border-slate-100 bg-slate-50 px-5 py-4 font-medium transition-all outline-none focus:border-slate-800 focus:bg-white"
						required
					/>
					<button
						type="button"
						onclick={() => (showPassword = !showPassword)}
						class="absolute top-1/2 right-4 -translate-y-1/2 text-slate-300"
					>
						{#if showPassword}
							<svg
								xmlns="http://www.w3.org/2000/svg"
								class="h-5 w-5"
								viewBox="0 0 24 24"
								fill="none"
								stroke="currentColor"
								stroke-width="2"
								><path
									d="M9.88 9.88 3 3m6.12 6.12a3 3 0 1 0 4.24 4.24m-4.24-4.24L13 13.56M10.73 5.08A10.43 10.43 0 0 1 12 5c7 0 10 7 10 7a13.16 13.16 0 0 1-1.67 2.68M6.61 6.61A13.52 13.52 0 0 0 2 12s3 7 10 7a9.74 9.74 0 0 0 5.39-1.61"
								/></svg
							>
						{:else}
							<svg
								xmlns="http://www.w3.org/2000/svg"
								class="h-5 w-5"
								viewBox="0 0 24 24"
								fill="none"
								stroke="currentColor"
								stroke-width="2"
								><path d="M2 12s3-7 10-7 10 7 10 7-3 7-10 7-10-7-10-7Z" /><circle
									cx="12"
									cy="12"
									r="3"
								/></svg
							>
						{/if}
					</button>
				</div>
			</div>

			<button
				type="submit"
				disabled={isLoading}
				class="flex w-full items-center justify-center gap-2 rounded-2xl bg-slate-800 py-4 text-sm font-black tracking-widest text-white uppercase shadow-xl transition-all hover:bg-black active:scale-95 disabled:bg-slate-500"
			>
				{#if isLoading}
					<span
						class="inline-block h-4 w-4 animate-spin rounded-full border-2 border-white border-t-transparent"
					></span>
				{:else}
					Authorize Login
				{/if}
			</button>
		</form>
	</div>
</div>
