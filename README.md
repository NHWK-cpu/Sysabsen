# Attendance Management System

## Project Overview

This repository implements a **full‑stack web‑based attendance system** for a private tutoring centre.

- **Backend** – Go (net/http) REST API with JWT authentication, role‑based access control, QR‑code & geofencing attendance, scheduled MySQL backups, Google Drive integration, and Swagger documentation.
- **Frontend** – SvelteKit (TypeScript) responsive UI for admins, teachers, and students; consumes the backend API, handles login, QR generation, attendance submission, dashboards, and file uploads.

Both parts communicate over HTTPS and are configurable via environment variables (`.env`).

---

## Directory Structure

```
Backend/                     # Go backend source
 ├─ config/        # env loader, DB connection
 ├─ controllers/   # HTTP handlers (auth, admin, guru, siswa, etc.)
 ├─ middlewares/   # JWT, role checks, CORS
 ├─ helpers/       # Google Drive, email, geofence, random utils
 ├─ models/        # DB structs (User, Siswa, Kelas, …)
 ├─ docs/          # Swagger generated files
 └─ main.go        # Server bootstrap

frontend/                    # SvelteKit frontend source
 ├─ src/           # Svelte components, stores, API wrappers
 ├─ static/        # Images, QR assets
 ├─ build/         # Production output (generated)
 ├─ package.json, tsconfig.json, vite.config.ts
 └─ panduan_deploy.txt  # Deployment notes

README.md          # <‑ this file
LICENSE             # MIT license
```

---

## Backend Specification

| Feature | Description |
|---|---|
| **Entry point** | `Backend/main.go` – loads `.env`, connects to MySQL, creates a default super‑admin (`admin_utama`/`admin123`), starts a daily scheduler that triggers a monthly MySQL dump and uploads it to Google Drive. |
| **Routing** | Uses `http.HandleFunc` with custom middle‑wares. Public routes: `/login`, `/login/siswa`, `/register/siswa`. Admin routes protected by `JWTMiddleware` + `AdminOnly`. Super‑admin routes (`SuperAdminOnly`) for hard‑delete, re‑activate, activity logs. Teacher routes (`GuruOnly`) for QR generation, manual attendance, schedule view, export, password reset. Student routes (`SiswaOnly`) for attendance marking and history. |
| **Authentication** | JWT (`Authorization: Bearer <token>`). Middleware validates token and injects user info; role‑based checks (`AdminOnly`, `GuruOnly`, `SiswaOnly`, `SuperAdminOnly`). |
| **Database** | MySQL accessed via `config/database.go`. Models include `User`, `Guru`, `Siswa`, `Kelas`, `Mapel`, `Periode`, `Sesi`, `Kehadiran`, `UserDevice`, etc. |
| **Core Features** | • **QR & Geofencing** – teachers generate QR codes per session; students submit attendance via QR or GPS radius check.<br>• **Automatic backup** – on day 1 of each month a dump is created, uploaded to Google Drive, and the local file removed.<br>• **Import/Export** – bulk import of students via Excel, export of attendance data for teachers.<br>• **Device registration** – admin can approve/reject teacher devices for push notifications.<br>• **Password reset flow** – token‑based email reset for teachers. |
| **Helpers** | `helpers/gdrive.go` (Google Drive API), `helpers/email.go` (SMTP), `helpers/geofence.go` (distance check), `helpers/random.go` (OTP). |
| **Swagger** | Docs at `/swagger/` via `github.com/swaggo/http-swagger`. |
| **CORS** | Dynamically built from `FRONTEND_URL` env variable, defaulting to `http://localhost:5173`. |
| **Configuration** | `.env` (example fields): `DB_NAME`, `DB_PASSWORD`, `FRONTEND_URL`, `GOOGLE_DRIVE_CLIENT_ID`, `GOOGLE_DRIVE_CLIENT_SECRET`, `EMAIL_HOST`, `EMAIL_USER`, `EMAIL_PASS`, `PORT`. |
| **Testing** | Unit tests live alongside each controller (`*_test.go`). A11y warnings are suppressed per user request. |
| **Build & Run** | ```bash\ncd Backend\ngo run main.go\n``` (or `go build`). |

