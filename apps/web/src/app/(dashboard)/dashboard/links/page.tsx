import Link from "next/link";

import { apiFetchAuthed } from "@/lib/api-server";
import { isCurrentlyActive } from "@/lib/links";
import { CopyLinkButton } from "@/components/copy-link-button";
import type { Link as ShortLink } from "@/types/link";

import { DeleteLinkButton } from "./delete-link-button";

export default async function LinksPage() {
  const links = await apiFetchAuthed<ShortLink[]>("/api/v1/links");

  return (
    <main className="flex flex-1 flex-col gap-6 px-8 py-10">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-extrabold tracking-tight text-deep">Your links</h1>
        <Link
          href="/dashboard/links/new"
          className="rounded-full bg-deep px-5 py-2.5 text-sm font-semibold text-bright transition-transform duration-200 hover:scale-105 active:scale-95"
        >
          New link
        </Link>
      </div>

      {links.length === 0 ? (
        <p className="text-deep/70">You haven&apos;t created any links yet.</p>
      ) : (
        <div className="overflow-hidden rounded-2xl border border-deep/10 bg-white shadow-sm">
          <table className="w-full text-left text-sm">
            <thead>
              <tr className="border-b border-deep/10 text-deep/70">
                <th className="px-5 py-3 font-medium">Short link</th>
                <th className="px-5 py-3 font-medium">Destination</th>
                <th className="px-5 py-3 font-medium">Clicks</th>
                <th className="px-5 py-3 font-medium">Status</th>
                <th className="px-5 py-3" />
              </tr>
            </thead>
            <tbody>
              {links.map((link) => (
                <tr key={link.id} className="border-b border-deep/10 last:border-0">
                  <td className="px-5 py-3">
                    <div className="flex items-center gap-2">
                      <a
                        href={link.shortUrl}
                        target="_blank"
                        rel="noreferrer"
                        className="font-medium text-deep underline"
                      >
                        {link.shortUrl.replace(/^https?:\/\//, "")}
                      </a>
                      <CopyLinkButton url={link.shortUrl} />
                    </div>
                    <Link
                      href={`/dashboard/links/${link.id}`}
                      className="text-xs text-deep/60 underline"
                    >
                      View analytics
                    </Link>
                  </td>
                  <td className="max-w-xs truncate px-5 py-3 text-deep/80">
                    {link.destination}
                  </td>
                  <td className="px-5 py-3 text-deep/80">
                    {link.clickCount}
                    {link.maxClicks ? ` / ${link.maxClicks}` : ""}
                  </td>
                  <td className="px-5 py-3 text-deep/80">
                    {isCurrentlyActive(link) ? "Active" : "Inactive"}
                  </td>
                  <td className="px-5 py-3">
                    <div className="flex items-center gap-3">
                      <Link
                        href={`/dashboard/links/${link.id}/edit`}
                        className="text-sm text-deep/70 hover:underline"
                      >
                        Edit
                      </Link>
                      <DeleteLinkButton linkId={link.id} />
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </main>
  );
}
