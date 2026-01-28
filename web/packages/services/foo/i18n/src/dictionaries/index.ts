import { normalizeLocale, DEFAULT_LOCALE } from "../config";
import { baseDictionary, dictionaries, type DictionaryLocale } from "./registry";
import type { Dictionary } from "./types";

export function getDictionary(localeInput?: string | null): Dictionary {
  const locale = normalizeLocale(localeInput || DEFAULT_LOCALE);
  const short = locale.split("-")[0];
  const match = dictionaries[short as DictionaryLocale];
  return match ?? baseDictionary;
}

export type { Dictionary } from "./types";
export type { DictionaryKey } from "./keys";
export { DICTIONARY_LOCALES } from "./registry";
