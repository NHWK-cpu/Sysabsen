<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { browser } from '$app/environment';
	import { goto } from '$app/navigation';

	let scanResult = $state<string | null>(null);
	let scannerState = $state<'loading' | 'scanning' | 'success' | 'error'>('loading');
	let html5Qrcode: any;
	let errorMessage = $state<string>('');

	// PERUBAHAN: Ganti id menjadi school untuk menampung Asal Sekolah
	let student = $state({ name: 'Memuat...', school: '...', role: 'Siswa' });

	onMount(async () => {
		if (browser) {
			const token = localStorage.getItem('jwt_token');

			if (!token) {
				goto('/login/student');
				return;
			}

			try {
				// Bongkar (Decode) bagian payload dari JWT
				const payloadBase64 = token.split('.')[1];
				// atob() adalah fungsi bawaan browser untuk decode Base64
				const decodedPayload = JSON.parse(atob(payloadBase64));

				// PERUBAHAN: Masukkan data asli ke state student!
				// Pastikan key-nya sesuai dengan yang digenerate backend di dalam token
				student = {
					name: decodedPayload.nama_lengkap || decodedPayload.username || 'Siswa',
					school: decodedPayload.nama_sekolah || 'Asal Sekolah Tidak Diketahui',
					role: 'Siswa'
				};
			} catch (e) {
				console.error('Token tidak valid atau korup');
				localStorage.removeItem('jwt_token');
				goto('/login/student');
				return;
			}
			try {
				const { Html5Qrcode } = await import('html5-qrcode');
				html5Qrcode = new Html5Qrcode('qr-reader');
				startScanner();
			} catch (err: any) {
				errorMessage = 'Gagal memuat modul kamera.';
				scannerState = 'error';
			}
		}
	});

	onDestroy(() => {
		if (html5Qrcode && html5Qrcode.isScanning) {
			html5Qrcode.stop().catch(console.error);
		}
	});

	const startScanner = async () => {
		if (!html5Qrcode) return;
		scannerState = 'loading';
		scanResult = null;

		const config = { fps: 10, qrbox: { width: 250, height: 250 }, aspectRatio: 1.0 };

		const onSuccess = async (decodedText: string) => {
			html5Qrcode.stop(); // Matikan kamera segera setelah dapat QR
			scannerState = 'loading'; // Ubah state ke loading sambil nunggu GPS & API
			scanResult = null;

			// 1. Ambil Token Login User (Sesuaikan dengan tempat kamu menyimpan token JWT saat login)
			const userToken = localStorage.getItem('jwt_token');

			// 2. Ekstrak SesiID dari QR Token
			// (Karena backend butuh sesi_id, kita bongkar saja dari token QR-nya)
			let sesiId = 0;
			try {
				const payloadBase64 = decodedText.split('.')[1];
				const decodedPayload = JSON.parse(atob(payloadBase64));
				sesiId = parseInt(decodedPayload.sesi_id);
			} catch (e) {
				errorMessage = 'Format kode QR tidak valid!';
				scannerState = 'error';
				return;
			}

			// 3. Minta Akses GPS (Wajib untuk Geofencing)
			if (!navigator.geolocation) {
				errorMessage = 'GPS tidak didukung di browser ini.';
				scannerState = 'error';
				return;
			}

			navigator.geolocation.getCurrentPosition(
				async (position) => {
					const lat = position.coords.latitude;
					const lon = position.coords.longitude;

					// 4. Kirim Data Lengkap ke Backend (Vercel Proxy)
					try {
						const API_URL = import.meta.env.VITE_API_URL;
						const res = await fetch(`${API_URL}/siswa/absen/submit`, {
							method: 'POST',
							headers: {
								'Content-Type': 'application/json',
								Authorization: `Bearer ${userToken}` // Kunci masuk backend
							},
							body: JSON.stringify({
								sesi_id: sesiId,
								qr_token: decodedText,
								latitude: lat,
								longitude: lon
							})
						});

						const data = await res.json();

						if (res.ok) {
							scanResult = data.message; // Menampilkan pesan sukses dari Go
							scannerState = 'success';
						} else {
							errorMessage = data.error || 'Gagal melakukan absensi.';
							scannerState = 'error';
						}
					} catch (err) {
						errorMessage = 'Gagal terhubung ke server backend.';
						scannerState = 'error';
					}
				},
				(geoErr) => {
					errorMessage = 'Gagal mendapatkan lokasi. Pastikan izin GPS aktif!';
					scannerState = 'error';
				},
				{
					enableHighAccuracy: true // Minta akurasi GPS paling tinggi
				}
			);
		};
		const onFail = () => {};

		try {
			await html5Qrcode.start({ facingMode: 'environment' }, config, onSuccess, onFail);
			scannerState = 'scanning';
		} catch (err: any) {
			console.warn('Kamera belakang gagal, mencoba kamera default...', err);
			try {
				const { Html5Qrcode } = await import('html5-qrcode');
				const devices = await Html5Qrcode.getCameras();

				if (devices && devices.length > 0) {
					await html5Qrcode.start(devices[0].id, config, onSuccess, onFail);
					scannerState = 'scanning';
				} else {
					throw new Error('Kamera tidak terdeteksi.');
				}
			} catch (fallbackErr: any) {
				errorMessage =
					fallbackErr.message || 'Izin kamera ditolak atau kamera tidak bisa digunakan';
				scannerState = 'error';
			}
		}
	};

	const retryScan = () => startScanner();

	const handleLogout = () => {
		if (browser) {
			localStorage.removeItem('jwt_token');
			goto('/login/student'); // Pastikan route ini sesuai dengan halaman login kamu
		}
	};
