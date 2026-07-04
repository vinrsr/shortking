import { afterEach, describe, expect, it, vi } from "vitest";

import { apiFetch, ApiError } from "./api-client";

function mockFetchOnce(response: Partial<Response> & { text?: () => Promise<string> }) {
  const fetchMock = vi.fn().mockResolvedValue(response as Response);
  vi.stubGlobal("fetch", fetchMock);
  return fetchMock;
}

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("apiFetch", () => {
  it("returns parsed JSON on success", async () => {
    mockFetchOnce({
      ok: true,
      status: 200,
      json: async () => ({ hello: "king" }),
    });

    const data = await apiFetch<{ hello: string }>("/api/v1/whatever");

    expect(data).toEqual({ hello: "king" });
  });

  it("returns undefined for a 204 No Content response", async () => {
    mockFetchOnce({ ok: true, status: 204 });

    const data = await apiFetch("/api/v1/whatever");

    expect(data).toBeUndefined();
  });

  it("extracts the .error field from a JSON error body instead of double-encoding it", async () => {
    // Regression: apiFetch used to throw the raw JSON text as the message,
    // so callers that wrapped it in another {"error": message} response
    // produced a nested, double-encoded JSON string instead of clean text.
    mockFetchOnce({
      ok: false,
      status: 401,
      statusText: "Unauthorized",
      text: async () => JSON.stringify({ error: "service: invalid credentials" }),
    });

    await expect(apiFetch("/api/v1/auth/login")).rejects.toMatchObject({
      status: 401,
      message: "service: invalid credentials",
    });
  });

  it("falls back to the raw body when the error response isn't JSON", async () => {
    mockFetchOnce({
      ok: false,
      status: 500,
      statusText: "Internal Server Error",
      text: async () => "boom",
    });

    await expect(apiFetch("/api/v1/whatever")).rejects.toMatchObject({
      status: 500,
      message: "boom",
    });
  });

  it("falls back to statusText when the error body is empty", async () => {
    mockFetchOnce({
      ok: false,
      status: 503,
      statusText: "Service Unavailable",
      text: async () => "",
    });

    await expect(apiFetch("/api/v1/whatever")).rejects.toMatchObject({
      status: 503,
      message: "Service Unavailable",
    });
  });

  it("throws an ApiError instance, not just an object shape", async () => {
    mockFetchOnce({
      ok: false,
      status: 404,
      statusText: "Not Found",
      text: async () => JSON.stringify({ error: "not found" }),
    });

    await expect(apiFetch("/api/v1/whatever")).rejects.toBeInstanceOf(ApiError);
  });

  it("sends the Authorization header when an access token is given", async () => {
    const fetchMock = mockFetchOnce({ ok: true, status: 200, json: async () => ({}) });

    await apiFetch("/api/v1/links", { accessToken: "token-123" });

    const [, init] = fetchMock.mock.calls[0];
    expect(init.headers.Authorization).toBe("Bearer token-123");
  });

  it("omits the Authorization header when no access token is given", async () => {
    const fetchMock = mockFetchOnce({ ok: true, status: 200, json: async () => ({}) });

    await apiFetch("/api/v1/stats");

    const [, init] = fetchMock.mock.calls[0];
    expect(init.headers.Authorization).toBeUndefined();
  });
});
