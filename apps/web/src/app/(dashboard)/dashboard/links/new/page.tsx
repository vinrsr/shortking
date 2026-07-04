"use client";

import { useRouter } from "next/navigation";
import { useState } from "react";

export default function NewLinkPage() {
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

    const res = await fetch("/api/links", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        destination: form.get("destination"),
        customAlias: form.get("alias") || undefined,
        expiresAt: expiresAt ? new Date(expiresAt).toISOString() : undefined,
        maxClicks: maxClicks ? Number(maxClicks) : undefined,
      }),
    });

    setSubmitting(false);

    if (!res.ok) {
      const data = await res.json().catch(() => ({}));
      setError(data.error ?? "Failed to create link");
      return;
    }

    router.push("/dashboard/links");
    router.refresh();
  }

  return (
    <main className="flex flex-1 flex-col gap-6 px-8 py-10">
      <h1 className="text-2xl font-extrabold tracking-tight text-deep">Create a link</h1>
      <form onSubmit={handleSubmit} className="flex max-w-md flex-col gap-4">
        {error && <p className="text-sm text-red-600">{error}</p>}
        <input
          type="url"
          name="destination"
          placeholder="Destination URL"
          required
          className="rounded-2xl border border-deep/10 bg-bright/40 px-4 py-3 text-deep placeholder:text-deep/40 focus:border-deep/30 focus:outline-none"
        />
        <input
          type="text"
          name="alias"
          placeholder="Custom alias (optional)"
          className="rounded-2xl border border-deep/10 bg-bright/40 px-4 py-3 text-deep placeholder:text-deep/40 focus:border-deep/30 focus:outline-none"
        />
        <label className="flex flex-col gap-1 text-sm text-deep/70">
          Expires at (optional)
          <input
            type="date"
            name="expiresAt"
            className="rounded-2xl border border-deep/10 bg-bright/40 px-4 py-3 text-deep focus:border-deep/30 focus:outline-none"
          />
        </label>
        <label className="flex flex-col gap-1 text-sm text-deep/70">
          Max clicks (optional)
          <input
            type="number"
            name="maxClicks"
            min={1}
            className="rounded-2xl border border-deep/10 bg-bright/40 px-4 py-3 text-deep placeholder:text-deep/40 focus:border-deep/30 focus:outline-none"
          />
        </label>
        <button
          type="submit"
          disabled={submitting}
          className="mt-2 rounded-full bg-deep px-4 py-3 font-semibold text-bright transition-transform duration-200 hover:scale-105 active:scale-95 disabled:opacity-50 disabled:hover:scale-100"
        >
          {submitting ? "Creating..." : "Create"}
        </button>
      </form>
    </main>
  );
}
