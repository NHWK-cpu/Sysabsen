<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';

	const API_BASE_URL = import.meta.env.VITE_API_URL;
	let token = '';

	let teacherProfile = $state({ name: 'Memuat...', role: 'Guru', school: '' });
	let classes = $state<any[]>([]);
	let students = $state<any[]>([]);
	let selectedClassId = $state<number | ''>('');
	let searchQuery = $state('');
	let isLoading = $state(false);

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

			teacherProfile = {
				name: decodedPayload.nama_lengkap || decodedPayload.username || 'Guru',
				role: decodedPayload.role === 'admin' ? 'Administrator' : 'Pengajar Utama',
				school: decodedPayload.nama_sekolah || ''
			};
		} catch (e) {
			console.error('Gagal membaca profil dari token');
		}

		await fetchClasses();
	});

	const fetchClasses = async () => {
		try {
			// Mengambil HANYA kelas yang diampu guru ini (sama seperti di absensi)
			const res = await fetch(`${API_BASE_URL}/guru/kelas`, {
				headers: { Authorization: `Bearer ${token}` }
			});
			if (res.ok) {
				classes = (await res.json()) || [];
			}
		} catch (error) {
			console.error('Gagal mengambil data kelas:', error);
		}
	};

	$effect(() => {
		if (selectedClassId !== '') {
			fetchStudents(Number(selectedClassId));
		} else {
			students = [];
		}
	});

	const fetchStudents = async (kelasId: number) => {
		isLoading = true;
		try {
			// PERUBAHAN: Disesuaikan dengan pola backend query parameter
			const res = await fetch(`${API_BASE_URL}/guru/kelas/siswa?kelas_id=${kelasId}`, {
				headers: { Authorization: `Bearer ${token}` }
			});

			if (res.ok) {
				students = await res.json();
			} else {
				students = [];
				console.error('Gagal memuat siswa');
			}
		} catch (error) {
			alert('Terjadi kesalahan jaringan saat memuat daftar siswa.');
		} finally {
			isLoading = false;
		}
	};

	// Filter siswa berdasarkan pencarian nama atau NIS
	let filteredStudents = $derived(
		students.filter(
			(student) =>
				student.nama.toLowerCase().includes(searchQuery.toLowerCase()) ||
				(student.nis && student.nis.includes(searchQuery))
		)
	);

	let showAddModal = $state(false);
	let allAvailableStudents = $state<any[]>([]);
	let selectedNewStudentId = $state('');
	let isProcessing = $state(false);

	// Panggil fungsi ini saat komponen dimuat atau saat modal akan dibuka
	const fetchAllAvailableStudents = async () => {
		try {
			const res = await fetch(`${API_BASE_URL}/guru/siswa/all`, {
				headers: { Authorization: `Bearer ${token}` }
			});
			if (res.ok) {
				allAvailableStudents = await res.json();
			}
		} catch (error) {
			console.error('Gagal memuat opsi siswa:', error);
		}
	};

	const handleAssignStudent = async () => {
		if (!selectedNewStudentId || !selectedClassId) return;
		isProcessing = true;

		try {
			const res = await fetch(`${API_BASE_URL}/admin/siswa-kelas/assign`, {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
					Authorization: `Bearer ${token}`
				},
				body: JSON.stringify({
					siswa_id: Number(selectedNewStudentId),
					kelas_id: Number(selectedClassId)
				})
			});

			if (res.ok) {
				alert('Siswa berhasil ditambahkan ke kelas!');
				showAddModal = false;
				selectedNewStudentId = '';
				await fetchStudents(Number(selectedClassId)); // Refresh tabel
			} else {
				const errText = await res.text();
				alert(`Gagal: ${errText}`);
			}
		} catch (error) {
			alert('Kesalahan jaringan.');
		} finally {
			isProcessing = false;
		}
	};

	// --- STATE BARU UNTUK MODAL REMOVE & RESET ---
	let showRemoveModal = $state(false);
	let studentToRemove = $state({ id: 0, name: '' });
	let deleteLogsChecked = $state(false);
	let isRemoving = $state(false);
	let isResetting = $state(false);

	// --- FUNGSI MUNCULKAN MODAL KELUARKAN SISWA ---
	const openRemoveModal = (studentUserId: number, studentName: string) => {
		studentToRemove = { id: studentUserId, name: studentName };
		deleteLogsChecked = false; // Checkbox default tidak tercentang agar aman
		showRemoveModal = true;
	};

	// --- FUNGSI EKSEKUSI KELUARKAN SISWA ---
	const executeRemoveStudent = async () => {
		if (!studentToRemove.id || !selectedClassId) return;
		isRemoving = true;

		try {
			const res = await fetch(
				`${API_BASE_URL}/admin/siswa-kelas/remove?siswa_id=${studentToRemove.id}&kelas_id=${selectedClassId}&delete_logs=${deleteLogsChecked}`,
				{
					method: 'DELETE',
					headers: { Authorization: `Bearer ${token}` }
				}
			);

			if (res.ok) {
				alert(`${studentToRemove.name} berhasil dikeluarkan dari kelas.`);
				showRemoveModal = false;
				await fetchStudents(Number(selectedClassId)); // Refresh data tabel
			} else {
				const errText = await res.text();
				alert(`Sistem Menolak: ${errText}`);
			}
		} catch (error) {
			alert('Gagal menghubungi server.');
		} finally {
			isRemoving = false;
		}
	};

	// --- FUNGSI RESET ABSENSI (Tanpa mengeluarkan siswa) ---
	const handleResetAbsen = async (studentUserId: number, studentName: string) => {
		if (
			!confirm(
				`Peringatan: Yakin ingin mereset/menghapus seluruh riwayat kehadiran ${studentName} di kelas ini?\n\nData yang dihapus tidak bisa dikembalikan.`
			)
		)
			return;

		isResetting = true;
		try {
			const res = await fetch(
				`${API_BASE_URL}/admin/siswa-kelas/reset-absen?siswa_id=${studentUserId}&kelas_id=${selectedClassId}`,
				{
					method: 'DELETE',
					headers: { Authorization: `Bearer ${token}` }
				}
			);

			if (res.ok) {
				alert(`Riwayat absensi ${studentName} berhasil di-reset.`);
				// Tidak perlu refresh tabel siswa karena dia masih di kelas,
				// tapi riwayatnya di server sudah bersih.
			} else {
				const errText = await res.text();
				alert(`Gagal mereset absensi: ${errText}`);
			}
		} catch (error) {
			alert('Terjadi kesalahan jaringan.');
		} finally {
			isResetting = false;
		}
	};

	// --- STATE UNTUK IMPORT EXCEL ---
    let showImportModal = $state(false);
    let selectedFile = $state<File | null>(null);
    let isUploading = $state(false);
    let uploadResult = $state<{ sukses: number; gagal: number; message: string } | null>(null);

    // --- FUNGSI IMPORT EXCEL ---
    const handleFileChange = (event: Event) => {
        const target = event.target as HTMLInputElement;
        if (target.files && target.files.length > 0) {
            const file = target.files[0];
            if (!file.name.endsWith('.xlsx')) {
                alert('Unggah file berformat .xlsx (Microsoft Excel)');
                target.value = ''; 
                selectedFile = null;
                return;
            }
            selectedFile = file;
            uploadResult = null; 
        }
    };

    const handleUpload = async () => {
        if (!selectedFile) return;

        isUploading = true;
        uploadResult = null;

        const formData = new FormData();
        formData.append('file_excel', selectedFile);

        try {
            const res = await fetch(`${API_BASE_URL}/admin/siswa/import`, {
                method: 'POST',
                headers: {
                    Authorization: `Bearer ${token}`
                },
                body: formData
            });

            if (res.ok) {
                const data = await res.json();
                uploadResult = data;
                selectedFile = null; 
                
                const fileInput = document.getElementById('file-upload') as HTMLInputElement;
                if (fileInput) fileInput.value = '';

                // Opsional: Refresh daftar siswa di dropdown tambah jika sukses
                if (data.sukses > 0) {
                    fetchAllAvailableStudents();
                }
            } else {
                const errText = await res.text();
                alert(`Gagal mengunggah: ${errText}`);
            }
        } catch (error) {
            alert('Terjadi kesalahan jaringan saat mengunggah file.');
        } finally {
            isUploading = false;
        }
    };
