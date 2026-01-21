"use client";

import { useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import Input from "../component/ui/input";
import Button from "../component/ui/button";
import { useAuth } from "../component/AuthContext";

type FormState = {
  identifier: string;
  password: string;
  remember: boolean;
};

type FormErrors = Partial<Record<keyof FormState | "submit", string>>;

export default function LoginPage() {
  const [formData, setFormData] = useState<FormState>({
    identifier: "",
    password: "",
    remember: false,
  });
  const [errors, setErrors] = useState<FormErrors>({});
  const [isSubmitting, setIsSubmitting] = useState(false);
  const { login } = useAuth();
  const router = useRouter();

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value, type, checked } = e.target;
    const nextValue = type === "checkbox" ? checked : value;
    setFormData((prev) => ({ ...prev, [name]: nextValue }));
    if (errors[name as keyof FormState]) {
      setErrors((prev) => ({ ...prev, [name]: undefined }));
    }
  };

  const validateForm = (data: FormState) => {
    const nextErrors: FormErrors = {};
    if (!data.identifier.trim()) {
      nextErrors.identifier = "Email or username is required.";
    }
    if (!data.password) {
      nextErrors.password = "Password is required.";
    }
    return nextErrors;
  };

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (isSubmitting) return;

    const nextErrors = validateForm(formData);
    if (Object.keys(nextErrors).length > 0) {
      setErrors(nextErrors);
      return;
    }

    setIsSubmitting(true);
    setErrors({});

    const identifier = formData.identifier.trim();
    const apiBaseUrl = process.env.NEXT_PUBLIC_API_BASE_URL || "";
    const endpoint = apiBaseUrl ? `${apiBaseUrl}/auth/login` : "/api/auth/login";

    const payload: Record<string, string> = {
      identifier,
      password: formData.password,
    };

    if (identifier.includes("@")) {
      payload.email = identifier;
    } else {
      payload.username = identifier;
    }

    try {
      const response = await fetch(endpoint, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(payload),
      });

      const data = await response.json().catch(() => ({}));
      if (!response.ok) {
        setErrors({
          submit: data?.message || "Login failed. Please try again.",
        });
        return;
      }

      const token = data?.token || data?.accessToken;
      if (!token) {
        setErrors({ submit: "Login failed. Missing token in response." });
        return;
      }

      const user = {
        username: data?.username || data?.user?.username,
        email: data?.email || data?.user?.email,
      };

      login(token, user, formData.remember);
      router.push("/");
    } catch {
      setErrors({ submit: "Network error. Please try again." });
    } finally {
      setIsSubmitting(false);
    }
  };

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

            <form className="space-y-4 text-left" onSubmit={handleSubmit} noValidate>
              <div className="space-y-2">
                <label htmlFor="identifier" className="text-sm font-medium block">
                  Email or Username
                </label>
                <Input
                  id="identifier"
                  name="identifier"
                  type="text"
                  placeholder="Enter your email or username"
                  autoComplete="username"
                  className="w-full h-10 rounded-md border-slate-300 text-sm focus:ring-1 focus:ring-blue-500"
                  value={formData.identifier}
                  onChange={handleInputChange}
                />
                {errors.identifier ? (
                  <p className="text-xs text-red-600">{errors.identifier}</p>
                ) : null}
              </div>

              <div className="space-y-2">
                <label htmlFor="password" className="text-sm font-medium block">
                  Password
                </label>
                <Input
                  id="password"
                  name="password"
                  type="password"
                  placeholder="Enter your password"
                  autoComplete="current-password"
                  className="w-full h-10 rounded-md border-slate-300 text-sm focus:ring-1 focus:ring-blue-500"
                  value={formData.password}
                  onChange={handleInputChange}
                />
                {errors.password ? (
                  <p className="text-xs text-red-600">{errors.password}</p>
                ) : null}
              </div>

              <label className="flex items-center gap-2 text-sm text-slate-700">
                <input
                  type="checkbox"
                  name="remember"
                  className="h-4 w-4 rounded border-slate-300"
                  checked={formData.remember}
                  onChange={handleInputChange}
                />
                Remember Me
              </label>

              <Button
                type="submit"
                className="w-full rounded-md bg-gradient-to-b from-blue-500 to-blue-700 py-2.5 text-sm shadow-sm hover:from-blue-600 hover:to-blue-800"
                disabled={isSubmitting}
              >
                {isSubmitting ? "Logging In..." : "Log In"}
              </Button>
              {errors.submit ? (
                <p className="text-sm text-red-600">{errors.submit}</p>
              ) : null}

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
