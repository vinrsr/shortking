"use client";

import { gsap } from "gsap";
import { ArrowRight, Link2, Loader2 } from "lucide-react";
import Link from "next/link";
import { useEffect, useRef, useState } from "react";

import { CopyLinkButton } from "@/components/copy-link-button";
import type { Link as ShortLink } from "@/types/link";

export function ShortenForm() {
  const [url, setUrl] = useState("");
  const [result, setResult] = useState<ShortLink | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [limitReached, setLimitReached] = useState(false);
  const [submitting, setSubmitting] = useState(false);

  const buttonRef = useRef<HTMLButtonElement>(null);
  const spinnerRef = useRef<SVGSVGElement>(null);
  const resultRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!submitting || !spinnerRef.current) return;
    const tween = gsap.to(spinnerRef.current, {
      rotate: 360,
      duration: 0.7,
      repeat: -1,
      ease: "linear",
    });
    return () => {
      tween.kill();
    };
  }, [submitting]);

  useEffect(() => {
    if (result && resultRef.current) {
      gsap.fromTo(
        resultRef.current,
        { opacity: 0, y: 12, scale: 0.98 },
        { opacity: 1, y: 0, scale: 1, duration: 0.5, ease: "power3.out" },
      );
    }
  }, [result]);

  async function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setError(null);
    setLimitReached(false);
    setSubmitting(true);

    if (buttonRef.current) {
      gsap.to(buttonRef.current, {
        scale: 0.9,
        duration: 0.12,
        yoyo: true,
        repeat: 1,
        ease: "power1.inOut",
        onComplete: () => gsap.set(buttonRef.current, { clearProps: "transform" }),
      });
    }

    try {
      const res = await fetch("/api/shorten", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ destination: url }),
      });

      if (!res.ok) {
        const data = await res.json().catch(() => ({}));
        setError(data.error ?? "Failed to shorten link");
        setLimitReached(res.status === 429);
        return;
      }

      setResult(await res.json());
    } finally {
      setSubmitting(false);
    }
  }

  if (result) {
    return (
      <div
        ref={resultRef}
        className="flex w-full max-w-xl flex-col items-center gap-3 rounded-2xl border border-deep/15 bg-deep/5 p-6"
      >
        <div className="flex items-center gap-2">
          <a
            href={result.shortUrl}
            target="_blank"
            rel="noreferrer"
            className="text-xl font-semibold underline"
          >
            {result.shortUrl.replace(/^https?:\/\//, "")}
          </a>
          <CopyLinkButton url={result.shortUrl} />
        </div>
        <p className="text-sm text-deep/70">
          Expires in 2 days,{" "}
          <Link href="/signup" className="underline">
            sign up
          </Link>{" "}
          to make it permanent, pick a custom alias, and track clicks.
        </p>
        <button
          onClick={() => {
            setResult(null);
            setUrl("");
          }}
          className="text-sm text-deep/70 underline"
        >
          Shorten another link
        </button>
      </div>
    );
  }

  return (
    <form onSubmit={handleSubmit} className="flex w-full max-w-xl flex-col items-center gap-3">
      <div className="relative w-full">
        <Link2 className="pointer-events-none absolute top-1/2 left-5 h-5 w-5 -translate-y-1/2 text-deep/40" />
        <input
          type="url"
          value={url}
          onChange={(e) => setUrl(e.target.value)}
          placeholder="Paste your long URL here"
          required
          className="w-full rounded-full border border-deep/10 bg-white py-4 pl-12 pr-16 text-deep shadow-sm placeholder:text-deep/40 focus:border-deep/30 focus:outline-none sm:pr-36"
        />
        <button
          ref={buttonRef}
          type="submit"
          disabled={submitting}
          className="absolute top-1/2 right-2 flex h-9 w-9 -translate-y-1/2 cursor-pointer items-center justify-center rounded-full bg-deep text-sm font-bold tracking-wide text-bright uppercase transition-transform duration-300 ease-out hover:scale-105 active:scale-95 active:duration-150 disabled:cursor-not-allowed disabled:opacity-50 sm:h-10 sm:w-20"
        >
          {submitting ? (
            <Loader2 ref={spinnerRef} className="h-4 w-4" />
          ) : (
            <>
              <ArrowRight className="h-4 w-4 sm:hidden" />
              <span className="hidden sm:inline">Shorten</span>
            </>
          )}
        </button>
      </div>
      {error &&
        (limitReached ? (
          <div className="flex w-full flex-col items-center gap-3 rounded-2xl border border-accent bg-accent/20 p-5 text-center">
            <p className="text-sm font-medium text-deep">{error}</p>
            <Link
              href="/signup"
              className="rounded-full bg-deep px-6 py-2.5 text-sm font-semibold text-bright transition-transform duration-200 hover:scale-105 active:scale-95"
            >
              Sign up free
            </Link>
          </div>
        ) : (
          <p className="text-sm text-deep">{error}</p>
        ))}
    </form>
  );
}
