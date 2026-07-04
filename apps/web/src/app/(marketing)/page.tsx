import {
  BarChart3,
  ChevronDown,
  MousePointerClick,
  QrCode,
  ShieldCheck,
  Tag,
  Users,
  Zap,
} from "lucide-react";
import Image from "next/image";
import Link from "next/link";

import { apiFetch } from "@/lib/api-client";

import { AnimatedSlogan } from "./animated-slogan";
import { ShortenForm } from "./shorten-form";

const featureHighlights = [
  {
    icon: Tag,
    title: "Custom Aliases",
    description: "Pick a short code that matches your brand or campaign.",
  },
  {
    icon: BarChart3,
    title: "Click Analytics",
    description: "See clicks, referrers, and timestamps for every link.",
  },
  {
    icon: QrCode,
    title: "QR Codes",
    description: "Generate and download a scannable QR code for any link.",
  },
];

const steps = [
  {
    number: "1",
    title: "Paste your links",
    description: "Simply drop your long destination URL into the dashboard.",
  },
  {
    number: "2",
    title: "Get a short URL",
    description: "Our engine generates a powerful, compact link instantly.",
  },
  {
    number: "3",
    title: "Share and track",
    description: "Deploy your link across the web and monitor its performance.",
  },
];

const noAccountRows = [
  { label: "Shorten a link instantly", included: true },
  { label: "Fixed 2-day expiry", included: true },
  { label: "Custom alias", included: false },
  { label: "Max click limit", included: false },
  { label: "Click analytics", included: false },
  { label: "QR code", included: false },
];

const accountRows = [
  "Shorten unlimited links",
  "Custom or no expiry",
  "Custom alias",
  "Max click limit",
  "Click analytics",
  "QR code",
];

const faqs = [
  {
    question: "Is ShortKing free?",
    answer:
      "Yes. Shortening links without an account is completely free, and creating an account (also free) unlocks custom aliases, analytics, and QR codes.",
  },
  {
    question: "How long do links last without an account?",
    answer:
      "Links created without logging in expire automatically after 2 days. Sign up to set a custom expiry, or no expiry at all.",
  },
  {
    question: "Can I track clicks on my links?",
    answer:
      "Yes, but only for links created with an account. Anonymous links aren't tied to any dashboard, so there's nowhere to view their analytics.",
  },
  {
    question: "Can I edit a link after creating it?",
    answer:
      "Not yet, links can currently be created and deleted, but not edited after creation.",
  },
];

interface Stats {
  totalLinks: number;
  activeLinks: number;
  totalClicks: number;
  totalQrCodes: number;
  totalUsers: number;
}

function formatStatValue(value: number): string {
  if (value < 10) return String(value);
  return `${Math.floor(value / 10) * 10}+`;
}

