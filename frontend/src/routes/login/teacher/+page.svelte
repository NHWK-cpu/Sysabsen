<script lang="ts">
	import { goto } from '$app/navigation';

	let username = $state('');
	let password = $state('');
	let showPassword = $state(false);

	// Tambahan state untuk API Login
	let isLoading = $state(false);
	let errorMessage = $state('');

	// --- STATE UNTUK MODAL LUPA PASSWORD ---
	let showForgotModal = $state(false);
	let forgotUsername = $state(''); // <-- UBAH INI
	let forgotLoading = $state(false);
	let forgotMessage = $state('');
	let forgotError = $state('');

	const API_BASE_URL = import.meta.env.VITE_API_URL;

	// Logika Submit ke API Backend (Login)
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

			localStorage.setItem('jwt_token', data.token);
			localStorage.setItem('user_role', data.role);

			goto('/teacher/dashboard');
		} catch (err) {
			console.error('Network Error:', err);
			errorMessage = 'Tidak dapat terhubung ke server.';
		} finally {
			isLoading = false;
		}
	};

	// --- LOGIKA SUBMIT LUPA PASSWORD ---
	const handleForgotPassword = async (e: Event) => {
		e.preventDefault();
		forgotLoading = true;
		forgotMessage = '';
		forgotError = '';

		try {
			const res = await fetch(`${API_BASE_URL}/guru/forgot-password`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				// <-- UBAH INI (Kirim username ke backend)
				body: JSON.stringify({ username: forgotUsername })
			});

			if (!res.ok) {
				const errorText = await res.text();
				try {
					const errJson = JSON.parse(errorText);
					forgotError = errJson.error || 'Terjadi kesalahan.';
				} catch {
					forgotError = 'Gagal mengirim permintaan.';
				}
				return;
			}

			const data = await res.json();
			forgotMessage = data.message || 'Link reset password telah dikirim.';
			forgotUsername = ''; // <-- Kosongkan input setelah sukses
		} catch (err) {
			console.error('Network Error:', err);
			forgotError = 'Tidak dapat terhubung ke server.';
		} finally {
			forgotLoading = false;
		}
	};
</script>

<div class="flex min-h-screen w-full items-center justify-center bg-brand-blue p-4">
	<a
		href="/"
		class="absolute top-6 left-6 flex items-center gap-2 text-sm font-semibold text-white/70 transition-all hover:text-white"
	>
		<svg
			xmlns="http://www.w3.org/2000/svg"
			class="h-4 w-4"
			viewBox="0 0 24 24"
			fill="none"
			stroke="currentColor"
			stroke-width="3"
			stroke-linecap="round"
			stroke-linejoin="round"><path d="m15 18-6-6 6-6" /></svg
		>
		Kembali
	</a>

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
					><path d="M4 19.5v-15A2.5 2.5 0 0 1 6.5 2H20v20H6.5a2.5 2.5 0 0 1-2.5-2.5Z" /><path
						d="M8 7h6"
					/><path d="M8 11h8" /></svg
				>
			</div>
			<h1 class="text-2xl font-black tracking-tight text-slate-800 uppercase">Login Teacher</h1>
			<p class="text-sm font-medium text-slate-400">Masuk untuk mengelola absensi kelas</p>
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
					class="ml-1 text-xs font-black tracking-wider text-slate-500 uppercase">Username</label
				>
				<input
					type="text"
					id="username"
					bind:value={username}
					placeholder="Masukkan Username"
					class="w-full rounded-2xl border-2 border-slate-100 bg-slate-50 px-5 py-4 font-medium transition-all outline-none focus:border-brand-blue focus:bg-white"
					required
				/>
			</div>

			<div class="flex flex-col gap-2">
				<div class="flex items-center justify-between px-1">
					<label for="password" class="text-xs font-black tracking-wider text-slate-500 uppercase"
						>Password</label
					>
					<button
						type="button"
						onclick={() => (showForgotModal = true)}
						class="text-xs font-extrabold text-brand-blue hover:text-blue-700"
					>
						Lupa?
					</button>
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
				class="flex w-full items-center justify-center gap-2 rounded-2xl bg-brand-blue py-4 text-sm font-black tracking-widest text-white uppercase shadow-xl shadow-blue-500/20 transition-all hover:bg-blue-700 active:scale-95 disabled:bg-blue-300"
			>
				{#if isLoading}
					<span
						class="inline-block h-4 w-4 animate-spin rounded-full border-2 border-white border-t-transparent"
					></span>
				{:else}
					Sign In as Teacher
				{/if}
			</button>
		</form>
	</div>

	{#if showForgotModal}
		<div
			class="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/70 p-4 backdrop-blur-md"
		>
			<div
				class="scale-in-center w-full max-w-sm rounded-[2.5rem] bg-white p-8 text-center shadow-2xl"
			>
				<h3 class="mb-2 text-xl font-black tracking-tight text-slate-900 uppercase">
					Lupa Password?
				</h3>
				<p class="mb-6 text-xs leading-relaxed font-medium text-slate-500">
					Masukkan Username Anda. Kami akan mengirimkan tautan pemulihan ke alamat email yang
					terhubung dengan akun tersebut.
				</p>

				{#if forgotMessage}
					<div
						class="mb-6 rounded-2xl border border-green-100 bg-green-50 p-3 text-xs font-black tracking-widest text-green-600 uppercase"
					>
						{forgotMessage}
					</div>
				{/if}

				{#if forgotError}
					<div
						class="mb-6 rounded-2xl border border-red-100 bg-red-50 p-3 text-xs font-black tracking-widest text-red-600 uppercase"
					>
						{forgotError}
					</div>
				{/if}

				<form onsubmit={handleForgotPassword} class="space-y-4 text-left">
					<div class="flex flex-col gap-2">
						<label class="ml-1 text-[10px] font-black tracking-widest text-slate-400 uppercase"
							>Username</label
						>
						<input
							type="text"
							bind:value={forgotUsername}
							placeholder="Masukkan Username"
							class="w-full rounded-2xl border-2 border-slate-100 bg-slate-50 px-5 py-4 font-medium transition-all outline-none focus:border-brand-blue focus:bg-white"
							required
						/>
					</div>

					<div class="flex gap-2 pt-4">
						<button
							type="button"
							onclick={() => {
								showForgotModal = false;
								forgotMessage = '';
								forgotError = '';
							}}
							class="flex-1 rounded-2xl bg-slate-100 py-4 text-[10px] font-black tracking-widest text-slate-500 uppercase transition-all hover:bg-slate-200"
						>
							Tutup
						</button>
						<button
							type="submit"
							disabled={forgotLoading}
							class="flex-1 rounded-2xl bg-brand-blue py-4 text-[10px] font-black tracking-widest text-white uppercase shadow-lg shadow-blue-500/20 transition-all hover:bg-blue-700 disabled:bg-blue-300"
						>
							{#if forgotLoading}
								<span
									class="inline-block h-3 w-3 animate-spin rounded-full border-2 border-white border-t-transparent"
								></span>
							{:else}
								Kirim Link
							{/if}
						</button>
					</div>
				</form>
			</div>
		</div>
	{/if}
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
