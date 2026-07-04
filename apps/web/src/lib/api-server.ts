import "server-only";

import { apiFetch } from "@/lib/api-client";
import { getAccessToken } from "@/lib/session";

// Server Components/Route Handlers can call the Go API directly (server to
// server, no CORS involved) by forwarding the httpOnly access token cookie
// as a Bearer header. Never expose this to client code.
export async function apiFetchAuthed<T>(path: string, options: RequestInit = {}): Promise<T> {
  const accessToken = await getAccessToken();
  return apiFetch<T>(path, { ...options, accessToken });
}