---

## Frontend Specification

| Feature | Description |
|---|---|
| **Framework** | SvelteKit (Vite) with TypeScript. |
| **Structure** | `src/` contains page components (`+page.svelte`), stores, and reusable UI components. `static/` holds images, icons, and generated QR assets. |
| **Entry point** | `src/App.svelte` (root component) bootstraps routing and global stores. |
| **Routing** | SvelteKit file‑system routing (e.g., `/src/routes/admin/*`, `/src/routes/guru/*`, `/src/routes/siswa/*`). |
| **State Management** | Svelte stores (`writable`, `derived`) keep auth token, user profile, current class/period data. |
| **API Interaction** | Wrapper functions in `src/lib/api.ts` call backend endpoints with the JWT token stored in localStorage. |
| **Key UI Features** | • Login / Register pages for admins, teachers, students.<br>• Admin dashboard – user, class, period, mapel management, backup triggers.<br>• Teacher dashboard – generate QR, view schedule, export attendance, password reset.<br>• Student view – submit attendance (QR scan or geofence), view history.<br>• Responsive design – works on desktop and mobile browsers. |
| **QR Generation** | Uses `qrcode` npm package; QR image rendered as data‑URL on teacher dashboard. |
| **File Upload** | Excel files for bulk student import sent via multipart/form‑data to `/admin/siswa/import`. |
| **Environment** | `VITE_BACKEND_URL` read from `.env` at build time; defaults to `http://localhost:8080`. |
| **Build** | ```bash\ncd frontend\nnpm install\nnpm run build   # output in build/\n``` |
| **Deployment** | `vercel.json` for Vercel deployment; `panduan_deploy.txt` describes steps for Fly.io (backend) and Vercel (frontend). |
| **Linting / Formatting** | ESLint + Prettier (`eslint.config.js`, `.prettierrc`). A11y warnings are hidden per user feedback. |
| **Testing** | Vitest for unit tests; Playwright for end‑to‑end flows (login, QR scan, attendance). |

---

## Project Highlights & Advantages

1. **Role‑Based Security** – Granular middle‑wares ensure each endpoint is only accessible by the appropriate user type.
2. **Automatic, Cloud‑Backed Backups** – Monthly MySQL dumps are automatically uploaded to Google Drive, reducing data‑loss risk.
3. **QR & Geofencing Attendance** – Teachers generate session‑specific QR codes; students can also attest attendance via GPS radius checks, covering both online and on‑site classes.
4. **Swagger Documentation** – Self‑hosted OpenAPI UI (`/swagger/`) keeps API contracts transparent.
5. **Scalable Frontend** – SvelteKit provides fast hydration and small bundle size, ideal for low‑spec devices used by students.
6. **Bulk Operations** – Excel import for student lists and CSV export of attendance streamline administrative workload.
7. **Extensible Design** – Clear separation (`controllers`, `models`, `middlewares`, `helpers`) enables adding new features (e.g., payment integration) without touching core routing.

---

## Getting Started

### Prerequisites
- **Go 1.22+**
- **Node.js 20+**, npm
- **MySQL** instance (create a DB; credentials go in `.env`)
- **Google Drive API** credentials (service‑account JSON placed as `client_secret.json`)

### Backend
```bash
cd Backend
cp .env.example .env   # edit values
go mod tidy
go run main.go
```
Server listens on port from `PORT` env (default **8080**). Swagger UI at `http://localhost:8080/swagger/`.

### Frontend
```bash
cd frontend
cp .env.example .env   # set VITE_BACKEND_URL if needed
npm install
npm run dev   # dev server at http://localhost:5173
npm run build  # static site in build/
```

### Deployment
- **Backend** – Dockerfile provided; works on Fly.io or any container platform.
- **Frontend** – Deploy via Vercel (`vercel.json`) or serve the static `build/` folder behind any web server (NGINX, Apache, etc.).

---
