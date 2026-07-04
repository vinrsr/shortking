import { NextResponse } from "next/server";

import { apiFetch, ApiError } from "@/lib/api-client";
import { setAuthCookies } from "@/lib/auth-cookies";
import { getRefreshToken } from "@/lib/session";

export async function POST() {
  const refreshToken = await getRefreshToken();
  if (!refreshToken) {
    return NextResponse.json({ error: "no refresh token" }, { status: 401 });
  }

  try {
    const data = await apiFetch<{ accessToken: string; refreshToken: string }>(
      "/api/v1/auth/refresh",
      {
        method: "POST",
        body: JSON.stringify({ refreshToken }),
      },
    );

    await setAuthCookies(data.accessToken, data.refreshToken);
    return NextResponse.json({ ok: true });
  } catch (err) {
    if (err instanceof ApiError) {
      return NextResponse.json({ error: err.message }, { status: err.status });
    }
    return NextResponse.json({ error: "refresh failed" }, { status: 500 });
  }
}