</script>

<div class="min-h-screen bg-slate-50 font-sans">
	<header
		class="flex items-center justify-between border-b border-slate-100 bg-white px-4 py-4 shadow-sm md:px-8"
	>
		<div class="flex items-center gap-6">
			<div class="flex items-center gap-3">
				<div
					class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-brand-blue font-bold text-white"
				>
					S
				</div>
				<div>
					<h1 class="text-lg leading-none font-black tracking-tighter text-slate-800">SISWA</h1>
					<p class="text-[10px] font-bold tracking-widest text-slate-400 uppercase">
						Manajemen Kelas
					</p>
				</div>
			</div>

			<nav class="hidden gap-4 border-l border-slate-100 pl-6 md:flex">
				<button
					onclick={() => goto('/teacher/dashboard')}
					class="text-xs font-bold text-slate-400 transition-colors hover:text-brand-blue"
					>Absensi</button
				>
				<button class="border-b-2 border-brand-blue pb-1 text-xs font-black text-brand-blue"
					>Daftar Siswa</button
				>
			</nav>
		</div>

		<div class="flex items-center gap-4 md:gap-6">
			<div class="hidden border-r border-slate-100 pr-6 text-right md:block">
				<p class="text-sm font-black tracking-tight text-slate-800 uppercase">
					{teacherProfile.name}
				</p>
				<p class="mt-0.5 text-[10px] font-black tracking-widest text-brand-blue uppercase">
					{teacherProfile.role}
				</p>
			</div>
			<button
				onclick={() => goto('/login/teacher')}
				class="rounded-xl bg-red-50 px-4 py-2 text-[10px] font-black tracking-widest text-red-600 uppercase transition-colors hover:bg-red-100 md:px-5 md:py-2.5"
			>
				Keluar
			</button>
		</div>
	</header>

	<main class="mx-auto max-w-6xl p-4 md:p-8">
		<div
			class="mb-6 rounded-[2rem] border border-slate-100 bg-white p-6 shadow-sm md:mb-8 md:rounded-[2.5rem] md:p-8"
		>
			<div class="flex flex-col gap-8 md:flex-row md:items-end md:justify-between">
				<div class="space-y-2">
					<h2 class="text-2xl font-black tracking-tight text-slate-900">Daftar Siswa & Kelas 📚</h2>
					<p class="text-sm leading-relaxed font-medium text-slate-500 md:max-w-xs md:text-base">
						Pantau daftar siswa pada kelas yang menjadi tanggung jawab Anda.
					</p>
				</div>

				<div class="flex w-full flex-col gap-4 md:w-auto md:flex-row md:items-end">
					<div class="flex flex-1 flex-col gap-2 md:w-56 md:flex-none">
						<label
							for="select-class"
							class="ml-1 text-[10px] font-black tracking-widest text-slate-400 uppercase"
							>Pilih Kelas Diampu</label
						>
						<select
							bind:value={selectedClassId}
							class="h-[54px] cursor-pointer rounded-2xl border-2 border-transparent bg-slate-50 px-4 font-bold text-slate-700 transition-all outline-none focus:border-brand-blue/20"
						>
							<option value="" disabled selected>-- Pilih Kelas --</option>
							{#each classes as c}
								<option value={c.id}>{c.nama_kelas}</option>
							{/each}
						</select>
					</div>

					<div class="flex flex-1 flex-col gap-2 md:w-64 md:flex-none">
						<label
							for="search-student"
							class="ml-1 text-[10px] font-black tracking-widest text-slate-400 uppercase"
							>Cari Siswa</label
						>
						<input
							type="text"
							placeholder="Ketik nama atau NIS..."
							bind:value={searchQuery}
							disabled={!selectedClassId}
							class="h-[54px] rounded-2xl border-2 border-transparent bg-slate-50 px-4 font-bold text-slate-700 transition-all outline-none focus:border-brand-blue/20 disabled:cursor-not-allowed disabled:opacity-50"
						/>
					</div>
				</div>
			</div>
		</div>

		<div
			class="overflow-hidden rounded-[2rem] border border-slate-100 bg-white shadow-sm md:rounded-[2.5rem]"
		>
			<div
				class="flex flex-col items-start gap-4 border-b border-slate-50 bg-white p-4 md:p-6 lg:flex-row lg:items-center lg:justify-between"
			>
				<div class="flex flex-wrap items-center gap-3 sm:ml-2">
					<h3 class="text-xs font-black tracking-widest text-slate-800 uppercase">Daftar Siswa</h3>
					<div class="flex w-full flex-wrap items-center justify-end gap-2 lg:w-auto">
						{#if selectedClassId !== ''}
							<span
								class="rounded-full border border-blue-100 bg-blue-50 px-3 py-1 text-[10px] font-black tracking-widest text-brand-blue uppercase"
							>
								Total {filteredStudents.length} siswa
							</span>
							<button
								onclick={() => {
									showAddModal = true;
									fetchAllAvailableStudents();
								}}
								class="flex items-center gap-2 rounded-xl bg-brand-blue px-5 py-2 text-[10px] font-black tracking-[0.2em] text-white uppercase shadow-lg shadow-blue-500/20 transition-all hover:bg-blue-700"
							>
								+ Tambah Siswa
							</button>
						{/if}
					</div>
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
								>NIS / Identitas</th
							>
							<th
								class="px-6 py-4 text-[10px] font-black tracking-[0.2em] text-slate-400 uppercase md:px-8 md:py-5"
								>Nama Lengkap</th
							>
							<th
								class="px-6 py-4 text-center text-[10px] font-black tracking-[0.2em] text-slate-400 uppercase md:px-8 md:py-5"
								>Jenis kelamin</th
							>
							<th
								class="px-6 py-4 text-right text-[10px] font-black tracking-[0.2em] text-slate-400 uppercase md:px-8 md:py-5"
								>Aksi</th
							>
						</tr>
					</thead>
					<tbody class="divide-y divide-slate-50">
						{#if isLoading}
							<tr>
								<td
									colspan="5"
									class="animate-pulse py-10 text-center text-sm font-bold text-slate-400"
									>Memuat data siswa...</td
								>
							</tr>
						{:else if filteredStudents.length === 0 && selectedClassId !== ''}
							<tr>
								<td colspan="5" class="py-10 text-center text-sm font-bold text-slate-400"
									>Tidak ada siswa ditemukan di kelas ini.</td
								>
							</tr>
						{:else if selectedClassId === ''}
							<tr>
								<td colspan="5" class="py-10 text-center text-sm font-bold text-slate-400"
									>Silakan pilih kelas terlebih dahulu.</td
								>
							</tr>
						{:else}
							{#each filteredStudents as student, i}
								<tr class="transition-colors hover:bg-slate-50/30">
									<td class="px-6 py-4 text-sm font-bold text-slate-400 md:px-8 md:py-5">{i + 1}</td
									>
									<td class="px-6 py-4 md:px-8 md:py-5">
										<span class="rounded-lg bg-slate-100 px-3 py-1 text-xs font-bold text-slate-600"
											>{student.nis || '-'}</span
										>
									</td>
									<td class="px-6 py-4 md:px-8 md:py-5">
										<p class="font-bold text-slate-800">{student.nama}</p>
									</td>
									<td class="px-6 py-4 text-center md:px-8 md:py-5">
										<span
											class="rounded-full border px-4 py-1.5 text-[9px] font-black tracking-widest uppercase {student.jenis_kelamin ===
											'L'
												? 'border-blue-200 bg-blue-50 text-blue-600'
												: 'border-pink-200 bg-pink-50 text-pink-600'}"
										>
											{student.jenis_kelamin === 'L' ? 'Laki-Laki' : 'Perempuan'}
										</span>
									</td>
									<td class="px-6 py-4 md:px-8 md:py-5">
										<div class="flex justify-end gap-2">
											<button
												onclick={() => handleResetAbsen(student.user_id, student.nama)}
												disabled={isResetting}
												class="rounded-xl border border-amber-200 bg-amber-50 px-4 py-2 text-[10px] font-black tracking-widest text-amber-600 uppercase transition-all hover:bg-amber-500 hover:text-white disabled:opacity-50"
											>
												Reset absensi
											</button>
											<button
												onclick={() => openRemoveModal(student.user_id, student.nama)}
												class="rounded-xl border border-red-200 bg-red-50 px-4 py-2 text-[10px] font-black tracking-widest text-red-600 uppercase transition-all hover:bg-red-600 hover:text-white"
											>
												Keluarkan
											</button>
										</div>
									</td>
								</tr>
							{/each}
						{/if}
					</tbody>
				</table>
			</div>
		</div>
	</main>

	{#if showAddModal}
		<div
			class="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/70 p-4 backdrop-blur-md md:p-6"
		>
			<div class="w-full max-w-lg rounded-[2.5rem] bg-white p-8 shadow-2xl md:p-10">
				<div class="mb-6 flex items-center justify-between">
					<h3 class="text-xl font-black tracking-tight text-slate-900 uppercase">Tambahkan siswa ke kelas</h3>
					<button
						onclick={() => (showAddModal = false)}
						class="rounded-full bg-slate-100 p-2 text-slate-400 hover:text-red-500"
					>
						<svg
							xmlns="http://www.w3.org/2000/svg"
							class="h-5 w-5"
							viewBox="0 0 24 24"
							fill="none"
							stroke="currentColor"
							stroke-width="2"
							stroke-linecap="round"
							stroke-linejoin="round"><path d="M18 6 6 18" /><path d="m6 6 12 12" /></svg
						>
					</button>
				</div>

				<div class="mb-6 space-y-4">
					<div class="flex flex-col gap-2">
						<label
							for="select-new-student"
							class="ml-1 text-[10px] font-black tracking-widest text-slate-400 uppercase"
							>Pilih Siswa</label
						>
						<select
							id="select-new-student"
							bind:value={selectedNewStudentId}
							class="h-[54px] w-full rounded-2xl border-2 border-slate-100 bg-slate-50 px-4 font-bold text-slate-700 outline-none focus:border-brand-blue/30"
						>
							<option value="" disabled selected>-- Cari & Pilih Siswa --</option>
							{#each allAvailableStudents as s}
								<option value={s.user_id}>{s.nama} ({s.nis})</option>
							{/each}
						</select>
						<p class="mt-1 ml-1 text-xs text-slate-400">
							Pastikan siswa belum terdaftar di kelas ini.
						</p>
					</div>
				</div>

				<div class="flex gap-3">
					<button
						onclick={() => (showAddModal = false)}
						class="flex-1 rounded-2xl bg-slate-100 py-4 text-[10px] font-black tracking-widest text-slate-500 uppercase transition-colors hover:bg-slate-200"
					>
						Batal
					</button>
					<button
						onclick={handleAssignStudent}
						disabled={!selectedNewStudentId || isProcessing}
						class="flex-1 rounded-2xl bg-brand-blue py-4 text-[10px] font-black tracking-widest text-white uppercase shadow-lg shadow-blue-500/20 transition-all hover:bg-blue-700 disabled:opacity-50"
					>
						{isProcessing ? 'Menyimpan...' : 'Tambahkan'}
					</button>
				</div>
			</div>
		</div>
	{/if}

	{#if showRemoveModal}
		<div
			class="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/70 p-4 backdrop-blur-md md:p-6"
		>
			<div
				class="scale-in-center w-full max-w-md rounded-[2.5rem] bg-white p-8 text-center shadow-2xl md:p-10"
			>
				<div
					class="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-red-100 text-red-500"
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
						><path
							d="m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3Z"
						/><path d="M12 9v4" /><path d="M12 17h.01" /></svg
					>
				</div>

				<h3 class="mb-2 text-xl font-black tracking-tight text-slate-900 uppercase">
					Keluarkan Siswa?
				</h3>
				<p class="mb-6 text-sm font-medium text-slate-500">
					Anda akan mengeluarkan <span class="font-bold text-slate-800">{studentToRemove.name}</span
					> dari kelas ini.
				</p>

				<label
					class="mb-8 flex cursor-pointer items-start gap-3 rounded-2xl border-2 border-slate-100 bg-slate-50 p-4 text-left transition-colors hover:border-red-200"
				>
					<div class="flex h-5 items-center">
						<input
							type="checkbox"
							bind:checked={deleteLogsChecked}
							class="h-5 w-5 rounded border-slate-300 text-red-600 focus:ring-red-500 focus:ring-offset-2"
						/>
					</div>
					<div class="flex flex-col">
						<span class="text-xs font-black tracking-widest text-slate-700 uppercase"
							>Hapus Riwayat Absensi</span
						>
						<span class="mt-0.5 text-[10px] leading-relaxed text-slate-500">
							Centang ini jika Anda juga ingin menghapus seluruh jejak absensi siswa ini pada
							sesi-sesi kelas ini. (Data tidak dapat dikembalikan).
						</span>
					</div>
				</label>

				<div class="flex gap-3">
					<button
						onclick={() => (showRemoveModal = false)}
						class="flex-1 rounded-2xl bg-slate-100 py-4 text-[10px] font-black tracking-widest text-slate-500 uppercase transition-colors hover:bg-slate-200"
					>
						Batal
					</button>
					<button
						onclick={executeRemoveStudent}
						disabled={isRemoving}
						class="flex-1 rounded-2xl bg-red-600 py-4 text-[10px] font-black tracking-widest text-white uppercase shadow-lg shadow-red-500/20 transition-all hover:bg-red-700 disabled:opacity-50"
					>
						{isRemoving ? 'Memproses...' : 'Keluarkan Siswa'}
					</button>
				</div>
			</div>
		</div>
	{/if}
</div>
