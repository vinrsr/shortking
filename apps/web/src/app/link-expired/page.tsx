import { Clock } from "lucide-react";

import { LinkStatusPage } from "@/components/link-status-page";

export default function LinkExpiredPage() {
  return (
    <LinkStatusPage
      icon={Clock}
      title="This link has expired"
      message="The short link you followed is no longer active. Reach out to whoever shared it with you and ask them for an updated link."
    />
  );
}