</script>

<div class="flex min-h-screen flex-col bg-slate-50 font-sans">
	<header
		class="sticky top-0 z-30 flex items-center justify-between border-b border-slate-100 bg-white px-6 py-4 shadow-sm"
	>
		<div class="flex items-center gap-3">
			<div
				class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-brand-blue font-bold text-white"
			>
				A
			</div>
			<div>
				<h1 class="text-lg leading-none font-black tracking-tighter text-slate-800">ABSENSI</h1>
				<p class="text-[10px] font-bold tracking-widest text-slate-400 uppercase">Aplikasi siswa</p>
			</div>
		</div>
		<button
			onclick={handleLogout}
			class="rounded-xl bg-slate-50 px-4 py-2 text-[10px] font-black tracking-widest text-slate-500 uppercase transition-colors hover:bg-slate-100"
		>
			Keluar
		</button>
	</header>

	<main class="mx-auto flex w-full max-w-md flex-1 flex-col items-center justify-center p-4 sm:p-8">
		<div
			class="mb-6 flex w-full items-center justify-between rounded-3xl border border-slate-100 bg-white p-5 shadow-sm"
		>
			<div>
				<p class="text-[10px] font-black tracking-widest text-slate-400 uppercase">Masuk sebagai</p>
				<h2 class="mt-0.5 text-base font-black tracking-tight text-slate-800">{student.name}</h2>
			</div>
			<div
				class="max-w-[120px] truncate rounded-lg bg-blue-50 px-3 py-1.5 text-right text-[10px] font-black tracking-widest text-brand-blue uppercase"
				title={student.school}
			>
				{student.school}
			</div>
		</div>

		<div
			class="relative w-full overflow-hidden rounded-[2.5rem] border border-slate-100 bg-white p-6 text-center shadow-xl shadow-slate-200/50"
		>
			<h3 class="mb-2 text-xl font-black tracking-tight text-slate-900 uppercase">Pindai QR kelas</h3>
			<p class="mb-6 px-4 text-xs font-medium text-slate-500">
				Arahkan kamera ke kode QR yang ditampilkan guru di depan kelas.
			</p>

			<div
				class="relative mx-auto flex aspect-square w-full max-w-[280px] items-center justify-center overflow-hidden rounded-[2rem] border-4 border-slate-50 bg-slate-100 shadow-inner"
			>
				<div id="qr-reader" class="absolute inset-0 h-full w-full"></div>

				{#if scannerState === 'loading'}
					<div
						class="absolute inset-0 z-10 flex flex-col items-center justify-center gap-3 bg-slate-100 text-slate-400"
					>
						<span
							class="h-8 w-8 animate-spin rounded-full border-4 border-slate-300 border-t-brand-blue"
						></span>
						<p class="text-xs font-bold tracking-widest uppercase">Mengakses kamera…</p>
					</div>
				{:else if scannerState === 'error'}
					<div
						class="absolute inset-0 z-10 flex flex-col items-center justify-center gap-3 bg-slate-100 p-6 text-red-500"
					>
						<svg
							xmlns="http://www.w3.org/2000/svg"
							class="h-10 w-10"
							viewBox="0 0 24 24"
							fill="none"
							stroke="currentColor"
							stroke-width="2"
							stroke-linecap="round"
							stroke-linejoin="round"
							><path
								d="M23 19a2 2 0 0 1-2 2H3a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h4l2-3h6l2 3h4a2 2 0 0 1 2 2z"
							/><line x1="12" y1="9" x2="12" y2="13" /><line
								x1="12"
								y1="17"
								x2="12.01"
								y2="17"
							/></svg
						>
						<p class="text-center text-[10px] font-bold tracking-widest uppercase">
							{errorMessage}
						</p>
					</div>
				{/if}

				{#if scannerState === 'scanning'}
					<div class="pointer-events-none absolute inset-0 z-20">
						<div
							class="absolute top-0 left-0 h-1 w-full animate-[scan_2s_ease-in-out_infinite] bg-brand-blue shadow-[0_0_15px_3px_rgba(37,99,235,0.5)]"
						></div>
					</div>
				{/if}
			</div>

			{#if scannerState === 'success'}
				<div
					class="animate-in fade-in absolute inset-0 z-30 flex flex-col items-center justify-center bg-white/95 p-6 backdrop-blur-sm duration-300"
				>
					<div
						class="mb-4 flex h-20 w-20 items-center justify-center rounded-full bg-green-100 shadow-lg shadow-green-100"
					>
						<svg
							xmlns="http://www.w3.org/2000/svg"
							class="h-10 w-10 text-green-600"
							viewBox="0 0 24 24"
							fill="none"
							stroke="currentColor"
							stroke-width="3"
							stroke-linecap="round"
							stroke-linejoin="round"><polyline points="20 6 9 17 4 12" /></svg
						>
					</div>
					<h3 class="mb-2 text-xl font-black tracking-tight text-slate-900 uppercase">
						Pemindaian berhasil!
					</h3>
					<p
						class="mb-8 rounded-xl border border-slate-100 bg-slate-50 px-4 py-3 text-center text-[10px] font-bold tracking-widest break-all text-slate-500 uppercase"
					>
						{scanResult}
					</p>
					<button
						onclick={retryScan}
						class="w-full rounded-2xl bg-slate-900 py-4 text-xs font-black tracking-widest text-white uppercase transition-all hover:bg-black"
					>
						Pindai lagi
					</button>
				</div>
			{/if}
		</div>
	</main>
</div>

<style>
	@keyframes scan {
		0%,
		100% {
			top: 0;
		}
		50% {
			top: 100%;
		}
	}

	:global(#qr-reader) {
		width: 100% !important;
		height: 100% !important;
		border: none !important;
	}
	:global(#qr-reader img) {
		display: none !important;
	}
	:global(#qr-reader__dashboard_section_csr span) {
		display: none !important;
	}
	:global(#qr-reader video) {
		width: 100% !important;
		height: 100% !important;
		object-fit: cover !important;
		border-radius: 2rem !important;
	}
</style>
