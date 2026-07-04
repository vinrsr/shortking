import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";

import { ShortenForm } from "./shorten-form";

vi.mock("gsap", () => ({
  gsap: {
    to: vi.fn(() => ({ kill: vi.fn() })),
    fromTo: vi.fn(),
    set: vi.fn(),
  },
}));

function mockFetchOnce(response: Partial<Response>) {
  vi.stubGlobal(
    "fetch",
    vi.fn().mockResolvedValue(response as Response),
  );
}

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("ShortenForm", () => {
  it("shows the shortened link on success", async () => {
    mockFetchOnce({
      ok: true,
      status: 201,
      json: async () => ({
        id: "1",
        shortCode: "abc1234",
        shortUrl: "https://sk.io/abc1234",
        destination: "https://example.com",
        clickCount: 0,
        isActive: true,
        createdAt: new Date().toISOString(),
      }),
    });

    render(<ShortenForm />);
    await userEvent.type(screen.getByPlaceholderText(/paste your long url/i), "https://example.com");
    await userEvent.click(screen.getByRole("button", { name: /shorten/i }));

    expect(await screen.findByText("sk.io/abc1234")).toBeInTheDocument();
  });

  it("shows a plain error message for a normal failure", async () => {
    mockFetchOnce({
      ok: false,
      status: 400,
      json: async () => ({ error: "destination is not reachable" }),
    });

    render(<ShortenForm />);
    await userEvent.type(screen.getByPlaceholderText(/paste your long url/i), "https://example.com");
    await userEvent.click(screen.getByRole("button", { name: /shorten/i }));

    expect(await screen.findByText("destination is not reachable")).toBeInTheDocument();
    expect(screen.queryByRole("link", { name: /sign up free/i })).not.toBeInTheDocument();
  });

  it("shows a highlighted signup CTA when the daily limit (429) is hit", async () => {
    mockFetchOnce({
      ok: false,
      status: 429,
      json: async () => ({
        error: "you've reached today's free shorten limit — sign up for unlimited links",
      }),
    });

    render(<ShortenForm />);
    await userEvent.type(screen.getByPlaceholderText(/paste your long url/i), "https://example.com");
    await userEvent.click(screen.getByRole("button", { name: /shorten/i }));

    await waitFor(() => {
      expect(screen.getByRole("link", { name: /sign up free/i })).toBeInTheDocument();
    });
    expect(screen.getByText(/reached today's free shorten limit/i)).toBeInTheDocument();
  });

  it("clears a previous limit-reached CTA on the next successful submit", async () => {
    mockFetchOnce({
      ok: false,
      status: 429,
      json: async () => ({ error: "daily limit reached" }),
    });
    render(<ShortenForm />);
    const input = screen.getByPlaceholderText(/paste your long url/i);
    await userEvent.type(input, "https://example.com");
    await userEvent.click(screen.getByRole("button", { name: /shorten/i }));
    await screen.findByRole("link", { name: /sign up free/i });

    mockFetchOnce({
      ok: true,
      status: 201,
      json: async () => ({
        id: "2",
        shortCode: "xyz9876",
        shortUrl: "https://sk.io/xyz9876",
        destination: "https://example.com",
        clickCount: 0,
        isActive: true,
        createdAt: new Date().toISOString(),
      }),
    });
    await userEvent.click(screen.getByRole("button", { name: /shorten/i }));

    expect(await screen.findByText("sk.io/xyz9876")).toBeInTheDocument();
    expect(screen.queryByRole("link", { name: /sign up free/i })).not.toBeInTheDocument();
  });
});
