import Link from "next/link";

import { apiFetchAuthed } from "@/lib/api-server";
import type { LinkDetail } from "@/types/link";

import { EditLinkForm } from "./edit-link-form";

export default async function EditLinkPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  const { link } = await apiFetchAuthed<LinkDetail>(`/api/v1/links/${id}`);

  return (
    <main className="flex flex-1 flex-col gap-6 px-8 py-10">
      <div>
        <Link href={`/dashboard/links/${id}`} className="text-sm text-deep/70 underline">
          &larr; Back to link
        </Link>
        <h1 className="mt-2 text-2xl font-extrabold tracking-tight text-deep">Edit link</h1>
      </div>
      <EditLinkForm link={link} />
    </main>
  );
}
