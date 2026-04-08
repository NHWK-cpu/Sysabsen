<script lang="ts">
	import { onMount } from 'svelte';
	import QRCode from 'qrcode';

	// State Svelte 5
	// Default tanggal diset ke hari ini
	let selectedDate = $state(new Date().toISOString().split('T')[0]);
	let selectedClass = $state('');
	let selectedSubject = $state('');
	let showModal = $state(false);
	let isDirty = $state(false);
	let qrCanvas = $state<HTMLCanvasElement>();

	const classes = ['Kelas 10 - IPA', 'Kelas 11 - IPS', 'Kelas 12 - Bahasa'];
	const subjects = ['Matematika', 'Bahasa Inggris', 'Fisika'];

	// Data siswa reaktif
	let students = $state([
		{ id: 1, name: 'Ahmad Hafizh', status: 'Hadir' },
		{ id: 2, name: 'Siti Aminah', status: 'Belum Absen' },
		{ id: 3, name: 'Budi Setiawan', status: 'Belum Absen' }
	]);

	// Fungsi Aksi Data
	const updateStatus = (id: number, newStatus: string) => {
		const index = students.findIndex((s) => s.id === id);
		if (index !== -1 && students[index].status !== newStatus) {
			students[index].status = newStatus;
			isDirty = true;
		}
	};

	const handleSave = () => {
		console.log('Saving to MariaDB...', { date: selectedDate, data: students });
		isDirty = false;
		alert('Log absensi berhasil diperbarui!');
	};

	// Fungsi Dummy untuk Ekstra Fitur
	const handleExport = () => alert(`Mengekspor data absensi tanggal ${selectedDate}...`);
	const handleBackup = () => alert('Mengupload file backup ke cloud storage ...');
	const handleRestore = () => alert('Membuka dialog untuk memulihkan backup...');

	$effect(() => {
		if (showModal && qrCanvas) {
			const sessionData = `ABS-${selectedClass}-${selectedSubject}-${selectedDate}-${Date.now()}`;
			QRCode.toCanvas(qrCanvas, sessionData, {
				width: 280,
				margin: 2,
				color: { dark: '#2563eb' }
			});
		}
	});
</script>

