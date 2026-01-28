import "server-only";

import { cookies, headers } from "next/headers";
import {
  DEFAULT_LOCALE,
  LOCALE_AUTO_DETECT,
  normalizeLocale,
  resolveLocaleFromAcceptLanguage,
  type Locale,
} from "./config";
import { getDictionary } from "./dictionaries";

export async function getRequestLocale(): Promise<Locale> {
  const hdrs = await headers();
  const headerLocale =
    hdrs.get("x-locale") ||
    (LOCALE_AUTO_DETECT ? resolveLocaleFromAcceptLanguage(hdrs.get("accept-language")) : null);
  const cookieStore = await cookies();
  const cookieLocale = cookieStore.get("locale")?.value;
  return normalizeLocale(cookieLocale || headerLocale || DEFAULT_LOCALE);
}

export async function getDictionaryForRequest() {
  const locale = await getRequestLocale();
  return { locale, dictionary: getDictionary(locale) };
}
