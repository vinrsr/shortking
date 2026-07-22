"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState } from "react";

import { Spinner } from "@/components/spinner";

export default function LoginPage() {
  const router = useRouter();
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  async function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setError(null);
    setSubmitting(true);

    const form = new FormData(e.currentTarget);
    try {
      const res = await fetch("/api/auth/login", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          email: form.get("email"),
          password: form.get("password"),
        }),
      });

      if (!res.ok) {
        const data = await res.json().catch(() => ({}));
        setError(data.error ?? "Login failed");
        return;
      }

      router.push("/dashboard");
      router.refresh();
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <main className="flex flex-1 flex-col items-center justify-center px-6 py-12">
      <div className="w-full max-w-sm rounded-3xl border border-deep/10 bg-white p-8 shadow-sm sm:p-10">
        <h1 className="text-3xl font-extrabold tracking-tight text-deep">Log in</h1>
        <p className="mt-1 text-sm text-deep/70">Welcome back, King.</p>

        <form onSubmit={handleSubmit} className="mt-6 flex flex-col gap-4">
          {error && <p className="text-sm text-red-600">{error}</p>}
          <input
            type="email"
            name="email"
            placeholder="Email"
            required
            className="rounded-2xl border border-deep/10 bg-bright/40 px-4 py-3 text-deep placeholder:text-deep/40 focus:border-deep/30 focus:outline-none"
          />
          <input
            type="password"
            name="password"
            placeholder="Password"
            required
            className="rounded-2xl border border-deep/10 bg-bright/40 px-4 py-3 text-deep placeholder:text-deep/40 focus:border-deep/30 focus:outline-none"
          />
          <button
            type="submit"
            disabled={submitting}
            className="mt-2 inline-flex items-center justify-center gap-2 rounded-full bg-deep px-4 py-3 font-semibold text-bright transition-transform duration-200 hover:scale-105 active:scale-95 disabled:opacity-50 disabled:hover:scale-100"
          >
            {submitting && <Spinner className="h-4 w-4" />}
            {submitting ? "Logging in..." : "Log in"}
          </button>
        </form>

        <p className="mt-4 text-center text-sm text-deep/70">
          <Link href="/forgot-password" className="underline">
            Forgot password?
          </Link>
        </p>

        <p className="mt-6 text-center text-sm text-deep/70">
          Don&apos;t have an account?{" "}
          <Link href="/signup" className="font-medium underline">
            Sign up
          </Link>
        </p>
      </div>
    </main>
  );
}
