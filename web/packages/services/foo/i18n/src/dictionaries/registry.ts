import { en } from "./en";
import { fi } from "./fi";
import type { Dictionary } from "./types";

export const dictionaries = {
  en,
  fi,
} satisfies Record<string, Dictionary>;

export type DictionaryLocale = keyof typeof dictionaries;

export const DICTIONARY_LOCALES = Object.keys(dictionaries) as DictionaryLocale[];

export const baseDictionary = dictionaries.en;
