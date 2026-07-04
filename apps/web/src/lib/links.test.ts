import { describe, expect, it } from "vitest";

import { isCurrentlyActive } from "./links";
import type { Link } from "@/types/link";

const baseLink: Link = {
  id: "1",
  shortCode: "abc1234",
  shortUrl: "https://sk.io/abc1234",
  destination: "https://example.com",
  clickCount: 0,
  isActive: true,
  qrGenerated: false,
  createdAt: new Date().toISOString(),
};

describe("isCurrentlyActive", () => {
  it("is false when isActive is false, regardless of everything else", () => {
    expect(isCurrentlyActive({ ...baseLink, isActive: false })).toBe(false);
  });

  it("is true for a link with no expiry or max-clicks limit", () => {
    expect(isCurrentlyActive(baseLink)).toBe(true);
  });

  it("is false once expiresAt is in the past", () => {
    const past = new Date(Date.now() - 1000).toISOString();
    expect(isCurrentlyActive({ ...baseLink, expiresAt: past })).toBe(false);
  });

  it("is true while expiresAt is still in the future", () => {
    const future = new Date(Date.now() + 1000 * 60 * 60).toISOString();
    expect(isCurrentlyActive({ ...baseLink, expiresAt: future })).toBe(true);
  });

  it("is false once clickCount reaches maxClicks", () => {
    expect(isCurrentlyActive({ ...baseLink, maxClicks: 5, clickCount: 5 })).toBe(false);
  });

  it("is false once clickCount exceeds maxClicks", () => {
    expect(isCurrentlyActive({ ...baseLink, maxClicks: 5, clickCount: 9 })).toBe(false);
  });

  it("is true while clickCount is still under maxClicks", () => {
    expect(isCurrentlyActive({ ...baseLink, maxClicks: 5, clickCount: 4 })).toBe(true);
  });

  it("treats maxClicks: 0 as a real limit, not 'no limit'", () => {
    expect(isCurrentlyActive({ ...baseLink, maxClicks: 0, clickCount: 0 })).toBe(false);
  });
});
