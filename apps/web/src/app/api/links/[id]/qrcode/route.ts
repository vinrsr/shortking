import { NextResponse } from "next/server";

import { getAccessToken } from "@/lib/session";

const API_URL = process.env.INTERNAL_API_URL ?? process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

export async function GET(
  _request: Request,
  { params }: { params: Promise<{ id: string }> },
) {
  const accessToken = await getAccessToken();
  if (!accessToken) {
    return NextResponse.json({ error: "not authenticated" }, { status: 401 });
  }

  const { id } = await params;

  const res = await fetch(`${API_URL}/api/v1/links/${id}/qrcode`, {
    headers: { Authorization: `Bearer ${accessToken}` },
  });

  if (!res.ok) {
    return NextResponse.json({ error: "failed to generate QR code" }, { status: res.status });
  }

  const png = await res.arrayBuffer();
  return new NextResponse(png, { headers: { "Content-Type": "image/png" } });
}
