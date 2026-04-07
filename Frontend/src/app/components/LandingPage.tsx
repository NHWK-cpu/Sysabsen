import { QRCodeSVG } from 'qrcode.react';
import { useNavigate } from 'react-router';
import { Button } from './ui/button';
import { Card } from './ui/card';
import { ClipboardCheck, Shield } from 'lucide-react';

export function LandingPage() {
  const navigate = useNavigate();
  
  return (
    <div className="min-h-screen flex flex-col items-center justify-center p-4 md:p-8">
      <div className="w-full max-w-2xl mx-auto text-center space-y-8">
        {/* Logo */}
        <div className="flex items-center justify-center gap-3 mb-8">
          <div className="w-12 h-12 bg-primary rounded-xl flex items-center justify-center">
            <ClipboardCheck className="w-7 h-7 text-primary-foreground" />
          </div>
          <h1 className="text-3xl font-semibold text-foreground">AttendEase</h1>
        </div>

        {/* Main Content */}
        <div className="space-y-6">
          <div>
            <h2 className="text-4xl md:text-5xl font-semibold text-foreground mb-4">
              Scan QR for Attendance
            </h2>
            <p className="text-lg text-muted-foreground max-w-lg mx-auto">
              Quick and easy attendance tracking. Students scan to check in, teachers manage records.
            </p>
          </div>

          {/* QR Code Card */}
          <Card className="p-8 md:p-12 shadow-lg">
            <div className="bg-white p-6 rounded-2xl inline-block">
              <QRCodeSVG 
                value="https://attendease.app/scan/class-2024-03-25" 
                size={256}
                level="H"
                includeMargin={true}
              />
            </div>
            <p className="mt-6 text-sm text-muted-foreground">
              Scan this code with your camera to mark attendance
            </p>
          </Card>

          {/* Action Buttons */}
          <div className="flex flex-col sm:flex-row gap-4 justify-center mt-8">
            <Button
              size="lg"
              className="text-lg px-8 h-14 rounded-xl shadow-md hover:shadow-lg transition-all"
              onClick={() => navigate('/login/student')}
            >
              Student Login
            </Button>
            <Button
              size="lg"
              variant="outline"
              className="text-lg px-8 h-14 rounded-xl border-2 hover:bg-accent transition-all"
              onClick={() => navigate('/login/teacher')}
            >
              Teacher Login
            </Button>
          </div>

          {/* Admin Access Link */}
          <div className="mt-6">
            <Button
              variant="ghost"
              size="sm"
              onClick={() => navigate('/login/admin')}
              className="gap-2 text-muted-foreground hover:text-foreground"
            >
              <Shield className="w-4 h-4" />
              Admin Access
            </Button>
          </div>
        </div>

        {/* Footer */}
        <p className="text-sm text-muted-foreground mt-12">
          Secure and reliable attendance management system
        </p>
      </div>
    </div>
  );
}
