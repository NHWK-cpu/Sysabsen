import { useState } from 'react';
import { useNavigate } from 'react-router';
import { Button } from './ui/button';
import { Card, CardContent, CardHeader, CardTitle } from './ui/card';
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
import { Label } from './ui/label';
import { 
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from './ui/select';
import {
  Users,
  CheckCircle,
  XCircle,
  LogOut,
  Edit,
  Download,
  Database,
  QrCode
} from 'lucide-react';

interface AttendanceRecord {
  id: string;
  studentName: string;
  status: 'present' | 'absent';
  time: string;
}

export function TeacherDashboard() {
  const navigate = useNavigate();
  const [teacherName] = useState('Dr. Sarah Johnson');
  const [editingRecord, setEditingRecord] = useState<AttendanceRecord | null>(null);
  const [selectedClass, setSelectedClass] = useState<string>('cs-101');
  const [showQRDialog, setShowQRDialog] = useState(false);
  const [records, setRecords] = useState<AttendanceRecord[]>([
    { id: '1', studentName: 'John Smith', status: 'present', time: '9:30 AM' },
    { id: '2', studentName: 'Emma Wilson', status: 'present', time: '9:28 AM' },
    { id: '3', studentName: 'Michael Brown', status: 'absent', time: '-' },
    { id: '4', studentName: 'Sophia Davis', status: 'present', time: '9:32 AM' },
    { id: '5', studentName: 'James Miller', status: 'present', time: '9:25 AM' },
    { id: '6', studentName: 'Olivia Garcia', status: 'absent', time: '-' },
    { id: '7', studentName: 'William Martinez', status: 'present', time: '9:29 AM' },
    { id: '8', studentName: 'Ava Rodriguez', status: 'present', time: '9:31 AM' },
    { id: '9', studentName: 'Ethan Anderson', status: 'absent', time: '-' },
    { id: '10', studentName: 'Isabella Taylor', status: 'present', time: '9:27 AM' },
  ]);

  const totalStudents = records.length;
  const presentCount = records.filter(r => r.status === 'present').length;
  const absentCount = records.filter(r => r.status === 'absent').length;

  const handleLogout = () => {
    navigate('/');
  };

  const handleEditRecord = (record: AttendanceRecord) => {
    setEditingRecord(record);
  };

  const handleSaveEdit = () => {
    if (editingRecord) {
      setRecords(records.map(r => 
        r.id === editingRecord.id ? editingRecord : r
      ));
      setEditingRecord(null);
    }
  };

  const handleExportCSV = () => {
    // Mock CSV export
    const csv = [
      ['Student Name', 'Status', 'Time'],
      ...records.map(r => [r.studentName, r.status, r.time])
    ].map(row => row.join(',')).join('\n');
    
    const blob = new Blob([csv], { type: 'text/csv' });
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `attendance-${new Date().toISOString().split('T')[0]}.csv`;
    a.click();
  };

  const handleBackup = () => {
    // Mock backup functionality
    alert('Attendance data backed up successfully!');
  };

  const classes = [
    { id: 'cs-101', name: 'CS 101 - Introduction to Programming' },
    { id: 'cs-201', name: 'CS 201 - Data Structures' },
    { id: 'cs-301', name: 'CS 301 - Algorithms' },
    { id: 'cs-401', name: 'CS 401 - Software Engineering' },
  ];

  return (
    <div className="min-h-screen flex flex-col bg-background">
      {/* Header */}
      <div className="bg-card border-b px-4 md:px-6 py-4 shadow-sm">
        <div className="max-w-7xl mx-auto flex items-center justify-between">
          <div>
            <h1 className="text-xl md:text-2xl font-semibold text-foreground">
              Attendance Dashboard
            </h1>
            <p className="text-sm text-muted-foreground">March 25, 2026</p>
          </div>
          <div className="flex items-center gap-3">
            <Avatar className="h-10 w-10">
              <AvatarFallback className="bg-primary text-primary-foreground">
                {teacherName.split(' ').map(n => n[0]).join('')}
              </AvatarFallback>
            </Avatar>
            <div className="hidden md:block">
              <p className="font-medium text-foreground text-sm">{teacherName}</p>
              <p className="text-xs text-muted-foreground">Teacher</p>
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
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4 md:gap-6">
            <Card className="shadow-md hover:shadow-lg transition-shadow">
              <CardHeader className="flex flex-row items-center justify-between pb-2">
                <CardTitle className="text-sm font-medium text-muted-foreground">
                  Total Students
                </CardTitle>
                <Users className="w-5 h-5 text-primary" />
              </CardHeader>
              <CardContent>
                <div className="text-3xl font-semibold text-foreground">{totalStudents}</div>
                <p className="text-xs text-muted-foreground mt-1">
                  Enrolled in class
                </p>
              </CardContent>
            </Card>

            <Card className="shadow-md hover:shadow-lg transition-shadow">
              <CardHeader className="flex flex-row items-center justify-between pb-2">
                <CardTitle className="text-sm font-medium text-muted-foreground">
                  Present Today
                </CardTitle>
                <CheckCircle className="w-5 h-5 text-green-600" />
              </CardHeader>
              <CardContent>
                <div className="text-3xl font-semibold text-green-600">{presentCount}</div>
                <p className="text-xs text-muted-foreground mt-1">
                  {Math.round((presentCount / totalStudents) * 100)}% attendance rate
                </p>
              </CardContent>
            </Card>

            <Card className="shadow-md hover:shadow-lg transition-shadow">
              <CardHeader className="flex flex-row items-center justify-between pb-2">
                <CardTitle className="text-sm font-medium text-muted-foreground">
                  Absent Today
                </CardTitle>
                <XCircle className="w-5 h-5 text-red-600" />
              </CardHeader>
              <CardContent>
                <div className="text-3xl font-semibold text-red-600">{absentCount}</div>
                <p className="text-xs text-muted-foreground mt-1">
                  Students not checked in
                </p>
              </CardContent>
            </Card>
          </div>

          {/* Action Buttons */}
          <div className="flex flex-col sm:flex-row gap-3">
            <Button 
              onClick={handleExportCSV}
              className="gap-2 rounded-xl"
              size="lg"
            >
              <Download className="w-4 h-4" />
              Export to CSV
            </Button>
            <Button 
              variant="outline"
              onClick={handleBackup}
              className="gap-2 rounded-xl border-2"
              size="lg"
            >
              <Database className="w-4 h-4" />
              Backup Data
            </Button>
          </div>

          {/* Attendance Table */}
          <Card className="shadow-lg">
            <CardHeader className="flex flex-row items-center justify-between">
              <CardTitle>Attendance Records</CardTitle>
              <div className="flex items-center gap-2">
                <Select value={selectedClass} onValueChange={setSelectedClass}>
                  <SelectTrigger className="w-[280px]">
                    <SelectValue placeholder="Select class" />
                  </SelectTrigger>
                  <SelectContent>
                    {classes.map((cls) => (
                      <SelectItem key={cls.id} value={cls.id}>
                        {cls.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setShowQRDialog(true)}
                  className="gap-2 border-2"
                >
                  <QrCode className="w-4 h-4" />
                  Show QR
                </Button>
              </div>
            </CardHeader>
            <CardContent className="p-0">
              <div className="overflow-x-auto">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Student Name</TableHead>
                      <TableHead>Status</TableHead>
                      <TableHead>Check-in Time</TableHead>
                      <TableHead className="text-right">Actions</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {records.map((record) => (
                      <TableRow key={record.id} className="hover:bg-muted/50">
                        <TableCell className="font-medium">
                          {record.studentName}
                        </TableCell>
                        <TableCell>
                          {record.status === 'present' ? (
                            <span className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-medium bg-green-100 text-green-700">
                              <CheckCircle className="w-3.5 h-3.5" />
                              Present
                            </span>
                          ) : (
                            <span className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-medium bg-red-100 text-red-700">
                              <XCircle className="w-3.5 h-3.5" />
                              Absent
                            </span>
                          )}
                        </TableCell>
                        <TableCell className="text-muted-foreground">
                          {record.time}
                        </TableCell>
                        <TableCell className="text-right">
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => handleEditRecord(record)}
                            className="gap-2"
                          >
                            <Edit className="w-4 h-4" />
                            Edit
                          </Button>
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

      {/* Edit Dialog */}
      <Dialog open={!!editingRecord} onOpenChange={() => setEditingRecord(null)}>
        <DialogContent className="sm:max-w-[425px]">
          <DialogHeader>
            <DialogTitle>Edit Attendance</DialogTitle>
            <DialogDescription>
              Update the attendance status for {editingRecord?.studentName}
            </DialogDescription>
          </DialogHeader>
          <div className="grid gap-4 py-4">
            <div className="grid gap-2">
              <Label htmlFor="status">Status</Label>
              <Select
                value={editingRecord?.status}
                onValueChange={(value: 'present' | 'absent') =>
                  setEditingRecord(prev => prev ? {...prev, status: value} : null)
                }
              >
                <SelectTrigger id="status">
                  <SelectValue placeholder="Select status" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="present">Present</SelectItem>
                  <SelectItem value="absent">Absent</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setEditingRecord(null)}>
              Cancel
            </Button>
            <Button onClick={handleSaveEdit}>
              Save Changes
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* QR Code Dialog */}
      <Dialog open={showQRDialog} onOpenChange={setShowQRDialog}>
        <DialogContent className="sm:max-w-[500px]">
          <DialogHeader>
            <DialogTitle>Class QR Code</DialogTitle>
            <DialogDescription>
              Students can scan this QR code to mark their attendance
            </DialogDescription>
          </DialogHeader>
          <div className="flex flex-col items-center justify-center py-8">
            <div className="bg-white p-8 rounded-2xl shadow-lg">
              <div className="w-64 h-64 bg-gray-100 rounded-xl flex items-center justify-center border-4 border-primary">
                <QrCode className="w-48 h-48 text-primary" />
              </div>
            </div>
            <p className="mt-6 text-center font-medium text-foreground">
              {classes.find(c => c.id === selectedClass)?.name}
            </p>
            <p className="text-sm text-muted-foreground mt-1">
              Session ID: {selectedClass.toUpperCase()}-{new Date().getTime().toString().slice(-6)}
            </p>
          </div>
          <DialogFooter>
            <Button onClick={() => setShowQRDialog(false)} className="w-full">
              Close
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
