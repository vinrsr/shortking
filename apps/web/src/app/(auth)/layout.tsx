import Image from "next/image";
import Link from "next/link";

export default function AuthLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="relative flex flex-1 flex-col overflow-hidden bg-gradient-to-b from-bright via-bright/70 to-white text-deep">
      <div
        className="pointer-events-none absolute inset-0 bg-[radial-gradient(circle,_rgba(61,59,120,0.14)_1px,_transparent_1px)] bg-[length:22px_22px]"
        aria-hidden
      />

      <header className="relative z-10 flex items-center justify-between px-6 py-6 sm:px-10">
        <Link href="/" className="shrink-0">
          <Image
            src="/logo-one-line.png"
            alt="Short King"
            width={1920}
            height={300}
            priority
            className="h-8 w-auto"
          />
        </Link>
        <Link
          href="/"
          className="text-sm font-medium text-deep transition-opacity hover:opacity-70"
        >
          ← Back to home
        </Link>
      </header>

      <div className="relative z-10 flex flex-1 flex-col">{children}</div>
    </div>
  );
}
