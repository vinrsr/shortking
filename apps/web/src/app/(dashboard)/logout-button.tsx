"use client";

import { useRouter } from "next/navigation";
import { useState } from "react";

import { Spinner } from "@/components/spinner";

export function LogoutButton() {
  const router = useRouter();
  const [pending, setPending] = useState(false);

  async function handleLogout() {
    setPending(true);
    try {
      await fetch("/api/auth/logout", { method: "POST" });
      router.push("/login");
      router.refresh();
    } finally {
      setPending(false);
    }
  }

  return (
    <button
      onClick={handleLogout}
      disabled={pending}
      className="inline-flex items-center justify-center gap-2 rounded-full border-2 border-deep/20 px-4 py-2 text-sm font-medium text-deep transition-transform duration-200 hover:scale-105 hover:border-deep/40 active:scale-95 disabled:opacity-50 disabled:hover:scale-100"
    >
      {pending && <Spinner className="h-4 w-4" />}
      {pending ? "Logging out..." : "Log out"}
    </button>
  );
}
