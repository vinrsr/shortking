"use client";

import { useRouter } from "next/navigation";
import { useState } from "react";

import type { Link as ShortLink } from "@/types/link";

export function EditLinkForm({ link }: { link: ShortLink }) {
  const router = useRouter();
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  async function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setError(null);
    setSubmitting(true);

    const form = new FormData(e.currentTarget);
    const expiresAt = form.get("expiresAt") as string;
    const maxClicks = form.get("maxClicks") as string;

    const res = await fetch(`/api/links/${link.id}`, {
      method: "PATCH",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        destination: form.get("destination"),
        expiresAt: expiresAt ? new Date(expiresAt).toISOString() : undefined,
        maxClicks: maxClicks ? Number(maxClicks) : undefined,
        isActive: form.get("isActive") === "on",
      }),
    });

    setSubmitting(false);

    if (!res.ok) {
      const data = await res.json().catch(() => ({}));
      setError(data.error ?? "Failed to update link");
      return;
    }

    router.push(`/dashboard/links/${link.id}`);
    router.refresh();
  }

  return (
    <form onSubmit={handleSubmit} className="flex max-w-md flex-col gap-4">
      {error && <p className="text-sm text-red-600">{error}</p>}
      <label className="flex flex-col gap-1 text-sm text-deep/70">
        Short link
        <input
          type="text"
          value={link.shortUrl.replace(/^https?:\/\//, "")}
          disabled
          className="rounded-2xl border border-deep/10 bg-deep/5 px-4 py-3 text-deep/50"
        />
      </label>
      <input
        type="url"
        name="destination"
        placeholder="Destination URL"
        defaultValue={link.destination}
        required
        className="rounded-2xl border border-deep/10 bg-bright/40 px-4 py-3 text-deep placeholder:text-deep/40 focus:border-deep/30 focus:outline-none"
      />
      <label className="flex flex-col gap-1 text-sm text-deep/70">
        Expires at (optional)
        <input
          type="date"
          name="expiresAt"
          defaultValue={link.expiresAt ? link.expiresAt.slice(0, 10) : ""}
          className="rounded-2xl border border-deep/10 bg-bright/40 px-4 py-3 text-deep focus:border-deep/30 focus:outline-none"
        />
      </label>
      <label className="flex flex-col gap-1 text-sm text-deep/70">
        Max clicks (optional)
        <input
          type="number"
          name="maxClicks"
          min={1}
          defaultValue={link.maxClicks ?? ""}
          className="rounded-2xl border border-deep/10 bg-bright/40 px-4 py-3 text-deep placeholder:text-deep/40 focus:border-deep/30 focus:outline-none"
        />
      </label>
      <label className="flex items-center gap-2 text-sm text-deep/70">
        <input type="checkbox" name="isActive" defaultChecked={link.isActive} className="h-4 w-4" />
        Active
      </label>
      <button
        type="submit"
        disabled={submitting}
        className="mt-2 rounded-full bg-deep px-4 py-3 font-semibold text-bright transition-transform duration-200 hover:scale-105 active:scale-95 disabled:opacity-50 disabled:hover:scale-100"
      >
        {submitting ? "Saving..." : "Save changes"}
      </button>
    </form>
  );
}
