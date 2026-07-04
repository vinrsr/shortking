"use client";

import { useRouter } from "next/navigation";
import { useState } from "react";

export function DeleteLinkButton({ linkId }: { linkId: string }) {
  const router = useRouter();
  const [pending, setPending] = useState(false);

  async function handleDelete() {
    if (!confirm("Delete this link?")) return;
    setPending(true);
    await fetch(`/api/links/${linkId}`, { method: "DELETE" });
    setPending(false);
    router.refresh();
  }

  return (
    <button
      onClick={handleDelete}
      disabled={pending}
      className="text-sm text-red-600 hover:underline disabled:opacity-50"
    >
      {pending ? "Deleting..." : "Delete"}
    </button>
  );
}
