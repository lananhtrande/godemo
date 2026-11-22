import type { Metadata } from "next";
import Link from "next/link";
import "./globals.css";

export const metadata: Metadata = {
  title: "AXOS Demo – Battery Sites",
  description: "Go + Next.js prototype: live state of charge of battery storage sites",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body>
        <header className="topbar">
          <Link href="/" className="brand">
            AXOS Demo
          </Link>
          <span className="muted">Go service · Next.js · Server-Sent Events</span>
        </header>
        <main className="container">{children}</main>
      </body>
    </html>
  );
}
