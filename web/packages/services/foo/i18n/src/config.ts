import {
  FALLBACK_LOCALE,
  FALLBACK_SUPPORTED_LOCALES,
  LOCALE_AUTO_DETECT,
  parseSupportedLocales,
  resolveLocale as resolveLocaleShared,
  resolveLocaleFromAcceptLanguage as resolveLocaleFromAcceptLanguageShared,
  type Locale,
} from "@api-boilerplate-core/i18n-shared/config";
import { DICTIONARY_LOCALES } from "./dictionaries/registry";

const dictionaryLocales = [...DICTIONARY_LOCALES] as Locale[];
const envSupported = parseSupportedLocales(
  process.env["NEXT_PUBLIC_SUPPORTED_LOCALES"]
) as Locale[];
const filtered = envSupported.filter((loc) => dictionaryLocales.includes(loc));

let supportedLocales: Locale[] =
  filtered.length > 0 ? filtered : dictionaryLocales;

if (
  process.env["NODE_ENV"] !== "production" &&
  filtered.length < envSupported.length
) {
  const unsupported = envSupported.filter(
    (loc) => !dictionaryLocales.includes(loc)
  );
  if (unsupported.length > 0) {
    console.warn(
      `[i18n] Ignoring unsupported locales (${unsupported.join(
        ", "
      )}). Add dictionaries before enabling them.`
    );
  }
}

export const DEFAULT_LOCALE: Locale = resolveLocaleShared(
  process.env["NEXT_PUBLIC_DEFAULT_LOCALE"],
  FALLBACK_LOCALE,
  supportedLocales
);

if (!supportedLocales.includes(DEFAULT_LOCALE)) {
  if (process.env["NODE_ENV"] !== "production") {
    console.warn(
      `[i18n] Adding default locale "${DEFAULT_LOCALE}" to supported locales because it was missing.`
    );
  }
  supportedLocales = [...supportedLocales, DEFAULT_LOCALE];
}

export const SUPPORTED_LOCALES: Locale[] = supportedLocales;

export function resolveLocale(
  input?: string | null,
  fallback: Locale = FALLBACK_LOCALE
): Locale {
  return resolveLocaleShared(input, fallback, SUPPORTED_LOCALES);
}

export function normalizeLocale(
  input?: string | null,
  defaultLocale: Locale = DEFAULT_LOCALE
): Locale {
  return resolveLocaleShared(input, defaultLocale, SUPPORTED_LOCALES);
}

export function resolveLocaleFromAcceptLanguage(
  input?: string | null,
  fallback: Locale = DEFAULT_LOCALE
): Locale {
  return resolveLocaleFromAcceptLanguageShared(
    input,
    fallback,
    SUPPORTED_LOCALES
  );
}

export {
  FALLBACK_LOCALE,
  FALLBACK_SUPPORTED_LOCALES,
  LOCALE_AUTO_DETECT,
  parseSupportedLocales,
  type Locale,
};
