import { SearchX } from "lucide-react";

import { LinkStatusPage } from "@/components/link-status-page";

export default function LinkNotFoundPage() {
  return (
    <LinkStatusPage
      icon={SearchX}
      title="Link not found"
      message="This short link doesn't exist, or has been deactivated by its owner. Double-check the URL, or reach out to whoever shared it with you."
    />
  );
}
