<script lang="ts">
	import { goto } from '$app/navigation';

	const API_BASE_URL = import.meta.env.VITE_API_URL;

	let username = $state('');
	let password = $state('');
	let confirmPassword = $state('');
	let namaLengkap = $state('');
	let namaSekolah = $state('');
	let labelKataKunci = $state('');
	let kataKunci = $state('');
	let showPassword = $state(false);

	let isLoading = $state(false);
	let errorMessage = $state('');
	let successMessage = $state('');

	const getDeviceToken = () => {
		let token = localStorage.getItem('device_token');
		if (!token) {
			token = 'DEV-' + Math.random().toString(36).substring(2, 15) + Date.now().toString(36);
			localStorage.setItem('device_token', token);
		}
		return token;
	};

	const handleSubmit = async (e: Event) => {
		e.preventDefault();
		errorMessage = '';
		successMessage = '';

		if (password !== confirmPassword) {
			errorMessage = 'Kata sandi dan konfirmasi tidak sama.';
			return;
		}

		isLoading = true;
		try {
			const res = await fetch(`${API_BASE_URL}/register/siswa`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					username,
					password,
					nama_lengkap: namaLengkap,
					nama_sekolah: namaSekolah,
					label_kata_kunci: labelKataKunci,
					kata_kunci: kataKunci,
					device_token: getDeviceToken()
				})
			});

			const text = await res.text();
			if (!res.ok) {
				try {
					const j = JSON.parse(text);
					errorMessage = j.error || text;
				} catch {
					errorMessage = text || 'Pendaftaran gagal.';
				}
				return;
			}

			successMessage =
				'Pendaftaran berhasil. Akun menunggu persetujuan admin. Setelah disetujui, gunakan halaman login dengan perangkat ini.';
			username = '';
			password = '';
			confirmPassword = '';
			namaLengkap = '';
			namaSekolah = '';
			labelKataKunci = '';
			kataKunci = '';
		} catch {
			errorMessage = 'Tidak dapat terhubung ke server.';
		} finally {
			isLoading = false;
		}
	};
</script>

<div class="flex min-h-screen w-full items-center justify-center bg-brand-blue p-4 py-10">
	<div class="w-full max-w-lg rounded-[2rem] bg-white p-8 shadow-2xl md:p-10">
		<div class="mb-8 text-center">
			<h1 class="text-2xl font-black tracking-tight text-slate-800 uppercase">Daftar Siswa</h1>
			<p class="mt-2 text-sm font-medium text-slate-400">
				Isi data berikut. Admin akan mengaktifkan akun Anda.
			</p>
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
				class="mb-6 rounded-2xl border border-green-100 bg-green-50 p-4 text-center text-xs font-bold leading-relaxed text-green-800"
			>
				{successMessage}
			</div>
		{/if}

		<form onsubmit={handleSubmit} class="space-y-5">
			<div class="flex flex-col gap-2">
				<label class="ml-1 text-xs font-black tracking-wider text-slate-500 uppercase" for="nl"
					>Nama Lengkap</label
				>
				<input
					id="nl"
					type="text"
					bind:value={namaLengkap}
					class="w-full rounded-2xl border-2 border-slate-100 bg-slate-50 px-5 py-4 font-medium outline-none focus:border-brand-blue focus:bg-white"
					required
				/>
			</div>
			<div class="flex flex-col gap-2">
				<label class="ml-1 text-xs font-black tracking-wider text-slate-500 uppercase" for="ns"
					>Asal Sekolah</label
				>
				<input
					id="ns"
					type="text"
					bind:value={namaSekolah}
					class="w-full rounded-2xl border-2 border-slate-100 bg-slate-50 px-5 py-4 font-medium outline-none focus:border-brand-blue focus:bg-white"
					required
				/>
			</div>
			<div class="flex flex-col gap-2">
				<label class="ml-1 text-xs font-black tracking-wider text-slate-500 uppercase" for="un"
					>Nama pengguna</label
				>
				<input
					id="un"
					autocomplete="username"
					type="text"
					bind:value={username}
					class="w-full rounded-2xl border-2 border-slate-100 bg-slate-50 px-5 py-4 font-medium outline-none focus:border-brand-blue focus:bg-white"
					required
				/>
			</div>
			<div class="flex flex-col gap-2">
				<label class="ml-1 text-xs font-black tracking-wider text-slate-500 uppercase" for="pw"
					>Kata sandi</label
				>
				<div class="relative">
					<input
						id="pw"
						autocomplete="new-password"
						type={showPassword ? 'text' : 'password'}
						bind:value={password}
						class="w-full rounded-2xl border-2 border-slate-100 bg-slate-50 px-5 py-4 font-medium outline-none focus:border-brand-blue focus:bg-white"
						required
						minlength="6"
					/>
					<button
						type="button"
						onclick={() => (showPassword = !showPassword)}
						class="absolute top-1/2 right-4 -translate-y-1/2 text-slate-400"
					>
						{showPassword ? 'Sembunyikan' : 'Lihat'}
					</button>
				</div>
			</div>
			<div class="flex flex-col gap-2">
				<label class="ml-1 text-xs font-black tracking-wider text-slate-500 uppercase" for="pw2"
					>Konfirmasi kata sandi</label
				>
				<input
					id="pw2"
					autocomplete="new-password"
					type={showPassword ? 'text' : 'password'}
					bind:value={confirmPassword}
					class="w-full rounded-2xl border-2 border-slate-100 bg-slate-50 px-5 py-4 font-medium outline-none focus:border-brand-blue focus:bg-white"
					required
				/>
			</div>
			<div class="flex flex-col gap-2">
				<label class="ml-1 text-xs font-black tracking-wider text-slate-500 uppercase" for="cl"
					>Clue Keamanan (untuk reset password oleh admin)</label
				>
				<input
					id="cl"
					type="text"
					bind:value={labelKataKunci}
					placeholder="Contoh: Nama hewan peliharaan?"
					class="w-full rounded-2xl border-2 border-slate-100 bg-slate-50 px-5 py-4 font-medium outline-none focus:border-brand-blue focus:bg-white"
					required
				/>
			</div>
			<div class="flex flex-col gap-2">
				<label class="ml-1 text-xs font-black tracking-wider text-slate-500 uppercase" for="kk"
					>Jawaban Keamanan</label
				>
				<input
					id="kk"
					type="text"
					bind:value={kataKunci}
					class="w-full rounded-2xl border-2 border-slate-100 bg-slate-50 px-5 py-4 font-medium outline-none focus:border-brand-blue focus:bg-white"
					required
				/>
			</div>

			<button
				type="submit"
				disabled={isLoading}
				class="mt-2 w-full rounded-2xl bg-brand-blue py-4 text-sm font-black tracking-widest text-white uppercase shadow-xl shadow-blue-500/20 transition-all hover:bg-blue-700 disabled:bg-blue-300"
			>
				{isLoading ? 'Mengirim…' : 'Kirim Pendaftaran'}
			</button>
		</form>

		<p class="mt-8 text-center text-sm font-medium text-slate-500">
			Sudah punya akun?
			<button
				type="button"
				onclick={() => goto('/login/student')}
				class="font-black text-brand-blue hover:underline"
			>
				Masuk sebagai siswa
			</button>
		</p>
	</div>
</div>
