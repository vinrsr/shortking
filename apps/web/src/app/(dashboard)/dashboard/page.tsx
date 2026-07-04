import Link from "next/link";

import { apiFetchAuthed } from "@/lib/api-server";
import { isCurrentlyActive } from "@/lib/links";
import type { Link as ShortLink } from "@/types/link";

export default async function DashboardOverviewPage() {
  const links = await apiFetchAuthed<ShortLink[]>("/api/v1/links");

  const totalLinks = links.length;
  const activeLinks = links.filter(isCurrentlyActive).length;
  const totalClicks = links.reduce((sum, l) => sum + l.clickCount, 0);

  const recentLinks = links.slice(0, 5);

  const stats = [
    { label: "Total links", value: totalLinks },
    { label: "Active links", value: activeLinks },
    { label: "Total clicks", value: totalClicks },
  ];

  return (
    <main className="flex flex-1 flex-col gap-8 px-8 py-10">
      <h1 className="text-2xl font-extrabold tracking-tight text-deep">Dashboard</h1>

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
        {stats.map((stat) => (
          <div
            key={stat.label}
            className="flex flex-col items-center gap-1 rounded-2xl border border-deep/10 bg-white px-4 py-8 text-center shadow-sm"
          >
            <span className="text-3xl font-bold text-deep">{stat.value.toLocaleString()}</span>
            <span className="text-sm text-deep/70">{stat.label}</span>
          </div>
        ))}
      </div>

      <div className="flex flex-col gap-3">
        <div className="flex items-center justify-between">
          <h2 className="text-lg font-semibold text-deep">Recent links</h2>
          <Link href="/dashboard/links" className="text-sm text-deep/70 underline">
            View all
          </Link>
        </div>

        {recentLinks.length === 0 ? (
          <p className="text-deep/70">
            You haven&apos;t created any links yet.{" "}
            <Link href="/dashboard/links/new" className="underline">
              Create one
            </Link>
            .
          </p>
        ) : (
          <div className="overflow-hidden rounded-2xl border border-deep/10 bg-white shadow-sm">
            <table className="w-full text-left text-sm">
              <thead>
                <tr className="border-b border-deep/10 text-deep/70">
                  <th className="px-5 py-3 font-medium">Short link</th>
                  <th className="px-5 py-3 font-medium">Destination</th>
                  <th className="px-5 py-3 font-medium">Clicks</th>
                  <th className="px-5 py-3 font-medium">Status</th>
                </tr>
              </thead>
              <tbody>
                {recentLinks.map((link) => (
                  <tr key={link.id} className="border-b border-deep/10 last:border-0">
                    <td className="px-5 py-3">
                      <Link
                        href={`/dashboard/links/${link.id}`}
                        className="font-medium text-deep underline"
                      >
                        {link.shortUrl.replace(/^https?:\/\//, "")}
                      </Link>
                    </td>
                    <td className="max-w-xs truncate px-5 py-3 text-deep/80">
                      {link.destination}
                    </td>
                    <td className="px-5 py-3 text-deep/80">{link.clickCount}</td>
                    <td className="px-5 py-3 text-deep/80">
                      {isCurrentlyActive(link) ? "Active" : "Inactive"}
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
