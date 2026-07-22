"use client";

import { useState } from "react";

import { Spinner } from "@/components/spinner";

export function VerifyEmailBanner() {
  const [status, setStatus] = useState<"idle" | "sending" | "sent">("idle");

  async function handleResend() {
    setStatus("sending");
    try {
      await fetch("/api/auth/resend-verification", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({}),
      });
      setStatus("sent");
    } catch {
      setStatus("idle");
    }
  }

  return (
    <div className="flex flex-wrap items-center justify-between gap-3 border-b border-accent/40 bg-accent/15 px-8 py-3 text-sm text-deep">
      <span>
        {status === "sent"
          ? "Verification email sent — check your inbox."
          : "Verify your email to secure your account."}
      </span>
      {status !== "sent" && (
        <button
          onClick={handleResend}
          disabled={status === "sending"}
          className="inline-flex shrink-0 items-center justify-center gap-2 rounded-full bg-deep px-4 py-1.5 font-semibold text-bright transition-transform duration-200 hover:scale-105 active:scale-95 disabled:opacity-50"
        >
          {status === "sending" && <Spinner className="h-3.5 w-3.5" />}
          {status === "sending" ? "Sending..." : "Resend email"}
        </button>
      )}
    </div>
  );
}
