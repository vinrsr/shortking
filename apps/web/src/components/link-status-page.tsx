import Image from "next/image";
import Link from "next/link";
import type { LucideIcon } from "lucide-react";

export function LinkStatusPage({
  icon: Icon,
  title,
  message,
}: {
  icon: LucideIcon;
  title: string;
  message: string;
}) {
  return (
    <main className="relative flex flex-1 flex-col items-center justify-center overflow-hidden bg-gradient-to-b from-bright via-bright/70 to-white px-6 py-20 text-center text-deep">
      <div
        className="pointer-events-none absolute inset-0 bg-[radial-gradient(circle,_rgba(61,59,120,0.14)_1px,_transparent_1px)] bg-[length:22px_22px]"
        aria-hidden
      />

      <div className="relative z-10 flex flex-col items-center gap-5">
        <Link href="/" className="shrink-0">
          <Image
            src="/logo-two-line.png"
            alt="Short King"
            width={1920}
            height={1080}
            priority
            className="h-20 w-auto"
          />
        </Link>

        <span className="flex h-14 w-14 items-center justify-center rounded-full bg-mist/40 text-deep">
          <Icon className="h-6 w-6" />
        </span>

        <h1 className="text-2xl font-extrabold tracking-tight sm:text-3xl">{title}</h1>
        <p className="max-w-md text-sm text-deep/70 sm:text-base">{message}</p>

        <Link
          href="/"
          className="mt-2 rounded-full border-2 border-deep/20 px-6 py-3 text-sm font-semibold text-deep transition-transform duration-200 hover:scale-105 hover:border-deep/40 active:scale-95"
        >
          Go to ShortKing
        </Link>
      </div>
    </main>
  );
}
