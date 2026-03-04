"use client";

import { useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { ArrowRight, Sparkles } from "lucide-react";
import { motion } from "framer-motion";
import { landingData } from "@/lib/data";
import { fadeUp, staggerContainer } from "@/components/Motion";
import { apiFetchJson, getApiBaseUrl } from "@/lib/api";
import { ApiResponse } from "@/lib/types";

type FormState = {
  firstName: string;
  lastName: string;
  email: string;
  password: string;
  confirmPassword: string;
  dob: string;
  nickname: string;
  about: string;
};

type FormErrors = Partial<Record<keyof FormState | "submit", string>>;
const initialFormData: FormState = {
  firstName: "",
  lastName: "",
  email: "",
  password: "",
  confirmPassword: "",
  dob: "",
  nickname: "",
  about: "",
};

export default function RegisterPage() {
  const [formData, setFormData] = useState<FormState>(initialFormData);
  const [errors, setErrors] = useState<FormErrors>({});
  const [avatarFile, setAvatarFile] = useState<File | null>(null);
  const [avatarWarning, setAvatarWarning] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isSuccess, setIsSuccess] = useState(false);
  const router = useRouter();

  const handleInputChange = (
    e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>,
  ) => {
    const { id, value } = e.target;
    const field = id as keyof FormState;
    setFormData((prev) => ({ ...prev, [field]: value }));
    setIsSuccess(false);
    if (errors[field]) {
      setErrors((prev) => ({ ...prev, [field]: undefined }));
    }
  };

  const handleAvatarChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0] ?? null;
    setAvatarFile(file);
    setIsSuccess(false);
    setAvatarWarning(null);
    if (errors.submit) {
      setErrors((prev) => ({ ...prev, submit: undefined }));
    }
  };

  const formatDateOfBirth = (value: string) => {
    const [year, month, day] = value.split("-");
    if (!year || !month || !day) return "";
    return `${day}/${month}/${year}`;
  };

  const validateForm = (data: FormState) => {
    const nextErrors: FormErrors = {};

    if (!data.firstName.trim()) {
      nextErrors.firstName = "First name is required.";
    }
    if (!data.lastName.trim()) {
      nextErrors.lastName = "Last name is required.";
    }
    if (!data.email.trim()) {
      nextErrors.email = "Email is required.";
    } else if (!/^\S+@\S+\.\S+$/.test(data.email)) {
      nextErrors.email = "Enter a valid email.";
    }
    if (!data.password) {
      nextErrors.password = "Password is required.";
    } else if (data.password.length < 8) {
      nextErrors.password = "Password must be at least 8 characters.";
    }
    if (!data.confirmPassword) {
      nextErrors.confirmPassword = "Please confirm your password.";
    } else if (data.password !== data.confirmPassword) {
      nextErrors.confirmPassword = "Passwords do not match.";
    }
    if (!data.dob) {
      nextErrors.dob = "Date of birth is required.";
    }

    return nextErrors;
  };

  const uploadAvatarAfterRegister = async (email: string, password: string, file: File) => {
    const apiBaseUrl = getApiBaseUrl();
    let loggedIn = false;

    try {
      const { response: loginRes, result: loginJson } = await apiFetchJson<
        ApiResponse<{ user?: { id?: number } }>
      >(
        `${apiBaseUrl}/auth/login`,
        {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ email, password }),
        },
        apiBaseUrl,
      );
      if (!loginRes.ok || !loginJson?.success) {
        throw new Error("Account created, but avatar upload failed. Please login and add an avatar.");
      }
      loggedIn = true;

      let userId = loginJson?.data?.user?.id;
      if (!userId) {
        const { response: meRes, result: meJson } = await apiFetchJson<ApiResponse<{ id?: number }>>(
          "/auth/me",
          {},
          apiBaseUrl,
        );
        if (!meRes.ok || !meJson?.success || !meJson.data?.id) {
          throw new Error("Account created, but avatar upload failed. Please login and add an avatar.");
        }
        userId = meJson.data.id;
      }

      const formData = new FormData();
      formData.append("file", file);
      formData.append("kind", "avatar");
      const { response: uploadRes, result: uploadJson } = await apiFetchJson<
        ApiResponse<{ path?: string }>
      >(
        "/uploads",
        {
          method: "POST",
          body: formData,
        },
        apiBaseUrl,
      );
      if (!uploadRes.ok || !uploadJson?.success || !uploadJson.data?.path) {
        throw new Error("Account created, but avatar upload failed. Please login and add an avatar.");
      }

      const { response: patchRes, result: patchJson } = await apiFetchJson<ApiResponse<unknown>>(
        `/profiles/${userId}`,
        {
          method: "PATCH",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ avatar_path: uploadJson.data.path }),
        },
        apiBaseUrl,
      );
      if (!patchRes.ok || !patchJson?.success) {
        throw new Error("Account created, but avatar upload failed. Please login and add an avatar.");
      }
    } finally {
      if (loggedIn) {
        try {
          await apiFetchJson<ApiResponse<unknown>>(
            "/auth/logout",
            { method: "POST" },
            apiBaseUrl,
          );
        } catch {
          // Best-effort logout.
        }
      }
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (isSubmitting) return;

    const nextErrors = validateForm(formData);
    if (Object.keys(nextErrors).length > 0) {
      setErrors(nextErrors);
      return;
    }

    setIsSubmitting(true);
    setErrors({});
    setAvatarWarning(null);

    const apiBaseUrl = getApiBaseUrl();
    const endpoint = `${apiBaseUrl}/auth/register`;

    const dateOfBirth = formatDateOfBirth(formData.dob);
    const payload = {
      email: formData.email.trim(),
      password: formData.password,
      first_name: formData.firstName.trim(),
      last_name: formData.lastName.trim(),
      date_of_birth: dateOfBirth,
      ...(formData.nickname.trim()
        ? { nickname: formData.nickname.trim() }
        : {}),
      ...(formData.about.trim() ? { about: formData.about.trim() } : {}),
    };

    try {
      const { response, result } = await apiFetchJson<ApiResponse<unknown>>(
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
          submit: result?.error || "Registration failed. Please try again.",
        });
        return;
      }

      if (avatarFile) {
        try {
          await uploadAvatarAfterRegister(formData.email.trim(), formData.password, avatarFile);
        } catch (err) {
          setAvatarWarning(
            err instanceof Error
              ? err.message
              : "Account created, but avatar upload failed. Please login and add an avatar.",
          );
        }
      }

      setFormData(initialFormData);
      setAvatarFile(null);
      setIsSuccess(true);
      setTimeout(() => {
        router.push("/login");
      }, 700);
    } catch {
      setErrors({
        submit: "Network error. Please try again.",
      });
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div
      className="relative min-h-screen overflow-hidden bg-[#2b2929] text-neutral-100"
      style={{
        backgroundImage: "url(/register-bg.png)",
        backgroundSize: "100%",
        backgroundPosition: "center",
      }}
    >
      <main className="mx-auto w-full max-w-3xl px-4 py-20 sm:px-6">
        <motion.section
          variants={staggerContainer}
          initial="hidden"
          animate="show"
          className="rounded-2xl border border-white/10 bg-[#2b2929]/80 p-6 shadow-2xl backdrop-blur-xl sm:p-8"
        >
          <motion.p
            variants={fadeUp}
            className="inline-flex items-center gap-2 rounded-full border border-white/20 bg-white/5 px-4 py-1.5 text-xs font-semibold uppercase tracking-[0.16em] text-neutral-400"
          >
            <Sparkles className="h-3.5 w-3.5" />
            Create account
          </motion.p>
          <motion.h1 variants={fadeUp} className="mt-4 text-3xl font-semibold tracking-tight text-white">
            Join {landingData.productName}
          </motion.h1>
          <motion.p variants={fadeUp} className="mt-2 text-sm text-neutral-400">
            Set up your profile and start participating in high-quality discussions.
          </motion.p>

          <motion.form variants={fadeUp} className="mt-8 space-y-5" onSubmit={handleSubmit} noValidate>
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <Field
                id="firstName"
                label="First name"
                value={formData.firstName}
                onChange={handleInputChange}
                error={errors.firstName}
                placeholder="First name"
              />
              <Field
                id="lastName"
                label="Last name"
                value={formData.lastName}
                onChange={handleInputChange}
                error={errors.lastName}
                placeholder="Last name"
              />
            </div>

            <Field
              id="email"
              label="Email"
              type="email"
              value={formData.email}
              onChange={handleInputChange}
              error={errors.email}
              placeholder="you@example.com"
              autoComplete="email"
            />

            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <Field
                id="password"
                label="Password"
                type="password"
                value={formData.password}
                onChange={handleInputChange}
                error={errors.password}
                placeholder="At least 8 characters"
                autoComplete="new-password"
              />
              <Field
                id="confirmPassword"
                label="Confirm password"
                type="password"
                value={formData.confirmPassword}
                onChange={handleInputChange}
                error={errors.confirmPassword}
                placeholder="Repeat password"
                autoComplete="new-password"
              />
            </div>

            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <Field
                id="dob"
                label="Date of birth"
                type="date"
                value={formData.dob}
                onChange={handleInputChange}
                error={errors.dob}
              />
              <Field
                id="nickname"
                label="Nickname (optional)"
                value={formData.nickname}
                onChange={handleInputChange}
                placeholder="Preferred display name"
              />
            </div>

            <div className="space-y-2">
              <label htmlFor="about" className="text-sm font-semibold text-neutral-300">
                About (optional)
              </label>
              <textarea
                id="about"
                value={formData.about}
                onChange={handleInputChange}
                rows={3}
                placeholder="Tell the community about your interests..."
                className="w-full resize-none rounded-xl border border-white/20 bg-white/5 px-4 py-3 text-sm text-white placeholder:text-neutral-500 focus:border-white/40 focus:outline-none focus-visible:ring-2 focus-visible:ring-white/30"
              />
            </div>

            <div className="space-y-2">
              <label htmlFor="avatar" className="text-sm font-semibold text-neutral-300">
                Avatar (optional)
              </label>
              <input
                id="avatar"
                type="file"
                accept="image/*"
                onChange={handleAvatarChange}
                className="h-12 w-full rounded-xl border border-white/20 bg-white/5 px-4 text-sm text-white file:mr-3 file:rounded-lg file:border-0 file:bg-white/10 file:px-3 file:py-2 file:text-xs file:font-semibold file:text-neutral-200 hover:file:bg-white/20"
              />
              <p className="text-xs text-neutral-500">
                Uploaded after registration completes.
              </p>
            </div>

            <button
              type="submit"
              className="group inline-flex h-12 w-full items-center justify-center gap-2 rounded-xl bg-white px-5 text-sm font-semibold text-black shadow-sm transition hover:-translate-y-0.5 hover:bg-neutral-100 hover:shadow-lg focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-white/50 focus-visible:ring-offset-2 focus-visible:ring-offset-[#2b2929] disabled:cursor-not-allowed disabled:opacity-70"
              disabled={isSubmitting}
            >
              <span>{isSubmitting ? "Creating account..." : "Create account"}</span>
              <ArrowRight className="h-4 w-4 transition-transform duration-200 group-hover:translate-x-0.5" />
            </button>

            {errors.submit ? (
              <p className="rounded-xl border border-rose-500/30 bg-rose-500/10 px-4 py-3 text-sm text-rose-400">
                {errors.submit}
              </p>
            ) : null}
            {avatarWarning ? (
              <p className="rounded-xl border border-amber-500/30 bg-amber-500/10 px-4 py-3 text-sm text-amber-300">
                {avatarWarning}
              </p>
            ) : null}
            {isSuccess ? (
              <p className="rounded-xl border border-emerald-500/30 bg-emerald-500/10 px-4 py-3 text-sm text-emerald-400">
                Account created successfully. Redirecting to login...
              </p>
            ) : null}
          </motion.form>

          <motion.p variants={fadeUp} className="mt-6 text-center text-sm text-neutral-400">
            Already have an account?{" "}
            <Link href="/login" className="font-semibold text-white hover:underline">
              Sign in
            </Link>
          </motion.p>
        </motion.section>
      </main>
    </div>
  );
}

function Field({
  id,
  label,
  type = "text",
  placeholder,
  value,
  onChange,
  error,
  autoComplete,
}: {
  id: string;
  label: string;
  type?: string;
  placeholder?: string;
  value: string;
  onChange: (event: React.ChangeEvent<HTMLInputElement>) => void;
  error?: string;
  autoComplete?: string;
}) {
  return (
    <div className="space-y-2">
      <label htmlFor={id} className="text-sm font-semibold text-neutral-300">
        {label}
      </label>
      <input
        id={id}
        type={type}
        value={value}
        onChange={onChange}
        placeholder={placeholder}
        autoComplete={autoComplete}
        className="h-12 w-full rounded-xl border border-white/20 bg-white/5 px-4 text-sm text-white placeholder:text-neutral-500 focus:border-white/40 focus:outline-none focus-visible:ring-2 focus-visible:ring-white/30"
      />
      {error ? <p className="text-xs text-rose-400">{error}</p> : null}
    </div>
  );
}
