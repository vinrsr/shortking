"use client";

import { useRouter } from "next/navigation";

export function LogoutButton() {
  const router = useRouter();

  async function handleLogout() {
    await fetch("/api/auth/logout", { method: "POST" });
    router.push("/login");
    router.refresh();
  }

  return (
    <button
      onClick={handleLogout}
      className="rounded-full border-2 border-deep/20 px-4 py-2 text-sm font-medium text-deep transition-transform duration-200 hover:scale-105 hover:border-deep/40 active:scale-95"
    >
      Log out
    </button>
  );
}
