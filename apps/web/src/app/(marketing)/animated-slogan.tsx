"use client";

import { gsap } from "gsap";
import { useEffect, useRef, useState } from "react";

//TODO: make these slogans more funnier
const SLOGANS = [
  "Doesn't matter if she's taller. Your links should be shorter",
  "We shorten links. Not your confidence",
  "No cap, your links are way too long",
  "Short king energy. Tiny links. Massive aura.",
  "If your link wraps to the next line, we need to talk",
  "Your link has more characters than your group chat",
  "She's taller. Your links aren't",
];

const TYPE_MS = 32;
const DELETE_MS = 18;
const HOLD_MS = 2200;
const HOLD_EMPTY_MS = 400;

type Phase = "typing" | "holding" | "deleting" | "holdingEmpty";

export function AnimatedSlogan() {
  const [index, setIndex] = useState(0);
  const [phase, setPhase] = useState<Phase>("typing");
  const [charCount, setCharCount] = useState(0);
  const cursorRef = useRef<HTMLSpanElement>(null);

  useEffect(() => {
    if (!cursorRef.current) return;
    const tween = gsap.to(cursorRef.current, {
      opacity: 0,
      duration: 0.5,
      repeat: -1,
      yoyo: true,
      ease: "power1.inOut",
    });
    return () => {
      tween.kill();
    };
  }, []);

  useEffect(() => {
    const text = SLOGANS[index];

    if (phase === "typing") {
      if (charCount < text.length) {
        const timeout = setTimeout(() => setCharCount((c) => c + 1), TYPE_MS);
        return () => clearTimeout(timeout);
      }
      const timeout = setTimeout(() => setPhase("holding"), 0);
      return () => clearTimeout(timeout);
    }

    if (phase === "holding") {
      const timeout = setTimeout(() => setPhase("deleting"), HOLD_MS);
      return () => clearTimeout(timeout);
    }

    if (phase === "deleting") {
      if (charCount > 0) {
        const timeout = setTimeout(() => setCharCount((c) => c - 1), DELETE_MS);
        return () => clearTimeout(timeout);
      }
      const timeout = setTimeout(() => setPhase("holdingEmpty"), 0);
      return () => clearTimeout(timeout);
    }

    const timeout = setTimeout(() => {
      setIndex((i) => (i + 1) % SLOGANS.length);
      setPhase("typing");
    }, HOLD_EMPTY_MS);
    return () => clearTimeout(timeout);
  }, [phase, charCount, index]);

  return (
    <p className="whitespace-nowrap text-[clamp(0.8rem,3vw,2rem)] text-deep/60">
      {SLOGANS[index].slice(0, charCount)}
      <span ref={cursorRef} className="text-accent">
        |
      </span>
    </p>
  );
}
