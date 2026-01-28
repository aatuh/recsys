import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import Link from "next/link";
import type { ReactNode } from "react";
import "./globals.css";
import {
  AppShell,
  SiteFooter,
  SiteHeader,
} from "@api-boilerplate-core/layouts";
import { Button } from "@api-boilerplate-core/ui";
import {
  ThemeProvider,
  ThemeVariantProvider,
} from "@api-boilerplate-core/theme";
import {
  ThemeSwitcher,
  ThemeVariantSwitcher,
} from "@api-boilerplate-core/widgets";
import { LocaleProvider } from "@foo/i18n/locale-context";
import { getDictionaryForRequest } from "@foo/i18n/locale.server";
import { appBaseUrl } from "@foo/config";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  metadataBase: new URL(appBaseUrl),
  title: "API Boilerplate",
  description: "Production-ready starter for Go APIs and Next.js apps.",
  icons: {
    icon: [
      { url: "/favicon.ico" },
      { url: "/favicon-32x32.png", sizes: "32x32", type: "image/png" },
      { url: "/favicon-16x16.png", sizes: "16x16", type: "image/png" },
    ],
    apple: [
      { url: "/apple-touch-icon.png", sizes: "180x180", type: "image/png" },
    ],
  },
  manifest: "/site.webmanifest",
};

export default async function RootLayout({
  children,
}: Readonly<{
  children: ReactNode;
}>) {
  const { locale, dictionary } = await getDictionaryForRequest();
  const themeScript = `
    (function() {
      try {
        var theme = localStorage.getItem('cx-theme');
        var variant = localStorage.getItem('cx-theme-variant');
        var root = document.documentElement;
        if (theme === 'light' || theme === 'dark') {
          root.setAttribute('data-theme', theme);
          root.style.colorScheme = theme;
        }
        if (variant) {
          root.setAttribute('data-theme-variant', variant);
        }
      } catch (e) {}
    })();
  `;

  return (
    <html lang={locale} suppressHydrationWarning data-theme-variant="slate">
      <body
        className={`${geistSans.variable} ${geistMono.variable} antialiased bg-grid text-foreground`}
      >
        <script dangerouslySetInnerHTML={{ __html: themeScript }} />
        <ThemeProvider>
          <ThemeVariantProvider>
            <LocaleProvider locale={locale}>
              <AppShell
                header={
                  <SiteHeader
                    brand={{
                      name: dictionary.brand.name,
                      shortName: dictionary.brand.shortName,
                    }}
                    signedOut={
                      <nav className="flex flex-wrap items-center gap-3 text-sm font-medium">
                        <Link
                          className="transition hover:text-primary"
                          href="/docs"
                        >
                          Docs
                        </Link>
                        <Link
                          className="transition hover:text-primary"
                          href="/foos"
                        >
                          Foo API
                        </Link>
                        <Button href="/foos" size="md">
                          Get Started
                        </Button>
                      </nav>
                    }
                  />
                }
                footer={
                  <SiteFooter
                    brand={{
                      name: dictionary.brand.name,
                      shortName: dictionary.brand.shortName,
                    }}
                    lead="Build and ship a typed API + web app without the glue work."
                    sections={[
                      {
                        title: "Product",
                        links: [
                          { href: "/", label: "Overview" },
                          { href: "/foos", label: "Foo API" },
                          { href: "/docs", label: "Docs" },
                        ],
                      },
                      {
                        title: "Resources",
                        links: [
                          {
                            href: "/docs/getting-started",
                            label: "Getting started",
                          },
                          { href: "/docs/architecture", label: "Architecture" },
                        ],
                      },
                    ]}
                    actions={
                      <div className="flex flex-wrap items-center gap-3">
                        <ThemeSwitcher compact />
                        <ThemeVariantSwitcher compact />
                      </div>
                    }
                  />
                }
                mainClassName="py-10"
              >
                {children}
              </AppShell>
            </LocaleProvider>
          </ThemeVariantProvider>
        </ThemeProvider>
      </body>
    </html>
  );
}
