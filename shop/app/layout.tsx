import "./globals.css";
import type { ReactNode } from "react";
import { UserPicker } from "@/components/UserPicker";
import { ThemeToggle } from "@/components/ThemeToggle";
import { ToastProvider } from "@/components/ToastProvider";
import ClickTelemetry from "@/components/ClickTelemetry";

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body className="min-h-screen">
        <script
          dangerouslySetInnerHTML={{
            __html: `(() => { try { const t = localStorage.getItem('theme'); const prefersDark = window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches; const dark = t ? t === 'dark' : prefersDark; document.documentElement.classList.toggle('dark', !!dark); } catch {} })();`,
          }}
        />
        <header className="border-b">
          <nav className="mx-auto max-w-6xl flex items-center gap-6 p-4">
            <a className="font-semibold" href="/">
              Recsys Shop
            </a>
            <a href="/cart">Cart</a>
            <a href="/orders">My Orders</a>
            <a href="/events">Events</a>
            <a href="/admin">Admin</a>
            <div className="ml-auto flex items-center gap-3">
              <ThemeToggle />
              <UserPicker />
            </div>
          </nav>
        </header>
        <ToastProvider>
          <ClickTelemetry />
          <div className="mx-auto max-w-6xl p-4">{children}</div>
        </ToastProvider>
      </body>
    </html>
  );
}
