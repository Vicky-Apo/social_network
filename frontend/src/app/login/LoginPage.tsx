"use client";

import { useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { ArrowRight, ShieldCheck } from "lucide-react";
import { useAuth } from "../component/AuthContext";
import { landingData } from "@/lib/data";
import { apiJson, asString, isRecord } from "@/lib/api";

type FormState = {
  email: string;
  password: string;
  remember: boolean;
};

type FormErrors = Partial<Record<keyof FormState | "submit", string>>;

export default function LoginPage() {
  const [formData, setFormData] = useState<FormState>({
    email: "",
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
    if (!data.email.trim()) {
      nextErrors.email = "Email is required.";
    } else if (!/^\S+@\S+\.\S+$/.test(data.email)) {
      nextErrors.email = "Enter a valid email.";
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

    const apiBaseUrl =
      process.env.NEXT_PUBLIC_API_BASE_URL?.trim().replace(/\/+$/, "") ||
      "http://localhost:8080";

    const payload = {
      email: formData.email.trim(),
      password: formData.password,
    };

    try {
      const response = await apiJson(apiBaseUrl, "/auth/login", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(payload),
      });

      if (!response.ok || !response.json?.success) {
        setErrors({
          submit: response.json?.error || "Login failed. Please try again.",
        });
        return;
      }

      // Backend sets an HttpOnly session cookie. Keep AuthContext for UI state only.
      const token = "session";
      const data = response.json?.data;
      const user = isRecord(data) && isRecord(data.user) ? data.user : null;
      login(
        token,
        {
          id: user && typeof user.id === "number" ? user.id : undefined,
          email: user ? asString(user.email) ?? undefined : undefined,
          firstName: user ? asString(user.first_name) ?? undefined : undefined,
          lastName: user ? asString(user.last_name) ?? undefined : undefined,
        },
        formData.remember,
      );
      router.push("/dashboard");
    } catch {
      setErrors({ submit: "Network error. Please try again." });
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="relative min-h-screen overflow-hidden bg-neutral-50 text-neutral-900">
      <div className="pointer-events-none absolute -left-32 top-8 h-80 w-80 rounded-full bg-indigo-200/35 blur-3xl" />
      <div className="pointer-events-none absolute -right-28 top-20 h-72 w-72 rounded-full bg-cyan-200/35 blur-3xl" />

      <main className="mx-auto grid min-h-screen w-full max-w-6xl items-center gap-8 px-4 py-24 sm:px-6 lg:grid-cols-[1fr_480px]">
        <section className="hidden rounded-[2rem] border border-neutral-200 bg-white/85 p-10 shadow-[0_40px_90px_-50px_rgba(2,6,23,0.45)] lg:block">
          <p className="inline-flex items-center gap-2 rounded-full border border-neutral-200 bg-white px-4 py-1.5 text-xs font-semibold uppercase tracking-[0.16em] text-neutral-600">
            <ShieldCheck className="h-3.5 w-3.5" />
            Secure Access
          </p>
          <h1 className="mt-5 text-4xl font-semibold tracking-tight text-neutral-900">
            Welcome back to {landingData.productName}
          </h1>
          <p className="mt-4 max-w-md text-sm leading-relaxed text-neutral-600">
            Sign in to continue your conversations, follow new threads, and manage your community space.
          </p>
          <div className="mt-8 grid grid-cols-2 gap-4">
            <div className="rounded-2xl border border-neutral-200 bg-neutral-50 p-4">
              <p className="text-2xl font-semibold tracking-tight text-neutral-900">99.9%</p>
              <p className="mt-1 text-xs text-neutral-600">Platform uptime</p>
            </div>
            <div className="rounded-2xl border border-neutral-200 bg-neutral-50 p-4">
              <p className="text-2xl font-semibold tracking-tight text-neutral-900">24/7</p>
              <p className="mt-1 text-xs text-neutral-600">Moderation visibility</p>
            </div>
          </div>
        </section>

        <section className="rounded-[2rem] border border-neutral-200 bg-white p-6 shadow-[0_35px_80px_-50px_rgba(2,6,23,0.45)] sm:p-8">
          <div>
            <p className="text-sm font-semibold text-neutral-500">{landingData.productName}</p>
            <h2 className="mt-2 text-3xl font-semibold tracking-tight text-neutral-900">Sign in</h2>
            <p className="mt-2 text-sm text-neutral-600">Access your dashboard and continue where you left off.</p>
          </div>

          <form className="mt-8 space-y-5" onSubmit={handleSubmit} noValidate>
            <div className="space-y-2">
              <label htmlFor="email" className="text-sm font-semibold text-neutral-700">
                Email
              </label>
              <input
                id="email"
                name="email"
                type="email"
                placeholder="you@example.com"
                autoComplete="email"
                className="h-12 w-full rounded-2xl border border-neutral-300 bg-white px-4 text-sm text-neutral-900 placeholder:text-neutral-400 focus:border-neutral-500 focus:outline-none focus-visible:ring-2 focus-visible:ring-neutral-900/30"
                value={formData.email}
                onChange={handleInputChange}
              />
              {errors.email ? <p className="text-xs text-rose-600">{errors.email}</p> : null}
            </div>

            <div className="space-y-2">
              <label htmlFor="password" className="text-sm font-semibold text-neutral-700">
                Password
              </label>
              <input
                id="password"
                name="password"
                type="password"
                placeholder="Enter your password"
                autoComplete="current-password"
                className="h-12 w-full rounded-2xl border border-neutral-300 bg-white px-4 text-sm text-neutral-900 placeholder:text-neutral-400 focus:border-neutral-500 focus:outline-none focus-visible:ring-2 focus-visible:ring-neutral-900/30"
                value={formData.password}
                onChange={handleInputChange}
              />
              {errors.password ? <p className="text-xs text-rose-600">{errors.password}</p> : null}
            </div>

            <div className="flex flex-wrap items-center justify-between gap-3">
              <label className="inline-flex items-center gap-2 text-sm text-neutral-600">
                <input
                  type="checkbox"
                  name="remember"
                  className="h-4 w-4 rounded border-neutral-300 text-neutral-900 focus:ring-neutral-900"
                  checked={formData.remember}
                  onChange={handleInputChange}
                />
                Remember me
              </label>
              <Link href="#" className="text-sm font-medium text-neutral-600 transition hover:text-neutral-900">
                Forgot password?
              </Link>
            </div>

            <button
              type="submit"
              className="brand-gradient group inline-flex h-12 w-full items-center justify-center gap-2 rounded-2xl px-5 text-sm font-semibold text-white shadow-sm transition hover:-translate-y-0.5 hover:shadow-lg focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-neutral-900/70 focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-70"
              disabled={isSubmitting}
            >
              <span>{isSubmitting ? "Signing in..." : "Sign in"}</span>
              <ArrowRight className="h-4 w-4 transition-transform duration-200 group-hover:translate-x-0.5" />
            </button>

            {errors.submit ? (
              <p className="rounded-2xl border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-700">
                {errors.submit}
              </p>
            ) : null}

            <p className="text-center text-sm text-neutral-600">
              Don&apos;t have an account?{" "}
              <Link href="/register" className="font-semibold text-neutral-900 hover:underline">
                Create one
              </Link>
            </p>
          </form>
        </section>
      </main>
    </div>
  );
}
