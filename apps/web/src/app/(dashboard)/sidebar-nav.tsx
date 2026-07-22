"use client";

import { LayoutDashboard, Link2, QrCode } from "lucide-react";
import Link, { useLinkStatus } from "next/link";
import { usePathname } from "next/navigation";

import { Spinner } from "@/components/spinner";

const navItems = [
  { href: "/dashboard", label: "Dashboard", icon: LayoutDashboard, exact: true },
  { href: "/dashboard/links", label: "Links", icon: Link2, exact: false },
  { href: "/dashboard/qr", label: "QR Codes", icon: QrCode, exact: false },
];

function NavPendingIndicator() {
  const { pending } = useLinkStatus();
  return pending ? <Spinner className="ml-auto h-4 w-4" /> : null;
}

export function SidebarNav() {
  const pathname = usePathname();

  return (
    <nav className="flex flex-col gap-1">
      {navItems.map((item) => {
        const isActive = item.exact
          ? pathname === item.href
          : pathname.startsWith(item.href);

        return (
          <Link
            key={item.href}
            href={item.href}
            className={`flex items-center gap-3 rounded-xl px-4 py-2.5 text-sm font-medium transition-colors ${
              isActive ? "bg-deep text-bright" : "text-deep hover:bg-deep/10"
            }`}
          >
            <item.icon className="h-5 w-5" />
            {item.label}
            <NavPendingIndicator />
          </Link>
        );
      })}
    </nav>
  );
}
