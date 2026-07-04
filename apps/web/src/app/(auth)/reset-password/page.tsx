"use client";

import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { Suspense, useState } from "react";

function ResetPasswordForm() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const token = searchParams.get("token") ?? "";
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  async function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setError(null);
    setSubmitting(true);

    const form = new FormData(e.currentTarget);
    const res = await fetch("/api/auth/reset-password", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ token, newPassword: form.get("newPassword") }),
    });

    setSubmitting(false);

    if (!res.ok) {
      const data = await res.json().catch(() => ({}));
      setError(data.error ?? "Failed to reset password");
      return;
    }

    router.push("/login");
  }

  if (!token) {
    return (
      <p className="mt-6 text-sm text-red-600">
        This reset link is missing a token. Request a new one from the{" "}
        <Link href="/forgot-password" className="underline">
          forgot password
        </Link>{" "}
        page.
      </p>
    );
  }

  return (
    <form onSubmit={handleSubmit} className="mt-6 flex flex-col gap-4">
      {error && <p className="text-sm text-red-600">{error}</p>}
      <input
        type="password"
        name="newPassword"
        placeholder="New password"
        required
        minLength={8}
        className="rounded-2xl border border-deep/10 bg-bright/40 px-4 py-3 text-deep placeholder:text-deep/40 focus:border-deep/30 focus:outline-none"
      />
      <button
        type="submit"
        disabled={submitting}
        className="mt-2 rounded-full bg-deep px-4 py-3 font-semibold text-bright transition-transform duration-200 hover:scale-105 active:scale-95 disabled:opacity-50 disabled:hover:scale-100"
      >
        {submitting ? "Saving..." : "Set new password"}
      </button>
    </form>
  );
}

export default function ResetPasswordPage() {
  return (
    <main className="flex flex-1 flex-col items-center justify-center px-6 py-12">
      <div className="w-full max-w-sm rounded-3xl border border-deep/10 bg-white p-8 shadow-sm sm:p-10">
        <h1 className="text-3xl font-extrabold tracking-tight text-deep">Reset password</h1>
        <p className="mt-1 text-sm text-deep/70">Choose a new password for your account.</p>
        <Suspense fallback={null}>
          <ResetPasswordForm />
        </Suspense>
      </div>
    </main>
  );
}
