import { NextRequest } from "next/server";
import { afterEach, describe, expect, it, vi } from "vitest";

import { POST } from "./route";

function makeRequest(body: unknown) {
  return new NextRequest("http://localhost/api/auth/verify-email", {
    method: "POST",
    body: JSON.stringify(body),
  });
}

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("POST /api/auth/verify-email", () => {
  it("returns 204 when the API confirms the token", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({ ok: true, status: 204, text: async () => "" }),
    );

    const res = await POST(makeRequest({ token: "good-token" }));

    expect(res.status).toBe(204);
  });

  it("forwards the API's error status and message for a bad token", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({
        ok: false,
        status: 401,
        statusText: "Unauthorized",
        text: async () => JSON.stringify({ error: "service: invalid or expired token" }),
      }),
    );

    const res = await POST(makeRequest({ token: "bad-token" }));

    expect(res.status).toBe(401);
    expect(await res.json()).toEqual({ error: "service: invalid or expired token" });
  });
});
