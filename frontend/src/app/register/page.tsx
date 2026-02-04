"use client";

import { useState } from "react";
import Input from "../component/ui/input";
import Button from "../component/ui/button";

export default function RegisterPage() {
  type ApiResponse<T> = {
    success?: boolean;
    data?: T;
    error?: string;
  };

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

  const [formData, setFormData] = useState<FormState>(initialFormData);
  const [errors, setErrors] = useState<FormErrors>({});
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { id, value } = e.target;
    const field = id as keyof FormState;
    setFormData((prev) => ({ ...prev, [field]: value }));
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
      const response = await fetch(endpoint, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        credentials: "include",
        body: JSON.stringify(payload),
      });

      const result = (await response.json().catch(() => null)) as
        | ApiResponse<null>
        | null;

      if (!response.ok || !result?.success) {
        setErrors({
          submit: result?.error || "Registration failed. Please try again.",
        });
        return;
      }

      setFormData(initialFormData);
    } catch (error) {
      setErrors({
        submit: "Network error. Please try again.",
      });
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="min-h-screen bg-white text-slate-900 flex items-center justify-center">
      <div className="w-full max-w-2xl px-6 py-16">
        <div className="space-y-6">
          <div className="text-center">
            <p className="text-2xl font-semibold tracking-tight text-slate-800">
              Social Network
            </p>
          </div>

          <div className="space-y-6">
            <h1 className="text-3xl font-semibold tracking-tight text-center">
              Sign Up
            </h1>

            <form className="space-y-4 text-left" onSubmit={handleSubmit} noValidate>
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-x-4 gap-y-4">
                <div className="space-y-2">
                  <label
                    htmlFor="firstName"
                    className="text-sm font-medium block"
                  >
                    First Name
                  </label>
                  <Input
                    id="firstName"
                    type="text"
                    placeholder="First Name"
                    className="w-full h-10 rounded-md border-slate-300 text-sm focus:ring-1 focus:ring-blue-500"
                    value={formData.firstName}
                    onChange={handleInputChange}
                  />
                  {errors.firstName ? (
                    <p className="text-xs text-red-600">{errors.firstName}</p>
                  ) : null}
                </div>
                <div className="space-y-2">
                  <label
                    htmlFor="lastName"
                    className="text-sm font-medium block"
                  >
                    Last Name
                  </label>
                  <Input
                    id="lastName"
                    type="text"
                    placeholder="Last Name"
                    className="w-full h-10 rounded-md border-slate-300 text-sm focus:ring-1 focus:ring-blue-500"
                    value={formData.lastName}
                    onChange={handleInputChange}
                  />
                  {errors.lastName ? (
                    <p className="text-xs text-red-600">{errors.lastName}</p>
                  ) : null}
                </div>
              </div>

              <div className="space-y-2">
                <label htmlFor="email" className="text-sm font-medium block">
                  Email
                </label>
                <Input
                  id="email"
                  type="email"
                  placeholder="Enter a valid email"
                  autoComplete="email"
                  className="w-full h-10 rounded-md border-slate-300 text-sm focus:ring-1 focus:ring-blue-500"
                  value={formData.email}
                  onChange={handleInputChange}
                />
                {errors.email ? (
                  <p className="text-xs text-red-600">{errors.email}</p>
                ) : null}
              </div>

              <div className="space-y-2">
                <label htmlFor="password" className="text-sm font-medium block">
                  Password
                </label>
                <Input
                  id="password"
                  type="password"
                  placeholder="Create a password"
                  autoComplete="new-password"
                  className="w-full h-10 rounded-md border-slate-300 text-sm focus:ring-1 focus:ring-blue-500"
                  value={formData.password}
                  onChange={handleInputChange}
                />
                {errors.password ? (
                  <p className="text-xs text-red-600">{errors.password}</p>
                ) : null}
              </div>
              <div className="space-y-2">
                <label
                  htmlFor="confirmPassword"
                  className="text-sm font-medium block"
                >
                  Confirm Password
                </label>
                <Input
                  id="confirmPassword"
                  type="password"
                  placeholder="Confirm your password"
                  autoComplete="new-password"
                  className="w-full h-10 rounded-md border-slate-300 text-sm focus:ring-1 focus:ring-blue-500"
                  value={formData.confirmPassword}
                  onChange={handleInputChange}
                />
                {errors.confirmPassword ? (
                  <p className="text-xs text-red-600">
                    {errors.confirmPassword}
                  </p>
                ) : null}
              </div>
              <div className="space-y-2">
                <label htmlFor="dob" className="text-sm font-medium block">
                  Date of Birth
                </label>
                <div className="relative">
                  <Input
                    id="dob"
                    type="date"
                    placeholder="DD / MM / YYYY"
                    className="w-full h-10 rounded-md border-slate-300 pr-10 text-sm focus:ring-1 focus:ring-blue-500"
                    value={formData.dob}
                    onChange={handleInputChange}
                  />
                  {errors.dob ? (
                    <p className="text-xs text-red-600">{errors.dob}</p>
                  ) : null}
                  <span className="pointer-events-none absolute right-3 top-1/2 -translate-y-1/2 text-slate-400">
                    <svg
                      aria-hidden="true"
                      viewBox="0 0 24 24"
                      className="h-4 w-4"
                      fill="none"
                      stroke="currentColor"
                      strokeWidth="2"
                    >
                      <rect x="3" y="5" width="18" height="16" rx="2" />
                      <path d="M16 3v4M8 3v4M3 11h18" />
                    </svg>
                  </span>
                </div>
              </div>

              <div className="grid grid-cols-1 sm:grid-cols-2 gap-x-4 gap-y-4">
                <div className="space-y-2">
                  <label
                    htmlFor="nickname"
                    className="text-sm font-medium block"
                  >
                    Nickname (Optional)
                  </label>
                  <Input
                    id="nickname"
                    type="text"
                    placeholder="Enter a nickname"
                    className="w-full h-10 rounded-md border-slate-300 text-sm focus:ring-1 focus:ring-blue-500"
                    value={formData.nickname}
                    onChange={handleInputChange}
                  />
                </div>
              </div>

              <div className="space-y-2">
                <label htmlFor="about" className="text-sm font-medium block">
                  About Me (Optional)
                </label>
                <Input
                  id="about"
                  type="text"
                  placeholder="Tell us about yourself..."
                  className="w-full h-10 rounded-md border-slate-300 text-sm focus:ring-1 focus:ring-blue-500"
                  value={formData.about}
                  onChange={handleInputChange}
                />
              </div>

              <Button
                type="submit"
                className="w-full rounded-md bg-gradient-to-b from-emerald-500 to-emerald-700 py-2.5 text-sm shadow-sm hover:from-emerald-600 hover:to-emerald-800"
                disabled={isSubmitting}
              >
                {isSubmitting ? "Signing Up..." : "Sign Up"}
              </Button>
              {errors.submit ? (
                <p className="text-sm text-red-600">{errors.submit}</p>
              ) : null}
            </form>

            <p className="text-xs text-slate-500 mt-4 text-center">
              By signing up, you agree to our{" "}
              <span className="text-slate-700">Terms &amp; Privacy Policy</span>
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}
