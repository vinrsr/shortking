import "server-only";

import { cookies } from "next/headers";

import { ACCESS_TOKEN_COOKIE, REFRESH_TOKEN_COOKIE } from "@/lib/session";

const ACCESS_TOKEN_MAX_AGE = 15 * 60; // seconds, matches the API's access token TTL
const REFRESH_TOKEN_MAX_AGE = 30 * 24 * 60 * 60; // seconds

export async function setAuthCookies(accessToken: string, refreshToken: string) {
  const store = await cookies();
  const shared = {
    httpOnly: true,
    secure: process.env.NODE_ENV === "production",
    sameSite: "lax" as const,
    path: "/",
  };
  store.set(ACCESS_TOKEN_COOKIE, accessToken, { ...shared, maxAge: ACCESS_TOKEN_MAX_AGE });
  store.set(REFRESH_TOKEN_COOKIE, refreshToken, { ...shared, maxAge: REFRESH_TOKEN_MAX_AGE });
}

export async function clearAuthCookies() {
  const store = await cookies();
  store.delete(ACCESS_TOKEN_COOKIE);
  store.delete(REFRESH_TOKEN_COOKIE);
}
