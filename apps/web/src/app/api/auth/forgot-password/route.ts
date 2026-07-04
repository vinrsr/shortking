import { NextRequest, NextResponse } from "next/server";

import { apiFetch, ApiError } from "@/lib/api-client";

export async function POST(request: NextRequest) {
  const body = await request.json();

  try {
    const data = await apiFetch<{ message: string }>("/api/v1/auth/forgot-password", {
      method: "POST",
      body: JSON.stringify(body),
    });
    return NextResponse.json(data);
  } catch (err) {
    if (err instanceof ApiError) {
      return NextResponse.json({ error: err.message }, { status: err.status });
    }
    return NextResponse.json({ error: "request failed" }, { status: 500 });
  }
}
