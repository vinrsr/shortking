"use client";

import { useState } from "react";

export function QrCodeReveal({
  linkId,
  shortCode,
  generated = false,
}: {
  linkId: string;
  shortCode: string;
  generated?: boolean;
}) {
  const [show, setShow] = useState(generated);

  if (!show) {
    return (
      <button
        onClick={() => {
          setShow(true);
          fetch(`/api/links/${linkId}/qrcode/generations`, { method: "POST" }).catch(() => {});
        }}
        className="w-fit rounded-full bg-deep px-4 py-2 text-sm font-semibold text-bright transition-transform duration-200 hover:scale-105 active:scale-95"
      >
        Generate QR code
      </button>
    );
  }

  const qrSrc = `/api/links/${linkId}/qrcode`;

  return (
    <div className="flex flex-col gap-3">
      {/* eslint-disable-next-line @next/next/no-img-element */}
      <img
        src={qrSrc}
        alt={`QR code for ${shortCode}`}
        width={200}
        height={200}
        className="rounded-2xl bg-sand p-3"
      />
      <a
        href={qrSrc}
        download={`${shortCode}-qrcode.png`}
        className="w-fit rounded-full border-2 border-deep/20 px-4 py-2 text-sm font-medium text-deep"
      >
        Download QR code
      </a>
    </div>
  );
}