export default async function LandingPage() {
  const stats = await apiFetch<Stats>("/api/v1/stats", { cache: "no-store" }).catch(() => null);

  const statItems = stats
    ? [
      { label: "Active Links", value: stats.activeLinks, icon: Tag },
      { label: "Total Clicks", value: stats.totalClicks, icon: MousePointerClick },
      { label: "QR Codes Generated", value: stats.totalQrCodes, icon: QrCode },
      { label: "Total Users", value: stats.totalUsers, icon: Users },
    ]
    : [];

  return (
    <>
      <main className="flex flex-1 flex-col items-center justify-center px-6 py-20 text-center sm:py-28">
        <div className="w-full max-w-5xl">
          <h1 className="whitespace-nowrap text-[clamp(1.25rem,5.2vw,3.25rem)] font-extrabold uppercase leading-tight tracking-tight">
            Shorten your <span className="text-accent">link</span> here, king.
          </h1>
        </div>

        <div className="mt-3 w-full max-w-5xl overflow-hidden">
          <AnimatedSlogan />
        </div>

        <div className="mt-4 flex w-full max-w-2xl flex-col items-center gap-6">
          <div className="flex w-full justify-center">
            <ShortenForm />
          </div>

          <div className="flex flex-wrap items-center justify-center gap-6 text-xs font-semibold tracking-wide text-deep/50 uppercase">
            <span className="flex items-center gap-1.5">
              <Zap className="h-3.5 w-3.5" /> Fast
            </span>
            <span className="flex items-center gap-1.5">
              <ShieldCheck className="h-3.5 w-3.5" /> Secure
            </span>
            <span className="flex items-center gap-1.5">
              <BarChart3 className="h-3.5 w-3.5" /> Analytics
            </span>
          </div>
        </div>
      </main>

      {stats && (
        <section id="analytics" className="w-full px-6 pb-16 sm:px-10">
          <div className="mx-auto max-w-5xl">
            <div className="flex flex-col items-center gap-1 text-center">
              <p className="text-xs font-bold tracking-widest text-deep/50 uppercase">
                Platform Stats
              </p>
              <h2 className="mt-1 text-2xl font-bold text-deep sm:text-3xl">
                Stats at a Glance
              </h2>
            </div>

            <div className="mt-8 grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-4">
              {statItems.map((item) => (
                <div
                  key={item.label}
                  className="group flex flex-col items-center gap-4 rounded-2xl border border-deep/10 bg-white p-6 text-center shadow-sm transition-all duration-200 hover:-translate-y-1 hover:shadow-lg"
                >
                  <span className="flex h-10 w-10 items-center justify-center rounded-full bg-mist/40 text-deep transition-transform duration-200 group-hover:scale-110">
                    <item.icon className="h-5 w-5" />
                  </span>
                  <div>
                    <p className="text-3xl font-extrabold text-deep sm:text-4xl">
                      {formatStatValue(item.value)}
                    </p>
                    <p className="text-sm font-medium text-deep/60">{item.label}</p>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </section>
      )}

      <section id="features" className="w-full px-6 py-16 sm:px-10">
        <div className="mx-auto max-w-4xl">
          <h2 className="text-center text-3xl font-bold tracking-tight text-deep sm:text-4xl">
            Everything You Need
          </h2>
          <p className="mt-2 text-center text-sm text-deep/60">
            Premium tools designed for the modern link sovereign.
          </p>

          <div className="mx-auto mt-10 grid grid-cols-1 gap-6 sm:grid-cols-3">
            {featureHighlights.map((feature) => (
              <div
                key={feature.title}
                className="group flex flex-col items-center gap-3 rounded-2xl border border-deep/10 bg-white p-8 text-center shadow-sm transition-all duration-200 hover:-translate-y-1 hover:shadow-lg"
              >
                <span className="flex h-14 w-14 items-center justify-center rounded-full bg-mist/40 text-deep transition-transform duration-200 group-hover:scale-110">
                  <feature.icon className="h-6 w-6" />
                </span>
                <h3 className="text-lg font-semibold text-deep">{feature.title}</h3>
                <p className="text-sm text-deep/70">{feature.description}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      <section className="w-full px-6 pb-16 sm:px-10">
        <div className="mx-auto max-w-4xl">
          <h2 className="text-center text-3xl font-bold tracking-tight text-deep sm:text-4xl">
            How It Works
          </h2>

          <div className="mx-auto mt-10 grid grid-cols-1 gap-6 sm:grid-cols-3">
            {steps.map((step) => (
              <div key={step.number} className="group flex flex-col items-center gap-3 text-center">
                <span className="flex h-12 w-12 items-center justify-center rounded-full bg-deep text-lg font-extrabold text-bright transition-all duration-200 group-hover:scale-110 group-hover:bg-accent group-hover:text-deep">
                  {step.number}
                </span>
                <h3 className="text-lg font-semibold text-deep">{step.title}</h3>
                <p className="text-sm text-deep/70">{step.description}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      <section id="compare" className="w-full px-6 py-16 sm:px-10">
        <div className="mx-auto max-w-4xl">
          <h2 className="text-center text-3xl font-bold tracking-tight text-deep sm:text-4xl">
            Free forever. More with an account.
          </h2>

          <div className="mx-auto mt-10 grid max-w-3xl grid-cols-1 gap-6 sm:grid-cols-2">
            <div className="flex flex-col gap-4 rounded-3xl border border-deep/10 bg-white p-8 text-left shadow-sm transition-all duration-200 hover:-translate-y-1 hover:shadow-lg">
              <h3 className="text-xl font-bold text-deep">No account</h3>
              <ul className="flex flex-col gap-2.5">
                {noAccountRows.map((row) => (
                  <li key={row.label} className="flex items-center gap-2 text-sm text-deep/80">
                    <span className="w-4">{row.included ? "✓" : "—"}</span>
                    <span>{row.label}</span>
                  </li>
                ))}
              </ul>
            </div>

            <div className="relative flex flex-col gap-4 rounded-3xl border-2 border-bright bg-deep p-8 text-left shadow-xl transition-shadow duration-300 hover:shadow-2xl sm:scale-105">
              <span className="absolute -top-3 left-1/2 -translate-x-1/2 rounded-full bg-bright px-4 py-1 text-xs font-bold tracking-wide text-deep uppercase">
                Recommended
              </span>
              <h3 className="text-xl font-bold text-bright">With an account</h3>
              <ul className="flex flex-col gap-2.5">
                {accountRows.map((label) => (
                  <li key={label} className="flex items-center gap-2 text-sm text-bright/90">
                    <span className="w-4">✓</span>
                    <span>{label}</span>
                  </li>
                ))}
              </ul>
              <Link
                href="/signup"
                className="mt-2 rounded-full bg-bright px-6 py-3 text-center font-semibold text-deep transition-transform duration-200 hover:scale-105 active:scale-95"
              >
                Sign up free
              </Link>
            </div>
          </div>
        </div>
      </section>

      <section id="faq" className="w-full px-6 py-16 sm:px-10">
        <div className="mx-auto max-w-3xl">
          <h2 className="text-center text-3xl font-bold tracking-tight text-deep sm:text-4xl">
            Frequently Asked Questions
          </h2>

          <div className="mx-auto mt-10 flex max-w-2xl flex-col gap-3">
            {faqs.map((faq) => (
              <details
                key={faq.question}
                className="group rounded-2xl border border-deep/10 bg-white p-5 shadow-sm transition-all duration-200 hover:border-deep/20 hover:shadow-md open:pb-5"
              >
                <summary className="flex cursor-pointer list-none items-center justify-between gap-4 font-semibold text-deep marker:content-none">
                  {faq.question}
                  <ChevronDown className="h-4 w-4 shrink-0 text-deep/50 transition-transform duration-200 group-open:rotate-180" />
                </summary>
                <p className="mt-2 text-sm text-deep/70">{faq.answer}</p>
              </details>
            ))}
          </div>
        </div>
      </section>

      <section id="about" className="w-full px-6 py-16 sm:px-10">
        <div className="mx-auto flex max-w-2xl flex-col items-center gap-4 text-center">
          <p className="text-xs font-bold tracking-widest text-deep/50 uppercase">About Us</p>
          <Image
            src="/logo-two-line.png"
            alt="Short King"
            width={1920}
            height={1080}
            className="h-28 w-auto"
          />
          <p className="text-sm text-deep/70 sm:text-base">
            Long, clunky links are a crime against good taste, so we started a weekend rebellion
            against them. ShortKing gives every link the kingly treatment it deserves: fast
            redirects, clean analytics, and zero nonsense. No paywalls on the basics, no dark
            patterns, no fine print. Just a tool built by people who got tired of ugly links and
            decided to do something about it.
          </p>
        </div>
      </section>

      <footer className="w-full bg-deep px-6 py-8 sm:px-10">
        <div className="mx-auto flex max-w-4xl flex-col items-center justify-between gap-4 text-sm text-bright/80 sm:flex-row">
          <span className="font-semibold text-bright">SHORT KING</span>
          <span>© 2026 ShortKing</span>
        </div>
      </footer>
    </>
  );
}
