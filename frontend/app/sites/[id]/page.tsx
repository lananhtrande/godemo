import Link from "next/link";
import { notFound } from "next/navigation";
import { getSite } from "@/lib/api";
import { LiveSite } from "./LiveSite";

export const dynamic = "force-dynamic";

// Server Component: loads the initial state on the server (fast first paint),
// then hands it to a Client Component that keeps it live.
export default async function SitePage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  const site = await getSite(id);
  if (!site) notFound();

  return (
    <>
      <Link href="/" className="back">
        ← All sites
      </Link>
      <LiveSite initial={site} />
    </>
  );
}
