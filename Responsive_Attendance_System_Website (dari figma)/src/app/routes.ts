import { createBrowserRouter } from 'react-router';
import { LandingPage } from './components/LandingPage';
import { StudentLogin } from './components/StudentLogin';
import { TeacherLogin } from './components/TeacherLogin';
import { StudentScanner } from './components/StudentScanner';
import { TeacherDashboard } from './components/TeacherDashboard';
import { AdminLogin } from './components/AdminLogin';
import { AdminDashboard } from './components/AdminDashboard';

export const router = createBrowserRouter([
  {
    path: '/',
    Component: LandingPage,
  },
  {
    path: '/login/student',
    Component: StudentLogin,
  },
  {
    path: '/login/teacher',
    Component: TeacherLogin,
  },
  {
    path: '/login/admin',
    Component: AdminLogin,
  },
  {
    path: '/student/scanner',
    Component: StudentScanner,
  },
  {
    path: '/teacher/dashboard',
    Component: TeacherDashboard,
  },
  {
    path: '/admin/dashboard',
    Component: AdminDashboard,
  },
]);
