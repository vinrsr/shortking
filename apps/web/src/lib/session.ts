import "server-only";

import { cookies } from "next/headers";

export const ACCESS_TOKEN_COOKIE = "shortking_access_token";
export const REFRESH_TOKEN_COOKIE = "shortking_refresh_token";

export async function getAccessToken(): Promise<string | undefined> {
  const store = await cookies();
  return store.get(ACCESS_TOKEN_COOKIE)?.value;
}

export async function getRefreshToken(): Promise<string | undefined> {
  const store = await cookies();
  return store.get(REFRESH_TOKEN_COOKIE)?.value;
}
