"use client";

import { useState } from "react";

export function CopyLinkButton({ url }: { url: string }) {
  const [copied, setCopied] = useState(false);

  async function handleCopy() {
    await navigator.clipboard.writeText(url);
    setCopied(true);
    setTimeout(() => setCopied(false), 1500);
  }

  return (
    <button
      onClick={handleCopy}
      className="rounded-md border border-gray-300 px-2 py-1 text-xs dark:border-gray-700"
    >
      {copied ? "Copied!" : "Copy"}
    </button>
  );
}
