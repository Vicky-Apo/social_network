"use client";

import { useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { ArrowRight } from "lucide-react";
import { motion } from "framer-motion";
import { useAuth } from "@/components/AuthContext";
import { landingData } from "@/lib/data";
import { fadeUp, staggerContainer } from "@/components/Motion";
import { apiFetchJson, getApiBaseUrl } from "@/lib/api";
import { ApiResponse } from "@/lib/types";

type FormState = {
  email: string;
  password: string;
  remember: boolean;
};

type FormErrors = Partial<Record<keyof FormState | "submit", string>>;
type LoginResponse = {
  token?: string;
  user?: {
    id?: number;
    email?: string;
    first_name?: string;
    last_name?: string;
  };
};

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

    const apiBaseUrl = getApiBaseUrl();
    const endpoint = `${apiBaseUrl}/auth/login`;

    const payload = {
      email: formData.email.trim(),
      password: formData.password,
    };

    try {
      const { response, result } = await apiFetchJson<ApiResponse<LoginResponse>>(
        endpoint,
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify(payload),
        },
        apiBaseUrl,
      );
      if (!response.ok || !result?.success) {
        setErrors({
          submit: result?.error || "Login failed. Please try again.",
        });
        return;
      }

      const token = result?.data?.token;
      if (!token) {
        setErrors({ submit: "Login failed. Missing token in response." });
        return;
      }

      const user = {
        id: result?.data?.user?.id,
        email: result?.data?.user?.email,
        firstName: result?.data?.user?.first_name,
        lastName: result?.data?.user?.last_name,
      };

      login(token, user, formData.remember);
      router.push("/dashboard");
    } catch {
      setErrors({ submit: "Network error. Please try again." });
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div
      className="relative min-h-screen overflow-hidden bg-[#2b2929] text-neutral-100"
      style={{
        backgroundImage: "url(/login-bg.png)",
        backgroundSize: "150%",
        backgroundPosition: "center",
      }}
    >
      <main className="mx-auto flex min-h-screen w-full max-w-6xl items-center justify-center px-4 py-24 sm:px-6">
        <motion.section
          variants={staggerContainer}
          initial="hidden"
          animate="show"
          className="w-full max-w-md rounded-2xl border border-white/10 bg-[#2b2929]/80 p-6 shadow-2xl backdrop-blur-xl sm:p-8"
        >
          <motion.div variants={fadeUp}>
            <p className="text-sm font-semibold text-neutral-400">{landingData.productName}</p>
            <h2 className="mt-2 text-3xl font-semibold tracking-tight text-white">Sign in</h2>
            <p className="mt-2 text-sm text-neutral-400">Access your dashboard and continue where you left off.</p>
          </motion.div>

          <motion.form variants={fadeUp} className="mt-8 space-y-5" onSubmit={handleSubmit} noValidate>
            <div className="space-y-2">
              <label htmlFor="email" className="text-sm font-semibold text-neutral-300">
                Email
              </label>
              <input
                id="email"
                name="email"
                type="email"
                placeholder="you@example.com"
                autoComplete="email"
                className="h-12 w-full rounded-xl border border-white/20 bg-white/5 px-4 text-sm text-white placeholder:text-neutral-500 focus:border-white/40 focus:outline-none focus-visible:ring-2 focus-visible:ring-white/30"
                value={formData.email}
                onChange={handleInputChange}
              />
              {errors.email ? <p className="text-xs text-rose-400">{errors.email}</p> : null}
            </div>

            <div className="space-y-2">
              <label htmlFor="password" className="text-sm font-semibold text-neutral-300">
                Password
              </label>
              <input
                id="password"
                name="password"
                type="password"
                placeholder="Enter your password"
                autoComplete="current-password"
                className="h-12 w-full rounded-xl border border-white/20 bg-white/5 px-4 text-sm text-white placeholder:text-neutral-500 focus:border-white/40 focus:outline-none focus-visible:ring-2 focus-visible:ring-white/30"
                value={formData.password}
                onChange={handleInputChange}
              />
              {errors.password ? <p className="text-xs text-rose-400">{errors.password}</p> : null}
            </div>

            <div className="flex flex-wrap items-center justify-between gap-3">
              <label className="inline-flex items-center gap-2 text-sm text-neutral-400">
                <input
                  type="checkbox"
                  name="remember"
                  className="h-4 w-4 rounded border-white/30 bg-white/5 text-white focus:ring-white/30"
                  checked={formData.remember}
                  onChange={handleInputChange}
                />
                Remember me
              </label>
              <Link href="#" className="text-sm font-medium text-neutral-400 transition hover:text-white">
                Forgot password?
              </Link>
            </div>

            <button
              type="submit"
              className="group inline-flex h-12 w-full items-center justify-center gap-2 rounded-xl bg-white px-5 text-sm font-semibold text-black shadow-sm transition hover:-translate-y-0.5 hover:bg-neutral-100 hover:shadow-lg focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-white/50 focus-visible:ring-offset-2 focus-visible:ring-offset-[#2b2929] disabled:cursor-not-allowed disabled:opacity-70"
              disabled={isSubmitting}
            >
              <span>{isSubmitting ? "Signing in..." : "Sign in"}</span>
              <ArrowRight className="h-4 w-4 transition-transform duration-200 group-hover:translate-x-0.5" />
            </button>

            {errors.submit ? (
              <p className="rounded-xl border border-rose-500/30 bg-rose-500/10 px-4 py-3 text-sm text-rose-400">
                {errors.submit}
              </p>
            ) : null}

            <p className="text-center text-sm text-neutral-400">
              Don&apos;t have an account?{" "}
              <Link href="/register" className="font-semibold text-white hover:underline">
                Create one
              </Link>
            </p>
          </motion.form>
        </motion.section>
      </main>
    </div>
  );
}
