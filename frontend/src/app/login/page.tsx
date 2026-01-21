import Link from "next/link";
import Input from "../component/ui/input";
import Button from "../component/ui/button";

export default function LoginPage() {
  return (
    <div className="min-h-screen bg-white text-slate-900 flex items-center justify-center">
      <div className="w-full max-w-md px-6 py-16">
        <div className="space-y-6">
          <div className="text-center">
            <p className="text-2xl font-semibold tracking-tight text-slate-800">
              Social Network
            </p>
          </div>
          <div className="space-y-6">
            <h1 className="text-3xl font-semibold tracking-tight text-center">
              Log In
            </h1>

            <form className="space-y-4 text-left">
              <div className="space-y-2">
                <label htmlFor="email" className="text-sm font-medium block">
                  Email
                </label>
                <Input
                  id="email"
                  type="email"
                  placeholder="Enter your email"
                  autoComplete="email"
                  className="w-full h-10 rounded-md border-slate-300 text-sm focus:ring-1 focus:ring-blue-500"
                />
              </div>

              <div className="space-y-2">
                <label htmlFor="password" className="text-sm font-medium block">
                  Password
                </label>
                <Input
                  id="password"
                  type="password"
                  placeholder="Enter your password"
                  autoComplete="current-password"
                  className="w-full h-10 rounded-md border-slate-300 text-sm focus:ring-1 focus:ring-blue-500"
                />
              </div>

              <label className="flex items-center gap-2 text-sm text-slate-700">
                <input
                  type="checkbox"
                  className="h-4 w-4 rounded border-slate-300"
                />
                Remember Me
              </label>

              <Button
                type="submit"
                className="w-full rounded-md bg-gradient-to-b from-blue-500 to-blue-700 py-2.5 text-sm shadow-sm hover:from-blue-600 hover:to-blue-800"
              >
                Log In
              </Button>

              <Link
                href="#"
                className="block text-sm text-blue-600 hover:text-blue-700"
              >
                Forgot Password?
              </Link>
            </form>

            <p className="text-sm text-slate-600 mt-6">
              Don&apos;t have an account?{" "}
              <Link href="/register" className="text-blue-600 hover:underline">
                Sign Up
              </Link>
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}
