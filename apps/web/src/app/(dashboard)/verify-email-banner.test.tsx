import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";

import { VerifyEmailBanner } from "./verify-email-banner";

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("VerifyEmailBanner", () => {
  it("prompts to verify and posts to resend-verification when clicked", async () => {
    const fetchMock = vi.fn().mockResolvedValue({ ok: true, status: 200, json: async () => ({}) });
    vi.stubGlobal("fetch", fetchMock);

    render(<VerifyEmailBanner />);
    expect(screen.getByText(/verify your email/i)).toBeInTheDocument();

    await userEvent.click(screen.getByRole("button", { name: /resend email/i }));

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/auth/resend-verification",
      expect.objectContaining({ method: "POST" }),
    );
    expect(await screen.findByText(/verification email sent/i)).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /resend email/i })).not.toBeInTheDocument();
  });
});
