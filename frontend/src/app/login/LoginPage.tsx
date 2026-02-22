"use client";

import { useState } from "react";
import Link from "next/link";
import Image from "next/image";
import { useRouter } from "next/navigation";
import { useAuth } from "../component/AuthContext";
import { Navbar } from "@/components/Navbar";
import { Footer } from "@/components/Footer";
import { Button3D } from "@/components/Button3D";
import { landingData } from "@/lib/data";
import { apiJson, asString, isRecord } from "@/lib/api";

type FormState = {
  email: string;
  password: string;
  remember: boolean;
};

type FormErrors = Partial<Record<keyof FormState | "submit", string>>;

const inputClass =
  "h-12 w-full rounded-sm border border-white/30 bg-white/5 px-4 text-sm text-white placeholder:text-white/50 transition focus:border-white/60 focus:outline-none focus:ring-2 focus:ring-white/30 focus:ring-offset-2 focus:ring-offset-[#2b2929]";

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
    <div className="min-h-screen bg-[#2b2929] text-neutral-900">
      <Navbar />

      <main className="relative mx-auto flex min-h-[calc(100vh-5rem)] w-full max-w-5xl flex-col items-center justify-center overflow-hidden px-4 py-24 sm:px-6 sm:py-28">
        <Image
          src="/gradient-2.png"
          alt=""
          fill
          className="object-cover object-center opacity-90"
          sizes="100vw"
        />
        <div className="absolute inset-0 bg-[#2b2929]/70" />
        <section className="relative z-10 w-full max-w-md rounded-sm border border-white/10 bg-[#2b2929]/40 p-8 backdrop-blur-sm sm:p-10">
          <header className="text-center">
            <Link
              href="/"
              className="mx-auto flex w-full justify-center focus:outline-none focus-visible:ring-2 focus-visible:ring-white/50 focus-visible:ring-offset-2 focus-visible:ring-offset-[#2b2929] rounded-sm"
            >
              <Image
                src="/vybez-logo-v2.png"
                alt={landingData.productName}
                width={200}
                height={80}
                className="h-14 w-auto sm:h-16"
              />
            </Link>
            <h1 className="mt-8 text-2xl font-semibold tracking-tight text-white sm:text-3xl">
              Welcome back
            </h1>
            <p className="mt-2 text-sm text-white/70">
              Sign in to your account to continue.
            </p>
          </header>

          <form className="mt-8 space-y-5" onSubmit={handleSubmit} noValidate>
            <div className="space-y-2">
              <label htmlFor="email" className="text-sm font-medium text-white/90">
                Email
              </label>
              <input
                id="email"
                name="email"
                type="email"
                placeholder="you@example.com"
                autoComplete="email"
                className={inputClass}
                value={formData.email}
                onChange={handleInputChange}
              />
              {errors.email ? (
                <p className="text-xs text-rose-400">{errors.email}</p>
              ) : null}
            </div>

            <div className="space-y-2">
              <label htmlFor="password" className="text-sm font-medium text-white/90">
                Password
              </label>
              <input
                id="password"
                name="password"
                type="password"
                placeholder="Enter your password"
                autoComplete="current-password"
                className={inputClass}
                value={formData.password}
                onChange={handleInputChange}
              />
              <div className="flex justify-end">
                <Link
                  href="#"
                  className="text-sm font-medium text-white/90 transition hover:text-white"
                >
                  Forgot password?
                </Link>
              </div>
              {errors.password ? (
                <p className="text-xs text-rose-400">{errors.password}</p>
              ) : null}
            </div>

            <label className="flex cursor-pointer items-center gap-3">
              <input
                type="checkbox"
                name="remember"
                className="h-4 w-4 rounded border-white/40 bg-white/5 text-[#2b2929] focus:ring-white/50"
                checked={formData.remember}
                onChange={handleInputChange}
              />
              <span className="text-sm text-white/70">Remember me</span>
            </label>

            {errors.submit ? (
              <p className="rounded-sm border border-rose-500/50 bg-rose-500/10 px-4 py-3 text-sm text-rose-300">
                {errors.submit}
              </p>
            ) : null}

            <button
              type="submit"
              disabled={isSubmitting}
              className="btn-3d-outer w-full disabled:pointer-events-none disabled:opacity-60"
            >
              <span className="btn-3d-inner flex h-12 w-full min-w-0 items-center justify-center bg-white px-6 font-semibold text-neutral-900">
                {isSubmitting ? "Signing in…" : "Sign in"}
              </span>
            </button>

            <p className="text-center text-sm text-white/70">
              Don&apos;t have an account?{" "}
              <Link
                href="/register"
                className="font-semibold text-white underline-offset-2 hover:underline"
              >
                Create one
              </Link>
            </p>
            <div className="flex justify-center [&_.btn-3d-inner]:bg-white [&_.btn-3d-inner]:text-neutral-900">
              <Button3D href={landingData.ctaUrl}>
                {landingData.ctaPrimary}
              </Button3D>
            </div>
          </form>
        </section>
      </main>

      <Footer />
    </div>
  );
}
