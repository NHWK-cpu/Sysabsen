<script lang="ts">
	// State Svelte 5
	let activeMenu = $state('dashboard');

	// Modals State
	let showAddModal = $state(false);
	let showResetModal = $state(false);
	let showPeriodModal = $state(false);

	let resetData = $state({ username: '', kataKunci: '', newPassword: '', confirmPassword: '' });

	const menuItems = [
		{ id: 'dashboard', label: 'Dashboard' },
		{ id: 'guru', label: 'Manajemen Guru' },
		{ id: 'siswa', label: 'Manajemen Siswa' },
		{ id: 'kelas', label: 'Manajemen Kelas' },
		{ id: 'mapel', label: 'Manajemen Mapel' },
		{ id: 'perangkat', label: 'Persetujuan Perangkat' }
	];

	const stats = { totalUsers: 370, activeUsers: 320, inactiveUsers: 50, pendingApproval: 2 };

	const recentActivities = [
		{ id: 1, time: '08:45 WIB', user: 'Ahmad Hafizh', role: 'Student', status: 'Online' },
		{ id: 2, time: '08:30 WIB', user: 'Bpk. Budi S.', role: 'Teacher', status: 'Online' },
		{
			id: 3,
			time: '07:15 WIB',
			user: 'Unknown User',
			role: 'Guest / New Device',
			status: 'Pending'
		}
	];

	// --- DATA STATE (MOCKUP DATABASE) ---
	let teachers = $state([
		{ id: 1, name: 'Bpk. Budi Santoso', nip: '19800101', subject: 'Fisika' },
		{ id: 2, name: 'Ibu Ratna Sari', nip: '19850212', subject: 'Matematika' }
	]);

	let students = $state([
		{ id: 1, name: 'Ahmad Hafizh', nis: '2201001', class: '12-IPA-1' },
		{ id: 2, name: 'Siti Aminah', nis: '2201002', class: '12-IPA-1' }
	]);

	let periods = $state([
		{ id: 1, tahunAjar: '2025/2026', statusAktif: 'Tidak Aktif' },
		{ id: 2, tahunAjar: '2026/2027', statusAktif: 'Aktif' }
	]);

	let classes = $state([
		{ id: 1, name: '10 - IPA 1', periode_id: 2 },
		{ id: 2, name: '11 - IPS 2', periode_id: 2 }
	]);

	// DATA UPDATE: Mapel (Tanpa Kode)
	let subjects = $state([
		{ id: 1, name: 'Matematika Wajib' },
		{ id: 2, name: 'Fisika Dasar' }
	]);

	// DATA BARU: Persetujuan Perangkat
	let pendingDevices = $state([
		{
			id: 101,
			name: 'Ahmad Hafizh',
			nis: '2201001',
			device: 'Xiaomi Poco F5',
			ip: '192.168.1.15',
			time: '07:15 WIB',
			status: 'Pending'
		},
		{
			id: 102,
			name: 'Siti Aminah',
			nis: '2201002',
			device: 'Samsung Galaxy A54',
			ip: '192.168.1.20',
			time: '08:00 WIB',
			status: 'Pending'
		}
	]);

	// --- FORM STATE ---
	let newUser = $state({
		username: '',
		password: '',
		identifier: '',
		namaLengkap: '',
		labelKataKunci: '',
		kataKunci: '',
		email: ''
	});
	let newClass = $state({ nama_kelas: '', periode_id: '' });
	let newSubject = $state({ name: '' }); // Update form state Mapel
	let newPeriod = $state({ tahun_ajar: '', status_aktif: 'Aktif' });

	// --- FUNGSI CRUD ---
	const handleAddEntity = (e: Event) => {
		e.preventDefault();
		if (activeMenu === 'guru') {
			teachers.push({
				id: Date.now(),
				name: newUser.namaLengkap,
				nip: newUser.identifier,
				subject: '-'
			});
		} else if (activeMenu === 'siswa') {
			students.push({
				id: Date.now(),
				name: newUser.namaLengkap,
				nis: newUser.identifier,
				class: '-'
			});
		} else if (activeMenu === 'kelas') {
			classes.push({
				id: Date.now(),
				name: newClass.nama_kelas,
				periode_id: parseInt(newClass.periode_id)
			});
		} else if (activeMenu === 'mapel') {
			// Menambah mapel hanya dengan nama
			subjects.push({ id: Date.now(), name: newSubject.name });
		}

		showAddModal = false;
		newUser = {
			username: '',
			password: '',
			identifier: '',
			namaLengkap: '',
			labelKataKunci: '',
			kataKunci: '',
			email: ''
		};
		newClass = { nama_kelas: '', periode_id: '' };
		newSubject = { name: '' };
	};

	const handleAddPeriod = (e: Event) => {
		e.preventDefault();
		periods.push({
			id: Date.now(),
			tahunAjar: newPeriod.tahun_ajar,
			statusAktif: newPeriod.status_aktif
		});
		showPeriodModal = false;
		newPeriod = { tahun_ajar: '', status_aktif: 'Aktif' };
	};

	const deletePeriod = (id: number) => {
		if (confirm('Yakin ingin menghapus periode ini?')) {
			periods = periods.filter((p) => p.id !== id);
		}
	};

	const openResetModal = (student: any) => {
		resetData.username = student.nis;
		resetData.kataKunci = '';
		resetData.newPassword = '';
		resetData.confirmPassword = '';
		showResetModal = true;
	};

	const handleResetPassword = (e: Event) => {
		e.preventDefault();
		if (resetData.newPassword !== resetData.confirmPassword) {
			alert('Password baru dan konfirmasi password tidak cocok!');
			return;
		}
		alert(`Password untuk ${resetData.username} berhasil direset!`);
		showResetModal = false;
	};

	// FUNGSI BARU: Persetujuan Perangkat
	const approveDevice = (id: number) => {
		const idx = pendingDevices.findIndex((d) => d.id === id);
		if (idx !== -1) pendingDevices[idx].status = 'Disetujui';
	};

	const rejectDevice = (id: number) => {
		const idx = pendingDevices.findIndex((d) => d.id === id);
		if (idx !== -1) pendingDevices[idx].status = 'Ditolak';
	};
