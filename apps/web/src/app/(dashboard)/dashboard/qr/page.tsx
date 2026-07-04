import { apiFetchAuthed } from "@/lib/api-server";
import { QrCodeReveal } from "@/components/qr-code-reveal";
import type { Link as ShortLink } from "@/types/link";

export default async function QrCodesPage() {
  const links = await apiFetchAuthed<ShortLink[]>("/api/v1/links");

  return (
    <main className="flex flex-1 flex-col gap-6 px-8 py-10">
      <h1 className="text-2xl font-extrabold tracking-tight text-deep">QR codes</h1>

      {links.length === 0 ? (
        <p className="text-deep/70">You haven&apos;t created any links yet.</p>
      ) : (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {links.map((link) => (
            <div
              key={link.id}
              className="flex flex-col gap-3 rounded-2xl border border-deep/10 bg-white p-6 shadow-sm"
            >
              <div>
                <p className="font-semibold text-deep">
                  {link.shortUrl.replace(/^https?:\/\//, "")}
                </p>
                <p className="truncate text-sm text-deep/60">{link.destination}</p>
              </div>
              <QrCodeReveal linkId={link.id} shortCode={link.shortCode} generated={link.qrGenerated} />
            </div>
          ))}
        </div>
      )}
    </main>
  );
}
