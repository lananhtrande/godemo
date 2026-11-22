// Same shape as the Go struct `Site` (backend/store.go).
export type SiteStatus = "charging" | "discharging" | "idle";

export interface Site {
  id: string;
  name: string;
  capacityMWh: number;
  maxPowerMW: number;
  soc: number; // state of charge, 0-100 %
  powerMW: number; // > 0 discharging, < 0 charging
  status: SiteStatus;
  updatedAt: string; // ISO timestamp, UTC
}

// Server side (inside Docker) the backend is reachable as http://backend:8080,
// the browser reaches it via NEXT_PUBLIC_API_URL (http://localhost:8080).
const SERVER_API_URL = process.env.API_URL ?? "http://localhost:8080";
export const BROWSER_API_URL =
  process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

export async function getSites(): Promise<Site[]> {
  const res = await fetch(`${SERVER_API_URL}/api/v1/sites`, {
    cache: "no-store", // operational data: always fresh, never cached
  });
  if (!res.ok) throw new Error(`Failed to load sites: ${res.status}`);
  return res.json();
}

export async function getSite(id: string): Promise<Site | null> {
  const res = await fetch(
    `${SERVER_API_URL}/api/v1/sites/${encodeURIComponent(id)}`,
    { cache: "no-store" },
  );
  if (res.status === 404) return null;
  if (!res.ok) throw new Error(`Failed to load site ${id}: ${res.status}`);
  return res.json();
}
