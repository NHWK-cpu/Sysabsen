import { useState } from 'react';
import { useNavigate } from 'react-router';
import { Button } from './ui/button';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from './ui/card';
import { Avatar, AvatarFallback } from './ui/avatar';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow
} from './ui/table';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle
} from './ui/dialog';
import { Input } from './ui/input';
import { Label } from './ui/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from './ui/select';
import { Badge } from './ui/badge';
import {
  Shield,
  LogOut,
  UserPlus,
  Key,
  Trash2,
  CheckCircle,
  XCircle,
  Clock,
  Users,
  AlertTriangle
} from 'lucide-react';

interface User {
  id: string;
  name: string;
  email: string;
  role: 'student' | 'teacher';
  lastLogin: string;
  status: 'active' | 'pending';
  fatherName?: string;
}

interface PendingLogin {
  id: string;
  userName: string;
  email: string;
  role: 'student' | 'teacher';
  requestTime: string;
  ipAddress: string;
}

export function AdminDashboard() {
  const navigate = useNavigate();
  const [adminName] = useState('System Administrator');

  // Dialogs state
  const [showAddUserDialog, setShowAddUserDialog] = useState(false);
  const [showChangePasswordDialog, setShowChangePasswordDialog] = useState(false);
  const [showDeleteConfirmDialog, setShowDeleteConfirmDialog] = useState(false);
  const [showApproveLoginDialog, setShowApproveLoginDialog] = useState(false);

  // Add user form state
  const [newUser, setNewUser] = useState({
    name: '',
    email: '',
    role: 'student' as 'student' | 'teacher',
    password: '',
    fatherName: ''
  });

  // Change password form state
  const [passwordForm, setPasswordForm] = useState({
    userId: '',
    currentPassword: '',
    keyword: '',
    newPassword: '',
    confirmPassword: ''
  });

  // Users state
  const [users, setUsers] = useState<User[]>([
    { id: '1', name: 'John Smith', email: 'john@example.com', role: 'student', lastLogin: '2026-04-01', status: 'active', fatherName: 'Robert Smith' },
    { id: '2', name: 'Emma Wilson', email: 'emma@example.com', role: 'student', lastLogin: '2026-03-28', status: 'active', fatherName: 'James Wilson' },
    { id: '3', name: 'Dr. Sarah Johnson', email: 'sarah@example.com', role: 'teacher', lastLogin: '2026-04-02', status: 'active', fatherName: 'Michael Johnson' },
    { id: '4', name: 'Michael Brown', email: 'michael@example.com', role: 'student', lastLogin: '2025-12-15', status: 'active', fatherName: 'David Brown' },
    { id: '5', name: 'Prof. David Lee', email: 'david@example.com', role: 'teacher', lastLogin: '2026-01-20', status: 'active', fatherName: 'Thomas Lee' },
  ]);

  // Pending login requests
  const [pendingLogins, setPendingLogins] = useState<PendingLogin[]>([
    { id: '1', userName: 'Alice Cooper', email: 'alice@example.com', role: 'student', requestTime: '2026-04-02 09:45', ipAddress: '192.168.1.100' },
    { id: '2', userName: 'Bob Martinez', email: 'bob@example.com', role: 'teacher', requestTime: '2026-04-02 10:30', ipAddress: '192.168.1.101' },
  ]);

  const [userToDelete, setUserToDelete] = useState<User | null>(null);

  // Calculate stats
  const totalUsers = users.length;
  const activeUsers = users.filter(u => u.status === 'active').length;
  const inactiveUsers = getInactiveUsers().length;
  const pendingApprovals = pendingLogins.length;

  function getInactiveUsers() {
    const threeMonthsAgo = new Date();
    threeMonthsAgo.setMonth(threeMonthsAgo.getMonth() - 3);
    return users.filter(user => {
      const lastLogin = new Date(user.lastLogin);
      return lastLogin < threeMonthsAgo;
    });
  }

  function getDaysSinceLogin(dateString: string) {
    const lastLogin = new Date(dateString);
    const today = new Date();
    const diffTime = Math.abs(today.getTime() - lastLogin.getTime());
    const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24));
    return diffDays;
  }

  const handleLogout = () => {
    navigate('/');
  };

  const handleAddUser = () => {
    const user: User = {
      id: (users.length + 1).toString(),
      name: newUser.name,
      email: newUser.email,
      role: newUser.role,
      lastLogin: new Date().toISOString().split('T')[0],
      status: 'active',
      fatherName: newUser.fatherName
    };
    setUsers([...users, user]);
    setNewUser({ name: '', email: '', role: 'student', password: '', fatherName: '' });
    setShowAddUserDialog(false);
  };

  const handleChangePassword = () => {
    // In real app, verify keyword against user's stored keyword
    if (passwordForm.newPassword !== passwordForm.confirmPassword) {
      alert('Passwords do not match!');
      return;
    }
    // Mock verification
    alert('Password changed successfully!');
    setPasswordForm({ userId: '', currentPassword: '', keyword: '', newPassword: '', confirmPassword: '' });
    setShowChangePasswordDialog(false);
  };

  const handleDeleteUser = () => {
    if (userToDelete) {
      setUsers(users.filter(u => u.id !== userToDelete.id));
      setUserToDelete(null);
      setShowDeleteConfirmDialog(false);
    }
  };

  const handleApproveLogin = (loginId: string) => {
    setPendingLogins(pendingLogins.filter(l => l.id !== loginId));
  };

  const handleRejectLogin = (loginId: string) => {
    setPendingLogins(pendingLogins.filter(l => l.id !== loginId));
  };

  return (
    <div className="min-h-screen flex flex-col bg-background">
      {/* Header */}
      <div className="bg-card border-b px-4 md:px-6 py-4 shadow-sm">
        <div className="max-w-7xl mx-auto flex items-center justify-between">
          <div>
            <h1 className="text-xl md:text-2xl font-semibold text-foreground flex items-center gap-2">
              <Shield className="w-6 h-6 text-red-600" />
              Admin Dashboard
            </h1>
            <p className="text-sm text-muted-foreground">System Management Panel</p>
          </div>
          <div className="flex items-center gap-3">
            <Avatar className="h-10 w-10">
              <AvatarFallback className="bg-red-600 text-white">
                AD
              </AvatarFallback>
            </Avatar>
            <div className="hidden md:block">
              <p className="font-medium text-foreground text-sm">{adminName}</p>
              <p className="text-xs text-muted-foreground">Administrator</p>
            </div>
            <Button
              variant="ghost"
              size="sm"
              onClick={handleLogout}
              className="gap-2 ml-2"
            >
              <LogOut className="w-4 h-4" />
              <span className="hidden md:inline">Logout</span>
            </Button>
          </div>
        </div>
      </div>

      {/* Main Content */}
      <div className="flex-1 p-4 md:p-6 lg:p-8">
        <div className="max-w-7xl mx-auto space-y-6 md:space-y-8">
          {/* Summary Cards */}
          <div className="grid grid-cols-1 md:grid-cols-4 gap-4 md:gap-6">
            <Card className="shadow-md hover:shadow-lg transition-shadow">
              <CardHeader className="flex flex-row items-center justify-between pb-2">
                <CardTitle className="text-sm font-medium text-muted-foreground">
                  Total Users
                </CardTitle>
                <Users className="w-5 h-5 text-primary" />
              </CardHeader>
              <CardContent>
                <div className="text-3xl font-semibold text-foreground">{totalUsers}</div>
                <p className="text-xs text-muted-foreground mt-1">
                  All registered accounts
                </p>
              </CardContent>
            </Card>

            <Card className="shadow-md hover:shadow-lg transition-shadow">
              <CardHeader className="flex flex-row items-center justify-between pb-2">
                <CardTitle className="text-sm font-medium text-muted-foreground">
                  Active Users
                </CardTitle>
                <CheckCircle className="w-5 h-5 text-green-600" />
              </CardHeader>
              <CardContent>
                <div className="text-3xl font-semibold text-green-600">{activeUsers}</div>
                <p className="text-xs text-muted-foreground mt-1">
                  Currently active
                </p>
              </CardContent>
            </Card>

            <Card className="shadow-md hover:shadow-lg transition-shadow">
              <CardHeader className="flex flex-row items-center justify-between pb-2">
                <CardTitle className="text-sm font-medium text-muted-foreground">
                  Inactive Users
                </CardTitle>
                <AlertTriangle className="w-5 h-5 text-orange-600" />
              </CardHeader>
              <CardContent>
                <div className="text-3xl font-semibold text-orange-600">{inactiveUsers}</div>
                <p className="text-xs text-muted-foreground mt-1">
                  Not logged in 90+ days
                </p>
              </CardContent>
            </Card>

            <Card className="shadow-md hover:shadow-lg transition-shadow">
              <CardHeader className="flex flex-row items-center justify-between pb-2">
                <CardTitle className="text-sm font-medium text-muted-foreground">
                  Pending Approvals
                </CardTitle>
                <Clock className="w-5 h-5 text-blue-600" />
              </CardHeader>
              <CardContent>
                <div className="text-3xl font-semibold text-blue-600">{pendingApprovals}</div>
                <p className="text-xs text-muted-foreground mt-1">
                  Awaiting approval
                </p>
              </CardContent>
            </Card>
          </div>

          {/* Action Buttons */}
          <div className="flex flex-col sm:flex-row gap-3">
            <Button
              onClick={() => setShowAddUserDialog(true)}
              className="gap-2 rounded-xl"
              size="lg"
            >
              <UserPlus className="w-4 h-4" />
              Add New Account
            </Button>
            <Button
              variant="outline"
              onClick={() => setShowChangePasswordDialog(true)}
              className="gap-2 rounded-xl border-2"
              size="lg"
            >
              <Key className="w-4 h-4" />
              Change Password
            </Button>
            <Button
              variant="outline"
              onClick={() => setShowApproveLoginDialog(true)}
              className="gap-2 rounded-xl border-2"
              size="lg"
            >
              <CheckCircle className="w-4 h-4" />
              Approve Logins ({pendingLogins.length})
            </Button>
          </div>

          {/* Inactive Users Table */}
          {inactiveUsers > 0 && (
            <Card className="shadow-lg border-orange-200">
              <CardHeader>
                <CardTitle className="flex items-center gap-2 text-orange-700">
                  <AlertTriangle className="w-5 h-5" />
                  Inactive Accounts (90+ Days)
                </CardTitle>
                <CardDescription>
                  These accounts haven't logged in for over 90 days and may be eligible for deletion
                </CardDescription>
              </CardHeader>
              <CardContent className="p-0">
                <div className="overflow-x-auto">
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead>Name</TableHead>
                        <TableHead>Email</TableHead>
                        <TableHead>Role</TableHead>
                        <TableHead>Last Login</TableHead>
                        <TableHead>Days Inactive</TableHead>
                        <TableHead className="text-right">Actions</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {getInactiveUsers().map((user) => (
                        <TableRow key={user.id} className="hover:bg-muted/50">
                          <TableCell className="font-medium">{user.name}</TableCell>
                          <TableCell className="text-muted-foreground">{user.email}</TableCell>
                          <TableCell>
                            <Badge variant={user.role === 'teacher' ? 'default' : 'secondary'}>
                              {user.role}
                            </Badge>
                          </TableCell>
                          <TableCell className="text-muted-foreground">{user.lastLogin}</TableCell>
                          <TableCell>
                            <span className="text-orange-600 font-medium">
                              {getDaysSinceLogin(user.lastLogin)} days
                            </span>
                          </TableCell>
                          <TableCell className="text-right">
                            <Button
                              variant="ghost"
                              size="sm"
                              onClick={() => {
                                setUserToDelete(user);
                                setShowDeleteConfirmDialog(true);
                              }}
                              className="gap-2 text-red-600 hover:text-red-700 hover:bg-red-50"
                            >
                              <Trash2 className="w-4 h-4" />
                              Delete
                            </Button>
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </div>
              </CardContent>
            </Card>
          )}

          {/* All Users Table */}
          <Card className="shadow-lg">
            <CardHeader>
              <CardTitle>All User Accounts</CardTitle>
              <CardDescription>
                Manage all registered users in the system
              </CardDescription>
            </CardHeader>
            <CardContent className="p-0">
              <div className="overflow-x-auto">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Name</TableHead>
                      <TableHead>Email</TableHead>
                      <TableHead>Role</TableHead>
                      <TableHead>Last Login</TableHead>
                      <TableHead>Status</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {users.map((user) => (
                      <TableRow key={user.id} className="hover:bg-muted/50">
                        <TableCell className="font-medium">{user.name}</TableCell>
                        <TableCell className="text-muted-foreground">{user.email}</TableCell>
                        <TableCell>
                          <Badge variant={user.role === 'teacher' ? 'default' : 'secondary'}>
                            {user.role}
                          </Badge>
                        </TableCell>
                        <TableCell className="text-muted-foreground">{user.lastLogin}</TableCell>
                        <TableCell>
                          {user.status === 'active' ? (
                            <span className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-medium bg-green-100 text-green-700">
                              <CheckCircle className="w-3.5 h-3.5" />
                              Active
                            </span>
                          ) : (
                            <span className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-medium bg-yellow-100 text-yellow-700">
                              <Clock className="w-3.5 h-3.5" />
                              Pending
                            </span>
                          )}
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>

      {/* Add User Dialog */}
      <Dialog open={showAddUserDialog} onOpenChange={setShowAddUserDialog}>
        <DialogContent className="sm:max-w-[500px]">
          <DialogHeader>
            <DialogTitle>Add New User Account</DialogTitle>
            <DialogDescription>
              Create a new student or teacher account
            </DialogDescription>
          </DialogHeader>
          <div className="grid gap-4 py-4">
            <div className="grid gap-2">
              <Label htmlFor="name">Full Name</Label>
              <Input
                id="name"
                placeholder="Enter full name"
                value={newUser.name}
                onChange={(e) => setNewUser({ ...newUser, name: e.target.value })}
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="email">Email Address</Label>
              <Input
                id="email"
                type="email"
                placeholder="Enter email address"
                value={newUser.email}
                onChange={(e) => setNewUser({ ...newUser, email: e.target.value })}
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="role">Role</Label>
              <Select value={newUser.role} onValueChange={(value: 'student' | 'teacher') => setNewUser({ ...newUser, role: value })}>
                <SelectTrigger id="role">
                  <SelectValue placeholder="Select role" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="student">Student</SelectItem>
                  <SelectItem value="teacher">Teacher</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className="grid gap-2">
              <Label htmlFor="fatherName">Father's Name (Security Keyword)</Label>
              <Input
                id="fatherName"
                placeholder="Enter father's name"
                value={newUser.fatherName}
                onChange={(e) => setNewUser({ ...newUser, fatherName: e.target.value })}
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="password">Initial Password</Label>
              <Input
                id="password"
                type="password"
                placeholder="Enter initial password"
                value={newUser.password}
                onChange={(e) => setNewUser({ ...newUser, password: e.target.value })}
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowAddUserDialog(false)}>
              Cancel
            </Button>
            <Button onClick={handleAddUser}>
              Create Account
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Change Password Dialog */}
      <Dialog open={showChangePasswordDialog} onOpenChange={setShowChangePasswordDialog}>
        <DialogContent className="sm:max-w-[500px]">
          <DialogHeader>
            <DialogTitle>Change User Password</DialogTitle>
            <DialogDescription>
              Verify keyword and set a new password for the user
            </DialogDescription>
          </DialogHeader>
          <div className="grid gap-4 py-4">
            <div className="grid gap-2">
              <Label htmlFor="userId">Select User</Label>
              <Select value={passwordForm.userId} onValueChange={(value) => setPasswordForm({ ...passwordForm, userId: value })}>
                <SelectTrigger id="userId">
                  <SelectValue placeholder="Select user" />
                </SelectTrigger>
                <SelectContent>
                  {users.map((user) => (
                    <SelectItem key={user.id} value={user.id}>
                      {user.name} ({user.email})
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="grid gap-2">
              <Label htmlFor="keyword">Security Keyword (Father's Name)</Label>
              <Input
                id="keyword"
                placeholder="Enter father's name for verification"
                value={passwordForm.keyword}
                onChange={(e) => setPasswordForm({ ...passwordForm, keyword: e.target.value })}
              />
              <p className="text-xs text-muted-foreground">
                Required for verification before changing password
              </p>
            </div>
            <div className="grid gap-2">
              <Label htmlFor="newPassword">New Password</Label>
              <Input
                id="newPassword"
                type="password"
                placeholder="Enter new password"
                value={passwordForm.newPassword}
                onChange={(e) => setPasswordForm({ ...passwordForm, newPassword: e.target.value })}
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="confirmPassword">Confirm New Password</Label>
              <Input
                id="confirmPassword"
                type="password"
                placeholder="Confirm new password"
                value={passwordForm.confirmPassword}
                onChange={(e) => setPasswordForm({ ...passwordForm, confirmPassword: e.target.value })}
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowChangePasswordDialog(false)}>
              Cancel
            </Button>
            <Button onClick={handleChangePassword}>
              Change Password
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete Confirmation Dialog */}
      <Dialog open={showDeleteConfirmDialog} onOpenChange={setShowDeleteConfirmDialog}>
        <DialogContent className="sm:max-w-[425px]">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2 text-red-600">
              <AlertTriangle className="w-5 h-5" />
              Confirm Account Deletion
            </DialogTitle>
            <DialogDescription>
              This action cannot be undone. The user account and all associated data will be permanently deleted.
            </DialogDescription>
          </DialogHeader>
          <div className="py-4">
            <div className="bg-red-50 border border-red-200 rounded-xl p-4">
              <p className="font-medium text-foreground">Account to be deleted:</p>
              <p className="text-sm text-muted-foreground mt-2">
                Name: <span className="font-medium text-foreground">{userToDelete?.name}</span>
              </p>
              <p className="text-sm text-muted-foreground">
                Email: <span className="font-medium text-foreground">{userToDelete?.email}</span>
              </p>
              <p className="text-sm text-muted-foreground">
                Last Login: <span className="font-medium text-foreground">{userToDelete?.lastLogin}</span>
              </p>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowDeleteConfirmDialog(false)}>
              Cancel
            </Button>
            <Button
              variant="destructive"
              onClick={handleDeleteUser}
              className="gap-2"
            >
              <Trash2 className="w-4 h-4" />
              Delete Account
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Approve Login Dialog */}
      <Dialog open={showApproveLoginDialog} onOpenChange={setShowApproveLoginDialog}>
        <DialogContent className="sm:max-w-[700px]">
          <DialogHeader>
            <DialogTitle>Pending Login Approvals</DialogTitle>
            <DialogDescription>
              Review and approve or reject login requests
            </DialogDescription>
          </DialogHeader>
          <div className="py-4">
            {pendingLogins.length === 0 ? (
              <div className="text-center py-8 text-muted-foreground">
                <CheckCircle className="w-12 h-12 mx-auto mb-4 text-green-600" />
                <p>No pending login requests</p>
              </div>
            ) : (
              <div className="space-y-3 max-h-[400px] overflow-y-auto">
                {pendingLogins.map((login) => (
                  <Card key={login.id} className="border-2">
                    <CardContent className="p-4">
                      <div className="flex items-start justify-between gap-4">
                        <div className="flex-1 space-y-2">
                          <div className="flex items-center gap-2">
                            <p className="font-medium text-foreground">{login.userName}</p>
                            <Badge variant={login.role === 'teacher' ? 'default' : 'secondary'}>
                              {login.role}
                            </Badge>
                          </div>
                          <p className="text-sm text-muted-foreground">{login.email}</p>
                          <div className="flex gap-4 text-xs text-muted-foreground">
                            <span className="flex items-center gap-1">
                              <Clock className="w-3 h-3" />
                              {login.requestTime}
                            </span>
                            <span>IP: {login.ipAddress}</span>
                          </div>
                        </div>
                        <div className="flex gap-2">
                          <Button
                            size="sm"
                            onClick={() => handleApproveLogin(login.id)}
                            className="gap-2"
                          >
                            <CheckCircle className="w-4 h-4" />
                            Approve
                          </Button>
                          <Button
                            variant="outline"
                            size="sm"
                            onClick={() => handleRejectLogin(login.id)}
                            className="gap-2 border-red-200 text-red-600 hover:bg-red-50"
                          >
                            <XCircle className="w-4 h-4" />
                            Reject
                          </Button>
                        </div>
                      </div>
                    </CardContent>
                  </Card>
                ))}
              </div>
            )}
          </div>
          <DialogFooter>
            <Button onClick={() => setShowApproveLoginDialog(false)}>
              Close
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
