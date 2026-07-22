"use client";

import { useRouter } from "next/navigation";
import { useState } from "react";

import { Spinner } from "@/components/spinner";

export function DeleteLinkButton({ linkId }: { linkId: string }) {
  const router = useRouter();
  const [pending, setPending] = useState(false);

  async function handleDelete() {
    if (!confirm("Delete this link?")) return;
    setPending(true);
    try {
      await fetch(`/api/links/${linkId}`, { method: "DELETE" });
      router.refresh();
    } finally {
      setPending(false);
    }
  }

  return (
    <button
      onClick={handleDelete}
      disabled={pending}
      className="inline-flex items-center gap-1.5 text-sm text-red-600 hover:underline disabled:opacity-50"
    >
      {pending && <Spinner className="h-3.5 w-3.5" />}
      {pending ? "Deleting..." : "Delete"}
    </button>
  );
}
