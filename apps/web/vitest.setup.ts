import "@testing-library/jest-dom/vitest";

import { cleanup } from "@testing-library/react";
import { afterEach } from "vitest";

// @testing-library/react's auto-cleanup only self-registers under Jest;
// under Vitest it has to be wired up explicitly, or DOM from one test leaks
// into the next and queries like getByRole start matching multiple elements.
afterEach(() => {
  cleanup();
});
