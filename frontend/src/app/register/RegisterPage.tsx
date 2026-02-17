"use client";

import { useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { ArrowRight, Sparkles } from "lucide-react";
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
      ...(formData.nickname.trim()
        ? { nickname: formData.nickname.trim() }
        : {}),
      ...(formData.about.trim() ? { about: formData.about.trim() } : {}),
    };

    try {
      const response = await apiJson(apiBaseUrl, "/auth/register", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
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
    <div className="relative min-h-screen overflow-hidden bg-neutral-50 text-neutral-900">
      <div className="pointer-events-none absolute -left-24 top-8 h-72 w-72 rounded-full bg-indigo-200/35 blur-3xl" />
      <div className="pointer-events-none absolute -right-32 top-16 h-80 w-80 rounded-full bg-emerald-200/35 blur-3xl" />

      <main className="mx-auto w-full max-w-3xl px-4 py-20 sm:px-6">
        <section className="rounded-[2rem] border border-neutral-200 bg-white p-6 shadow-[0_35px_80px_-50px_rgba(2,6,23,0.45)] sm:p-8">
          <p className="inline-flex items-center gap-2 rounded-full border border-neutral-200 bg-white px-4 py-1.5 text-xs font-semibold uppercase tracking-[0.16em] text-neutral-600">
            <Sparkles className="h-3.5 w-3.5" />
            Create account
          </p>
          <h1 className="mt-4 text-3xl font-semibold tracking-tight text-neutral-900">
            Join {landingData.productName}
          </h1>
          <p className="mt-2 text-sm text-neutral-600">
            Set up your profile and start participating in high-quality discussions.
          </p>

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
                placeholder="Preferred display name"
              />
            </div>

            <div className="space-y-2">
              <label htmlFor="about" className="text-sm font-semibold text-neutral-700">
                About (optional)
              </label>
              <textarea
                id="about"
                value={formData.about}
                onChange={handleInputChange}
                rows={3}
                placeholder="Tell the community about your interests..."
                className="w-full resize-none rounded-2xl border border-neutral-300 bg-white px-4 py-3 text-sm text-neutral-900 placeholder:text-neutral-400 focus:border-neutral-500 focus:outline-none focus-visible:ring-2 focus-visible:ring-neutral-900/30"
              />
            </div>

            <button
              type="submit"
              className="brand-gradient group inline-flex h-12 w-full items-center justify-center gap-2 rounded-2xl px-5 text-sm font-semibold text-white shadow-sm transition hover:-translate-y-0.5 hover:shadow-lg focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-neutral-900/70 focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-70"
              disabled={isSubmitting}
            >
              <span>{isSubmitting ? "Creating account..." : "Create account"}</span>
              <ArrowRight className="h-4 w-4 transition-transform duration-200 group-hover:translate-x-0.5" />
            </button>

            {errors.submit ? (
              <p className="rounded-2xl border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-700">
                {errors.submit}
              </p>
            ) : null}
            {isSuccess ? (
              <p className="rounded-2xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-700">
                Account created successfully. Redirecting to login...
              </p>
            ) : null}
          </form>

          <p className="mt-6 text-center text-sm text-neutral-600">
            Already have an account?{" "}
            <Link href="/login" className="font-semibold text-neutral-900 hover:underline">
              Sign in
            </Link>
          </p>
        </section>
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
      <label htmlFor={id} className="text-sm font-semibold text-neutral-700">
        {label}
      </label>
      <input
        id={id}
        type={type}
        value={value}
        onChange={onChange}
        placeholder={placeholder}
        autoComplete={autoComplete}
        className="h-12 w-full rounded-2xl border border-neutral-300 bg-white px-4 text-sm text-neutral-900 placeholder:text-neutral-400 focus:border-neutral-500 focus:outline-none focus-visible:ring-2 focus-visible:ring-neutral-900/30"
      />
      {error ? <p className="text-xs text-rose-600">{error}</p> : null}
    </div>
  );
}
