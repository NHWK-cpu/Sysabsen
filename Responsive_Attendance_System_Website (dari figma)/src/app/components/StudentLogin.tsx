import { useState } from 'react';
import { useNavigate } from 'react-router';
import { Button } from './ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from './ui/card';
import { Input } from './ui/input';
import { Label } from './ui/label';
import { ClipboardCheck, ArrowLeft } from 'lucide-react';

export function StudentLogin() {
  const navigate = useNavigate();
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');

  const handleLogin = (e: React.FormEvent) => {
    e.preventDefault();
    // Mock login - in production, this would authenticate with a backend
    setTimeout(() => {
      navigate('/student/scanner');
    }, 500);
  };

  return (
    <div className="min-h-screen flex flex-col items-center justify-center p-4 md:p-8">
      <div className="w-full max-w-md mx-auto">
        {/* Back Button */}
        <Button 
          variant="ghost" 
          className="mb-6"
          onClick={() => navigate('/')}
        >
          <ArrowLeft className="w-4 h-4 mr-2" />
          Back
        </Button>

        {/* Logo */}
        <div className="flex items-center justify-center gap-3 mb-8">
          <div className="w-12 h-12 bg-primary rounded-xl flex items-center justify-center">
            <ClipboardCheck className="w-7 h-7 text-primary-foreground" />
          </div>
          <h1 className="text-2xl font-semibold text-foreground">AttendEase</h1>
        </div>

        {/* Login Card */}
        <Card className="shadow-xl">
          <CardHeader className="space-y-2 text-center">
            <CardTitle className="text-2xl">Student Login</CardTitle>
            <CardDescription className="text-base">
              Sign in to mark your attendance
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-6">
            <form onSubmit={handleLogin} className="space-y-5">
              <div className="space-y-2">
                <Label htmlFor="username">Username</Label>
                <Input
                  id="username"
                  type="text"
                  placeholder="username"
                  value={username}
                  onChange={(e) => setUsername(e.target.value)}
                  className="h-11 rounded-lg"
                  required
                />
              </div>

              <div className="space-y-2">
                <div className="flex items-center justify-between">
                  <Label htmlFor="password">Password</Label>
                </div>
                <Input
                  id="password"
                  type="password"
                  placeholder="••••••••"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  className="h-11 rounded-lg"
                  required
                />
              </div>

              <Button
                type="submit"
                size="lg"
                className="w-full h-12 text-base rounded-xl"
              >
                Login
              </Button>
            </form>

            <p className="text-sm text-center text-muted-foreground px-4">
              By logging in, you agree to our Terms of Service and Privacy Policy
            </p>
          </CardContent>
        </Card>

        <p className="text-center text-sm text-muted-foreground mt-6">
          Are you a teacher?{' '}
          <button 
            className="text-primary hover:underline font-medium"
            onClick={() => navigate('/login/teacher')}
          >
            Login here
          </button>
        </p>
      </div>
    </div>
  );
}
