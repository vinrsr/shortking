import { NextRequest, NextResponse } from "next/server";

import { apiFetch, ApiError } from "@/lib/api-client";

export async function POST(request: NextRequest) {
  const body = await request.json();

  try {
    const link = await apiFetch("/api/v1/shorten", {
      method: "POST",
      body: JSON.stringify(body),
    });
    return NextResponse.json(link, { status: 201 });
  } catch (err) {
    if (err instanceof ApiError) {
      return NextResponse.json({ error: err.message }, { status: err.status });
    }
    return NextResponse.json({ error: "failed to shorten link" }, { status: 500 });
  }
}
