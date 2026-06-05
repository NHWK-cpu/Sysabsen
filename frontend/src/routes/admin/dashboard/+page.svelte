<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';

	// State Svelte 5
	let activeMenu = $state('dashboard');

	// Modals State
	let showAddModal = $state(false);
	let showResetModal = $state(false);
	let showPeriodModal = $state(false);
	let showAssignModal = $state(false);
	let showAssignGuruModal = $state(false);

	// --- STATE EDIT & RESET ---
	let isEditing = $state(false);
	let currentEditId = $state<number | null>(null);

	let resetData = $state({
		username: '',
		labelKataKunci: '',
		kataKunci: '',
		newPassword: '',
		confirmPassword: ''
	});

	// --- STATE PLOTTING KELAS SISWA ---
	let assignData = $state({
		siswaId: 0,
		namaSiswa: '',
		action: 'assign',
		oldKelasId: '',
		kelasId: ''
	});

	// --- STATE PLOTTING WALI KELAS GURU ---
	let assignGuruData = $state({
		guruUserId: 0,
		namaGuru: '',
		kelasId: ''
	});

	let menuItems = $state([
		{ id: 'dashboard', label: 'Beranda' },
		{ id: 'guru', label: 'Manajemen Guru' },
		{ id: 'siswa', label: 'Manajemen Siswa' },
		{ id: 'pendaftar_siswa', label: 'Pendaftar Baru' },
		{ id: 'kelas', label: 'Manajemen Kelas' },
		{ id: 'mapel', label: 'Manajemen Mapel' },
		{ id: 'perangkat', label: 'Persetujuan Perangkat' }
	]);

	let adminList = $state<any[]>([]);

	// --- API CONFIG & STATE ---
	const API_BASE_URL = import.meta.env.VITE_API_URL;
	let token = '';

	// --- STATE DATA DINAMIS ---
	let adminProfile = $state({ name: 'Memuat...', role: 'Admin utama' });
	let stats = $state({
		totalUsers: 0,
		activeUsers: 0,
		inactiveUsers: 0,
		pendingApproval: 0,
		pendingSiswaRegs: 0
	});
	let recentActivities = $state<any[]>([]);

	let teachers = $state<any[]>([]);
	let students = $state<any[]>([]);
	let studentCurrentClasses = $state<any[]>([]);
	let periods = $state<any[]>([]);
	let classes = $state<any[]>([]);
	let subjects = $state<any[]>([]);
	let pendingDevices = $state<any[]>([]);
	let pendingStudentRegs = $state<any[]>([]);

	// --- FORM STATE ---
	let newUser = $state({
		username: '',
		password: '',
		identifier: '', // Dipakai untuk NIP (Guru) atau Asal Sekolah (Siswa)
		namaLengkap: '',
		labelKataKunci: '',
		kataKunci: '',
		email: ''
	});
	let newClass = $state({ nama_kelas: '', periode_id: '' });
	let newSubject = $state({ name: '' });

	// PERBAIKAN: Tambah field semester dan default status_aktif ke angka 1
	let newPeriod = $state({ tahun_ajar: '', semester: 'Ganjil', status_aktif: 1 });

	// --- STATE BARU: DAFTAR SISWA DALAM KELAS ---
	let showClassStudentsModal = $state(false);
	let selectedClassForStudents = $state({ id: '', name: '' });
	let classStudentsList = $state<any[]>([]);

	// --- STATE BARU UNTUK DROPDOWN TAMBAH SISWA ---
	let selectedStudentToAdd = $state('');

	// --- LIFECYCLE ---
	onMount(async () => {
		token = localStorage.getItem('jwt_token') || '';
		const role = localStorage.getItem('user_role');

		// PERUBAHAN: Izinkan admin DAN super_admin
		if (!token || (role !== 'admin' && role !== 'super_admin')) {
			goto('/login/admin');
			return;
		}

		try {
			const payloadBase64 = token.split('.')[1];
			const decodedPayload = JSON.parse(atob(payloadBase64));

			adminProfile = {
				name: decodedPayload.nama || decodedPayload.username || 'Pengelola',
				role: role === 'super_admin' ? 'Admin utama' : 'Admin standar'
			};

			// Jika Admin utama, tambahkan menu khusus!
			if (role === 'super_admin') {
				menuItems.push({ id: 'admin_users', label: 'Manajemen Admin' });
			}
		} catch (e) {
			console.error('Gagal membaca profil dari token');
		}

		await fetchDashboardData();
	});

	// --- REAKTIF LOAD DATA BERDASARKAN MENU ---
	$effect(() => {
		if (token) {
			if (activeMenu === 'dashboard') fetchDashboardData();
			else if (activeMenu === 'guru' || activeMenu === 'siswa') {
				fetchUsersAll();
				if (activeMenu === 'siswa' || activeMenu === 'guru') fetchClasses();
			} else if (activeMenu === 'kelas') {
				fetchPeriods();
				fetchClasses();
			} 			else if (activeMenu === 'mapel') fetchSubjects();
			else if (activeMenu === 'perangkat') fetchPendingDevices();
			else if (activeMenu === 'pendaftar_siswa') fetchPendingStudentRegistrations();
			else if (activeMenu === 'admin_users') fetchAdminList();
		}
	});

	// --- FUNGSI BACA DATA (READ) ---
	const fetchDashboardData = async () => {
		try {
			const res = await fetch(`${API_BASE_URL}/admin/dashboard/stats`, {
				headers: { Authorization: `Bearer ${token}` }
			});
			if (res.ok) {
				const data = await res.json();
				stats = {
					totalUsers: data.total_users || 0,
					activeUsers: data.active_users || 0,
					inactiveUsers: data.inactive_users || 0,
					pendingApproval: data.pending_devices || 0,
					pendingSiswaRegs: data.pending_siswa_registrations || 0
				};
				recentActivities = data.recent_logins || [];
			}
		} catch (error) {
			console.error('Fetch Stats Error:', error);
		}
	};

	const fetchUsersAll = async () => {
		try {
			const res = await fetch(`${API_BASE_URL}/admin/users/all`, {
				headers: { Authorization: `Bearer ${token}` }
			});
			if (res.ok) {
				const data = await res.json();
				teachers = data
					.filter((u: any) => u.role === 'guru')
					.map((u: any) => ({
						id: u.id,
						name: u.nama_lengkap,
						nip: u.identifier || u.nip || '-',
						username: u.username,
						email: u.email || '',
						subject: '-',
						status:
							u.is_active === 0 || u.is_active === false || u.is_active === '0'
								? 'Nonaktif'
								: 'Aktif'
					}));

				// PERBAIKAN: Mapping untuk siswa
				students = data
					.filter((u: any) => u.role === 'siswa')
					.map((u: any) => ({
						id: u.id,
						name: u.nama_lengkap,
						namaSekolah: u.nama_sekolah || u.identifier || '-',
						username: u.username,
						class: '-',
						status:
							u.is_active === 0 || u.is_active === false || u.is_active === '0'
								? 'Nonaktif'
								: 'Aktif'
					}));
			}
		} catch (error) {
			console.error('Fetch Users Error:', error);
		}
	};

	const fetchClasses = async () => {
		try {
			const res = await fetch(`${API_BASE_URL}/admin/kelas/all`, {
				headers: { Authorization: `Bearer ${token}` }
			});
			if (res.ok) {
				const data = (await res.json()) || [];
				classes = data.map((c: any) => {
					// Ambil nilai wali_guru_id secara aman
					let rawWaliId = c.WaliGuruID !== undefined ? c.WaliGuruID : c.wali_guru_id;

					return {
						id: Number(c.ID || c.id),
						name: c.NamaKelas || c.nama_kelas,
						periode_id: Number(c.PeriodeID || c.periode_id),
						// Pastikan jika bukan null, ubah ke Number. Jika null/0, jadikan null.
						wali_guru_id: rawWaliId !== null && rawWaliId !== 0 ? Number(rawWaliId) : null
					};
				});
			}
		} catch (error) {
			console.error('Fetch Classes Error:', error);
		}
	};

	const fetchPeriods = async () => {
		try {
			const res = await fetch(`${API_BASE_URL}/admin/periode/all`, {
				headers: { Authorization: `Bearer ${token}` }
			});
			if (res.ok) {
				const data = (await res.json()) || [];
				periods = data.map((p: any) => ({
					id: p.ID || p.id,
					tahunAjar: p.TahunAjar || p.tahun_ajaran || p.tahun_ajar,
					semester: p.Semester || p.semester || '-',
					statusAktif: p.StatusAktif === 1 || p.status_aktif === 1 ? 'Aktif' : 'Tidak Aktif'
				}));
			}
		} catch (error) {
			console.error('Fetch Periods Error:', error);
		}
	};

	const fetchSubjects = async () => {
		try {
			const res = await fetch(`${API_BASE_URL}/admin/mapel/all`, {
				headers: { Authorization: `Bearer ${token}` }
			});
			if (res.ok) {
				const data = (await res.json()) || [];
				subjects = data.map((s: any) => ({
					id: s.id,
					name: s.nama_mapel,
					statusAktif: s.is_active === 1 ? 'Aktif' : 'Nonaktif'
				}));
			}
		} catch (error) {
			console.error('Fetch Subjects Error:', error);
		}
	};

	const toggleMapelStatus = async (id: number, currentStatus: string) => {
		const newStatus = currentStatus === 'Aktif' ? 0 : 1;
		const confirmMsg =
			newStatus === 1
				? 'Yakin ingin MENGAKTIFKAN mapel ini?'
				: 'Yakin ingin MENONAKTIFKAN mapel ini? (Mapel tidak akan muncul di opsi absensi)';

		if (!confirm(confirmMsg)) return;

		try {
			const res = await fetch(`${API_BASE_URL}/admin/mapel/status`, {
				method: 'PUT',
				headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
				body: JSON.stringify({ id: id, is_active: newStatus })
			});

			if (res.ok) {
				alert('Status mapel berhasil diubah!');
				fetchSubjects();
			} else {
				alert(`Gagal: ${await res.text()}`);
			}
		} catch (error) {
			alert('Kesalahan jaringan saat mengubah status mapel.');
		}
	};

	const fetchPendingStudentRegistrations = async () => {
		try {
			const res = await fetch(`${API_BASE_URL}/admin/siswa/pending-registrations`, {
				headers: { Authorization: `Bearer ${token}` }
			});
			if (res.ok) {
				pendingStudentRegs = (await res.json()) || [];
			}
		} catch (e) {
			console.error('Fetch pending registrations:', e);
		}
	};

	const approveStudentRegistration = async (userId: number) => {
		if (!confirm('Setujui pendaftar ini? Akun aktif dan perangkat yang didaftarkan langsung boleh login.')) return;
		try {
			const res = await fetch(`${API_BASE_URL}/admin/siswa/approve-registration`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
				body: JSON.stringify({ user_id: userId })
			});
			if (res.ok) {
				alert('Pendaftar disetujui.');
				fetchPendingStudentRegistrations();
				fetchDashboardData();
				fetchUsersAll();
			} else {
				let msg = await res.text();
				try {
					msg = JSON.parse(msg).error || msg;
				} catch {
					/* plain text */
				}
				alert(`Gagal: ${msg}`);
			}
		} catch {
			alert('Kesalahan jaringan.');
		}
	};

	const rejectStudentRegistration = async (userId: number) => {
		if (
			!confirm(
				'Tolak dan hapus permanen data pendaftar ini? Siswa bisa mendaftar ulang dengan username lain.'
			)
		)
			return;
		try {
			const res = await fetch(`${API_BASE_URL}/admin/siswa/reject-registration`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
				body: JSON.stringify({ user_id: userId })
			});
			if (res.ok) {
				alert('Pendaftaran ditolak.');
				fetchPendingStudentRegistrations();
				fetchDashboardData();
			} else {
				let msg = await res.text();
				try {
					msg = JSON.parse(msg).error || msg;
				} catch {
					/* */
				}
				alert(`Gagal: ${msg}`);
			}
		} catch {
			alert('Kesalahan jaringan.');
		}
	};

	const fetchPendingDevices = async () => {
		try {
			const res = await fetch(`${API_BASE_URL}/admin/device/pending`, {
				headers: { Authorization: `Bearer ${token}` }
			});
			if (res.ok) {
				const data = (await res.json()) || [];
				pendingDevices = data.map((d: any) => ({
					id: d.id,
					name: d.nama_siswa,
					namaSekolah: '-',
					device: d.user_agent.length > 40 ? d.user_agent.substring(0, 40) + '...' : d.user_agent,
					ip: 'Token: ' + (d.device_cookie_token.substring(0, 8) || 'N/A'),
					time: new Date(d.created_at).toLocaleString('id-ID'),
					status: 'Menunggu'
				}));
			}
		} catch (error) {
			console.error('Fetch Pending Devices Error:', error);
		}
	};

	// --- FUNGSI BUKA MODAL ---
	const openAddModal = () => {
		isEditing = false;
		currentEditId = null;
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
		showAddModal = true;
	};

	const openEditModal = (item: any) => {
		isEditing = true;
		currentEditId = item.id;
		if (activeMenu === 'guru' || activeMenu === 'siswa') {
			newUser = {
				username: item.username || '',
				password: '',
				identifier: item.namaSekolah || item.nip || '',
				namaLengkap: item.name || '',
				labelKataKunci: '',
				kataKunci: '',
				email: item.email || ''
			};
		} else if (activeMenu === 'kelas') {
			newClass = { nama_kelas: item.name || '', periode_id: item.periode_id || '' };
		} else if (activeMenu === 'mapel') {
			newSubject = { name: item.name || '' };
		} else if (activeMenu === 'admin_users') {
			newUser = {
				username: item.username || '',
				password: '', // Dikosongkan agar jika tidak diisi, backend tidak update password
				identifier: '',
				namaLengkap: '',
				labelKataKunci: '',
				kataKunci: '',
				email: ''
			};
		}
		showAddModal = true;
	};

	const openAssignModal = async (student: any) => {
		assignData = {
			siswaId: student.id,
			namaSiswa: student.name,
			action: 'assign',
			oldKelasId: '',
			kelasId: ''
		};
		studentCurrentClasses = [];
		showAssignModal = true;

		try {
			const res = await fetch(`${API_BASE_URL}/admin/siswa-kelas/list?siswa_id=${student.id}`, {
				headers: { Authorization: `Bearer ${token}` }
			});
			if (res.ok) {
				studentCurrentClasses = (await res.json()) || [];
			}
		} catch (error) {
			console.error('Gagal load kelas siswa:', error);
		}
	};

	const openAssignGuruModal = (guru: any) => {
		assignGuruData = {
			guruUserId: guru.id,
			namaGuru: guru.name,
			kelasId: ''
		};
		showAssignGuruModal = true;
	};

	const handleAssignGuru = async (e: Event) => {
		e.preventDefault();
		if (!assignGuruData.kelasId) return;

		try {
			const res = await fetch(`${API_BASE_URL}/admin/kelas/set-wali`, {
				method: 'PUT',
				headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
				body: JSON.stringify({
					kelas_id: Number(assignGuruData.kelasId),
					guru_user_id: assignGuruData.guruUserId
				})
			});

			if (res.ok) {
				alert('Berhasil menugaskan wali kelas!');
				assignGuruData.kelasId = '';
				fetchClasses();
			} else {
				alert(`Gagal: ${await res.text()}`);
			}
		} catch (error) {
			alert('Terjadi kesalahan jaringan.');
		}
	};

	const removeWaliKelas = async (kelasId: number) => {
		if (!confirm('Cabut guru ini dari wali kelas?')) return;
		try {
			const res = await fetch(`${API_BASE_URL}/admin/kelas/remove-wali`, {
				method: 'PUT',
				headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
				body: JSON.stringify({ kelas_id: kelasId })
			});

			if (res.ok) {
				fetchClasses();
			} else {
				alert(`Gagal: ${await res.text()}`);
			}
		} catch (error) {
			alert('Terjadi kesalahan jaringan.');
		}
	};

	const openClassStudentsModal = async (cls: any) => {
		selectedClassForStudents = { id: cls.id, name: cls.name };
		showClassStudentsModal = true;
		classStudentsList = [];

		await fetchClassStudents(cls.id);
	};

	const fetchClassStudents = async (kelasId: string | number) => {
		try {
			const res = await fetch(`${API_BASE_URL}/admin/kelas/siswa?kelas_id=${kelasId}`, {
				headers: { Authorization: `Bearer ${token}` }
			});
			if (res.ok) {
				classStudentsList = (await res.json()) || [];
			}
		} catch (error) {
			console.error('Fetch Class Students Error:', error);
		}
	};

	const removeStudentFromClassModal = async (userId: number, studentName: string) => {
		if (!confirm(`Keluarkan ${studentName} dari kelas ini?`)) return;

		try {
			const res = await fetch(
				`${API_BASE_URL}/admin/siswa-kelas/remove?siswa_id=${userId}&kelas_id=${selectedClassForStudents.id}`,
				{
					method: 'DELETE',
					headers: { Authorization: `Bearer ${token}` }
				}
			);

			if (res.ok) {
				fetchClassStudents(selectedClassForStudents.id);
			} else {
				alert(`Gagal: ${await res.text()}`);
			}
		} catch (error) {
			alert('Terjadi kesalahan jaringan.');
		}
	};

	// --- FUNGSI UNTUK MEMASUKKAN SISWA KE KELAS (DARI DALAM MODAL) ---
	const addStudentToClass = async () => {
		if (!selectedStudentToAdd) {
			alert('Pilih siswa yang ingin dimasukkan terlebih dahulu!');
			return;
		}

		try {
			const res = await fetch(`${API_BASE_URL}/admin/siswa-kelas/assign`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
				body: JSON.stringify({
					siswa_id: Number(selectedStudentToAdd),
					kelas_id: Number(selectedClassForStudents.id)
				})
			});

			if (res.ok) {
				// Jika sukses, kosongkan pilihan dan refresh daftar siswa di dalam ruangan
				selectedStudentToAdd = '';
				await fetchClassStudents(selectedClassForStudents.id);
			} else {
				const errText = await res.text();
				alert(`Gagal memasukkan siswa: ${errText}`);
			}
		} catch (error) {
			alert('Terjadi kesalahan jaringan saat memproses data.');
		}
	};

	// --- (OPSIONAL TAPI KEREN) FILTER SISWA ---
	// Menyaring daftar 'students' agar yang muncul di dropdown hanya siswa
	// yang belum ada di dalam kelas ini.
	let availableStudents = $derived(
		students.filter((s) => !classStudentsList.some((cs) => cs.user_id === s.id))
	);

	// --- FUNGSI MANAJEMEN DATABASE ---
	const handleBackup = async () => {
		try {
			const res = await fetch(`${API_BASE_URL}/admin/backup`, {
				headers: { Authorization: `Bearer ${token}` }
			});
			if (res.ok) alert(await res.text());
			else alert(`Gagal: ${await res.text()}`);
		} catch (err) {
			alert('Gagal menjalankan pencadangan.');
		}
	};

	const handleRestore = async () => {
		const fileId = prompt(
			'Masukkan ID file Google Drive untuk pemulihan:\n(PERINGATAN: data saat ini akan diganti seluruhnya!)'
		);
		if (!fileId) return;
		try {
			const res = await fetch(`${API_BASE_URL}/admin/restore?file_id=${fileId}`, {
				method: 'POST',
				headers: { Authorization: `Bearer ${token}` }
			});
			if (res.ok) {
				alert(await res.text());
				fetchDashboardData();
			} else {
				alert(`Gagal: ${await res.text()}`);
			}
		} catch (err) {
			alert('Gagal memulihkan database.');
		}
	};

	// --- FUNGSI SUBMIT (TAMBAH/EDIT DATA) ---
	const handleAddEntity = async (e: Event) => {
		e.preventDefault();
		const method = isEditing ? 'PUT' : 'POST';
		let endpoint = '';
		let payload: any = {};

		if (activeMenu === 'guru') {
			endpoint = isEditing ? `/admin/guru/update?id=${currentEditId}` : `/admin/guru/create`;
			payload = {
				username: newUser.username,
				password: newUser.password,
				nama_lengkap: newUser.namaLengkap,
				identifier: newUser.identifier,
				nip: newUser.identifier,
				email: newUser.email
			};
		} else if (activeMenu === 'siswa') {
			endpoint = isEditing ? `/admin/siswa/update?id=${currentEditId}` : `/admin/siswa/create`;
			payload = {
				username: newUser.username,
				password: newUser.password,
				nama_lengkap: newUser.namaLengkap,
				nama_sekolah: newUser.identifier,
				label_kata_kunci: newUser.labelKataKunci,
				kata_kunci: newUser.kataKunci
			};
		} else if (activeMenu === 'kelas') {
			endpoint = isEditing ? `/admin/kelas/update` : `/admin/kelas/create`;
			payload = {
				id: currentEditId,
				nama_kelas: newClass.nama_kelas,
				periode_id: Number(newClass.periode_id)
			};
		} else if (activeMenu === 'mapel') {
			endpoint = isEditing ? `/admin/mapel/update` : `/admin/mapel/create`;
			payload = { id: currentEditId, nama_mapel: newSubject.name };
		} else if (activeMenu === 'admin_users') {
			endpoint = isEditing
				? `/superadmin/admin/update?id=${currentEditId}`
				: `/superadmin/admin/create`;
			payload = {
				username: newUser.username,
				password: newUser.password // Akan diproses backend (diabaikan jika kosong saat update)
			};
		}

		try {
			const res = await fetch(`${API_BASE_URL}${endpoint}`, {
				method: method,
				headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
				body: JSON.stringify(payload)
			});

			if (res.ok) {
				alert(`Data berhasil ${isEditing ? 'diperbarui' : 'ditambahkan'}!`);
				showAddModal = false;
				if (activeMenu === 'guru' || activeMenu === 'siswa') fetchUsersAll();
				else if (activeMenu === 'kelas') fetchClasses();
				else if (activeMenu === 'mapel') fetchSubjects();
				else if (activeMenu === 'admin_users') fetchAdminList();
			} else {
				alert(`Gagal menyimpan data: ${await res.text()}`);
			}
		} catch (error) {
			alert('Terjadi kesalahan jaringan saat menyimpan data.');
		}
	};

	const handleAddPeriod = async (e: Event) => {
		e.preventDefault();
		try {
			const res = await fetch(`${API_BASE_URL}/admin/periode/create`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
				body: JSON.stringify({
					tahun_ajaran: newPeriod.tahun_ajar,
					semester: newPeriod.semester,
					status_aktif: Number(newPeriod.status_aktif)
				})
			});
			if (res.ok) {
				alert('Periode berhasil ditambahkan!');
				showPeriodModal = false;
				newPeriod = { tahun_ajar: '', semester: 'Ganjil', status_aktif: 1 };
				fetchPeriods();
			} else {
				const errText = await res.text();
				alert(`Gagal menambahkan periode: ${errText}`);
			}
		} catch (error) {
			alert('Kesalahan jaringan.');
		}
	};

	const handleAssignSubmit = async (e: Event) => {
		e.preventDefault();
		let endpoint = '';
		let method = '';
		let payload: any = null;

		if (assignData.action === 'assign') {
			endpoint = '/admin/siswa-kelas/assign';
			method = 'POST';
			payload = { siswa_id: assignData.siswaId, kelas_id: Number(assignData.kelasId) };
		} else if (assignData.action === 'update') {
			endpoint = '/admin/siswa-kelas/update';
			method = 'POST';
			payload = {
				siswa_id: assignData.siswaId,
				old_kelas_id: Number(assignData.oldKelasId),
				new_kelas_id: Number(assignData.kelasId)
			};
		} else if (assignData.action === 'remove') {
			endpoint = `/admin/siswa-kelas/remove?siswa_id=${assignData.siswaId}&kelas_id=${assignData.kelasId}`;
			method = 'DELETE';
		}

		try {
			const res = await fetch(`${API_BASE_URL}${endpoint}`, {
				method: method,
				headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
				body: payload ? JSON.stringify(payload) : null
			});
			if (res.ok) {
				alert('Status plotting siswa berhasil diperbarui!');
				showAssignModal = false;
			} else {
				alert(`Gagal memproses: ${await res.text()}`);
			}
		} catch (error) {
			alert('Terjadi kesalahan jaringan saat memproses plotting kelas.');
		}
	};

	// --- FUNGSI HAPUS DATA (DELETE) ---
	const deleteUser = async (id: number, role: string) => {
		if (!confirm(`Yakin ingin menonaktifkan akun ini?`)) return;
		try {
			const res = await fetch(`${API_BASE_URL}/admin/${role}/delete?id=${id}`, {
				method: 'DELETE',
				headers: { Authorization: `Bearer ${token}` }
			});
			if (res.ok) {
				alert(`Data berhasil dinonaktifkan!`);
				fetchUsersAll();
			} else {
				alert(`Gagal menonaktifkan data: ${await res.text()}`);
			}
		} catch (error) {
			alert('Terjadi kesalahan jaringan.');
		}
	};

	const deletePeriod = async (id: number) => {
		if (
			!confirm(
				'Yakin ingin menghapus periode ini? Periode yang terhubung dengan kelas tidak bisa dihapus.'
			)
		)
			return;
		try {
			const res = await fetch(`${API_BASE_URL}/admin/periode/delete?id=${id}`, {
				method: 'DELETE',
				headers: { Authorization: `Bearer ${token}` }
			});
			if (res.ok) {
				alert('Periode berhasil dihapus!');
				fetchPeriods();
			} else {
				alert(`Gagal: ${await res.text()}`);
			}
		} catch (error) {
			alert('Kesalahan jaringan.');
		}
	};

	const togglePeriodStatus = async (id: number, currentStatus: string) => {
		const newStatus = currentStatus === 'Aktif' ? 0 : 1;
		const confirmMsg =
			newStatus === 1
				? 'Yakin ingin MENGAKTIFKAN periode ini?'
				: 'Yakin ingin MENONAKTIFKAN periode ini?';

		if (!confirm(confirmMsg)) return;

		try {
			const res = await fetch(`${API_BASE_URL}/admin/periode/update`, {
				method: 'PUT',
				headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
				body: JSON.stringify({ id: id, status_aktif: newStatus })
			});

			if (res.ok) {
				alert('Status periode berhasil diubah!');
				fetchPeriods();
			} else {
				alert(`Gagal: ${await res.text()}`);
			}
		} catch (error) {
			alert('Kesalahan jaringan saat mengubah status.');
		}
	};

	const deleteClass = async (id: number) => {
		if (!confirm('Yakin ingin menghapus kelas ini?')) return;
		try {
			const res = await fetch(`${API_BASE_URL}/admin/kelas/delete?id=${id}`, {
				method: 'DELETE',
				headers: { Authorization: `Bearer ${token}` }
			});
			if (res.ok) {
				alert('Kelas berhasil dihapus!');
				fetchClasses();
			} else {
				alert(`Gagal: ${await res.text()}`);
			}
		} catch (error) {
			alert('Kesalahan jaringan.');
		}
	};

	const deleteSubject = async (id: number) => {
		if (!confirm('Yakin ingin menghapus mata pelajaran ini?')) return;
		try {
			const res = await fetch(`${API_BASE_URL}/admin/mapel/delete?id=${id}`, {
				method: 'DELETE',
				headers: { Authorization: `Bearer ${token}` }
			});
			if (res.ok) {
				alert('Mapel berhasil dihapus!');
				fetchSubjects();
			} else {
				alert(`Gagal: ${await res.text()}`);
			}
		} catch (error) {
			alert('Kesalahan jaringan.');
		}
	};

	const openResetModal = async (student: any) => {
		resetData = {
			username: student.username,
			labelKataKunci: 'Memuat Pertanyaan...',
			kataKunci: '',
			newPassword: '',
			confirmPassword: ''
		};
		showResetModal = true;
		try {
			const res = await fetch(`${API_BASE_URL}/admin/siswa/clue?username=${student.username}`, {
				headers: { Authorization: `Bearer ${token}` }
			});
			if (res.ok) {
				const data = await res.json();
				resetData.labelKataKunci = data.label_kata_kunci;
			} else {
				resetData.labelKataKunci = 'Gagal memuat pertanyaan keamanan (Belum Diatur)';
			}
		} catch (error) {
			resetData.labelKataKunci = 'Terjadi kesalahan saat memuat';
		}
	};

	const handleResetPassword = async (e: Event) => {
		e.preventDefault();
		if (resetData.newPassword !== resetData.confirmPassword) {
			alert('Kata sandi baru dan konfirmasi tidak cocok!');
			return;
		}
		try {
			const res = await fetch(`${API_BASE_URL}/admin/siswa/reset-password`, {
				method: 'PUT',
				headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
				body: JSON.stringify({
					username: resetData.username,
					kata_kunci: resetData.kataKunci,
					new_password: resetData.newPassword
				})
			});
			if (res.ok) {
				alert('Berhasil: kata sandi siswa telah diatur ulang!');
				showResetModal = false;
			} else {
				const errData = await res.json();
				alert(`Gagal: ${errData.error || 'Terjadi kesalahan'}`);
			}
		} catch (error) {
			alert('Terjadi kesalahan jaringan saat mereset password.');
		}
	};

	const approveDevice = async (id: number) => {
		if (!confirm('Yakin ingin MENYETUJUI akses login perangkat ini?')) return;
		try {
			const res = await fetch(`${API_BASE_URL}/admin/device/approve`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
				body: JSON.stringify({ device_id: id })
			});
			if (res.ok) {
				alert('Perangkat disetujui.');
				fetchPendingDevices();
				fetchDashboardData();
			} else {
				alert(`Gagal: ${await res.text()}`);
			}
		} catch (error) {
			alert('Terjadi kesalahan jaringan.');
		}
	};

	const rejectDevice = async (id: number) => {
		if (!confirm('Yakin ingin MENOLAK perangkat ini secara permanen?')) return;
		try {
			const res = await fetch(`${API_BASE_URL}/admin/device/reject`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
				body: JSON.stringify({ device_id: id })
			});
			if (res.ok) {
				alert('Berhasil ditolak!');
				fetchPendingDevices();
				fetchDashboardData();
			} else {
				alert(`Gagal: ${await res.text()}`);
			}
		} catch (error) {
			alert('Terjadi kesalahan jaringan.');
		}
	};

	// --- FUNGSI SUPER ADMIN ---
	const fetchAdminList = async () => {
		try {
			const res = await fetch(`${API_BASE_URL}/superadmin/admin/all`, {
				headers: { Authorization: `Bearer ${token}` }
			});
			if (res.ok) {
				const data = await res.json();
				adminList = data.map((a: any) => ({
					id: a.id,
					username: a.username,
					status: a.is_active === 1 ? 'Aktif' : 'Nonaktif'
				}));
			}
		} catch (error) {
			console.error('Gagal mengambil daftar admin');
		}
	};

	const toggleAdminStatus = async (id: number) => {
		if (!confirm('Yakin ingin mengubah status Admin ini?')) return;
		try {
			const res = await fetch(`${API_BASE_URL}/superadmin/admin/toggle?id=${id}`, {
				method: 'PUT',
				headers: { Authorization: `Bearer ${token}` }
			});
			if (res.ok) fetchAdminList();
			else alert(`Gagal: ${await res.text()}`);
		} catch (error) {
			alert('Kesalahan Jaringan');
		}
	};

	const reactivateUser = async (id: number) => {
		if (!confirm('Aktifkan kembali akun ini agar bisa login?')) return;
		try {
			const res = await fetch(`${API_BASE_URL}/superadmin/users/reactivate?id=${id}`, {
				method: 'PUT',
				headers: { Authorization: `Bearer ${token}` }
			});
			if (res.ok) {
				alert('Akun berhasil diaktifkan kembali!');
				fetchUsersAll();
			} else alert(`Gagal: ${await res.text()}`);
		} catch (error) {
			alert('Kesalahan Jaringan');
		}
	};

	const hardDeleteUser = async (id: number) => {
		if (
			!confirm(
				'PERINGATAN KERAS!\n\nHapus permanen user ini beserta seluruh histori absensi dan perangkatnya?\nAksi ini tidak bisa dibatalkan.'
			)
		)
			return;
		try {
			const res = await fetch(`${API_BASE_URL}/superadmin/users/hard-delete?id=${id}`, {
				method: 'DELETE',
				headers: { Authorization: `Bearer ${token}` }
			});

			if (res.ok) {
				alert('Data berhasil dimusnahkan secara permanen!');

				// PERBAIKAN: Refresh data sesuai tab yang sedang dibuka!
				if (activeMenu === 'admin_users') {
					fetchAdminList();
				} else {
					fetchUsersAll();
				}
			} else {
				alert(`Gagal: ${await res.text()}`);
			}
		} catch (error) {
			alert('Kesalahan Jaringan');
		}
	};

	const logout = () => {
		localStorage.removeItem('jwt_token');
		localStorage.removeItem('user_role');
		goto('/login/admin');
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
					Authorization: `Bearer ${token}` // Jangan tambahkan 'Content-Type' di sini
				},
				body: formData
			});

			if (res.ok) {
				// CEK HEADER DULU, JANGAN BACA BODY DULU
				const contentType = res.headers.get('content-type');

				if (contentType && contentType.includes('application/json')) {
					// KONDISI 1: Backend kirim JSON, baca sebagai JSON
					const data = await res.json();
					alert(`${data.message}\nTotal Duplikat di-skip: ${data.duplikat}`);
				} else {
					// KONDISI 2: Backend kirim FILE EXCEL, baca sebagai Blob
					const blob = await res.blob(); // Buat link download otomatis

					const url = window.URL.createObjectURL(blob);
					const a = document.createElement('a');
					a.href = url;
					a.download = 'Hasil_Akun_Siswa.xlsx';
					document.body.appendChild(a);
					a.click();
					document.body.removeChild(a);
					window.URL.revokeObjectURL(url); // Ambil rekap data dari Header buatan Golang

					const sukses = res.headers.get('X-Import-Success');
					const duplikat = res.headers.get('X-Import-Duplicate');

					alert(
						`Impor selesai!\nSiswa baru: ${sukses}\nDuplikat: ${duplikat}\n\nFile akun otomatis diunduh.`
					); // Refresh tabel

					fetchUsersAll();
				} // Reset input form

				selectedFile = null;
				const fileInput = document.getElementById('file-upload') as HTMLInputElement;
				if (fileInput) fileInput.value = '';
			} else {
				// Kalau tidak OK, baca error dari backend sebagai text
				const responseText = await res.text();
				alert(`Gagal mengunggah: ${responseText}`);
			}
		} catch (error) {
			console.error('Error saat upload:', error);
			alert(`Terjadi kesalahan jaringan saat mengunggah file. Silakan cek console.`);
		} finally {
			isUploading = false;
		}
	};

	function downloadTemplate() {
		// Karena file ada di folder static, path-nya langsung dari root '/'
		const fileUrl = '/template_import_siswa.xlsx';

		// Membuat elemen <a> sementara untuk memicu download
		const link = document.createElement('a');
		link.href = fileUrl;
		link.setAttribute('download', 'Template_Data_Siswa.xlsx'); // Nama file saat diunduh
		document.body.appendChild(link);
		link.click();
		document.body.removeChild(link);
	}
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
				<p class="text-[10px] font-bold tracking-widest text-slate-400 uppercase">Pengelola</p>
			</div>
		</div>
		<div class="flex items-center gap-4 md:gap-6">
			<div class="hidden border-r border-slate-100 pr-6 text-right md:block">
				<p class="text-sm font-black tracking-tight text-slate-800 uppercase">
					{adminProfile.name}
				</p>
				<p class="mt-0.5 text-[10px] font-black tracking-widest text-slate-500 uppercase">
					{adminProfile.role}
				</p>
			</div>
			<button
				class="rounded-xl bg-red-50 px-4 py-2 text-[10px] font-black tracking-widest text-red-600 uppercase transition-colors hover:bg-red-100 md:px-5 md:py-2.5"
				onclick={logout}>Keluar</button
			>
		</div>
	</header>

	<main class="mx-auto flex w-full max-w-7xl flex-1 flex-col gap-8 p-4 md:p-8 lg:flex-row">
		<aside class="w-full shrink-0 lg:w-72">
			<div class="sticky top-28 rounded-[2rem] border border-slate-100 bg-white p-6 shadow-sm">
				<h3 class="mb-6 px-2 text-xs font-black tracking-widest text-slate-800 uppercase">
					Menu pengelola
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
				<div
					class="flex flex-col justify-between gap-4 rounded-[2.5rem] border border-slate-100 bg-white p-6 shadow-sm md:flex-row md:items-center md:p-8"
				>
					<div>
						<h2 class="text-2xl font-black tracking-tight text-slate-900 md:text-3xl">
							Halo, Admin 👋
						</h2>
						<p class="mt-2 text-sm leading-relaxed font-medium text-slate-500 md:text-base">
							Ringkasan statistik sistem absensi hari ini.
						</p>
					</div>

					<div class="flex items-center gap-3">
						<button
							onclick={handleBackup}
							class="rounded-xl border border-slate-200 bg-slate-50 px-5 py-2.5 text-[10px] font-black tracking-widest text-slate-600 uppercase transition-all hover:bg-slate-100 hover:text-slate-800"
						>
							Cadangkan data
						</button>
						<button
							onclick={handleRestore}
							class="rounded-xl bg-red-50 px-5 py-2.5 text-[10px] font-black tracking-widest text-red-600 uppercase transition-all hover:bg-red-100"
						>
							Pulihkan data
						</button>
					</div>
				</div>
				<div class="grid grid-cols-1 gap-4 sm:grid-cols-2 md:gap-6 xl:grid-cols-3 2xl:grid-cols-5">
					<div class="rounded-[2rem] border border-slate-100 bg-white p-5 shadow-sm xl:p-6">
						<p class="mb-2 text-[10px] font-black tracking-[0.2em] text-slate-400 uppercase">
							Total pengguna
						</p>
						<h2 class="text-3xl font-black text-slate-900 xl:text-4xl">{stats.totalUsers}</h2>
					</div>
					<div class="rounded-[2rem] border border-slate-100 bg-white p-5 shadow-sm xl:p-6">
						<p class="mb-2 text-[10px] font-black tracking-[0.2em] text-slate-400 uppercase">
							Pengguna aktif
						</p>
						<h2 class="text-3xl font-black text-green-600 xl:text-4xl">{stats.activeUsers}</h2>
					</div>
					<div class="rounded-[2rem] border border-slate-100 bg-white p-5 shadow-sm xl:p-6">
						<p class="mb-2 text-[10px] font-black tracking-[0.2em] text-slate-400 uppercase">
							Tidak aktif
						</p>
						<h2 class="text-3xl font-black text-slate-400 xl:text-4xl">{stats.inactiveUsers}</h2>
					</div>
					<div
						class="relative flex flex-col justify-between overflow-hidden rounded-[2rem] border border-transparent bg-teal-600 p-5 text-white shadow-lg shadow-teal-500/20 xl:p-6"
					>
						<div class="relative z-10">
							<p class="mb-2 text-[10px] font-black tracking-[0.2em] text-teal-200 uppercase">
								Pendaftar Siswa
							</p>
							<h2 class="text-3xl font-black xl:text-4xl">{stats.pendingSiswaRegs}</h2>
						</div>
					</div>
					<div
						class="relative flex flex-col justify-between overflow-hidden rounded-[2rem] border border-transparent bg-brand-blue p-5 text-white shadow-lg shadow-blue-500/20 xl:p-6"
					>
						<div class="relative z-10">
							<p class="mb-2 text-[10px] font-black tracking-[0.2em] text-blue-200 uppercase">
								Antre perangkat
							</p>
							<h2 class="text-3xl font-black xl:text-4xl">{stats.pendingApproval}</h2>
						</div>
					</div>
				</div>
				<div class="overflow-hidden rounded-[2.5rem] border border-slate-100 bg-white shadow-sm">
					<div class="border-b border-slate-50 bg-white p-6 md:p-8">
						<h3 class="text-xs font-black tracking-widest text-slate-800 uppercase">
							Masuk terbaru
						</h3>
					</div>
					<div class="w-full overflow-x-auto">
						<table class="w-full min-w-[700px] border-collapse text-left">
							<thead>
								<tr class="bg-slate-50/50">
									<th
										class="w-32 px-6 py-4 text-[10px] font-black tracking-[0.2em] text-slate-400 uppercase md:px-8 md:py-5"
										>Waktu masuk</th
									>
									<th
										class="px-6 py-4 text-[10px] font-black tracking-[0.2em] text-slate-400 uppercase md:px-8 md:py-5"
										>Nama pengguna</th
									>
									<th
										class="px-6 py-4 text-[10px] font-black tracking-[0.2em] text-slate-400 uppercase md:px-8 md:py-5"
										>Peran</th
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
													: 'border-amber-200 bg-amber-100 text-amber-700'}">{activity.status === 'Online'
													? 'Aktif'
													: activity.status}</span
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
					<div class="flex flex-col gap-3 sm:flex-row sm:items-center">
						<!-- Tombol Tambah Baru -->
						<button
							onclick={openAddModal}
							class="w-full rounded-2xl bg-brand-blue px-6 py-3.5 text-[10px] font-black tracking-widest text-white uppercase shadow-lg shadow-blue-500/20 transition-all hover:scale-105 sm:w-auto"
						>
							+ Tambah Baru
						</button>

						<!-- Tombol Import Data -->
						{#if activeMenu === 'siswa'}
							<button
								onclick={() => (showImportModal = true)}
								class="flex w-full items-center justify-center gap-2 rounded-2xl border-2 border-slate-200 bg-white px-6 py-3 text-[10px] font-black tracking-widest text-slate-600 uppercase transition-all hover:scale-105 hover:border-brand-blue hover:text-brand-blue sm:w-auto"
							>
								<svg
									xmlns="http://www.w3.org/2000/svg"
									class="h-4 w-4"
									viewBox="0 0 24 24"
									fill="none"
									stroke="currentColor"
									stroke-width="2"
									stroke-linecap="round"
									stroke-linejoin="round"
								>
									<path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
									<polyline points="17 8 12 3 7 8" />
									<line x1="12" x2="12" y1="3" y2="15" />
								</svg>
								Impor data
							</button>
						{/if}
					</div>
				</div>
				<div class="overflow-hidden rounded-[2.5rem] border border-slate-100 bg-white shadow-sm">
					<div class="w-full overflow-x-auto">
						<table class="w-full min-w-[800px] border-collapse text-left">
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
										>{activeMenu === 'guru' ? 'NIP' : 'Asal Sekolah'}</th
									>
									<th
										class="px-6 py-4 text-[10px] font-black tracking-[0.2em] text-slate-400 uppercase md:px-8 md:py-5"
										>{activeMenu === 'guru' ? 'Kelas (Wali)' : 'Kelas'}</th
									>
									<th
										class="px-6 py-4 text-[10px] font-black tracking-[0.2em] text-slate-400 uppercase md:px-8 md:py-5"
										>Status</th
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
											>{user.nip || user.namaSekolah}</td
										>
										<td class="px-6 py-4 font-medium md:px-8 md:py-5">
											{#if activeMenu === 'guru'}
												{@const guruClasses = classes.filter((c) => c.wali_guru_id === user.id)}
												{#if guruClasses.length > 0}
													<div class="flex flex-wrap gap-1">
														{#each guruClasses as cls}
															<span class="rounded-md bg-indigo-50 px-2.5 py-1 text-[10px] font-bold text-indigo-600">
																{cls.name}
															</span>
														{/each}
													</div>
												{:else}
													<span class="text-slate-400">-</span>
												{/if}
											{:else}
												<span class="text-slate-500">{user.class}</span>
											{/if}
										</td>
										<td class="px-6 py-4 font-medium md:px-8 md:py-5">
											<span
												class="rounded-full px-3 py-1 text-[9px] font-black tracking-widest uppercase {user.status ===
												'Aktif'
													? 'bg-green-100 text-green-700'
													: 'bg-slate-200 text-slate-500'}">{user.status || 'Aktif'}</span
											>
										</td>
										<td class="px-6 py-4 text-right md:px-8 md:py-5">
											<div class="flex justify-end gap-2">
												{#if activeMenu === 'siswa'}
													<button
														onclick={() => openAssignModal(user)}
														class="rounded-xl bg-purple-50 px-4 py-2 text-[10px] font-black tracking-widest text-purple-600 uppercase transition-all hover:bg-purple-100"
														>Atur Kelas</button
													>
													<button
														onclick={() => openResetModal(user)}
														class="rounded-xl bg-amber-50 px-4 py-2 text-[10px] font-black tracking-widest text-amber-600 uppercase transition-all hover:bg-amber-100"
														>Reset sandi</button
													>
												{:else if activeMenu === 'guru'}
													<button
														onclick={() => openAssignGuruModal(user)}
														class="rounded-xl bg-indigo-50 px-4 py-2 text-[10px] font-black tracking-widest text-indigo-600 uppercase transition-all hover:bg-indigo-100"
														>Atur Kelas (Wali)</button
													>
												{/if}
												<button
													onclick={() => openEditModal(user)}
													class="rounded-xl bg-slate-50 px-4 py-2 text-[10px] font-black tracking-widest text-brand-blue uppercase transition-all hover:bg-blue-50"
													>Ubah</button
												>
												{#if user.status !== 'Nonaktif'}
													<button
														onclick={() => deleteUser(user.id, activeMenu)}
														class="rounded-xl bg-slate-50 px-4 py-2 text-[10px] font-black tracking-widest text-red-500 uppercase transition-all hover:bg-red-50"
														>Nonaktifkan</button
													>
												{:else if adminProfile.role === 'Admin utama'}
													<button
														onclick={() => reactivateUser(user.id)}
														class="rounded-xl bg-green-50 px-4 py-2 text-[10px] font-black tracking-widest text-green-600 uppercase transition-all hover:bg-green-100"
														>Aktifkan</button
													>
													<button
														onclick={() => hardDeleteUser(user.id)}
														class="rounded-xl bg-red-100 px-4 py-2 text-[10px] font-black tracking-widest text-red-700 uppercase shadow-sm transition-all hover:bg-red-200"
														>Musnahkan</button
													>
												{/if}
											</div>
										</td>
									</tr>
								{/each}
							</tbody>
						</table>
					</div>
				</div>
			{:else if activeMenu === 'pendaftar_siswa'}
				<div
					class="flex flex-col justify-between gap-4 rounded-[2.5rem] border border-slate-100 bg-white p-6 shadow-sm sm:flex-row sm:items-center md:p-8"
				>
					<div>
						<h2 class="text-2xl font-black tracking-tight text-slate-900 uppercase">
							Pendaftar Siswa
						</h2>
						<p class="mt-1 text-sm font-medium text-slate-500">
							Siswa yang mendaftar mandiri menunggu persetujuan. Setelah disetujui, perangkat yang
							didaftarkan saat registrasi boleh langsung login.
						</p>
					</div>
				</div>
				<div class="overflow-hidden rounded-[2.5rem] border border-slate-100 bg-white shadow-sm">
					{#if pendingStudentRegs.length === 0}
						<p class="p-10 text-center text-sm font-medium text-slate-400">
							Tidak ada pendaftar yang menunggu.
						</p>
					{:else}
						<div class="w-full overflow-x-auto">
							<table class="w-full min-w-[800px] border-collapse text-left">
								<thead>
									<tr class="bg-slate-50/50">
										<th
											class="px-6 py-4 text-[10px] font-black tracking-[0.2em] text-slate-400 uppercase md:px-8 md:py-5"
											>Nama pengguna</th
										>
										<th
											class="px-6 py-4 text-[10px] font-black tracking-[0.2em] text-slate-400 uppercase md:px-8 md:py-5"
											>Nama</th
										>
										<th
											class="px-6 py-4 text-[10px] font-black tracking-[0.2em] text-slate-400 uppercase md:px-8 md:py-5"
											>Sekolah</th
										>
										<th
											class="px-6 py-4 text-[10px] font-black tracking-[0.2em] text-slate-400 uppercase md:px-8 md:py-5"
											>Petunjuk</th
										>
										<th
											class="px-6 py-4 text-[10px] font-black tracking-[0.2em] text-slate-400 uppercase md:px-8 md:py-5"
											>Waktu Daftar</th
										>
										<th
											class="px-6 py-4 text-right text-[10px] font-black tracking-[0.2em] text-slate-400 uppercase md:px-8 md:py-5"
											>Aksi</th
										>
									</tr>
								</thead>
								<tbody class="divide-y divide-slate-50">
									{#each pendingStudentRegs as row}
										<tr class="transition-colors hover:bg-slate-50/30">
											<td class="px-6 py-4 font-bold text-slate-800 md:px-8 md:py-5"
												>{row.username}</td
											>
											<td class="px-6 py-4 font-medium text-slate-600 md:px-8 md:py-5"
												>{row.nama_lengkap}</td
											>
											<td class="px-6 py-4 font-medium text-slate-600 md:px-8 md:py-5"
												>{row.nama_sekolah}</td
											>
											<td class="px-6 py-4 text-sm text-slate-500 md:px-8 md:py-5"
												>{row.label_kata_kunci}</td
											>
											<td class="px-6 py-4 text-xs font-medium text-slate-400 md:px-8 md:py-5"
												>{row.created_at || '—'}</td
											>
											<td class="px-6 py-4 text-right md:px-8 md:py-5">
												<div class="flex justify-end gap-2">
													<button
														onclick={() => approveStudentRegistration(row.user_id)}
														class="rounded-xl bg-brand-blue px-4 py-2 text-[10px] font-black tracking-widest text-white uppercase shadow-lg shadow-blue-500/20"
														>Setujui</button
													>
													<button
														onclick={() => rejectStudentRegistration(row.user_id)}
														class="rounded-xl border border-slate-200 bg-slate-50 px-4 py-2 text-[10px] font-black tracking-widest text-red-500 uppercase"
														>Tolak</button
													>
												</div>
											</td>
										</tr>
									{/each}
								</tbody>
							</table>
						</div>
					{/if}
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
											<p class="font-black tracking-tight text-slate-800">
												{period.tahunAjar}
												<span class="text-xs text-brand-blue">({period.semester})</span>
											</p>
											<span
												class="mt-1 inline-block rounded-full px-3 py-1 text-[8px] font-black tracking-widest uppercase {period.statusAktif ===
												'Aktif'
													? 'bg-green-100 text-green-700'
													: 'bg-slate-200 text-slate-500'}">{period.statusAktif}</span
											>
										</div>
										<button
											onclick={() => togglePeriodStatus(period.id, period.statusAktif)}
											class="p-2 transition-all {period.statusAktif === 'Aktif'
												? 'text-green-500 hover:text-slate-400'
												: 'text-slate-300 hover:text-green-500'}"
											title={period.statusAktif === 'Aktif'
												? 'Nonaktifkan Periode'
												: 'Aktifkan Periode'}
										>
											{#if period.statusAktif === 'Aktif'}
												<svg
													xmlns="http://www.w3.org/2000/svg"
													class="h-5 w-5"
													viewBox="0 0 24 24"
													fill="none"
													stroke="currentColor"
													stroke-width="2.5"
													stroke-linecap="round"
													stroke-linejoin="round"
													><rect width="20" height="12" x="2" y="6" rx="6" ry="6" /><circle
														cx="16"
														cy="12"
														r="2"
													/></svg
												>
											{:else}
												<svg
													xmlns="http://www.w3.org/2000/svg"
													class="h-5 w-5"
													viewBox="0 0 24 24"
													fill="none"
													stroke="currentColor"
													stroke-width="2.5"
													stroke-linecap="round"
													stroke-linejoin="round"
													><rect width="20" height="12" x="2" y="6" rx="6" ry="6" /><circle
														cx="8"
														cy="12"
														r="2"
													/></svg
												>
											{/if}
										</button>
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
								onclick={openAddModal}
								class="rounded-2xl bg-brand-blue px-6 py-3.5 text-[10px] font-black tracking-widest text-white uppercase shadow-lg shadow-blue-500/20 transition-all hover:scale-105"
								>+ Tambah Kelas</button
							>
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
												>Referensi periode</th
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
														{periods.find((p) => p.id === item.periode_id)?.tahunAjar || 'Tidak diketahui'}
														({periods.find((p) => p.id === item.periode_id)?.semester || '-'})
													</span>
												</td>
												<td class="px-6 py-5 text-right">
													<div class="flex justify-end gap-2">
														<button
															onclick={() => openClassStudentsModal(item)}
															class="rounded-xl bg-purple-50 px-4 py-2 text-[10px] font-black tracking-widest text-purple-600 uppercase transition-all hover:bg-purple-100"
															>Lihat Siswa</button
														>
														<button
															onclick={() => openEditModal(item)}
															class="rounded-xl bg-slate-50 px-4 py-2 text-[10px] font-black tracking-widest text-brand-blue uppercase transition-all hover:bg-blue-50"
															>Ubah</button
														>
														<button
															onclick={() => deleteClass(item.id)}
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
						onclick={openAddModal}
						class="w-full rounded-2xl bg-brand-blue px-6 py-3.5 text-[10px] font-black tracking-widest text-white uppercase shadow-lg shadow-blue-500/20 transition-all hover:scale-105 sm:w-auto"
						>+ Tambah Mapel</button
					>
				</div>
				<div class="overflow-hidden rounded-[2.5rem] border border-slate-100 bg-white shadow-sm">
					<div class="w-full overflow-x-auto">
						<table class="w-full min-w-[600px] border-collapse text-left">
							<thead>
								<tr class="bg-slate-50/50">
									<th
										class="w-16 px-6 py-4 text-[10px] font-black tracking-[0.2em] text-slate-400 uppercase md:px-8 md:py-5"
										>ID (otomatis)</th
									>
									<th
										class="px-6 py-4 text-[10px] font-black tracking-[0.2em] text-slate-400 uppercase md:px-8 md:py-5"
										>Nama Mata Pelajaran</th
									>
									<th
										class="px-6 py-4 text-[10px] font-black tracking-[0.2em] text-slate-400 uppercase md:px-8 md:py-5"
										>Status</th
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
										<td class="px-6 py-4 md:px-8 md:py-5">
											<span
												class="inline-block rounded-full px-3 py-1 text-[9px] font-black tracking-widest uppercase {item.statusAktif ===
												'Aktif'
													? 'bg-green-100 text-green-700'
													: 'bg-slate-200 text-slate-500'}"
											>
												{item.statusAktif}
											</span>
										</td>
										<td class="px-6 py-4 text-right md:px-8 md:py-5">
											<div class="flex justify-end gap-1">
												<button
													onclick={() => toggleMapelStatus(item.id, item.statusAktif)}
													class="p-2 transition-all {item.statusAktif === 'Aktif'
														? 'text-green-500 hover:text-slate-400'
														: 'text-slate-300 hover:text-green-500'}"
													title={item.statusAktif === 'Aktif'
														? 'Nonaktifkan Mapel'
														: 'Aktifkan Mapel'}
												>
													{#if item.statusAktif === 'Aktif'}
														<svg
															xmlns="http://www.w3.org/2000/svg"
															class="h-5 w-5"
															viewBox="0 0 24 24"
															fill="none"
															stroke="currentColor"
															stroke-width="2.5"
															stroke-linecap="round"
															stroke-linejoin="round"
															><rect width="20" height="12" x="2" y="6" rx="6" ry="6" /><circle
																cx="16"
																cy="12"
																r="2"
															/></svg
														>
													{:else}
														<svg
															xmlns="http://www.w3.org/2000/svg"
															class="h-5 w-5"
															viewBox="0 0 24 24"
															fill="none"
															stroke="currentColor"
															stroke-width="2.5"
															stroke-linecap="round"
															stroke-linejoin="round"
															><rect width="20" height="12" x="2" y="6" rx="6" ry="6" /><circle
																cx="8"
																cy="12"
																r="2"
															/></svg
														>
													{/if}
												</button>

												<button
													onclick={() => openEditModal(item)}
													class="p-2 text-slate-400 transition-colors hover:text-brand-blue"
													title="Ubah nama"
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
														><path d="M12 20h9" /><path
															d="M16.5 3.5a2.12 2.12 0 0 1 3 3L7 19l-4 1 1-4Z"
														/></svg
													>
												</button>

												<button
													onclick={() => deleteSubject(item.id)}
													class="p-2 text-slate-300 transition-colors hover:text-red-600"
													title="Hapus Permanen"
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
														><path d="M3 6h18" /><path
															d="M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6"
														/><path d="M8 6V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2" /></svg
													>
												</button>
											</div>
										</td>
									</tr>
								{/each}
							</tbody>
						</table>
					</div>
				</div>
			{:else if activeMenu === 'admin_users'}
				<div
					class="flex flex-col justify-between gap-4 rounded-[2.5rem] border border-slate-100 bg-white p-6 shadow-sm sm:flex-row sm:items-center md:p-8"
				>
					<div>
						<h2 class="text-2xl font-black tracking-tight text-slate-900 uppercase">
							Manajemen admin standar
						</h2>
						<p class="mt-1 text-sm font-medium text-slate-500">Hanya bisa diakses oleh admin utama.</p>
					</div>
					<button
						onclick={openAddModal}
						class="w-full rounded-2xl bg-brand-blue px-6 py-3.5 text-[10px] font-black tracking-widest text-white uppercase shadow-lg shadow-blue-500/20 transition-all hover:scale-105 sm:w-auto"
					>
						+ Tambah Admin
					</button>
				</div>
				<div class="overflow-hidden rounded-[2.5rem] border border-slate-100 bg-white shadow-sm">
					<div class="w-full overflow-x-auto">
						<table class="w-full min-w-[600px] border-collapse text-left">
							<thead class="bg-slate-50/50">
								<tr>
									<th
										class="px-6 py-5 text-[10px] font-black tracking-[0.2em] text-slate-400 uppercase"
										>Nama pengguna</th
									>
									<th
										class="px-6 py-5 text-[10px] font-black tracking-[0.2em] text-slate-400 uppercase"
										>Status</th
									>
									<th
										class="px-6 py-5 text-right text-[10px] font-black tracking-[0.2em] text-slate-400 uppercase"
										>Aksi</th
									>
								</tr>
							</thead>
							<tbody class="divide-y divide-slate-50">
								{#each adminList as admin}
									<tr class="transition-colors hover:bg-slate-50/30">
										<td class="px-6 py-5 font-black text-slate-800">{admin.username}</td>
										<td class="px-6 py-5">
											<span
												class="rounded-full px-3 py-1 text-[9px] font-black tracking-widest uppercase {admin.status ===
												'Aktif'
													? 'bg-green-100 text-green-700'
													: 'bg-slate-200 text-slate-500'}"
											>
												{admin.status}
											</span>
										</td>
										<td class="px-6 py-5 text-right">
											<button
												onclick={() => toggleAdminStatus(admin.id)}
												class="rounded-xl px-4 py-2 text-[10px] font-black tracking-widest uppercase transition-all {admin.status ===
												'Aktif'
													? 'bg-red-50 text-red-500 hover:bg-red-100'
													: 'bg-green-50 text-green-600 hover:bg-green-100'}"
											>
												{admin.status === 'Aktif' ? 'Nonaktifkan' : 'Aktifkan'}
											</button>
											<button
												onclick={() => openEditModal(admin)}
												class="ml-2 rounded-xl bg-slate-50 px-4 py-2 text-[10px] font-black tracking-widest text-brand-blue uppercase transition-all hover:bg-blue-50"
											>
												Ubah data
											</button>

											{#if admin.status === 'Nonaktif'}
												<button
													onclick={() => hardDeleteUser(admin.id)}
													class="ml-2 rounded-xl bg-red-100 px-4 py-2 text-[10px] font-black tracking-widest text-red-700 uppercase shadow-sm transition-all hover:bg-red-200"
												>
													Musnahkan
												</button>
											{/if}
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
							Kelola izin masuk dari perangkat baru milik siswa.
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
										>Perangkat / User Agent</th
									>
									<th
										class="px-6 py-4 text-[10px] font-black tracking-[0.2em] text-slate-400 uppercase md:px-8 md:py-5"
										>Waktu permintaan</th
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
										<td class="px-6 py-4 md:px-8 md:py-5"
											><p class="font-bold text-slate-800">{dev.name}</p></td
										>
										<td class="px-6 py-4 md:px-8 md:py-5"
											><p class="font-bold text-slate-700">{dev.device}</p></td
										>
										<td class="px-6 py-4 text-sm font-medium text-slate-500 md:px-8 md:py-5"
											>{dev.time}</td
										>
										<td class="px-6 py-4 text-center md:px-8 md:py-5">
											<span
												class="rounded-full border px-4 py-1.5 text-[9px] font-black tracking-widest whitespace-nowrap uppercase {dev.status ===
												'Disetujui'
													? 'border-green-200 bg-green-100 text-green-700'
													: dev.status === 'Ditolak'
														? 'border-red-200 bg-red-100 text-red-700'
														: 'border-amber-200 bg-amber-100 text-amber-700'}">{dev.status}</span
											>
										</td>
										<td class="px-6 py-4 text-right md:px-8 md:py-5">
											{#if dev.status === 'Menunggu'}
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
						{isEditing ? 'Ubah' : 'Tambah'}
						{activeMenu === 'guru'
							? 'Guru'
							: activeMenu === 'siswa'
								? 'Siswa'
								: activeMenu === 'kelas'
									? 'Kelas'
									: activeMenu === 'admin_users'
										? 'Admin standar'
										: 'Mata Pelajaran'}
					</h3>
					<p class="mt-2 text-[10px] font-bold tracking-widest text-slate-400 uppercase">
						{isEditing ? 'Perbarui data yang ada' : 'Isi sesuai petunjuk kolom'}
					</p>
				</div>
				<form onsubmit={handleAddEntity} class="space-y-5">
					{#if activeMenu === 'guru' || activeMenu === 'siswa'}
						<div
							class="grid grid-cols-1 gap-5 {isEditing && activeMenu === 'siswa'
								? ''
								: 'sm:grid-cols-2'}"
						>
							<div class="space-y-5">
								<div class="flex flex-col gap-2">
									<label
										class="ml-1 text-[10px] font-black tracking-widest text-slate-400 uppercase"
										>Nama Lengkap</label
									>
									<input
										type="text"
										bind:value={newUser.namaLengkap}
										class="rounded-2xl border-2 border-transparent bg-slate-50 px-5 py-4 font-bold text-slate-700 outline-none focus:border-brand-blue/20"
										required
									/>
								</div>
								<div class="flex flex-col gap-2">
									<label
										class="ml-1 text-[10px] font-black tracking-widest text-slate-400 uppercase"
										>Nama pengguna</label
									>
									<input
										type="text"
										bind:value={newUser.username}
										class="rounded-2xl border-2 border-transparent bg-slate-50 px-5 py-4 font-bold text-slate-700 outline-none focus:border-brand-blue/20"
										required
									/>
								</div>
								<div class="flex flex-col gap-2">
									<label
										class="ml-1 text-[10px] font-black tracking-widest text-slate-400 uppercase"
										>{activeMenu === 'guru' ? 'NIP (Opsional)' : 'Asal Sekolah'}</label
									>
									<input
										type="text"
										bind:value={newUser.identifier}
										class="rounded-2xl border-2 border-transparent bg-slate-50 px-5 py-4 font-bold text-slate-700 outline-none focus:border-brand-blue/20"
										required={activeMenu === 'siswa'}
									/>
								</div>
							</div>
							<div class="space-y-5">
								{#if !isEditing}
									<div class="flex flex-col gap-2">
										<label
											class="ml-1 text-[10px] font-black tracking-widest text-slate-400 uppercase"
											>Kata sandi awal</label
										>
										<input
											type="password"
											bind:value={newUser.password}
											class="rounded-2xl border-2 border-transparent bg-slate-50 px-5 py-4 font-bold text-slate-700 outline-none focus:border-brand-blue/20"
											required
										/>
									</div>
								{/if}
								{#if activeMenu === 'siswa'}
									{#if !isEditing}
										<div class="flex flex-col gap-2">
											<label
												class="ml-1 text-[10px] font-black tracking-widest text-slate-400 uppercase"
												>Label petunjuk keamanan</label
											>
											<input
												type="text"
												bind:value={newUser.labelKataKunci}
												class="rounded-2xl border-2 border-transparent bg-slate-50 px-5 py-4 font-bold text-slate-700 outline-none focus:border-brand-blue/20"
												required
											/>
										</div>
										<div class="flex flex-col gap-2">
											<label
												class="ml-1 text-[10px] font-black tracking-widest text-slate-400 uppercase"
												>Jawaban Keamanan</label
											>
											<input
												type="text"
												bind:value={newUser.kataKunci}
												class="rounded-2xl border-2 border-transparent bg-slate-50 px-5 py-4 font-bold text-slate-700 outline-none focus:border-brand-blue/20"
												required
											/>
										</div>
									{/if}
								{:else}
									<div class="flex flex-col gap-2">
										<label
											class="ml-1 text-[10px] font-black tracking-widest text-slate-400 uppercase"
											>Alamat email</label
										>
										<input
											type="email"
											bind:value={newUser.email}
											class="rounded-2xl border-2 border-transparent bg-slate-50 px-5 py-4 font-bold text-slate-700 outline-none focus:border-brand-blue/20"
										/>
									</div>
								{/if}
							</div>
						</div>
					{:else if activeMenu === 'kelas'}
						<div class="space-y-5">
							<div class="flex flex-col gap-2">
								<label class="ml-1 text-[10px] font-black tracking-widest text-slate-400 uppercase"
									>Nama Kelas</label
								>
								<input
									type="text"
									bind:value={newClass.nama_kelas}
									class="rounded-2xl border-2 border-transparent bg-slate-50 px-5 py-4 font-bold text-slate-700 outline-none focus:border-brand-blue/20"
									required
								/>
							</div>
							<div class="flex flex-col gap-2">
								<label class="ml-1 text-[10px] font-black tracking-widest text-slate-400 uppercase"
									>Pilih Periode Belajar</label
								>
								<select
									bind:value={newClass.periode_id}
									class="cursor-pointer rounded-2xl border-2 border-transparent bg-slate-50 px-5 py-4 font-bold text-slate-700 outline-none focus:border-brand-blue/20"
									required
								>
									<option value="" disabled selected>-- Pilih Tahun Ajar --</option>
									{#each periods as period}
										<option value={period.id}
											>{period.tahunAjar} ({period.semester}) - {period.statusAktif}</option
										>
									{/each}
								</select>
							</div>
						</div>
					{:else if activeMenu === 'mapel'}
						<div class="space-y-5">
							<div class="flex flex-col gap-2">
								<label class="ml-1 text-[10px] font-black tracking-widest text-slate-400 uppercase"
									>Nama Mata Pelajaran</label
								>
								<input
									type="text"
									bind:value={newSubject.name}
									class="rounded-2xl border-2 border-transparent bg-slate-50 px-5 py-4 font-bold text-slate-700 outline-none focus:border-brand-blue/20"
									required
								/>
							</div>
						</div>
					{:else if activeMenu === 'admin_users'}
						<div class="space-y-5">
							<div class="flex flex-col gap-2">
								<label class="ml-1 text-[10px] font-black tracking-widest text-slate-400 uppercase"
									>Nama pengguna</label
								>
								<input
									type="text"
									bind:value={newUser.username}
									class="rounded-2xl border-2 border-transparent bg-slate-50 px-5 py-4 font-bold text-slate-700 outline-none focus:border-brand-blue/20"
									required
								/>
							</div>
							<div class="flex flex-col gap-2">
								<label class="ml-1 text-[10px] font-black tracking-widest text-slate-400 uppercase">
									{isEditing ? 'Kata sandi baru (kosongkan jika tidak diubah)' : 'Kata sandi awal'}
								</label>
								<input
									type="password"
									bind:value={newUser.password}
									class="rounded-2xl border-2 border-transparent bg-slate-50 px-5 py-4 font-bold text-slate-700 outline-none focus:border-brand-blue/20"
									required={!isEditing}
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
							>{isEditing ? 'Perbarui data' : 'Simpan data'}</button
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
						<label class="ml-1 text-[10px] font-black tracking-widest text-slate-400 uppercase"
							>Tahun Ajar</label
						>
						<input
							type="text"
							bind:value={newPeriod.tahun_ajar}
							placeholder="Contoh: 2024/2025"
							class="rounded-2xl border-2 border-transparent bg-slate-50 px-5 py-4 text-center font-bold text-slate-700 outline-none focus:border-brand-blue/20"
							required
						/>
					</div>
					<div class="flex flex-col gap-2">
						<label class="ml-1 text-[10px] font-black tracking-widest text-slate-400 uppercase"
							>Semester</label
						>
						<select
							bind:value={newPeriod.semester}
							class="cursor-pointer rounded-2xl border-2 border-transparent bg-slate-50 px-5 py-4 font-bold text-slate-700 outline-none focus:border-brand-blue/20"
						>
							<option value="Ganjil">Ganjil</option>
							<option value="Genap">Genap</option>
						</select>
					</div>
					<div class="flex flex-col gap-2">
						<label class="ml-1 text-[10px] font-black tracking-widest text-slate-400 uppercase"
							>Status</label
						>
						<select
							bind:value={newPeriod.status_aktif}
							class="cursor-pointer rounded-2xl border-2 border-transparent bg-slate-50 px-5 py-4 font-bold text-slate-700 outline-none focus:border-brand-blue/20"
						>
							<option value={1}>Aktif</option>
							<option value={0}>Tidak Aktif</option>
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

	{#if showAssignModal}
		<div
			class="fixed inset-0 z-50 flex items-center justify-center overflow-y-auto bg-slate-900/70 p-4 backdrop-blur-md md:p-6"
		>
			<div
				class="scale-in-center my-auto w-full max-w-md rounded-[3rem] border-t-8 border-purple-500 bg-white p-8 shadow-2xl md:p-10"
			>
				<div class="mb-8 text-center">
					<h3 class="text-2xl font-black tracking-tight text-slate-900 uppercase">
						Plotting Kelas
					</h3>
					<p class="mt-2 text-[10px] font-bold tracking-widest text-slate-400 uppercase">
						Atur penempatan kelas untuk siswa
					</p>
				</div>
				<form onsubmit={handleAssignSubmit} class="space-y-5">
					<div class="rounded-2xl border-2 border-purple-100 bg-purple-50 px-5 py-4">
						<p class="mb-1 text-[10px] font-black tracking-widest text-purple-400 uppercase">
							Siswa Terpilih
						</p>
						<p class="font-bold text-purple-800">{assignData.namaSiswa}</p>
					</div>
					<div class="flex flex-col gap-2">
						<label class="ml-1 text-[10px] font-black tracking-widest text-slate-400 uppercase"
							>Pilih Aksi</label
						>
						<select
							bind:value={assignData.action}
							class="cursor-pointer rounded-2xl border-2 border-transparent bg-slate-50 px-5 py-4 font-bold text-slate-700 outline-none focus:border-purple-200"
							required
						>
							<option value="assign">Masukkan ke Kelas (Baru)</option>
							<option value="update">Mutasi / Pindah Kelas</option>
							<option value="remove">Keluarkan dari Kelas (Hapus)</option>
						</select>
					</div>
					{#if assignData.action === 'update'}
						<div class="flex flex-col gap-2">
							<label class="ml-1 text-[10px] font-black tracking-widest text-amber-500 uppercase"
								>Kelas Asal (Yang Ingin Ditinggalkan)</label
							>
							<select
								bind:value={assignData.oldKelasId}
								class="cursor-pointer rounded-2xl border-2 border-transparent bg-amber-50 px-5 py-4 font-bold text-amber-700 outline-none focus:border-amber-200"
								required
							>
								<option value="" disabled selected>-- Pilih Kelas Asal --</option>
								{#each studentCurrentClasses as cls}
									<option value={cls.id}>{cls.name} ({cls.tahun_ajaran} - {cls.semester})</option>
								{/each}
								{#if studentCurrentClasses.length === 0}
									<option value="" disabled>Siswa belum terdaftar di kelas manapun</option>
								{/if}
							</select>
						</div>
					{/if}
					<div class="flex flex-col gap-2">
						<label class="ml-1 text-[10px] font-black tracking-widest text-slate-400 uppercase"
							>{assignData.action === 'update' ? 'Kelas Tujuan (Baru)' : 'Pilih Kelas'}</label
						>
						<select
							bind:value={assignData.kelasId}
							class="cursor-pointer rounded-2xl border-2 border-transparent bg-slate-50 px-5 py-4 font-bold text-slate-700 outline-none focus:border-purple-200"
							required
						>
							<option value="" disabled selected>-- Daftar Kelas --</option>
							{#if assignData.action === 'remove'}
								{#each studentCurrentClasses as cls}
									<option value={cls.id}>{cls.name} ({cls.tahun_ajaran} - {cls.semester})</option>
								{/each}
								{#if studentCurrentClasses.length === 0}
									<option value="" disabled>Siswa ini sedang tidak berada di kelas manapun</option>
								{/if}
							{:else}
								{#each classes as cls}
									<option value={cls.id}>{cls.name}</option>
								{/each}
							{/if}
						</select>
					</div>
					<div class="flex gap-3 pt-6">
						<button
							type="button"
							onclick={() => (showAssignModal = false)}
							class="flex-1 rounded-2xl bg-slate-100 py-4 text-[10px] font-black tracking-widest text-slate-500 uppercase transition-all hover:bg-slate-200"
							>Batal</button
						>
						<button
							type="submit"
							class="flex-1 rounded-2xl bg-purple-500 py-4 text-[10px] font-black tracking-widest text-white uppercase shadow-lg shadow-purple-500/30 transition-all hover:bg-purple-600"
							>Simpan Status</button
						>
					</div>
				</form>
			</div>
		</div>
	{/if}

	{#if showAssignGuruModal}
		<div
			class="fixed inset-0 z-50 flex items-center justify-center overflow-y-auto bg-slate-900/70 p-4 backdrop-blur-md md:p-6"
		>
			<div
				class="scale-in-center my-auto w-full max-w-md rounded-[3rem] border-t-8 border-indigo-500 bg-white p-8 shadow-2xl md:p-10"
			>
				<div class="mb-8 text-center">
					<h3 class="text-2xl font-black tracking-tight text-slate-900 uppercase">
						Plotting Wali Kelas
					</h3>
					<p class="mt-2 text-[10px] font-bold tracking-widest text-slate-400 uppercase">
						Atur tanggung jawab kelas untuk guru
					</p>
				</div>

				<div class="mb-6 rounded-2xl border-2 border-indigo-100 bg-indigo-50 px-5 py-4">
					<p class="mb-1 text-[10px] font-black tracking-widest text-indigo-400 uppercase">
						Pengajar Terpilih
					</p>
					<p class="font-bold text-indigo-800">{assignGuruData.namaGuru}</p>
				</div>

				<div class="mb-6 space-y-3">
					<p class="ml-1 text-[10px] font-black tracking-widest text-slate-400 uppercase">
						Kelas Tanggung Jawab (Saat Ini)
					</p>

					{#each classes.filter((c) => Number(c.wali_guru_id) === Number(assignGuruData.guruUserId)) as cls}
						<div
							class="flex items-center justify-between rounded-2xl border border-slate-100 bg-white p-4 shadow-sm"
						>
							<div>
								<p class="font-bold text-slate-800">{cls.name}</p>
								<p class="text-[10px] font-medium text-slate-400">
									{periods.find((p) => p.id === cls.periode_id)?.tahunAjar}
								</p>
							</div>
							<button
								onclick={() => removeWaliKelas(cls.id)}
								class="rounded-xl bg-red-50 px-3 py-1.5 text-[10px] font-black tracking-widest text-red-500 uppercase hover:bg-red-100"
								>Cabut</button
							>
						</div>
					{/each}

					{#if classes.filter((c) => Number(c.wali_guru_id) === Number(assignGuruData.guruUserId)).length === 0}
						<div
							class="rounded-2xl border border-dashed border-slate-200 bg-slate-50 p-4 text-center"
						>
							<p class="text-[10px] font-bold text-slate-400">
								Guru ini belum di-assign ke kelas manapun.
							</p>
						</div>
					{/if}
				</div>

				<form onsubmit={handleAssignGuru} class="space-y-4 border-t border-slate-100 pt-4">
					<div class="flex flex-col gap-2">
						<label class="ml-1 text-[10px] font-black tracking-widest text-slate-400 uppercase"
							>Tambah Tanggung Jawab Kelas</label
						>
						<select
							bind:value={assignGuruData.kelasId}
							class="cursor-pointer rounded-2xl border-2 border-transparent bg-slate-50 px-5 py-4 font-bold text-slate-700 outline-none focus:border-indigo-200"
							required
						>
							<option value="" disabled selected>-- Pilih Kelas Kosong --</option>
							{#each classes.filter((c) => c.wali_guru_id !== assignGuruData.guruUserId) as cls}
								<option value={cls.id}
									>{cls.name}
									{#if cls.wali_guru_id}(Sudah ada wali){/if}</option
								>
							{/each}
						</select>
					</div>
					<div class="flex gap-3 pt-4">
						<button
							type="button"
							onclick={() => (showAssignGuruModal = false)}
							class="flex-1 rounded-2xl bg-slate-100 py-4 text-[10px] font-black tracking-widest text-slate-500 uppercase transition-all hover:bg-slate-200"
							>Selesai</button
						>
						<button
							type="submit"
							class="flex-1 rounded-2xl bg-indigo-500 py-4 text-[10px] font-black tracking-widest text-white uppercase shadow-lg shadow-indigo-500/30 transition-all hover:bg-indigo-600"
							>Simpan Status</button
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
						Atur ulang kata sandi
					</h3>
				</div>
				<form onsubmit={handleResetPassword} class="space-y-5">
					<div class="flex flex-col gap-2">
						<label class="ml-1 text-[10px] font-black tracking-widest text-slate-400 uppercase"
							>Nama pengguna / asal sekolah</label
						>
						<input
							type="text"
							bind:value={resetData.username}
							class="cursor-not-allowed rounded-2xl border-2 border-transparent bg-slate-100 px-5 py-4 font-bold text-slate-500 outline-none"
							readonly
						/>
					</div>
					<div class="rounded-2xl border-2 border-blue-100 bg-blue-50 px-5 py-4">
						<p class="mb-1 text-[10px] font-black tracking-widest text-blue-400 uppercase">
							Petunjuk keamanan
						</p>
						<p class="font-bold text-blue-800">{resetData.labelKataKunci}</p>
					</div>
					<div class="flex flex-col gap-2">
						<label class="ml-1 text-[10px] font-black tracking-widest text-amber-500 uppercase"
							>Jawaban Keamanan</label
						>
						<input
							type="text"
							bind:value={resetData.kataKunci}
							class="rounded-2xl border-2 border-transparent bg-amber-50 px-5 py-4 font-bold text-amber-700 transition-all outline-none focus:border-amber-400/30"
							required
						/>
					</div>
					<div class="flex flex-col gap-2 pt-2">
						<label class="ml-1 text-[10px] font-black tracking-widest text-slate-400 uppercase"
							>Kata sandi baru</label
						>
						<input
							type="password"
							bind:value={resetData.newPassword}
							class="rounded-2xl border-2 border-transparent bg-slate-50 px-5 py-4 font-bold text-slate-700 transition-all outline-none focus:border-brand-blue/20"
							required
						/>
					</div>
					<div class="flex flex-col gap-2">
						<label class="ml-1 text-[10px] font-black tracking-widest text-slate-400 uppercase"
							>Konfirmasi kata sandi</label
						>
						<input
							type="password"
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

	{#if showClassStudentsModal}
		<div
			class="fixed inset-0 z-50 flex items-center justify-center overflow-y-auto bg-slate-900/70 p-4 backdrop-blur-md md:p-6"
		>
			<div
				class="scale-in-center my-auto w-full max-w-3xl rounded-[3rem] border-t-8 border-brand-blue bg-white p-8 shadow-2xl md:p-10"
			>
				<div class="mb-6 flex items-start justify-between">
					<div>
						<h3 class="text-2xl font-black tracking-tight text-slate-900 uppercase">
							Daftar Siswa
						</h3>
						<p class="mt-1 text-[10px] font-bold tracking-widest text-slate-400 uppercase">
							Kelas: <span class="text-brand-blue">{selectedClassForStudents.name}</span>
						</p>
					</div>
					<button
						type="button"
						onclick={() => {
							showClassStudentsModal = false;
							selectedStudentToAdd = '';
						}}
						class="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-slate-100 text-slate-500 transition-colors hover:bg-slate-200 hover:text-slate-800"
					>
						<svg
							xmlns="http://www.w3.org/2000/svg"
							class="h-5 w-5"
							viewBox="0 0 24 24"
							fill="none"
							stroke="currentColor"
							stroke-width="2.5"
							stroke-linecap="round"
							stroke-linejoin="round"><path d="M18 6 6 18" /><path d="m6 6 12 12" /></svg
						>
					</button>
				</div>

				<div class="mb-6 flex flex-col gap-3 sm:flex-row">
					<select
						bind:value={selectedStudentToAdd}
						class="flex-1 cursor-pointer rounded-2xl border-2 border-slate-100 bg-slate-50 px-5 py-3 font-bold text-slate-700 transition-all outline-none focus:border-brand-blue/30"
					>
						<option value="" disabled selected>-- Cari & Pilih Siswa untuk Dimasukkan --</option>
						{#each availableStudents as student}
							<option value={student.id}>{student.name} ({student.namaSekolah})</option>
						{/each}
					</select>
					<button
						onclick={addStudentToClass}
						class="rounded-2xl bg-brand-blue px-6 py-3 text-[10px] font-black tracking-widest whitespace-nowrap text-white uppercase shadow-lg shadow-blue-500/20 transition-all hover:scale-105 active:scale-95 disabled:cursor-not-allowed disabled:opacity-50"
						disabled={!selectedStudentToAdd}
					>
						+ Masukkan Siswa
					</button>
				</div>

				<div class="overflow-hidden rounded-[2rem] border border-slate-100 bg-white shadow-inner">
					<div class="max-h-[350px] w-full overflow-y-auto">
						<table class="w-full min-w-[500px] border-collapse text-left">
							<thead class="sticky top-0 z-10 bg-slate-50/90 backdrop-blur-sm">
								<tr>
									<th
										class="px-6 py-4 text-[10px] font-black tracking-[0.2em] text-slate-400 uppercase"
										>Asal Sekolah</th
									>
									<th
										class="px-6 py-4 text-[10px] font-black tracking-[0.2em] text-slate-400 uppercase"
										>Nama Lengkap</th
									>
									<th
										class="px-6 py-4 text-right text-[10px] font-black tracking-[0.2em] text-slate-400 uppercase"
										>Aksi</th
									>
								</tr>
							</thead>
							<tbody class="divide-y divide-slate-50">
								{#each classStudentsList as student}
									<tr class="transition-colors hover:bg-slate-50/50">
										<td class="px-6 py-4 text-sm font-bold text-slate-500"
											>{student.nama_sekolah}</td
										>
										<td class="px-6 py-4 font-bold text-slate-800">{student.nama_lengkap}</td>
										<td class="px-6 py-4 text-right">
											<button
												onclick={() =>
													removeStudentFromClassModal(student.user_id, student.nama_lengkap)}
												class="rounded-xl border border-red-100 bg-red-50 px-4 py-2 text-[10px] font-black tracking-widest text-red-500 uppercase transition-colors hover:bg-red-100 hover:text-red-700"
												>Keluarkan</button
											>
										</td>
									</tr>
								{/each}
								{#if classStudentsList.length === 0}
									<tr>
										<td colspan="3" class="px-6 py-10 text-center">
											<p class="text-sm font-bold text-slate-400">
												Ruangan kelas masih kosong. Silakan tambahkan siswa di atas.
											</p>
										</td>
									</tr>
								{/if}
							</tbody>
						</table>
					</div>
				</div>
			</div>
		</div>
	{/if}

	<!-- MODAL IMPORT EXCEL -->
	{#if showImportModal}
		<div
			class="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/70 p-4 backdrop-blur-md md:p-6"
		>
			<div class="scale-in-center w-full max-w-lg rounded-[2.5rem] bg-white p-8 shadow-2xl md:p-10">
				<div class="mb-6 flex items-center justify-between">
					<div>
						<h3 class="text-xl font-black tracking-tight text-slate-900 uppercase">
							Impor data siswa
						</h3>
						<p class="mt-1 text-xs font-medium text-slate-500">
							Pastikan file sesuai dengan template.
						</p>
					</div>
					<button
						onclick={() => (showImportModal = false)}
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

				<div class="mb-6 rounded-2xl border border-blue-100 bg-blue-50/50 p-4">
					<div class="mb-2 flex items-center justify-between">
						<span class="text-[10px] font-black tracking-widest text-brand-blue uppercase"
							>Petunjuk</span
						>
						<a
							href="/template_import_siswa.xlsx"
							download
							class="flex items-center gap-1.5 rounded-lg bg-white px-3 py-1.5 text-[10px] font-bold text-brand-blue shadow-sm ring-1 ring-blue-200 transition-all hover:bg-brand-blue hover:text-white"
						>
							<svg
								xmlns="http://www.w3.org/2000/svg"
								class="h-3 w-3"
								viewBox="0 0 24 24"
								fill="none"
								stroke="currentColor"
								stroke-width="2.5"
								stroke-linecap="round"
								stroke-linejoin="round"
							>
								<path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
								<polyline points="7 10 12 15 17 10" />
								<line x1="12" x2="12" y1="15" y2="3" />
							</svg>
							Unduh Template
						</a>
					</div>
					<ul class="space-y-1.5 text-[10px] text-slate-600">
						<li><span class="font-bold text-slate-800">Kolom A:</span> Nama Lengkap (Wajib)</li>
						<li><span class="font-bold text-slate-800">Kolom B:</span> Asal Sekolah (Opsional)</li>
						<li>
							<span class="font-bold text-slate-800">Kolom C & D:</span> petunjuk & kata kunci (opsional)
						</li>
					</ul>
				</div>

				<!-- AREA UPLOAD -->
				<div
					class="mb-6 flex flex-col items-center justify-center rounded-2xl border-2 border-dashed border-slate-200 bg-slate-50 p-6 text-center transition-colors hover:border-brand-blue/50"
				>
					<label
						for="file-upload"
						class="cursor-pointer text-sm font-bold text-brand-blue hover:underline"
					>
						Pilih File Excel (.xlsx)
					</label>
					<input
						id="file-upload"
						type="file"
						accept=".xlsx"
						class="hidden"
						onchange={handleFileChange}
					/>
					{#if selectedFile}
						<p class="mt-2 text-[10px] font-medium text-slate-500">
							Terpilih: <span class="font-bold text-slate-800">{selectedFile.name}</span>
						</p>
					{:else}
						<p class="mt-2 text-[10px] font-medium text-slate-400">Belum ada file yang dipilih</p>
					{/if}
				</div>

				<!-- HASIL UPLOAD -->
				{#if uploadResult}
					<div class="mb-6 rounded-2xl border border-green-100 bg-green-50 p-4 text-center">
						<p class="mb-1 text-xs font-black tracking-widest text-green-600 uppercase">
							{uploadResult.message}
						</p>
						<div class="flex justify-center gap-4 text-[10px] font-bold text-slate-600">
							<span>Sukses: {uploadResult.sukses}</span>
							<span>Gagal: {uploadResult.gagal}</span>
						</div>
					</div>
				{/if}

				<div class="flex gap-3">
					<button
						onclick={handleUpload}
						disabled={!selectedFile || isUploading}
						class="flex-1 rounded-2xl bg-slate-900 py-4 text-[10px] font-black tracking-widest text-white uppercase shadow-lg shadow-slate-900/20 transition-all hover:bg-slate-800 disabled:opacity-50"
					>
						{isUploading ? 'Memproses...' : 'Impor data'}
					</button>
				</div>
			</div>
		</div>
	{/if}
</div>
