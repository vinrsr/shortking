import { Spinner } from "@/components/spinner";

export default function Loading() {
  return (
    <div className="flex flex-1 items-center justify-center py-24">
      <Spinner className="h-8 w-8 text-deep/40" />
    </div>
  );
}
