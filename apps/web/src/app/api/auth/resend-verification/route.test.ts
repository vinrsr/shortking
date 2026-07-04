import { NextRequest } from "next/server";
import { afterEach, describe, expect, it, vi } from "vitest";

import { getAccessToken } from "@/lib/session";

import { POST } from "./route";

vi.mock("@/lib/session", () => ({
  getAccessToken: vi.fn(),
}));

function makeRequest(body: unknown) {
  return new NextRequest("http://localhost/api/auth/resend-verification", {
    method: "POST",
    body: JSON.stringify(body),
  });
}

afterEach(() => {
  vi.unstubAllGlobals();
  vi.mocked(getAccessToken).mockReset();
});

describe("POST /api/auth/resend-verification", () => {
  it("uses the email from the request body when one is given", async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValue({ ok: true, status: 200, json: async () => ({ message: "sent" }) });
    vi.stubGlobal("fetch", fetchMock);

    const res = await POST(makeRequest({ email: "explicit@example.com" }));

    expect(res.status).toBe(200);
    const [, init] = fetchMock.mock.calls[0];
    expect(JSON.parse(init.body)).toEqual({ email: "explicit@example.com" });
    expect(getAccessToken).not.toHaveBeenCalled();
  });

  it("falls back to the logged-in user's email when the body has none", async () => {
    vi.mocked(getAccessToken).mockResolvedValue("token-123");
    const fetchMock = vi.fn().mockImplementation((url: string) => {
      if (url.includes("/api/v1/me")) {
        return Promise.resolve({ ok: true, status: 200, json: async () => ({ email: "me@example.com" }) });
      }
      return Promise.resolve({ ok: true, status: 200, json: async () => ({ message: "sent" }) });
    });
    vi.stubGlobal("fetch", fetchMock);

    const res = await POST(makeRequest({}));

    expect(res.status).toBe(200);
    const resendCall = fetchMock.mock.calls.find((call) =>
      String(call[0]).includes("/api/v1/auth/resend-verification"),
    );
    expect(resendCall).toBeDefined();
    expect(JSON.parse(resendCall![1].body)).toEqual({ email: "me@example.com" });
  });

  it("returns 401 when there's no email and no authenticated session", async () => {
    vi.mocked(getAccessToken).mockResolvedValue(undefined);

    const res = await POST(makeRequest({}));

    expect(res.status).toBe(401);
  });
});