<div class="min-h-screen bg-slate-50 font-sans">
	<header
		class="flex items-center justify-between border-b border-slate-100 bg-white px-4 py-4 shadow-sm md:px-8"
	>
		<div class="flex items-center gap-3">
			<div
				class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-brand-blue font-bold text-white"
			>
				A
			</div>
			<div>
				<h1 class="text-lg leading-none font-black tracking-tighter text-slate-800">ABSENSI</h1>
				<p class="text-[10px] font-bold tracking-widest text-slate-400 uppercase">Teacher Panel</p>
			</div>
		</div>

		<div class="flex items-center gap-4 md:gap-6">
			<div class="hidden border-r border-slate-100 pr-6 text-right md:block">
				<p class="text-sm font-black tracking-tight text-slate-800 uppercase">Bpk. Hafizh, S.Kom</p>
				<p class="mt-0.5 text-[10px] font-black tracking-widest text-brand-blue uppercase">
					Pengajar Utama
				</p>
			</div>
			<button
				class="rounded-xl bg-red-50 px-4 py-2 text-[10px] font-black tracking-widest text-red-600 uppercase transition-colors hover:bg-red-100 md:px-5 md:py-2.5"
			>
				Logout
			</button>
		</div>
	</header>

	<main class="mx-auto max-w-6xl p-4 md:p-8">
		<div
			class="mb-6 rounded-[2rem] border border-slate-100 bg-white p-6 shadow-sm md:mb-8 md:rounded-[2.5rem] md:p-8"
		>
			<div class="flex flex-col gap-8 md:flex-row md:items-end md:justify-between">
				<div class="space-y-2">
					<h2 class="text-2xl font-black tracking-tight text-slate-900">Halo, Kakak Pengajar 👋</h2>
					<p class="text-sm leading-relaxed font-medium text-slate-500 md:max-w-xs md:text-base">
						Tentukan jadwal, kelas, dan mapel untuk mengelola absensi.
					</p>
				</div>

				<div class="flex w-full flex-col gap-4 md:w-auto md:flex-row md:items-end">
					<div class="flex flex-1 flex-col gap-2 md:w-36 md:flex-none">
						<label
							for="select-date"
							class="ml-1 text-[10px] font-black tracking-widest text-slate-400 uppercase"
							>Tanggal</label
						>
						<input
							type="date"
							id="select-date"
							bind:value={selectedDate}
							class="h-[54px] cursor-pointer rounded-2xl border-2 border-transparent bg-slate-50 px-4 font-bold text-slate-700 transition-all outline-none focus:border-brand-blue/20"
						/>
					</div>

					<div class="flex flex-1 flex-col gap-2 md:w-40 md:flex-none">
						<label
							for="select-class"
							class="ml-1 text-[10px] font-black tracking-widest text-slate-400 uppercase"
							>Pilih Kelas</label
						>
						<select
							id="select-class"
							bind:value={selectedClass}
							class="h-[54px] cursor-pointer rounded-2xl border-2 border-transparent bg-slate-50 px-4 font-bold text-slate-700 transition-all outline-none focus:border-brand-blue/20"
						>
							<option value="" disabled selected>-- Kelas --</option>
							{#each classes as c}
								<option value={c}>{c}</option>
							{/each}
						</select>
					</div>

					<div class="flex flex-1 flex-col gap-2 md:w-44 md:flex-none">
						<label
							for="select-subject"
							class="ml-1 text-[10px] font-black tracking-widest text-slate-400 uppercase"
							>Mata Pelajaran</label
						>
						<select
							id="select-subject"
							bind:value={selectedSubject}
							class="h-[54px] cursor-pointer rounded-2xl border-2 border-transparent bg-slate-50 px-4 font-bold text-slate-700 transition-all outline-none focus:border-brand-blue/20"
						>
							<option value="" disabled selected>-- Mapel --</option>
							{#each subjects as s}
								<option value={s}>{s}</option>
							{/each}
						</select>
					</div>

					<div class="md:flex-none">
						{#if selectedClass && selectedSubject}
							<button
								onclick={() => (showModal = true)}
								class="h-[54px] w-full rounded-2xl bg-brand-blue px-6 text-xs font-black tracking-widest text-white uppercase shadow-lg shadow-blue-500/20 transition-all hover:scale-105 active:scale-95 md:w-auto"
							>
								Show QR
							</button>
						{:else}
							<div class="hidden h-[54px] w-full md:block md:w-[120px]"></div>
						{/if}
					</div>
				</div>
			</div>
		</div>

		<div
			class="overflow-hidden rounded-[2rem] border border-slate-100 bg-white shadow-sm md:rounded-[2.5rem]"
		>
			<div
				class="sticky top-0 z-10 flex flex-col items-start gap-4 border-b border-slate-50 bg-white p-4 md:p-6 lg:flex-row lg:items-center lg:justify-between"
			>
				<div class="flex flex-wrap items-center gap-3 sm:ml-2">
					<h3 class="text-xs font-black tracking-widest text-slate-800 uppercase">
						Kehadiran Siswa
					</h3>
					{#if selectedClass}
						<span
							class="rounded-full border border-blue-100 bg-blue-50 px-3 py-1 text-[10px] font-black tracking-widest text-brand-blue uppercase"
						>
							{selectedClass}
						</span>
					{/if}
				</div>

				<div class="flex w-full flex-wrap items-center justify-end gap-2 lg:w-auto">
					<button
						onclick={handleRestore}
						class="rounded-xl border border-slate-100 bg-slate-50 px-4 py-2 text-[10px] font-black tracking-widest text-slate-500 uppercase transition-all hover:bg-slate-100 hover:text-slate-800"
					>
						Restore
					</button>
					<button
						onclick={handleBackup}
						class="rounded-xl border border-slate-100 bg-slate-50 px-4 py-2 text-[10px] font-black tracking-widest text-slate-500 uppercase transition-all hover:bg-slate-100 hover:text-slate-800"
					>
						Backup
					</button>
					<button
						onclick={handleExport}
						class="rounded-xl bg-indigo-50 px-4 py-2 text-[10px] font-black tracking-widest text-indigo-600 uppercase transition-all hover:bg-indigo-100"
					>
						Export Data
					</button>

					{#if isDirty}
						<button
							onclick={handleSave}
							class="flex items-center justify-center gap-2 rounded-xl bg-green-600 px-5 py-2 text-[10px] font-black tracking-[0.2em] text-white uppercase shadow-lg shadow-green-500/20 transition-all hover:bg-green-700"
						>
							<span class="relative flex h-2 w-2">
								<span
									class="absolute inline-flex h-full w-full animate-ping rounded-full bg-green-200 opacity-75"
								></span>
								<span class="relative inline-flex h-2 w-2 rounded-full bg-white"></span>
							</span>
							Save
						</button>
					{/if}
				</div>
			</div>

			<div class="w-full overflow-x-auto">
				<table class="w-full min-w-[700px] border-collapse text-left">
					<thead>
						<tr class="bg-slate-50/50">
							<th
								class="w-16 px-6 py-4 text-[10px] font-black tracking-[0.2em] text-slate-400 uppercase md:px-8 md:py-5"
								>No</th
							>
							<th
								class="px-6 py-4 text-[10px] font-black tracking-[0.2em] text-slate-400 uppercase md:px-8 md:py-5"
								>Nama Lengkap</th
							>
							<th
								class="px-6 py-4 text-center text-[10px] font-black tracking-[0.2em] text-slate-400 uppercase md:px-8 md:py-5"
								>Status</th
							>
							<th
								class="px-6 py-4 text-right text-[10px] font-black tracking-[0.2em] text-slate-400 uppercase md:px-8 md:py-5"
								>Aksi</th
							>
						</tr>
					</thead>
					<tbody class="divide-y divide-slate-50">
						{#each students as student, i}
							<tr class="transition-colors hover:bg-slate-50/30">
								<td class="px-6 py-4 text-sm font-bold text-slate-400 md:px-8 md:py-5">{i + 1}</td>
								<td class="px-6 py-4 md:px-8 md:py-5">
									<p class="font-bold text-slate-800">{student.name}</p>
									<p class="text-[10px] font-medium text-slate-400">ID: SLC-00{student.id}</p>
								</td>
								<td class="px-6 py-4 text-center md:px-8 md:py-5">
									<span
										class="rounded-full px-4 py-1.5 text-[9px] font-black tracking-widest whitespace-nowrap uppercase
                                        {student.status === 'Hadir'
											? 'border border-green-200 bg-green-100 text-green-700'
											: student.status === 'Belum Absen'
												? 'border border-slate-200 bg-slate-100 text-slate-500'
												: 'border border-red-200 bg-red-100 text-red-700'}"
									>
										{student.status}
									</span>
								</td>
								<td class="px-6 py-4 md:px-8 md:py-5">
									<div class="flex justify-end gap-2">
										<button
											onclick={() => updateStatus(student.id, 'Hadir')}
											class="rounded-xl px-4 py-2 text-[10px] font-black tracking-widest uppercase transition-all {student.status ===
											'Hadir'
												? 'bg-green-600 text-white shadow-md'
												: 'bg-slate-100 text-slate-500 hover:bg-green-50 hover:text-green-600'}"
											>Hadir</button
										>
										<button
											onclick={() => updateStatus(student.id, 'Alpa')}
											class="rounded-xl px-4 py-2 text-[10px] font-black tracking-widest uppercase transition-all {student.status ===
											'Alpa'
												? 'bg-red-600 text-white shadow-md'
												: 'bg-slate-100 text-slate-500 hover:bg-red-50 hover:text-red-600'}"
											>Alpa</button
										>
									</div>
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		</div>
	</main>

	{#if showModal}
		<div
			class="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/70 p-4 backdrop-blur-md md:p-6"
		>
			<div
				class="scale-in-center w-full max-w-md rounded-[2.5rem] bg-white p-8 text-center shadow-2xl md:rounded-[3rem] md:p-12"
			>
				<h3 class="mb-2 text-xl font-black tracking-tight text-slate-900 uppercase md:text-2xl">
					Scan Kehadiran
				</h3>
				<p
					class="mb-6 text-[10px] font-bold tracking-widest text-slate-400 uppercase md:mb-8 md:text-xs"
				>
					{selectedDate} • {selectedSubject}
				</p>
				<div
					class="mb-8 inline-block rounded-[2rem] border-2 border-slate-100 bg-slate-50 p-4 shadow-inner md:mb-10 md:rounded-[2.5rem] md:p-6"
				>
					<canvas bind:this={qrCanvas} class="max-w-full"></canvas>
				</div>
				<button
					onclick={() => (showModal = false)}
					class="w-full rounded-2xl bg-slate-900 py-3 text-[10px] font-black tracking-widest text-white uppercase transition-all hover:bg-black md:py-4 md:text-xs"
				>
					Tutup Jendela QR
				</button>
			</div>
		</div>
	{/if}
</div>
