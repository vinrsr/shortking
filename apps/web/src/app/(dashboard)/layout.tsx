import Image from "next/image";
import Link from "next/link";

import { apiFetchAuthed } from "@/lib/api-server";

import { LogoutButton } from "./logout-button";
import { SidebarNav } from "./sidebar-nav";
import { VerifyEmailBanner } from "./verify-email-banner";

interface Me {
  id: string;
  email: string;
  displayName: string;
  emailVerified: boolean;
}

export default async function DashboardLayout({ children }: { children: React.ReactNode }) {
  const me = await apiFetchAuthed<Me>("/api/v1/me").catch(() => null);

  return (
    <div className="relative flex min-h-screen flex-1 overflow-hidden bg-gradient-to-b from-bright via-bright/60 to-white text-deep">
      <div
        className="pointer-events-none absolute inset-0 bg-[radial-gradient(circle,_rgba(61,59,120,0.14)_1px,_transparent_1px)] bg-[length:22px_22px]"
        aria-hidden
      />

      <aside className="relative z-10 flex w-64 shrink-0 flex-col gap-6 border-r border-deep/10 bg-white/70 p-6 backdrop-blur-sm">
        <Link href="/dashboard" className="shrink-0">
          <Image
            src="/logo-one-line.png"
            alt="Short King"
            width={1920}
            height={300}
            priority
            className="h-8 w-auto"
          />
        </Link>
        <SidebarNav />
      </aside>

      <div className="relative z-10 flex flex-1 flex-col">
        <header className="flex items-center justify-end gap-4 border-b border-deep/10 bg-white/50 px-8 py-4 backdrop-blur-sm">
          {(me?.displayName || me?.email) && (
            <span className="text-sm font-medium text-deep">
              Hi, {me?.displayName || me?.email}
            </span>
          )}
          <LogoutButton />
        </header>
        {me && !me.emailVerified && <VerifyEmailBanner />}
        <div className="flex flex-1 flex-col">{children}</div>
      </div>
    </div>
  );
}
