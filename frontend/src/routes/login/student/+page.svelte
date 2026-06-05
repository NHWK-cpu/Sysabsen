<script lang="ts">
	import { goto } from '$app/navigation';

	let username = $state('');
	let password = $state('');
	let showPassword = $state(false);

	// Tambahan state untuk API
	let isLoading = $state(false);
	let errorMessage = $state('');
	let infoMessage = $state('');

	const API_BASE_URL = import.meta.env.VITE_API_URL;

	// Logika Device Token
	const getDeviceToken = () => {
		let token = localStorage.getItem('device_token');
		if (!token) {
			token = 'DEV-' + Math.random().toString(36).substring(2, 15) + Date.now().toString(36);
			localStorage.setItem('device_token', token);
		}
		return token;
	};

	// Logika Submit ke API Backend
	const handleSubmit = async (e: Event) => {
		e.preventDefault();
		isLoading = true;
		errorMessage = '';
		infoMessage = '';

		try {
			const payload = {
				username,
				password,
				device_token: getDeviceToken()
			};

			const res = await fetch(`${API_BASE_URL}/login/siswa`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(payload)
			});

			if (!res.ok) {
				const errorText = await res.text();
				// 403 untuk status Pending/Ditolak dari Admin
				if (res.status === 403) {
					infoMessage = errorText;
				} else {
					try {
						const errJson = JSON.parse(errorText);
						errorMessage = errJson.error || 'Gagal masuk.';
					} catch {
						errorMessage = errorText || 'Nama pengguna atau kata sandi salah.';
					}
				}
				isLoading = false;
				return;
			}

			const data = await res.json();

			// Simpan token dan rute ke scanner
			localStorage.setItem('jwt_token', data.data.token);
			localStorage.setItem('user_role', data.data.role);

			goto('/student/scanner');
		} catch (err) {
			console.error('Network Error:', err);
			errorMessage = 'Tidak dapat terhubung ke server.';
		} finally {
			isLoading = false;
		}
	};
</script>

<div class="flex min-h-screen w-full items-center justify-center bg-brand-blue p-4">
	<div class="w-full max-w-md rounded-[2rem] bg-white p-8 shadow-2xl md:p-10">
		<div class="mb-10 text-center">
			<div class="mb-4 inline-flex h-16 w-16 items-center justify-center rounded-2xl bg-blue-50">
				<svg
					xmlns="http://www.w3.org/2000/svg"
					class="h-8 w-8 text-brand-blue"
					viewBox="0 0 24 24"
					fill="none"
					stroke="currentColor"
					stroke-width="2"
					stroke-linecap="round"
					stroke-linejoin="round"
					><path d="M19 21v-2a4 4 0 0 0-4-4H9a4 4 0 0 0-4 4v2" /><circle
						cx="12"
						cy="7"
						r="4"
					/></svg
				>
			</div>
			<h1 class="text-2xl font-black tracking-tight text-slate-800 uppercase">Masuk siswa</h1>
			<p class="text-sm font-medium text-slate-400">Masuk untuk melakukan absensi</p>
		</div>

		{#if errorMessage}
			<div
				class="mb-6 rounded-2xl border border-red-100 bg-red-50 p-4 text-center text-xs font-black tracking-widest text-red-600 uppercase"
			>
				{errorMessage}
			</div>
		{/if}

		{#if infoMessage}
			<div
				class="mb-6 rounded-2xl border border-amber-100 bg-amber-50 p-4 text-center text-xs leading-relaxed font-black tracking-widest text-amber-700 uppercase"
			>
				{infoMessage}
			</div>
		{/if}

		<p class="mb-6 text-center text-sm font-medium text-slate-500">
			Belum punya akun?
			<a href="/register/siswa" class="font-black text-brand-blue hover:underline">Daftar siswa</a>
		</p>

		<form onsubmit={handleSubmit} class="space-y-6">
			<div class="flex flex-col gap-2">
				<label
					for="username"
					class="ml-1 text-xs font-black tracking-wider text-slate-500 uppercase">Nama pengguna</label
				>
				<input
					type="text"
					id="username"
					bind:value={username}
					placeholder="Nama pengguna"
					class="w-full rounded-2xl border-2 border-slate-100 bg-slate-50 px-5 py-4 font-medium transition-all outline-none focus:border-brand-blue focus:bg-white"
					required
				/>
			</div>

			<div class="flex flex-col gap-2">
				<div class="flex items-center justify-between px-1">
					<label for="password" class="text-xs font-black tracking-wider text-slate-500 uppercase"
						>Kata sandi</label
					>
				</div>
				<div class="relative">
					<input
						type={showPassword ? 'text' : 'password'}
						id="password"
						bind:value={password}
						placeholder="••••••••"
						class="w-full rounded-2xl border-2 border-slate-100 bg-slate-50 px-5 py-4 font-medium transition-all outline-none focus:border-brand-blue focus:bg-white"
						required
					/>
					<button
						type="button"
						onclick={() => (showPassword = !showPassword)}
						class="absolute top-1/2 right-4 -translate-y-1/2 text-slate-300 hover:text-brand-blue"
					>
						{#if showPassword}
							<svg
								xmlns="http://www.w3.org/2000/svg"
								class="h-5 w-5"
								viewBox="0 0 24 24"
								fill="none"
								stroke="currentColor"
								stroke-width="2"
								stroke-linecap="round"
								stroke-linejoin="round"
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
								stroke-linecap="round"
								stroke-linejoin="round"
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
				class="flex w-full items-center justify-center gap-2 rounded-2xl bg-brand-blue py-4 text-sm font-black tracking-widest text-white uppercase shadow-xl shadow-blue-500/20 transition-all hover:bg-blue-700 active:scale-95 disabled:bg-blue-300"
			>
				{#if isLoading}
					<span
						class="inline-block h-4 w-4 animate-spin rounded-full border-2 border-white border-t-transparent"
					></span>
				{:else}
					Masuk
				{/if}
			</button>
		</form>

		<div class="mt-8 text-center">
			<p class="text-xs font-medium text-slate-400">© 2026 sistem absensi sekolah</p>
		</div>
	</div>
</div>
