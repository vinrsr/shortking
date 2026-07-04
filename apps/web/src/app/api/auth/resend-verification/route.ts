import { NextRequest, NextResponse } from "next/server";

import { apiFetch, ApiError } from "@/lib/api-client";
import { getAccessToken } from "@/lib/session";

export async function POST(request: NextRequest) {
  let body = await request.json().catch(() => ({}));

  if (!body.email) {
    const accessToken = await getAccessToken();
    if (!accessToken) {
      return NextResponse.json({ error: "not authenticated" }, { status: 401 });
    }
    const me = await apiFetch<{ email: string }>("/api/v1/me", { accessToken }).catch(() => null);
    if (!me) {
      return NextResponse.json({ error: "not authenticated" }, { status: 401 });
    }
    body = { email: me.email };
  }

  try {
    const data = await apiFetch<{ message: string }>("/api/v1/auth/resend-verification", {
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
