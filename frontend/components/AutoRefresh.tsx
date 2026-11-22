"use client";

import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";

// Re-runs the Server Component on the server and swaps in the new result,
// without a full page reload. Good enough for an overview that may lag a few seconds.
export function AutoRefresh({ seconds }: { seconds: number }) {
  const router = useRouter();
  const [lastRefresh, setLastRefresh] = useState<Date | null>(null);

  useEffect(() => {
    const timer = setInterval(() => {
      router.refresh();
      setLastRefresh(new Date());
    }, seconds * 1000);
    return () => clearInterval(timer); // cleanup when leaving the page
  }, [router, seconds]);

  return (
    <span className="muted small">
      refreshes every {seconds}s
      {lastRefresh && ` · last ${lastRefresh.toLocaleTimeString()}`}
    </span>
  );
}
