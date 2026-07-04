"use client";

import { gsap } from "gsap";
import { ScrollToPlugin } from "gsap/ScrollToPlugin";

gsap.registerPlugin(ScrollToPlugin);

const links = [
  { href: "#features", label: "Features" },
  { href: "#analytics", label: "Analytics" },
  { href: "#compare", label: "Compare" },
  { href: "#faq", label: "FAQ" },
];

export function NavLinks() {
  function handleClick(e: React.MouseEvent<HTMLAnchorElement>, href: string) {
    e.preventDefault();
    gsap.to(window, {
      duration: 0.9,
      scrollTo: { y: href, offsetY: 24 },
      ease: "power2.inOut",
    });
  }

  return (
    <nav className="hidden items-center gap-8 text-sm font-medium text-deep/70 sm:flex">
      {links.map((link) => (
        <a
          key={link.href}
          href={link.href}
          onClick={(e) => handleClick(e, link.href)}
          className="transition-opacity hover:opacity-70"
        >
          {link.label}
        </a>
      ))}
    </nav>
  );
}
