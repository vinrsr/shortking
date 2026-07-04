import { NextRequest, NextResponse } from "next/server";

import { apiFetch, ApiError } from "@/lib/api-client";

export async function POST(request: NextRequest) {
  const body = await request.json();

  try {
    await apiFetch("/api/v1/auth/reset-password", {
      method: "POST",
      body: JSON.stringify(body),
    });
    return new NextResponse(null, { status: 204 });
  } catch (err) {
    if (err instanceof ApiError) {
      return NextResponse.json({ error: err.message }, { status: err.status });
    }
    return NextResponse.json({ error: "reset failed" }, { status: 500 });
  }
}
