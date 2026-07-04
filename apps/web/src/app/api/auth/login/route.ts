import { NextRequest, NextResponse } from "next/server";

import { apiFetch, ApiError } from "@/lib/api-client";
import { setAuthCookies } from "@/lib/auth-cookies";

export async function POST(request: NextRequest) {
  const body = await request.json();

  try {
    const data = await apiFetch<{
      user: { id: string; email: string; displayName: string };
      accessToken: string;
      refreshToken: string;
    }>("/api/v1/auth/login", {
      method: "POST",
      body: JSON.stringify(body),
    });

    await setAuthCookies(data.accessToken, data.refreshToken);
    return NextResponse.json({ user: data.user });
  } catch (err) {
    if (err instanceof ApiError) {
      return NextResponse.json({ error: err.message }, { status: err.status });
    }
    return NextResponse.json({ error: "login failed" }, { status: 500 });
  }
}
