<script lang="ts">
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';

	// State untuk form
	let newPassword = $state('');
	let confirmPassword = $state('');
	let showPassword = $state(false);

	// State untuk API dan Token
	let token = $state('');
	let isLoading = $state(false);
	let successMessage = $state('');
	let errorMessage = $state('');

	const API_BASE_URL = import.meta.env.VITE_API_URL;

	// Menangkap token dari URL persis saat halaman dimuat
	onMount(() => {
		// Mengambil nilai ?token=XYZ dari URL
		token = $page.url.searchParams.get('token') || '';

		if (!token) {
			errorMessage = 'Akses ditolak: Token reset password tidak ditemukan di URL.';
		}
	});

	const handleSubmit = async (e: Event) => {
		e.preventDefault();
		isLoading = true;
		errorMessage = '';
		successMessage = '';

		// Validasi kecocokan password di frontend
		if (newPassword !== confirmPassword) {
			errorMessage = 'Konfirmasi kata sandi tidak cocok!';
			isLoading = false;
			return;
		}

		try {
			// Tembak API Tahap 2: ExecuteResetPassword
			const res = await fetch(`${API_BASE_URL}/guru/reset-password`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					token: token,
					new_password: newPassword
				})
			});

			if (!res.ok) {
				const errorText = await res.text();
				try {
					const errJson = JSON.parse(errorText);
					errorMessage = errJson.error || 'Gagal mengatur ulang kata sandi.';
				} catch {
					errorMessage = errorText || 'Terjadi kesalahan pada server.';
				}
				return;
			}

			const data = await res.json();
			successMessage = data.message || 'Kata sandi berhasil diubah!';

			// Redirect otomatis ke halaman login guru setelah 3 detik
			setTimeout(() => {
				goto('/login/teacher'); // Sesuaikan rute ini jika URL login gurumu berbeda
			}, 3000);
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
			<div
				class="mb-4 inline-flex h-16 w-16 items-center justify-center rounded-2xl bg-amber-50 text-amber-500"
			>
				<svg
					xmlns="http://www.w3.org/2000/svg"
					class="h-8 w-8"
					viewBox="0 0 24 24"
					fill="none"
					stroke="currentColor"
					stroke-width="2"
					stroke-linecap="round"
					stroke-linejoin="round"
				>
					<rect width="18" height="11" x="3" y="11" rx="2" ry="2" />
					<path d="M7 11V7a5 5 0 0 1 10 0v4" />
				</svg>
			</div>
			<h1 class="text-2xl font-black tracking-tight text-slate-800 uppercase">Buat Sandi Baru</h1>
			<p class="text-sm font-medium text-slate-400">Masukkan kata sandi baru untuk akun Anda</p>
		</div>

		{#if errorMessage}
			<div
				class="mb-6 rounded-2xl border border-red-100 bg-red-50 p-4 text-center text-xs font-black tracking-widest text-red-600 uppercase"
			>
				{errorMessage}
			</div>
		{/if}

		{#if successMessage}
			<div
				class="mb-6 rounded-2xl border border-green-100 bg-green-50 p-6 text-center text-xs leading-relaxed font-black tracking-widest text-green-600 uppercase"
			>
				{successMessage} <br /><br /> Mengalihkan ke halaman login...
			</div>
		{:else if token}
			<form onsubmit={handleSubmit} class="space-y-6">
				<div class="flex flex-col gap-2">
					<label
						for="newPassword"
						class="ml-1 text-xs font-black tracking-wider text-slate-500 uppercase"
						>Kata sandi baru</label
					>
					<div class="relative">
						<input
							type={showPassword ? 'text' : 'password'}
							id="newPassword"
							bind:value={newPassword}
							placeholder="••••••••"
							class="w-full rounded-2xl border-2 border-slate-100 bg-slate-50 px-5 py-4 font-medium transition-all outline-none focus:border-amber-400/50 focus:bg-white"
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

				<div class="flex flex-col gap-2">
					<label
						for="confirmPassword"
						class="ml-1 text-xs font-black tracking-wider text-slate-500 uppercase"
						>Konfirmasi kata sandi</label
					>
					<input
						type={showPassword ? 'text' : 'password'}
						id="confirmPassword"
						bind:value={confirmPassword}
						placeholder="••••••••"
						class="w-full rounded-2xl border-2 border-slate-100 bg-slate-50 px-5 py-4 font-medium transition-all outline-none focus:border-amber-400/50 focus:bg-white"
						required
					/>
				</div>

				<button
					type="submit"
					disabled={isLoading}
					class="flex w-full items-center justify-center gap-2 rounded-2xl bg-slate-900 py-4 text-sm font-black tracking-widest text-white uppercase shadow-xl shadow-slate-900/20 transition-all hover:bg-black active:scale-95 disabled:bg-slate-300"
				>
					{#if isLoading}
						<span
							class="inline-block h-4 w-4 animate-spin rounded-full border-2 border-white border-t-transparent"
						></span>
					{:else}
						Simpan kata sandi
					{/if}
				</button>
			</form>
		{/if}
	</div>
</div>

<style>
	.scale-in-center {
		animation: scale-in-center 0.3s cubic-bezier(0.25, 0.46, 0.45, 0.94) both;
	}
	@keyframes scale-in-center {
		0% {
			transform: scale(0.9);
			opacity: 0;
		}
		100% {
			transform: scale(1);
			opacity: 1;
		}
	}
</style>
