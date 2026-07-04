import { NextRequest, NextResponse } from "next/server";

import { apiFetch, ApiError } from "@/lib/api-client";
import { getAccessToken } from "@/lib/session";

export async function POST(request: NextRequest) {
  const accessToken = await getAccessToken();
  if (!accessToken) {
    return NextResponse.json({ error: "not authenticated" }, { status: 401 });
  }

  const body = await request.json();

  try {
    const link = await apiFetch("/api/v1/links", {
      method: "POST",
      body: JSON.stringify(body),
      accessToken,
    });
    return NextResponse.json(link, { status: 201 });
  } catch (err) {
    if (err instanceof ApiError) {
      return NextResponse.json({ error: err.message }, { status: err.status });
    }
    return NextResponse.json({ error: "failed to create link" }, { status: 500 });
  }
}
