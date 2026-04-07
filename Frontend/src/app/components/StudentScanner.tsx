import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router';
import { Button } from './ui/button';
import { Card } from './ui/card';
import { Avatar, AvatarFallback } from './ui/avatar';
import { Camera, CheckCircle2, XCircle, User, LogOut } from 'lucide-react';

export function StudentScanner() {
  const navigate = useNavigate();
  const [scanStatus, setScanStatus] = useState<'idle' | 'scanning' | 'success' | 'error'>('idle');
  const [studentName] = useState('John Smith');

  useEffect(() => {
    // Auto-start scanning when component mounts
    setScanStatus('scanning');
  }, []);

  const handleScan = () => {
    setScanStatus('scanning');
    
    // Simulate QR scan after 2 seconds
    setTimeout(() => {
      // Randomly succeed or fail for demo purposes
      const success = Math.random() > 0.2;
      setScanStatus(success ? 'success' : 'error');
      
      // Reset after 3 seconds
      setTimeout(() => {
        setScanStatus('idle');
      }, 3000);
    }, 2000);
  };

  const handleLogout = () => {
    navigate('/');
  };

  return (
    <div className="min-h-screen flex flex-col bg-background">
      {/* Header */}
      <div className="bg-card border-b px-4 py-4 shadow-sm">
        <div className="max-w-2xl mx-auto flex items-center justify-between">
          <div className="flex items-center gap-3">
            <Avatar className="h-10 w-10">
              <AvatarFallback className="bg-primary text-primary-foreground">
                {studentName.split(' ').map(n => n[0]).join('')}
              </AvatarFallback>
            </Avatar>
            <div>
              <p className="font-medium text-foreground">{studentName}</p>
              <p className="text-xs text-muted-foreground">Student</p>
            </div>
          </div>
          <Button 
            variant="ghost" 
            size="sm"
            onClick={handleLogout}
            className="gap-2"
          >
            <LogOut className="w-4 h-4" />
            Logout
          </Button>
        </div>
      </div>

      {/* Main Content */}
      <div className="flex-1 flex items-center justify-center p-4 md:p-8">
        <div className="w-full max-w-2xl mx-auto space-y-8">
          <div className="text-center space-y-3">
            <h1 className="text-3xl md:text-4xl font-semibold text-foreground">
              QR Scanner
            </h1>
            <p className="text-muted-foreground text-lg">
              Position the QR code within the frame to mark your attendance
            </p>
          </div>

          {/* Scanner Frame */}
          <Card className="p-6 md:p-10 shadow-xl">
            <div className="relative aspect-square max-w-md mx-auto bg-gradient-to-br from-slate-50 to-slate-100 rounded-3xl overflow-hidden border-4 border-dashed border-border">
              {/* Camera Icon Background */}
              <div className="absolute inset-0 flex items-center justify-center">
                <Camera className="w-24 h-24 text-muted-foreground/30" />
              </div>

              {/* Corner Markers */}
              <div className="absolute top-4 left-4 w-12 h-12 border-t-4 border-l-4 border-primary rounded-tl-2xl" />
              <div className="absolute top-4 right-4 w-12 h-12 border-t-4 border-r-4 border-primary rounded-tr-2xl" />
              <div className="absolute bottom-4 left-4 w-12 h-12 border-b-4 border-l-4 border-primary rounded-bl-2xl" />
              <div className="absolute bottom-4 right-4 w-12 h-12 border-b-4 border-r-4 border-primary rounded-br-2xl" />

              {/* Scanning Animation */}
              {scanStatus === 'scanning' && (
                <div className="absolute inset-0 flex items-center justify-center">
                  <div className="absolute w-full h-1 bg-primary/40 animate-pulse" 
                       style={{ 
                         animation: 'scan 2s ease-in-out infinite',
                       }} 
                  />
                </div>
              )}

              {/* Status Overlay */}
              {(scanStatus === 'success' || scanStatus === 'error') && (
                <div className={`absolute inset-0 flex flex-col items-center justify-center ${
                  scanStatus === 'success' ? 'bg-green-500/20' : 'bg-red-500/20'
                }`}>
                  {scanStatus === 'success' ? (
                    <>
                      <CheckCircle2 className="w-20 h-20 text-green-600 mb-4" />
                      <p className="text-xl font-semibold text-green-900">Attendance Marked!</p>
                      <p className="text-sm text-green-700 mt-2">March 25, 2026 • 9:30 AM</p>
                    </>
                  ) : (
                    <>
                      <XCircle className="w-20 h-20 text-red-600 mb-4" />
                      <p className="text-xl font-semibold text-red-900">Scan Failed</p>
                      <p className="text-sm text-red-700 mt-2">Please try again</p>
                    </>
                  )}
                </div>
              )}
            </div>

            {/* Instructions */}
            <div className="mt-8 space-y-4">
              {scanStatus === 'idle' && (
                <Button 
                  size="lg"
                  className="w-full h-14 text-lg rounded-xl"
                  onClick={handleScan}
                >
                  Start Scanning
                </Button>
              )}
              
              {scanStatus === 'scanning' && (
                <div className="text-center">
                  <div className="inline-flex items-center gap-2 text-primary">
                    <div className="w-2 h-2 bg-primary rounded-full animate-pulse" />
                    <p className="font-medium">Scanning...</p>
                  </div>
                </div>
              )}

              {(scanStatus === 'success' || scanStatus === 'error') && (
                <Button 
                  size="lg"
                  variant="outline"
                  className="w-full h-14 text-lg rounded-xl border-2"
                  onClick={handleScan}
                >
                  Scan Again
                </Button>
              )}

              <div className="bg-muted/50 rounded-xl p-4 space-y-2">
                <p className="text-sm font-medium text-foreground">Tips:</p>
                <ul className="text-sm text-muted-foreground space-y-1">
                  <li>• Ensure good lighting for best results</li>
                  <li>• Hold camera steady and within frame</li>
                  <li>• QR code should be clearly visible</li>
                </ul>
              </div>
            </div>
          </Card>
        </div>
      </div>

      <style>{`
        @keyframes scan {
          0%, 100% { transform: translateY(-100%); }
          50% { transform: translateY(100%); }
        }
      `}</style>
    </div>
  );
}
