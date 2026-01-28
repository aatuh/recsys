import type { LegalDoc, LegalTokens } from "@api-boilerplate-core/legal";
import {
  composeLegalDoc,
  getVanillaLegalSnippets,
  getVanillaLegalTemplate,
} from "@api-boilerplate-core/legal";
import {
  DEFAULT_LOCALE,
  SUPPORTED_LOCALES,
  normalizeLocale,
  type Locale,
} from "@foo/i18n/config";

export { getCookieConsentConfig } from "./consent";

export const LEGAL_SLUGS = ["terms", "privacy", "cookies"] as const;
export type LegalSlug = (typeof LEGAL_SLUGS)[number];

type BaseMeta = {
  eyebrow: string;
  tocLabel: string;
  updatedLabel: string;
};

const DEFAULT_META: BaseMeta = {
  eyebrow: "Legal",
  tocLabel: "Contents",
  updatedLabel: "Last updated",
};

const baseMetaByLocale: Record<string, BaseMeta> = {
  en: DEFAULT_META,
  fi: DEFAULT_META,
};

export const tokensByLocale: Record<string, LegalTokens> = {
  en: {
    COMPANY_NAME: "Example, Inc.",
    SERVICE_NAME: "API Boilerplate",
    SERVICE_DESCRIPTION:
      "API Boilerplate helps you ship Go APIs with a modern Next.js frontend.",
    CONTACT_URL: "/contact",
    GOVERNING_LAW: "Delaware",
    GOVERNING_VENUE: "Wilmington, Delaware",
    COMPANY_ADDRESS: "Wilmington, Delaware",
    SERVICE_DATA_CATEGORY: "Application data",
    SERVICE_DATA_EXAMPLES:
      "Account profiles, usage events, and support tickets",
    SERVICE_DATA_TYPE: "Service records",
    SERVICE_DATA_TYPE_DESCRIPTION:
      "Stored while your account is active and for up to 24 months after closure.",
    PAYMENTS_PROVIDER: "Stripe",
    DATA_REGION_PRIMARY: "US",
    DATA_REGION_BACKUP: "EU",
  },
  fi: {
    COMPANY_NAME: "Example, Inc.",
    SERVICE_NAME: "API Boilerplate",
    SERVICE_DESCRIPTION:
      "API Boilerplate helps you ship Go APIs with a modern Next.js frontend.",
    CONTACT_URL: "/contact",
    GOVERNING_LAW: "Delaware",
    GOVERNING_VENUE: "Wilmington, Delaware",
    COMPANY_ADDRESS: "Wilmington, Delaware",
    SERVICE_DATA_CATEGORY: "Application data",
    SERVICE_DATA_EXAMPLES:
      "Account profiles, usage events, and support tickets",
    SERVICE_DATA_TYPE: "Service records",
    SERVICE_DATA_TYPE_DESCRIPTION:
      "Stored while your account is active and for up to 24 months after closure.",
    PAYMENTS_PROVIDER: "Stripe",
    DATA_REGION_PRIMARY: "US",
    DATA_REGION_BACKUP: "EU",
  },
};

export function resolveLegalLocale(input?: string | null): Locale {
  const normalized = normalizeLocale(input);
  if (getVanillaLegalTemplate(normalized, "terms")) return normalized;
  if (process.env["NODE_ENV"] !== "production") {
    console.warn(
      `[legal] Missing legal content for locale "${normalized}". Falling back to "${DEFAULT_LOCALE}".`
    );
  }
  return getVanillaLegalTemplate(DEFAULT_LOCALE, "terms")
    ? DEFAULT_LOCALE
    : "en";
}

export function getLegalDoc(
  locale: string | null | undefined,
  slug: LegalSlug
): LegalDoc {
  const resolvedLocale = resolveLegalLocale(locale);
  const meta = baseMetaByLocale[resolvedLocale] ?? DEFAULT_META;
  const template =
    getVanillaLegalTemplate(resolvedLocale, slug) ??
    getVanillaLegalTemplate("en", slug);
  if (!template) {
    throw new Error(`[legal] Missing vanilla template for slug "${slug}".`);
  }
  const tokens = tokensByLocale[resolvedLocale] ?? tokensByLocale["en"] ?? {};
  const vanillaSnippets = getVanillaLegalSnippets(resolvedLocale);
  return composeLegalDoc({
    template: {
      ...template,
      eyebrow: meta.eyebrow,
      tocLabel: meta.tocLabel,
      updatedLabel: meta.updatedLabel,
    },
    snippets: vanillaSnippets,
    include: [],
    tokens,
  });
}

export function getLegalLocaleCoverage() {
  const localeMap = new Map<Locale, LegalSlug[]>();
  SUPPORTED_LOCALES.forEach((locale) => {
    const available = LEGAL_SLUGS.filter((slug) =>
      getVanillaLegalTemplate(locale, slug)
    );
    localeMap.set(locale, available);
  });
  return localeMap;
}
