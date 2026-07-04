"use client";

import Link from "next/link";
import { useSearchParams } from "next/navigation";
import { Suspense, useEffect, useState } from "react";

type Status = "verifying" | "success" | "error";

function VerifyEmailStatus() {
  const searchParams = useSearchParams();
  const token = searchParams.get("token") ?? "";
  const [status, setStatus] = useState<Status>(token ? "verifying" : "error");
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!token) return;

    let cancelled = false;

    (async () => {
      const res = await fetch("/api/auth/verify-email", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ token }),
      });

      if (cancelled) return;

      if (!res.ok) {
        const data = await res.json().catch(() => ({}));
        setError(data.error ?? "Failed to verify email");
        setStatus("error");
        return;
      }

      setStatus("success");
    })();

    return () => {
      cancelled = true;
    };
  }, [token]);

  if (!token) {
    return (
      <p className="mt-6 text-sm text-red-600">
        This verification link is missing a token. Request a new one from your dashboard.
      </p>
    );
  }

  if (status === "verifying") {
    return <p className="mt-6 text-sm text-deep/70">Verifying your email...</p>;
  }

  if (status === "error") {
    return <p className="mt-6 text-sm text-red-600">{error}</p>;
  }

  return (
    <>
      <p className="mt-6 text-sm text-deep/80">Your email is verified.</p>
      <Link
        href="/dashboard"
        className="mt-6 inline-block rounded-full bg-deep px-6 py-3 text-center font-semibold text-bright transition-transform duration-200 hover:scale-105 active:scale-95"
      >
        Go to dashboard
      </Link>
    </>
  );
}

export default function VerifyEmailPage() {
  return (
    <main className="flex flex-1 flex-col items-center justify-center px-6 py-12">
      <div className="w-full max-w-sm rounded-3xl border border-deep/10 bg-white p-8 text-center shadow-sm sm:p-10">
        <h1 className="text-3xl font-extrabold tracking-tight text-deep">Verify email</h1>
        <Suspense fallback={<p className="mt-6 text-sm text-deep/70">Verifying your email...</p>}>
          <VerifyEmailStatus />
        </Suspense>
      </div>
    </main>
  );
}
