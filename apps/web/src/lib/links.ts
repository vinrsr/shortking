import type { Link } from "@/types/link";

// A link can be inactive without isActive being false: it can also be
// expired by date, or have hit its max-click limit. Anywhere that shows a
// link's status should go through this, not just check link.isActive.
export function isCurrentlyActive(link: Link): boolean {
  if (!link.isActive) return false;
  if (link.expiresAt && new Date(link.expiresAt) <= new Date()) return false;
  if (link.maxClicks != null && link.clickCount >= link.maxClicks) return false;
  return true;
}
