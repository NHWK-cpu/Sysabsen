<script lang="ts">
	import { onMount } from 'svelte';
	import QRCode from 'qrcode';
	import { goto } from '$app/navigation';

	let selectedDate = $state(new Date().toISOString().split('T')[0]);
	let selectedClassId = $state<number | ''>('');
	let selectedSubjectId = $state<number | ''>('');
	let activeSessionId = $state<number | null>(null);

	let showModal = $state(false);
	let isDirty = $state(false);
	let qrCanvas = $state<HTMLCanvasElement>();

	const API_BASE_URL = import.meta.env.VITE_API_URL;
	let token = '';

	// PERUBAHAN: Menambahkan field school ke dalam state teacherProfile
	let teacherProfile = $state({ name: 'Memuat...', role: 'Guru', school: '' });
	let classes = $state<any[]>([]);
	let subjects = $state<any[]>([]);
	let students = $state<any[]>([]);
	let stats = $state({
		total_students: 0,
		present_today: 0,
		absent_today: 0,
		attendance_rate: 0
	});

	onMount(async () => {
		token = localStorage.getItem('jwt_token') || '';
		const role = localStorage.getItem('user_role');

		if (!token || (role !== 'guru' && role !== 'teacher')) {
			goto('/login/teacher');
			return;
		}

		try {
			const payloadBase64 = token.split('.')[1];
			const decodedPayload = JSON.parse(atob(payloadBase64));

			// PERUBAHAN: Menggunakan nama_lengkap dan mengambil nama_sekolah dari JWT
			teacherProfile = {
				name: decodedPayload.nama_lengkap || decodedPayload.username || 'Guru',
				role: decodedPayload.role === 'admin' ? 'Administrator' : 'Pengajar Utama',
				school: decodedPayload.nama_sekolah || ''
			};
		} catch (e) {
			console.error('Gagal membaca profil dari token');
		}

		try {
			const [resKelas, resMapel] = await Promise.all([
				fetch(`${API_BASE_URL}/guru/kelas`, { headers: { Authorization: `Bearer ${token}` } }),
				fetch(`${API_BASE_URL}/guru/mapel`, { headers: { Authorization: `Bearer ${token}` } })
			]);

			if (resKelas.ok) classes = (await resKelas.json()) || [];
			if (resMapel.ok) subjects = (await resMapel.json()) || [];
		} catch (error) {
			console.error('Gagal mengambil data referensi:', error);
		}
	});

	$effect(() => {
		if (selectedClassId !== '' && selectedSubjectId !== '' && selectedDate && token) {
			initiateClassSession();
		} else {
			activeSessionId = null;
			students = [];
		}
	});

	const initiateClassSession = async () => {
		try {
			const res = await fetch(`${API_BASE_URL}/guru/sesi/init`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
				body: JSON.stringify({
					kelas_id: Number(selectedClassId),
					mapel_id: Number(selectedSubjectId),
					tanggal: selectedDate
				})
			});

			if (res.ok) {
				const data = await res.json();
				activeSessionId = data.sesi_id;
				fetchAttendanceList(activeSessionId);
				fetchStats(activeSessionId);
			} else {
				// PENANGKAP ERROR ANTI-BADAI
				const errText = await res.text();
				let errMsg = errText;
				try {
					const errJson = JSON.parse(errText);
					if (errJson.error) errMsg = errJson.error;
				} catch (e) {} // Abaikan kalau teks bukan JSON

				alert(`Sistem Menolak: ${errMsg}`);
				activeSessionId = null;
			}
		} catch (error) {
			alert('Gagal menghubungi server database.');
		}
	};

	const fetchStats = async (sesiId: number) => {
		try {
			const res = await fetch(`${API_BASE_URL}/guru/dashboard/stats?sesi_id=${sesiId}`, {
				headers: { Authorization: `Bearer ${token}` }
			});
			if (res.ok) {
				stats = await res.json();
			}
		} catch (error) {
			console.error('Gagal mengambil statistik:', error);
		}
	};

	const fetchAttendanceList = async (sesiId: number) => {
		try {
			const res = await fetch(
				`${API_BASE_URL}/guru/dashboard/attendance-list?sesi_id=${sesiId}&tanggal=${selectedDate}`,
				{ headers: { Authorization: `Bearer ${token}` } }
			);

			if (res.ok) {
				const responseData = await res.json();
				const uniqueStudents = [];
				const seenId = new Set();

				for (const s of responseData) {
					if (!seenId.has(s.id)) {
						seenId.add(s.id);
						const rawStatus = (s.status || '').toLowerCase();
						let finalStatus = 'Belum Absen';

						if (rawStatus === 'hadir') finalStatus = 'Hadir';
						else if (rawStatus === 'alpa') finalStatus = 'Alpa';
						else if (rawStatus === 'izin') finalStatus = 'Izin';
						else if (rawStatus === 'sakit') finalStatus = 'Sakit';

						uniqueStudents.push({
							id: s.nama_sekolah,
							db_id: s.id,
							name: s.nama_lengkap || s.nama,
							status: finalStatus,
							waktu_absen: s.waktu_absen
						});
					}
				}
				students = uniqueStudents;
				isDirty = false;
			}
		} catch (error) {
			console.error('Gagal memuat daftar absensi:', error);
		}
	};

	const updateStatus = (id: string | number, newStatus: string) => {
		const index = students.findIndex((s) => s.id === id);
		if (index !== -1 && students[index].status !== newStatus) {
			students[index].status = newStatus;
			isDirty = true;
		}
	};

	const handleSave = async () => {
		if (!activeSessionId) return;

		const studentsToUpdate = students.filter((s) => s.status !== 'Belum Absen');
		if (studentsToUpdate.length === 0) {
			alert('Tidak ada data yang diubah.');
			return;
		}

		try {
			const promises = studentsToUpdate.map((student) => {
				return fetch(`${API_BASE_URL}/guru/absen`, {
					method: 'POST',
					headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
					body: JSON.stringify({
						sesi_id: activeSessionId,
						siswa_id: student.db_id,
						status_kehadiran: student.status,
						tanggal: selectedDate
					})
				});
			});

			const results = await Promise.all(promises);
			if (results.every((res) => res.ok)) {
				isDirty = false;
				alert('Absensi manual tersimpan!');
				fetchStats(activeSessionId);
			} else {
				alert('Beberapa data gagal dicatat.');
			}
		} catch (error) {
			alert('Terjadi kesalahan jaringan.');
		}
	};

	const handleExport = async () => {
		try {
			const res = await fetch(`${API_BASE_URL}/guru/export`, {
				method: 'GET',
				headers: { Authorization: `Bearer ${token}` }
			});

			if (res.ok) {
				const blob = await res.blob();
				const url = window.URL.createObjectURL(blob);
				const a = document.createElement('a');
				a.href = url;
				a.download = `Laporan_Absen_${selectedDate}.xlsx`;
				document.body.appendChild(a);
				a.click();
				a.remove();
				window.URL.revokeObjectURL(url);
			} else {
				const errText = await res.text();
				alert(`Gagal mengekspor data: ${errText}`);
			}
		} catch (err) {
			alert('Terjadi kesalahan jaringan saat mengunduh laporan.');
		}
	};

	// const handleBackup = async () => {
	// 	try {
	// 		const res = await fetch(`${API_BASE_URL}/guru/backup`, {
	// 			headers: { Authorization: `Bearer ${token}` }
	// 		});
	// 		if (res.ok) {
	// 			alert(await res.text());
	// 		}
	// 	} catch (err) {
	// 		alert('Gagal trigger backup.');
	// 	}
	// };

	let isRefreshing = $state(false);

	const handleRefresh = async () => {
		if (activeSessionId) {
			isRefreshing = true;
			await fetchAttendanceList(activeSessionId);
			await fetchStats(activeSessionId);

			// Jeda 500ms agar animasi loading pada tombol terlihat
			setTimeout(() => {
				isRefreshing = false;
			}, 500);
		} else if (selectedClassId !== '' && selectedSubjectId !== '' && selectedDate) {
			isRefreshing = true;
			await initiateClassSession();
			isRefreshing = false;
		} else {
			alert('Pilih kelas, mata pelajaran, dan tanggal terlebih dahulu!');
		}
	};

	$effect(() => {
		if (showModal && qrCanvas && activeSessionId && token) {
			fetch(`${API_BASE_URL}/guru/generate-qr?sesi_id=${activeSessionId}`, {
				headers: { Authorization: `Bearer ${token}` }
			})
				.then((res) => res.json())
				.then((data) => {
					if (data.qr_token) {
						QRCode.toCanvas(qrCanvas, data.qr_token, {
							width: 280,
							margin: 2,
							color: { dark: '#2563eb' }
						});
					}
				})
				.catch((err) => console.error('Gagal generate QR', err));
		}
	});

	const logout = () => {
		localStorage.removeItem('jwt_token');
		localStorage.removeItem('user_role');
		goto('/login/teacher');
	};
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
				<p class="text-sm font-black tracking-tight text-slate-800 uppercase">
					{teacherProfile.name}
				</p>
				<p class="mt-0.5 text-[10px] font-black tracking-widest text-brand-blue uppercase">
					{teacherProfile.role}
					{#if teacherProfile.school}
						<span class="text-slate-300">•</span> {teacherProfile.school}
					{/if}
				</p>
			</div>
			<button
				onclick={logout}
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
							bind:value={selectedClassId}
							class="h-[54px] cursor-pointer rounded-2xl border-2 border-transparent bg-slate-50 px-4 font-bold text-slate-700 transition-all outline-none focus:border-brand-blue/20"
						>
							<option value="" disabled selected>-- Kelas --</option>
							{#each classes as c}
								<option value={c.id}>{c.nama_kelas}</option>
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
							bind:value={selectedSubjectId}
							class="h-[54px] cursor-pointer rounded-2xl border-2 border-transparent bg-slate-50 px-4 font-bold text-slate-700 transition-all outline-none focus:border-brand-blue/20"
						>
							<option value="" disabled selected>-- Mapel --</option>
							{#each subjects as s}
								<option value={s.id}>{s.nama_mapel}</option>
							{/each}
						</select>
					</div>

					<div class="md:flex-none">
						{#if activeSessionId}
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

		{#if activeSessionId}
			<div class="mb-6 grid grid-cols-2 gap-4 md:mb-8 md:grid-cols-4 md:gap-6">
				<div
					class="rounded-[2rem] border border-slate-100 bg-white p-6 shadow-sm transition-all hover:shadow-md"
				>
					<p class="text-[10px] font-black tracking-widest text-slate-400 uppercase">Total Siswa</p>
					<div class="mt-2 flex items-baseline gap-2">
						<span class="text-3xl font-black tracking-tighter text-slate-800"
							>{stats.total_students}</span
						>
						<span class="text-xs font-bold text-slate-400">Orang</span>
					</div>
				</div>

				<div
					class="rounded-[2rem] border border-blue-100 bg-blue-50 p-6 shadow-sm transition-all hover:shadow-md"
				>
					<p class="text-[10px] font-black tracking-widest text-brand-blue uppercase">
						Sudah Absen
					</p>
					<div class="mt-2 flex items-baseline gap-2">
						<span class="text-3xl font-black tracking-tighter text-brand-blue"
							>{stats.present_today}</span
						>
						<span class="text-xs font-bold text-blue-400">Hadir</span>
					</div>
				</div>

				<div
					class="rounded-[2rem] border border-rose-100 bg-rose-50 p-6 shadow-sm transition-all hover:shadow-md"
				>
					<p class="text-[10px] font-black tracking-widest text-rose-500 uppercase">Tidak Hadir</p>
					<div class="mt-2 flex items-baseline gap-2">
						<span class="text-3xl font-black tracking-tighter text-rose-600"
							>{stats.absent_today}</span
						>
						<span class="text-xs font-bold text-rose-400">Siswa</span>
					</div>
				</div>

				<div
					class="rounded-[2rem] border border-emerald-100 bg-emerald-50 p-6 shadow-sm transition-all hover:shadow-md"
				>
					<p class="text-[10px] font-black tracking-widest text-emerald-500 uppercase">
						Tingkat Kehadiran
					</p>
					<div class="mt-2 flex items-baseline gap-2">
						<span class="text-3xl font-black tracking-tighter text-emerald-600">
							{stats.attendance_rate.toFixed(1)}
						</span>
						<span class="text-lg font-black text-emerald-500">%</span>
					</div>
				</div>
			</div>
		{/if}

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
					{#if selectedClassId !== ''}
						<span
							class="rounded-full border border-blue-100 bg-blue-50 px-3 py-1 text-[10px] font-black tracking-widest text-brand-blue uppercase"
						>
							{classes.find((c) => c.id === selectedClassId)?.nama_kelas || ''}
						</span>
					{/if}
				</div>

				<div class="flex w-full flex-wrap items-center justify-end gap-2 lg:w-auto">
					<button
						onclick={handleRefresh}
						disabled={isRefreshing || (!activeSessionId && selectedClassId === '')}
						class="flex items-center gap-2 rounded-xl border border-slate-200 bg-white px-4 py-2 text-[10px] font-black tracking-widest text-slate-600 uppercase transition-all hover:bg-slate-50 hover:text-brand-blue disabled:cursor-not-allowed disabled:opacity-50"
					>
						<svg
							xmlns="http://www.w3.org/2000/svg"
							class="h-3.5 w-3.5 {isRefreshing ? 'animate-spin text-brand-blue' : ''}"
							viewBox="0 0 24 24"
							fill="none"
							stroke="currentColor"
							stroke-width="3"
							stroke-linecap="round"
							stroke-linejoin="round"
						>
							<path d="M3 12a9 9 0 1 0 9-9 9.75 9.75 0 0 0-6.74 2.74L3 8" />
							<path d="M3 3v5h5" />
						</svg>
						{isRefreshing ? 'Memuat...' : 'Perbarui Data'}
					</button>
					<!-- <button
						onclick={handleBackup}
						class="rounded-xl border border-slate-100 bg-slate-50 px-4 py-2 text-[10px] font-black tracking-widest text-slate-500 uppercase transition-all hover:bg-slate-100 hover:text-slate-800"
						>Backup</button
					> -->
					<button
						onclick={handleExport}
						class="rounded-xl bg-indigo-50 px-4 py-2 text-[10px] font-black tracking-widest text-indigo-600 uppercase transition-all hover:bg-indigo-100"
						>Export Data</button
					>

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
									<p class="text-[10px] font-medium text-slate-400">Sekolah: {student.id}</p>
								</td>
								<td class="px-6 py-4 text-center md:px-8 md:py-5">
									<span
										class="rounded-full px-4 py-1.5 text-[9px] font-black tracking-widest whitespace-nowrap uppercase {student.status ===
										'Hadir'
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
						{#if students.length === 0}
							<tr>
								<td colspan="4" class="py-10 text-center text-sm font-bold text-slate-400"
									>Pilih Kelas dan Mapel untuk menampilkan data siswa</td
								>
							</tr>
						{/if}
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
					{selectedDate} • {subjects.find((s) => s.id === selectedSubjectId)?.nama_mapel || ''}
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
