import Image from "next/image";
import Link from "next/link";

import { NavLinks } from "./nav-links";

export default function MarketingLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="relative flex flex-1 flex-col overflow-hidden bg-gradient-to-b from-bright via-bright/70 to-white text-deep">
      <div
        className="pointer-events-none absolute inset-0 bg-[radial-gradient(circle,_rgba(61,59,120,0.14)_1px,_transparent_1px)] bg-[length:22px_22px]"
        aria-hidden
      />

      <header className="relative z-10 flex items-center justify-between px-4 py-4 sm:px-10 sm:py-6">
        <Link href="/" className="shrink-0">
          <Image
            src="/logo-one-line.png"
            alt="Short King"
            width={1920}
            height={300}
            priority
            className="h-6 w-auto sm:h-8"
          />
        </Link>

        <NavLinks />

        <div className="flex items-center gap-2 sm:gap-4">
          <Link
            href="/login"
            className="text-xs font-medium text-deep transition-opacity hover:opacity-70 sm:text-sm"
          >
            Log in
          </Link>
          <Link
            href="/signup"
            className="rounded-full bg-deep px-3 py-2 text-xs font-semibold text-bright transition-transform duration-200 hover:scale-105 active:scale-95 sm:px-5 sm:py-2.5 sm:text-sm"
          >
            Get Started
          </Link>
        </div>
      </header>

      <div className="relative z-10 flex flex-1 flex-col">{children}</div>
    </div>
  );
}
