import Link from "next/link";

import { apiFetchAuthed } from "@/lib/api-server";
import { isCurrentlyActive } from "@/lib/links";
import type { LinkDetail } from "@/types/link";

import { CopyLinkButton } from "@/components/copy-link-button";
import { QrCodeReveal } from "@/components/qr-code-reveal";

export default async function LinkDetailPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  const { link, clicks } = await apiFetchAuthed<LinkDetail>(`/api/v1/links/${id}`);

  return (
    <main className="flex flex-1 flex-col gap-8 px-8 py-10">
      <div>
        <div className="flex items-center justify-between">
          <Link href="/dashboard/links" className="text-sm text-deep/70 underline">
            &larr; Back to links
          </Link>
          <Link href={`/dashboard/links/${link.id}/edit`} className="text-sm text-deep/70 underline">
            Edit
          </Link>
        </div>
        <div className="mt-2 flex items-center gap-2">
          <a
            href={link.shortUrl}
            target="_blank"
            rel="noreferrer"
            className="text-2xl font-extrabold tracking-tight text-deep underline"
          >
            {link.shortUrl}
          </a>
          <CopyLinkButton url={link.shortUrl} />
        </div>
        <p className="text-deep/70">&rarr; {link.destination}</p>
      </div>

      <div className="grid grid-cols-2 gap-4 text-sm sm:grid-cols-4">
        <Stat label="Clicks" value={String(link.clickCount)} />
        <Stat label="Max clicks" value={link.maxClicks ? String(link.maxClicks) : "—"} />
        <Stat
          label="Expires"
          value={link.expiresAt ? new Date(link.expiresAt).toLocaleDateString() : "—"}
        />
        <Stat label="Status" value={isCurrentlyActive(link) ? "Active" : "Inactive"} />
      </div>

      <div className="flex flex-col gap-3">
        <h2 className="text-lg font-semibold text-deep">QR code</h2>
        <QrCodeReveal linkId={link.id} shortCode={link.shortCode} generated={link.qrGenerated} />
      </div>

      <div className="flex flex-col gap-3">
        <h2 className="text-lg font-semibold text-deep">Recent clicks</h2>
        {clicks.length === 0 ? (
          <p className="text-deep/70">No clicks yet.</p>
        ) : (
          <div className="overflow-hidden rounded-2xl border border-deep/10 bg-white shadow-sm">
            <table className="w-full text-left text-sm">
              <thead>
                <tr className="border-b border-deep/10 text-deep/70">
                  <th className="px-5 py-3 font-medium">Time</th>
                  <th className="px-5 py-3 font-medium">Referrer</th>
                  <th className="px-5 py-3 font-medium">User agent</th>
                </tr>
              </thead>
              <tbody>
                {clicks.map((click) => (
                  <tr key={click.id} className="border-b border-deep/10 last:border-0">
                    <td className="px-5 py-3 text-deep/80">
                      {new Date(click.clickedAt).toLocaleString()}
                    </td>
                    <td className="px-5 py-3 text-deep/80">{click.referrer || "—"}</td>
                    <td className="max-w-xs truncate px-5 py-3 text-deep/80">
                      {click.userAgent}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </main>
  );
}

function Stat({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-2xl border border-deep/10 bg-white p-4 shadow-sm">
      <div className="text-deep/60">{label}</div>
      <div className="text-lg font-bold text-deep">{value}</div>
    </div>
  );
}
