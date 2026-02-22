"use client";

import { useState } from "react";
import Link from "next/link";
import Image from "next/image";
import { useRouter } from "next/navigation";
import { Navbar } from "@/components/Navbar";
import { Footer } from "@/components/Footer";
import { landingData } from "@/lib/data";
import { apiJson } from "@/lib/api";

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

const inputClass =
  "h-12 w-full rounded-sm border border-white/30 bg-white/5 px-4 text-sm text-white placeholder:text-white/50 transition focus:border-white/60 focus:outline-none focus:ring-2 focus:ring-white/30 focus:ring-offset-2 focus:ring-offset-[#2b2929]";

const textareaClass =
  "w-full resize-none rounded-sm border border-white/30 bg-white/5 px-4 py-3 text-sm text-white placeholder:text-white/50 transition focus:border-white/60 focus:outline-none focus:ring-2 focus:ring-white/30 focus:ring-offset-2 focus:ring-offset-[#2b2929]";

export default function RegisterPage() {
  const [formData, setFormData] = useState<FormState>(initialFormData);
  const [errors, setErrors] = useState<FormErrors>({});
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

  const formatDateOfBirth = (value: string) => {
    const [year, month, day] = value.split("-");
    if (!year || !month || !day) return "";
    return `${day}/${month}/${year}`;
  };

  const validateForm = (data: FormState) => {
    const nextErrors: FormErrors = {};

    if (!data.firstName.trim()) nextErrors.firstName = "First name is required.";
    if (!data.lastName.trim()) nextErrors.lastName = "Last name is required.";
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
    if (!data.dob) nextErrors.dob = "Date of birth is required.";

    return nextErrors;
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

    const apiBaseUrl =
      process.env.NEXT_PUBLIC_API_BASE_URL?.trim().replace(/\/+$/, "") ||
      "http://localhost:8080";

    const dateOfBirth = formatDateOfBirth(formData.dob);
    const payload = {
      email: formData.email.trim(),
      password: formData.password,
      first_name: formData.firstName.trim(),
      last_name: formData.lastName.trim(),
      date_of_birth: dateOfBirth,
      ...(formData.nickname.trim() ? { nickname: formData.nickname.trim() } : {}),
      ...(formData.about.trim() ? { about: formData.about.trim() } : {}),
    };

    try {
      const response = await apiJson(apiBaseUrl, "/auth/register", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });

      if (!response.ok || !response.json?.success) {
        setErrors({
          submit: response.json?.error || "Registration failed. Please try again.",
        });
        return;
      }

      setFormData(initialFormData);
      setIsSuccess(true);
      setTimeout(() => router.push("/login"), 700);
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
          src="/register-bg.png"
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
              className="mx-auto flex w-full justify-center rounded-sm focus:outline-none focus-visible:ring-2 focus-visible:ring-white/50 focus-visible:ring-offset-2 focus-visible:ring-offset-[#2b2929]"
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
              Join {landingData.productName}
            </h1>
            <p className="mt-2 text-sm text-white/70">
              Set up your profile and start connecting.
            </p>
          </header>

          <form className="mt-8 space-y-5" onSubmit={handleSubmit} noValidate>
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
                placeholder="Display name"
              />
            </div>

            <div className="space-y-2">
              <label htmlFor="about" className="text-sm font-medium text-white/90">
                About (optional)
              </label>
              <textarea
                id="about"
                value={formData.about}
                onChange={handleInputChange}
                rows={3}
                placeholder="Tell the community about your interests..."
                className={textareaClass}
              />
            </div>

            <button
              type="submit"
              disabled={isSubmitting}
              className="btn-3d-outer w-full disabled:pointer-events-none disabled:opacity-60"
            >
              <span className="btn-3d-inner flex h-12 w-full min-w-0 items-center justify-center bg-white px-6 font-semibold text-neutral-900">
                {isSubmitting ? "Creating account…" : "Create account"}
              </span>
            </button>

            {errors.submit ? (
              <p className="rounded-sm border border-rose-500/50 bg-rose-500/10 px-4 py-3 text-sm text-rose-300">
                {errors.submit}
              </p>
            ) : null}
            {isSuccess ? (
              <p className="rounded-sm border border-white/30 bg-white/10 px-4 py-3 text-sm text-white/95">
                Account created. Redirecting to login…
              </p>
            ) : null}
          </form>

          <p className="mt-6 text-center text-sm text-white/70">
            Already have an account?{" "}
            <Link href="/login" className="font-semibold text-white underline-offset-2 hover:underline">
              Sign in
            </Link>
          </p>
        </section>
      </main>

      <Footer />
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
  onChange: (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => void;
  error?: string;
  autoComplete?: string;
}) {
  return (
    <div className="space-y-2">
      <label htmlFor={id} className="text-sm font-medium text-white/90">
        {label}
      </label>
      <input
        id={id}
        type={type}
        value={value}
        onChange={onChange}
        placeholder={placeholder}
        autoComplete={autoComplete}
        className={inputClass}
      />
      {error ? <p className="text-xs text-rose-400">{error}</p> : null}
    </div>
  );
}