</script>

<div class="flex min-h-screen flex-col bg-slate-50 font-sans">
	<header
		class="sticky top-0 z-20 flex items-center justify-between border-b border-slate-100 bg-white px-4 py-4 shadow-sm md:px-8"
	>
		<div class="flex items-center gap-3">
			<div
				class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-slate-900 font-bold text-white"
			>
				A
			</div>
			<div>
				<h1 class="text-lg leading-none font-black tracking-tighter text-slate-800">ABSENSI</h1>
				<p class="text-[10px] font-bold tracking-widest text-slate-400 uppercase">Administrator</p>
			</div>
		</div>
		<div class="flex items-center gap-4 md:gap-6">
			<div class="hidden border-r border-slate-100 pr-6 text-right md:block">
				<p class="text-sm font-black tracking-tight text-slate-800 uppercase">Hafizh Admin</p>
				<p class="mt-0.5 text-[10px] font-black tracking-widest text-slate-500 uppercase">
					Super Admin
				</p>
			</div>
			<button
				class="rounded-xl bg-red-50 px-4 py-2 text-[10px] font-black tracking-widest text-red-600 uppercase transition-colors hover:bg-red-100 md:px-5 md:py-2.5"
			>
				Logout
			</button>
		</div>
	</header>

	<main class="mx-auto flex w-full max-w-7xl flex-1 flex-col gap-8 p-4 md:p-8 lg:flex-row">
		<aside class="w-full shrink-0 lg:w-72">
			<div class="sticky top-28 rounded-[2rem] border border-slate-100 bg-white p-6 shadow-sm">
				<h3 class="mb-6 px-2 text-xs font-black tracking-widest text-slate-800 uppercase">
					Admin Menu
				</h3>
				<nav class="flex flex-col gap-2">
					{#each menuItems as item}
						<button
							onclick={() => (activeMenu = item.id)}
							class="w-full rounded-2xl px-5 py-3.5 text-left text-[10px] font-black tracking-widest uppercase transition-all {activeMenu ===
							item.id
								? 'bg-slate-900 text-white shadow-lg shadow-slate-900/20'
								: 'text-slate-500 hover:bg-slate-50 hover:text-slate-800'}"
						>
							{item.label}
						</button>
					{/each}
				</nav>
			</div>
		</aside>

		<div class="flex-1 space-y-8">
			{#if activeMenu === 'dashboard'}
				<div class="rounded-[2.5rem] border border-slate-100 bg-white p-6 shadow-sm md:p-8">
					<h2 class="text-2xl font-black tracking-tight text-slate-900 md:text-3xl">
						Halo, Admin Utama 👋
					</h2>
					<p class="mt-2 text-sm leading-relaxed font-medium text-slate-500 md:text-base">
						Ringkasan statistik sistem absensi hari ini.
					</p>
				</div>
				<div class="grid grid-cols-1 gap-4 sm:grid-cols-2 md:gap-6 xl:grid-cols-4">
					<div class="rounded-[2rem] border border-slate-100 bg-white p-5 shadow-sm xl:p-6">
						<p class="mb-2 text-[10px] font-black tracking-[0.2em] text-slate-400 uppercase">
							Total Users
						</p>
						<h2 class="text-3xl font-black text-slate-900 xl:text-4xl">{stats.totalUsers}</h2>
					</div>
					<div class="rounded-[2rem] border border-slate-100 bg-white p-5 shadow-sm xl:p-6">
						<p class="mb-2 text-[10px] font-black tracking-[0.2em] text-slate-400 uppercase">
							Active Users
						</p>
						<h2 class="text-3xl font-black text-green-600 xl:text-4xl">{stats.activeUsers}</h2>
					</div>
					<div class="rounded-[2rem] border border-slate-100 bg-white p-5 shadow-sm xl:p-6">
						<p class="mb-2 text-[10px] font-black tracking-[0.2em] text-slate-400 uppercase">
							Inactive Users
						</p>
						<h2 class="text-3xl font-black text-slate-400 xl:text-4xl">{stats.inactiveUsers}</h2>
					</div>
					<div
						class="relative flex flex-col justify-between overflow-hidden rounded-[2rem] border border-transparent bg-brand-blue p-5 text-white shadow-lg shadow-blue-500/20 xl:p-6"
					>
						<div class="relative z-10">
							<p class="mb-2 text-[10px] font-black tracking-[0.2em] text-blue-200 uppercase">
								Pending Approval
							</p>
							<h2 class="text-3xl font-black xl:text-4xl">{stats.pendingApproval}</h2>
						</div>
					</div>
				</div>
				<div class="overflow-hidden rounded-[2.5rem] border border-slate-100 bg-white shadow-sm">
					<div class="border-b border-slate-50 bg-white p-6 md:p-8">
						<h3 class="text-xs font-black tracking-widest text-slate-800 uppercase">
							Login Terbaru
						</h3>
					</div>
					<div class="w-full overflow-x-auto">
						<table class="w-full min-w-[700px] border-collapse text-left">
							<thead>
								<tr class="bg-slate-50/50">
									<th
										class="w-32 px-6 py-4 text-[10px] font-black tracking-[0.2em] text-slate-400 uppercase md:px-8 md:py-5"
										>Waktu Login</th
									>
									<th
										class="px-6 py-4 text-[10px] font-black tracking-[0.2em] text-slate-400 uppercase md:px-8 md:py-5"
										>Nama Pengguna</th
									>
									<th
										class="px-6 py-4 text-[10px] font-black tracking-[0.2em] text-slate-400 uppercase md:px-8 md:py-5"
										>Role</th
									>
									<th
										class="px-6 py-4 text-center text-[10px] font-black tracking-[0.2em] text-slate-400 uppercase md:px-8 md:py-5"
										>Status</th
									>
								</tr>
							</thead>
							<tbody class="divide-y divide-slate-50">
								{#each recentActivities as activity}
									<tr class="transition-colors hover:bg-slate-50/30">
										<td class="px-6 py-4 text-xs font-bold text-slate-400 md:px-8 md:py-5"
											>{activity.time}</td
										>
										<td class="px-6 py-4 font-bold text-slate-800 md:px-8 md:py-5"
											>{activity.user}</td
										>
										<td class="px-6 py-4 text-sm font-medium text-slate-600 italic md:px-8 md:py-5"
											>{activity.role}</td
										>
										<td class="px-6 py-4 text-center md:px-8 md:py-5">
											<span
												class="rounded-full border px-4 py-1.5 text-[9px] font-black tracking-widest whitespace-nowrap uppercase {activity.status ===
												'Online'
													? 'border-green-200 bg-green-100 text-green-700'
													: 'border-amber-200 bg-amber-100 text-amber-700'}">{activity.status}</span
											>
										</td>
									</tr>
								{/each}
							</tbody>
						</table>
					</div>
				</div>
			{:else if activeMenu === 'guru' || activeMenu === 'siswa'}
				<div
					class="flex flex-col justify-between gap-4 rounded-[2.5rem] border border-slate-100 bg-white p-6 shadow-sm sm:flex-row sm:items-center md:p-8"
				>
					<div>
						<h2 class="text-2xl font-black tracking-tight text-slate-900 uppercase">
							Manajemen {activeMenu === 'guru' ? 'Guru' : 'Siswa'}
						</h2>
						<p class="mt-1 text-sm font-medium text-slate-500">
							Kelola data {activeMenu} terdaftar di sistem.
						</p>
					</div>
					<button
						onclick={() => (showAddModal = true)}
						class="w-full rounded-2xl bg-brand-blue px-6 py-3.5 text-[10px] font-black tracking-widest text-white uppercase shadow-lg shadow-blue-500/20 transition-all hover:scale-105 sm:w-auto"
					>
						+ Tambah Baru
					</button>
				</div>
				<div class="overflow-hidden rounded-[2.5rem] border border-slate-100 bg-white shadow-sm">
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
										class="px-6 py-4 text-[10px] font-black tracking-[0.2em] text-slate-400 uppercase md:px-8 md:py-5"
										>{activeMenu === 'guru' ? 'NIP' : 'NIS'}</th
									>
									<th
										class="px-6 py-4 text-[10px] font-black tracking-[0.2em] text-slate-400 uppercase md:px-8 md:py-5"
										>{activeMenu === 'guru' ? 'Mapel' : 'Kelas'}</th
									>
									<th
										class="px-6 py-4 text-right text-[10px] font-black tracking-[0.2em] text-slate-400 uppercase md:px-8 md:py-5"
										>Aksi</th
									>
								</tr>
							</thead>
							<tbody class="divide-y divide-slate-50">
								{#each activeMenu === 'guru' ? teachers : students as user, i}
									<tr class="transition-colors hover:bg-slate-50/30">
										<td class="px-6 py-4 text-sm font-bold text-slate-400 md:px-8 md:py-5"
											>{i + 1}</td
										>
										<td class="px-6 py-4 font-bold text-slate-800 md:px-8 md:py-5">{user.name}</td>
										<td class="px-6 py-4 font-medium text-slate-500 md:px-8 md:py-5"
											>{user.nip || user.nis}</td
										>
										<td class="px-6 py-4 font-medium text-slate-500 md:px-8 md:py-5"
											>{user.subject || user.class}</td
										>
										<td class="px-6 py-4 text-right md:px-8 md:py-5">
											<div class="flex justify-end gap-2">
												{#if activeMenu === 'siswa'}
													<button
														onclick={() => openResetModal(user)}
														class="rounded-xl bg-amber-50 px-4 py-2 text-[10px] font-black tracking-widest text-amber-600 uppercase transition-all hover:bg-amber-100"
														>Reset Pass</button
													>
												{/if}
												<button
													class="rounded-xl bg-slate-50 px-4 py-2 text-[10px] font-black tracking-widest text-brand-blue uppercase transition-all hover:bg-blue-50"
													>Edit</button
												>
												<button
													class="rounded-xl bg-slate-50 px-4 py-2 text-[10px] font-black tracking-widest text-red-500 uppercase transition-all hover:bg-red-50"
													>Hapus</button
												>
											</div>
										</td>
									</tr>
								{/each}
							</tbody>
						</table>
					</div>
				</div>
			{:else if activeMenu === 'kelas'}
				<div class="grid grid-cols-1 gap-8 lg:grid-cols-3">
					<div class="flex flex-col gap-6 lg:col-span-1">
						<div
							class="flex items-center justify-between rounded-[2.5rem] border border-slate-100 bg-white p-6 shadow-sm"
						>
							<h3 class="text-sm font-black tracking-widest text-slate-800 uppercase">
								Periode Ajar
							</h3>
							<button
								onclick={() => (showPeriodModal = true)}
								class="flex h-10 w-10 items-center justify-center rounded-full bg-brand-blue text-lg font-bold text-white shadow-lg shadow-blue-500/20 transition-all hover:scale-105"
								>+</button
							>
						</div>

						<div
							class="overflow-hidden rounded-[2rem] border border-slate-100 bg-white p-2 shadow-sm"
						>
							<div class="space-y-2">
								{#each periods as period}
									<div
										class="flex items-center justify-between rounded-2xl border border-slate-100 bg-slate-50 p-4"
									>
										<div>
											<p class="font-black tracking-tight text-slate-800">{period.tahunAjar}</p>
											<span
												class="mt-1 inline-block rounded-full px-3 py-1 text-[8px] font-black tracking-widest uppercase {period.statusAktif ===
												'Aktif'
													? 'bg-green-100 text-green-700'
													: 'bg-slate-200 text-slate-500'}"
											>
												{period.statusAktif}
											</span>
										</div>
										<button
											onclick={() => deletePeriod(period.id)}
											class="p-2 text-red-400 transition-colors hover:text-red-600"
										>
											<svg
												xmlns="http://www.w3.org/2000/svg"
												class="h-5 w-5"
												viewBox="0 0 24 24"
												fill="none"
												stroke="currentColor"
												stroke-width="2"
												stroke-linecap="round"
												stroke-linejoin="round"
												><path d="M3 6h18" /><path d="M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6" /><path
													d="M8 6V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2"
												/></svg
											>
										</button>
									</div>
								{/each}
							</div>
						</div>
					</div>

					<div class="flex flex-col gap-6 lg:col-span-2">
						<div
							class="flex flex-col justify-between gap-4 rounded-[2.5rem] border border-slate-100 bg-white p-6 shadow-sm sm:flex-row sm:items-center"
						>
							<div>
								<h2 class="text-2xl font-black tracking-tight text-slate-900 uppercase">
									Daftar Kelas
								</h2>
							</div>
							<button
								onclick={() => (showAddModal = true)}
								class="rounded-2xl bg-brand-blue px-6 py-3.5 text-[10px] font-black tracking-widest text-white uppercase shadow-lg shadow-blue-500/20 transition-all hover:scale-105"
							>
								+ Tambah Kelas
							</button>
						</div>

						<div
							class="overflow-hidden rounded-[2.5rem] border border-slate-100 bg-white shadow-sm"
						>
							<div class="w-full overflow-x-auto">
								<table class="w-full min-w-[500px] border-collapse text-left">
									<thead>
										<tr class="bg-slate-50/50">
											<th
												class="px-6 py-5 text-[10px] font-black tracking-[0.2em] text-slate-400 uppercase"
												>Nama Kelas</th
											>
											<th
												class="px-6 py-5 text-[10px] font-black tracking-[0.2em] text-slate-400 uppercase"
												>Periode Ref</th
											>
											<th
												class="px-6 py-5 text-right text-[10px] font-black tracking-[0.2em] text-slate-400 uppercase"
												>Aksi</th
											>
										</tr>
									</thead>
									<tbody class="divide-y divide-slate-50">
										{#each classes as item}
											<tr class="transition-colors hover:bg-slate-50/30">
												<td class="px-6 py-5 font-black text-slate-800">{item.name}</td>
												<td class="px-6 py-5 font-medium text-slate-500">
													<span class="rounded-lg bg-slate-100 px-3 py-1 text-xs font-bold">
														{periods.find((p) => p.id === item.periode_id)?.tahunAjar || 'Unknown'}
													</span>
												</td>
												<td class="px-6 py-5 text-right">
													<div class="flex justify-end gap-2">
														<button
															class="rounded-xl bg-slate-50 px-4 py-2 text-[10px] font-black tracking-widest text-red-500 uppercase transition-all hover:bg-red-50"
															>Hapus</button
														>
													</div>
												</td>
											</tr>
										{/each}
									</tbody>
								</table>
							</div>
						</div>
					</div>
				</div>
			{:else if activeMenu === 'mapel'}
				<div
					class="flex flex-col justify-between gap-4 rounded-[2.5rem] border border-slate-100 bg-white p-6 shadow-sm sm:flex-row sm:items-center md:p-8"
				>
					<div>
						<h2 class="text-2xl font-black tracking-tight text-slate-900 uppercase">
							Manajemen Mapel
						</h2>
						<p class="mt-1 text-sm font-medium text-slate-500">Kelola referensi mata pelajaran.</p>
					</div>
					<button
						onclick={() => (showAddModal = true)}
						class="w-full rounded-2xl bg-brand-blue px-6 py-3.5 text-[10px] font-black tracking-widest text-white uppercase shadow-lg shadow-blue-500/20 transition-all hover:scale-105 sm:w-auto"
					>
						+ Tambah Mapel
					</button>
				</div>
				<div class="overflow-hidden rounded-[2.5rem] border border-slate-100 bg-white shadow-sm">
					<div class="w-full overflow-x-auto">
						<table class="w-full min-w-[500px] border-collapse text-left">
							<thead>
								<tr class="bg-slate-50/50">
									<th
										class="w-16 px-6 py-4 text-[10px] font-black tracking-[0.2em] text-slate-400 uppercase md:px-8 md:py-5"
										>ID (Auto)</th
									>
									<th
										class="px-6 py-4 text-[10px] font-black tracking-[0.2em] text-slate-400 uppercase md:px-8 md:py-5"
										>Nama Mata Pelajaran</th
									>
									<th
										class="px-6 py-4 text-right text-[10px] font-black tracking-[0.2em] text-slate-400 uppercase md:px-8 md:py-5"
										>Aksi</th
									>
								</tr>
							</thead>
							<tbody class="divide-y divide-slate-50">
								{#each subjects as item}
									<tr class="transition-colors hover:bg-slate-50/30">
										<td class="px-6 py-4 text-sm font-bold text-slate-400 md:px-8 md:py-5"
											>{item.id}</td
										>
										<td class="px-6 py-4 font-bold text-slate-800 md:px-8 md:py-5">{item.name}</td>
										<td class="px-6 py-4 text-right md:px-8 md:py-5">
											<div class="flex justify-end gap-2">
												<button
													class="rounded-xl bg-slate-50 px-4 py-2 text-[10px] font-black tracking-widest text-red-500 uppercase transition-all hover:bg-red-50"
													>Hapus</button
												>
											</div>
										</td>
									</tr>
								{/each}
							</tbody>
						</table>
					</div>
				</div>
			{:else if activeMenu === 'perangkat'}
				<div
					class="flex flex-col justify-between gap-4 rounded-[2.5rem] border border-slate-100 bg-white p-6 shadow-sm sm:flex-row sm:items-center md:p-8"
				>
					<div>
						<h2 class="text-2xl font-black tracking-tight text-slate-900 uppercase">
							Persetujuan Perangkat
						</h2>
						<p class="mt-1 text-sm font-medium text-slate-500">
							Kelola akses *login* perangkat baru dari siswa.
						</p>
					</div>
				</div>
				<div class="overflow-hidden rounded-[2.5rem] border border-slate-100 bg-white shadow-sm">
					<div class="w-full overflow-x-auto">
						<table class="w-full min-w-[900px] border-collapse text-left">
							<thead>
								<tr class="bg-slate-50/50">
									<th
										class="w-16 px-6 py-4 text-[10px] font-black tracking-[0.2em] text-slate-400 uppercase md:px-8 md:py-5"
										>Siswa</th
									>
									<th
										class="px-6 py-4 text-[10px] font-black tracking-[0.2em] text-slate-400 uppercase md:px-8 md:py-5"
										>Perangkat & IP</th
									>
									<th
										class="px-6 py-4 text-[10px] font-black tracking-[0.2em] text-slate-400 uppercase md:px-8 md:py-5"
										>Waktu Request</th
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
								{#each pendingDevices as dev}
									<tr class="transition-colors hover:bg-slate-50/30">
										<td class="px-6 py-4 md:px-8 md:py-5">
											<p class="font-bold text-slate-800">{dev.name}</p>
											<p class="text-[10px] font-bold text-slate-400">NIS: {dev.nis}</p>
										</td>
										<td class="px-6 py-4 md:px-8 md:py-5">
											<p class="font-bold text-slate-700">{dev.device}</p>
											<p class="text-[10px] font-medium text-slate-500">IP: {dev.ip}</p>
										</td>
										<td class="px-6 py-4 text-sm font-medium text-slate-500 md:px-8 md:py-5"
											>{dev.time}</td
										>
										<td class="px-6 py-4 text-center md:px-8 md:py-5">
											<span
												class="rounded-full border px-4 py-1.5 text-[9px] font-black tracking-widest whitespace-nowrap uppercase
												{dev.status === 'Disetujui'
													? 'border-green-200 bg-green-100 text-green-700'
													: dev.status === 'Ditolak'
														? 'border-red-200 bg-red-100 text-red-700'
														: 'border-amber-200 bg-amber-100 text-amber-700'}"
											>
												{dev.status}
											</span>
										</td>
										<td class="px-6 py-4 text-right md:px-8 md:py-5">
											{#if dev.status === 'Pending'}
												<div class="flex justify-end gap-2">
													<button
														onclick={() => approveDevice(dev.id)}
														class="rounded-xl bg-brand-blue px-4 py-2 text-[10px] font-black tracking-widest text-white uppercase shadow-lg shadow-blue-500/20 transition-all hover:scale-105"
														>Setujui</button
													>
													<button
														onclick={() => rejectDevice(dev.id)}
														class="rounded-xl border border-slate-200 bg-slate-50 px-4 py-2 text-[10px] font-black tracking-widest text-red-500 uppercase transition-all hover:bg-red-50"
														>Tolak</button
													>
												</div>
											{/if}
										</td>
									</tr>
								{/each}
							</tbody>
						</table>
					</div>
				</div>
			{/if}
		</div>
	</main>

	{#if showAddModal}
		<div
			class="fixed inset-0 z-50 flex items-center justify-center overflow-y-auto bg-slate-900/70 p-4 backdrop-blur-md md:p-6"
		>
			<div
				class="scale-in-center my-auto w-full max-w-2xl rounded-[3rem] bg-white p-8 shadow-2xl md:p-10"
			>
				<div class="mb-8 text-center">
					<h3 class="text-2xl font-black tracking-tight text-slate-900 uppercase">
						Tambah {activeMenu === 'guru'
							? 'Guru'
							: activeMenu === 'siswa'
								? 'Siswa'
								: activeMenu === 'kelas'
									? 'Kelas'
									: 'Mata Pelajaran'}
					</h3>
					<p class="mt-2 text-[10px] font-bold tracking-widest text-slate-400 uppercase">
						Input sesuai struktur model database
					</p>
				</div>

				<form onsubmit={handleAddEntity} class="space-y-5">
					{#if activeMenu === 'guru' || activeMenu === 'siswa'}
						<div class="grid grid-cols-1 gap-5 sm:grid-cols-2">
							<div class="space-y-5">
								<div class="flex flex-col gap-2">
									<label
										for="namaLengkap"
										class="ml-1 text-[10px] font-black tracking-widest text-slate-400 uppercase"
										>Nama Lengkap</label
									>
									<input
										type="text"
										id="namaLengkap"
										bind:value={newUser.namaLengkap}
										class="rounded-2xl border-2 border-transparent bg-slate-50 px-5 py-4 font-bold text-slate-700 outline-none focus:border-brand-blue/20"
										required
									/>
								</div>
								<div class="flex flex-col gap-2">
									<label
										for="username"
										class="ml-1 text-[10px] font-black tracking-widest text-slate-400 uppercase"
										>Username</label
									>
									<input
										type="text"
										id="username"
										bind:value={newUser.username}
										class="rounded-2xl border-2 border-transparent bg-slate-50 px-5 py-4 font-bold text-slate-700 outline-none focus:border-brand-blue/20"
										required
									/>
								</div>
								<div class="flex flex-col gap-2">
									<label
										for="id"
										class="ml-1 text-[10px] font-black tracking-widest text-slate-400 uppercase"
										>{activeMenu === 'guru' ? 'NIP' : 'NIS'}</label
									>
									<input
										type="text"
										id="id"
										bind:value={newUser.identifier}
										class="rounded-2xl border-2 border-transparent bg-slate-50 px-5 py-4 font-bold text-slate-700 outline-none focus:border-brand-blue/20"
										required
									/>
								</div>
							</div>
							<div class="space-y-5">
								<div class="flex flex-col gap-2">
									<label
										for="pass"
										class="ml-1 text-[10px] font-black tracking-widest text-slate-400 uppercase"
										>Password Default</label
									>
									<input
										type="password"
										id="pass"
										bind:value={newUser.password}
										class="rounded-2xl border-2 border-transparent bg-slate-50 px-5 py-4 font-bold text-slate-700 outline-none focus:border-brand-blue/20"
										required
									/>
								</div>
								{#if activeMenu === 'siswa'}
									<div class="flex flex-col gap-2">
										<label
											for="labelKunci"
											class="ml-1 text-[10px] font-black tracking-widest text-slate-400 uppercase"
											>Label Clue Keamanan</label
										>
										<input
											type="text"
											id="labelKunci"
											bind:value={newUser.labelKataKunci}
											class="rounded-2xl border-2 border-transparent bg-slate-50 px-5 py-4 font-bold text-slate-700 outline-none focus:border-brand-blue/20"
											required
										/>
									</div>
									<div class="flex flex-col gap-2">
										<label
											for="kataKunci"
											class="ml-1 text-[10px] font-black tracking-widest text-slate-400 uppercase"
											>Jawaban Keamanan</label
										>
										<input
											type="text"
											id="kataKunci"
											bind:value={newUser.kataKunci}
											class="rounded-2xl border-2 border-transparent bg-slate-50 px-5 py-4 font-bold text-slate-700 outline-none focus:border-brand-blue/20"
											required
										/>
									</div>
								{:else}
									<div class="flex flex-col gap-2">
										<label
											for="email"
											class="ml-1 text-[10px] font-black tracking-widest text-slate-400 uppercase"
											>Email</label
										>
										<input
											type="email"
											id="email"
											bind:value={newUser.email}
											class="rounded-2xl border-2 border-transparent bg-slate-50 px-5 py-4 font-bold text-slate-700 outline-none focus:border-brand-blue/20"
											required
										/>
									</div>
								{/if}
							</div>
						</div>
					{:else if activeMenu === 'kelas'}
						<div class="space-y-5">
							<div class="flex flex-col gap-2">
								<label
									for="namaKelas"
									class="ml-1 text-[10px] font-black tracking-widest text-slate-400 uppercase"
									>Nama Kelas</label
								>
								<input
									type="text"
									id="namaKelas"
									bind:value={newClass.nama_kelas}
									placeholder="Misal: 10 - Rekayasa Perangkat Lunak 1"
									class="rounded-2xl border-2 border-transparent bg-slate-50 px-5 py-4 font-bold text-slate-700 outline-none focus:border-brand-blue/20"
									required
								/>
							</div>
							<div class="flex flex-col gap-2">
								<label
									for="periodeId"
									class="ml-1 text-[10px] font-black tracking-widest text-slate-400 uppercase"
									>Pilih Periode Belajar</label
								>
								<select
									id="periodeId"
									bind:value={newClass.periode_id}
									class="cursor-pointer rounded-2xl border-2 border-transparent bg-slate-50 px-5 py-4 font-bold text-slate-700 outline-none focus:border-brand-blue/20"
									required
								>
									<option value="" disabled selected>-- Pilih Tahun Ajar --</option>
									{#each periods as period}
										<option value={period.id}>{period.tahunAjar} ({period.statusAktif})</option>
									{/each}
								</select>
							</div>
						</div>
					{:else if activeMenu === 'mapel'}
						<div class="space-y-5">
							<div class="flex flex-col gap-2">
								<label
									for="namaMapel"
									class="ml-1 text-[10px] font-black tracking-widest text-slate-400 uppercase"
									>Nama Mata Pelajaran</label
								>
								<input
									type="text"
									id="namaMapel"
									bind:value={newSubject.name}
									placeholder="Misal: Matematika Terapan"
									class="rounded-2xl border-2 border-transparent bg-slate-50 px-5 py-4 font-bold text-slate-700 outline-none focus:border-brand-blue/20"
									required
								/>
							</div>
						</div>
					{/if}

					<div class="flex gap-3 pt-6">
						<button
							type="button"
							onclick={() => (showAddModal = false)}
							class="flex-1 rounded-2xl bg-slate-100 py-4 text-[10px] font-black tracking-widest text-slate-500 uppercase transition-all hover:bg-slate-200"
							>Batal</button
						>
						<button
							type="submit"
							class="flex-1 rounded-2xl bg-brand-blue py-4 text-[10px] font-black tracking-widest text-white uppercase shadow-lg shadow-blue-500/20 transition-all hover:bg-blue-700"
							>Simpan Data</button
						>
					</div>
				</form>
			</div>
		</div>
	{/if}

	{#if showPeriodModal}
		<div
			class="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/70 p-4 backdrop-blur-md md:p-6"
		>
			<div
				class="scale-in-center my-auto w-full max-w-sm rounded-[3rem] bg-white p-8 shadow-2xl md:p-10"
			>
				<div class="mb-8 text-center">
					<h3 class="text-xl font-black tracking-tight text-slate-900 uppercase">Tambah Periode</h3>
				</div>
				<form onsubmit={handleAddPeriod} class="space-y-5">
					<div class="flex flex-col gap-2">
						<label
							for="tahunAjar"
							class="ml-1 text-[10px] font-black tracking-widest text-slate-400 uppercase"
							>Tahun Ajar</label
						>
						<input
							type="text"
							id="tahunAjar"
							bind:value={newPeriod.tahun_ajar}
							placeholder="Misal: 2026/2027"
							class="rounded-2xl border-2 border-transparent bg-slate-50 px-5 py-4 text-center font-bold text-slate-700 outline-none focus:border-brand-blue/20"
							required
						/>
					</div>
					<div class="flex flex-col gap-2">
						<label
							for="statusAktif"
							class="ml-1 text-[10px] font-black tracking-widest text-slate-400 uppercase"
							>Status</label
						>
						<select
							id="statusAktif"
							bind:value={newPeriod.status_aktif}
							class="cursor-pointer rounded-2xl border-2 border-transparent bg-slate-50 px-5 py-4 font-bold text-slate-700 outline-none focus:border-brand-blue/20"
						>
							<option value="Aktif">Aktif</option>
							<option value="Tidak Aktif">Tidak Aktif</option>
						</select>
					</div>
					<div class="flex gap-3 pt-4">
						<button
							type="button"
							onclick={() => (showPeriodModal = false)}
							class="flex-1 rounded-2xl bg-slate-100 py-4 text-[10px] font-black tracking-widest text-slate-500 uppercase transition-all hover:bg-slate-200"
							>Batal</button
						>
						<button
							type="submit"
							class="flex-1 rounded-2xl bg-slate-900 py-4 text-[10px] font-black tracking-widest text-white uppercase shadow-lg shadow-slate-900/20 transition-all hover:bg-black"
							>Simpan</button
						>
					</div>
				</form>
			</div>
		</div>
	{/if}

	{#if showResetModal}
		<div
			class="fixed inset-0 z-50 flex items-center justify-center overflow-y-auto bg-slate-900/70 p-4 backdrop-blur-md md:p-6"
		>
			<div
				class="scale-in-center my-auto w-full max-w-md rounded-[3rem] border-t-8 border-amber-400 bg-white p-8 shadow-2xl md:p-10"
			>
				<div class="mb-8 text-center">
					<div
						class="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-amber-100 text-amber-500"
					>
						<svg
							xmlns="http://www.w3.org/2000/svg"
							class="h-8 w-8"
							viewBox="0 0 24 24"
							fill="none"
							stroke="currentColor"
							stroke-width="2.5"
							stroke-linecap="round"
							stroke-linejoin="round"
							><rect width="18" height="11" x="3" y="11" rx="2" ry="2" /><path
								d="M7 11V7a5 5 0 0 1 10 0v4"
							/></svg
						>
					</div>
					<h3 class="text-2xl font-black tracking-tight text-slate-900 uppercase">
						Reset Password
					</h3>
					<p class="mt-2 text-[10px] font-bold tracking-widest text-slate-400 uppercase">
						Ubah sandi menggunakan kata kunci
					</p>
				</div>
				<form onsubmit={handleResetPassword} class="space-y-5">
					<div class="flex flex-col gap-2">
						<label
							for="resetUser"
							class="ml-1 text-[10px] font-black tracking-widest text-slate-400 uppercase"
							>Username / NIS</label
						>
						<input
							type="text"
							id="resetUser"
							bind:value={resetData.username}
							class="cursor-not-allowed rounded-2xl border-2 border-transparent bg-slate-100 px-5 py-4 font-bold text-slate-500 outline-none"
							readonly
						/>
					</div>
					<div class="flex flex-col gap-2">
						<label
							for="resetKunci"
							class="ml-1 text-[10px] font-black tracking-widest text-amber-500 uppercase"
							>Jawaban Keamanan</label
						>
						<input
							type="text"
							id="resetKunci"
							bind:value={resetData.kataKunci}
							class="rounded-2xl border-2 border-transparent bg-amber-50 px-5 py-4 font-bold text-amber-700 transition-all outline-none placeholder:text-amber-300 focus:border-amber-400/30"
							required
						/>
					</div>
					<div class="flex flex-col gap-2 pt-2">
						<label
							for="newPass"
							class="ml-1 text-[10px] font-black tracking-widest text-slate-400 uppercase"
							>Password Baru</label
						>
						<input
							type="password"
							id="newPass"
							bind:value={resetData.newPassword}
							class="rounded-2xl border-2 border-transparent bg-slate-50 px-5 py-4 font-bold text-slate-700 transition-all outline-none focus:border-brand-blue/20"
							required
						/>
					</div>
					<div class="flex flex-col gap-2">
						<label
							for="confirmPass"
							class="ml-1 text-[10px] font-black tracking-widest text-slate-400 uppercase"
							>Konfirmasi Password</label
						>
						<input
							type="password"
							id="confirmPass"
							bind:value={resetData.confirmPassword}
							class="rounded-2xl border-2 border-transparent bg-slate-50 px-5 py-4 font-bold text-slate-700 transition-all outline-none focus:border-brand-blue/20"
							required
						/>
					</div>
					<div class="flex gap-3 pt-6">
						<button
							type="button"
							onclick={() => (showResetModal = false)}
							class="flex-1 rounded-2xl bg-slate-100 py-4 text-[10px] font-black tracking-widest text-slate-500 uppercase transition-all hover:bg-slate-200"
							>Batal</button
						>
						<button
							type="submit"
							class="flex-1 rounded-2xl bg-amber-400 py-4 text-[10px] font-black tracking-widest text-white uppercase shadow-lg shadow-amber-400/30 transition-all hover:bg-amber-500"
							>Reset Sandi</button
						>
					</div>
				</form>
			</div>
		</div>
	{/if}
</div>
