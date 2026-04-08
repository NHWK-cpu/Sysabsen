<script lang="ts">
	import { onMount } from 'svelte';
	import QRCode from 'qrcode';

	let currentTime = $state(new Date());
	let qrCanvas: HTMLCanvasElement;

	onMount(() => {
		const timer = setInterval(() => {
			currentTime = new Date();
		}, 1000);

		const loginUrl = `${window.location.origin}/login/student`;
		QRCode.toCanvas(qrCanvas, loginUrl, {
			width: 320,
			margin: 2,
			color: {
				dark: '#1e293b',
				light: '#ffffff'
			}
		});

		return () => clearInterval(timer);
	});

	const timeString = $derived(
		currentTime.toLocaleTimeString('id-ID', {
			hour: '2-digit',
			minute: '2-digit',
			second: '2-digit'
		})
	);

	const dateString = $derived(
		currentTime.toLocaleDateString('id-ID', {
			weekday: 'long',
			year: 'numeric',
			month: 'long',
			day: 'numeric'
		})
	);
</script>

<div class="flex min-h-screen flex-col items-center justify-center bg-slate-50 p-6">
	<header
		class="fixed top-0 flex w-full flex-col items-center justify-between gap-6 p-6 md:flex-row md:p-10"
	>
		<div class="flex items-center gap-3">
			<div
				class="flex h-12 w-12 items-center justify-center rounded-xl bg-brand-blue shadow-lg shadow-blue-200"
			>
				<span class="text-2xl font-bold text-white italic">A</span>
			</div>
			<div>
				<h1 class="text-xl font-black tracking-tighter text-slate-800">ABSENSI</h1>
				<p class="text-[10px] leading-none font-bold tracking-widest text-slate-400 uppercase">
					Smart Learning Center
				</p>
			</div>
		</div>

		<nav class="flex items-center gap-3 rounded-2xl border border-slate-100 bg-white p-2 shadow-sm">
			<a
				href="/login/teacher"
				class="rounded-xl px-5 py-2.5 text-sm font-bold text-slate-600 transition-all hover:bg-slate-50 hover:text-brand-blue"
			>
				Login Guru
			</a>
			<div class="h-6 w-[1px] bg-slate-100"></div>
			<a
				href="/login/admin"
				class="rounded-xl px-5 py-2.5 text-sm font-bold text-slate-600 transition-all hover:bg-slate-50 hover:text-slate-900"
			>
				Login Admin
			</a>
		</nav>

		<div class="hidden text-right md:block">
			<h2 class="text-3xl leading-none font-black text-slate-800 tabular-nums">{timeString}</h2>
			<p class="mt-1 text-[11px] font-bold tracking-widest text-slate-400 uppercase">
				{dateString}
			</p>
		</div>
	</header>

	<main class="flex flex-col items-center text-center">
		<div
			class="mb-10 rounded-[3rem] border border-slate-100 bg-white p-6 shadow-2xl shadow-slate-200 transition-transform hover:scale-[1.02] md:p-10"
		>
			<canvas bind:this={qrCanvas} class="rounded-3xl"></canvas>
		</div>

		<div class="space-y-3">
			<span
				class="rounded-full bg-blue-100 px-4 py-1.5 text-[11px] font-black tracking-[0.2em] text-brand-blue uppercase"
				>Portal Siswa</span
			>
			<h2 class="text-4xl font-black tracking-tight text-slate-900 md:text-5xl">
				Scan QR Untuk Absen
			</h2>
			<p class="mx-auto max-w-sm text-sm font-medium text-slate-400 md:text-base">
				Arahkan kamera smartphone Anda untuk masuk ke sistem absensi kehadiran siswa.
			</p>
		</div>
	</main>

	<div class="mt-12 text-center md:hidden">
		<h2 class="text-2xl leading-none font-black text-slate-800 tabular-nums">{timeString}</h2>
		<p class="mt-1 text-[10px] font-bold tracking-widest text-slate-400 uppercase">{dateString}</p>
	</div>

	<footer class="fixed bottom-8 text-[10px] font-bold tracking-[0.3em] text-slate-300 uppercase">
		&copy; 2026 SLC System • Production Environment
	</footer>
</div>

<style>
	:global(body) {
		overflow: hidden;
		@media (max-width: 768px) {
			overflow: auto;
		}
	}
</style>
