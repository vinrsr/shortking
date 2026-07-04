"use client";

import Link from "next/link";
import { useState } from "react";

export default function ForgotPasswordPage() {
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [sent, setSent] = useState(false);

  async function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setError(null);
    setSubmitting(true);

    const form = new FormData(e.currentTarget);
    const res = await fetch("/api/auth/forgot-password", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ email: form.get("email") }),
    });

    setSubmitting(false);

    if (!res.ok) {
      const data = await res.json().catch(() => ({}));
      setError(data.error ?? "Something went wrong");
      return;
    }

    setSent(true);
  }

  return (
    <main className="flex flex-1 flex-col items-center justify-center px-6 py-12">
      <div className="w-full max-w-sm rounded-3xl border border-deep/10 bg-white p-8 shadow-sm sm:p-10">
        <h1 className="text-3xl font-extrabold tracking-tight text-deep">Forgot password</h1>
        <p className="mt-1 text-sm text-deep/70">
          Enter your email and we&apos;ll send you a reset link.
        </p>

        {sent ? (
          <p className="mt-6 text-sm text-deep/80">
            If that email is registered, a reset link is on its way.
          </p>
        ) : (
          <form onSubmit={handleSubmit} className="mt-6 flex flex-col gap-4">
            {error && <p className="text-sm text-red-600">{error}</p>}
            <input
              type="email"
              name="email"
              placeholder="Email"
              required
              className="rounded-2xl border border-deep/10 bg-bright/40 px-4 py-3 text-deep placeholder:text-deep/40 focus:border-deep/30 focus:outline-none"
            />
            <button
              type="submit"
              disabled={submitting}
              className="mt-2 rounded-full bg-deep px-4 py-3 font-semibold text-bright transition-transform duration-200 hover:scale-105 active:scale-95 disabled:opacity-50 disabled:hover:scale-100"
            >
              {submitting ? "Sending..." : "Send reset link"}
            </button>
          </form>
        )}

        <p className="mt-6 text-center text-sm text-deep/70">
          <Link href="/login" className="font-medium underline">
            Back to log in
          </Link>
        </p>
      </div>
    </main>
  );
}
