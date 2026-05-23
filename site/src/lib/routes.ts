export type Locale = "en" | "fi";
export type PageKey =
  | "home"
  | "pricing"
  | "security"
  | "evaluation"
  | "contact"
  | "documentation"
  | "blog";

export const siteName = "RecSys";
export const siteUrl = "https://recsys.app";

export const pageRoutes: Record<PageKey, Record<Locale, string>> = {
  home: { en: "/", fi: "/fi/" },
  pricing: { en: "/pricing/", fi: "/fi/hinnoittelu/" },
  security: { en: "/security/", fi: "/fi/tietoturva/" },
  evaluation: { en: "/evaluation/", fi: "/fi/arviointi/" },
  contact: { en: "/contact/", fi: "/fi/yhteys/" },
  documentation: { en: "/documentation/", fi: "/fi/dokumentaatio/" },
  blog: { en: "/blog/", fi: "/fi/blogi/" },
};

export const navLabels: Record<Locale, Record<PageKey, string>> = {
  en: {
    home: "Product",
    pricing: "Pricing",
    security: "Security",
    evaluation: "Evaluation",
    contact: "Contact",
    documentation: "Documentation",
    blog: "Blog",
  },
  fi: {
    home: "Tuote",
    pricing: "Hinnoittelu",
    security: "Tietoturva",
    evaluation: "Arviointi",
    contact: "Yhteys",
    documentation: "Dokumentaatio",
    blog: "Blogi",
  },
};

export const navOrder: PageKey[] = [
  "home",
  "pricing",
  "security",
  "evaluation",
  "documentation",
  "blog",
  "contact",
];

export function absolutePath(path: string): string {
  return new URL(path, siteUrl).toString();
}

export function alternatesForPage(pageKey: PageKey): Record<string, string> {
  return {
    en: absolutePath(pageRoutes[pageKey].en),
    fi: absolutePath(pageRoutes[pageKey].fi),
    "x-default": absolutePath(pageRoutes[pageKey].en),
  };
}

export function languagePath(pageKey: PageKey, locale: Locale): string {
  return pageRoutes[pageKey][locale];
}
