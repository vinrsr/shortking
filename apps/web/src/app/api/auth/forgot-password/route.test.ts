import { NextRequest } from "next/server";
import { afterEach, describe, expect, it, vi } from "vitest";

import { POST } from "./route";

function makeRequest(body: unknown) {
  return new NextRequest("http://localhost/api/auth/forgot-password", {
    method: "POST",
    body: JSON.stringify(body),
  });
}

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("POST /api/auth/forgot-password", () => {
  it("forwards the API's generic message on success", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({
        ok: true,
        status: 200,
        json: async () => ({ message: "if that email exists, a reset link has been sent" }),
      }),
    );

    const res = await POST(makeRequest({ email: "someone@example.com" }));

    expect(res.status).toBe(200);
    expect(await res.json()).toEqual({
      message: "if that email exists, a reset link has been sent",
    });
  });

  it("forwards a non-200 error status and clean message from the API", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({
        ok: false,
        status: 400,
        statusText: "Bad Request",
        text: async () => JSON.stringify({ error: "email is required" }),
      }),
    );

    const res = await POST(makeRequest({}));

    expect(res.status).toBe(400);
    expect(await res.json()).toEqual({ error: "email is required" });
  });
});
