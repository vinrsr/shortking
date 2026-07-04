import { NextResponse } from "next/server";

import { apiFetch } from "@/lib/api-client";
import { clearAuthCookies } from "@/lib/auth-cookies";
import { getRefreshToken } from "@/lib/session";

export async function POST() {
  const refreshToken = await getRefreshToken();

  if (refreshToken) {
    try {
      await apiFetch("/api/v1/auth/logout", {
        method: "POST",
        body: JSON.stringify({ refreshToken }),
      });
    } catch {
      // best-effort, still clear local cookies even if the API call fails
    }
  }

  await clearAuthCookies();
  return NextResponse.json({ ok: true });
}
