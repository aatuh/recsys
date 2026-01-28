"use client";

import { useMemo } from "react";
import {
  createLocaleProvider,
  useLocale as useSharedLocale,
} from "@api-boilerplate-core/i18n-shared/locale-context";
import { getDictionary } from "./dictionaries";
import type { DictionaryKey } from "./dictionaries/keys";

export type Translator = (
  key: DictionaryKey,
  params?: Record<string, string | number>
) => string;
export type RawTranslator = (
  key: string,
  params?: Record<string, string | number>
) => string;
export type ValueTranslator = (
  key: string,
  params?: Record<string, string | number>
) => unknown;

export const LocaleProvider = createLocaleProvider(getDictionary);

export function useLocale() {
  const { locale, t, tRaw } = useSharedLocale();
  const tValue: ValueTranslator = tRaw;
  const tLoose = useMemo<RawTranslator>(() => {
    return (key, params) => {
      const value = tRaw(key, params);
      return typeof value === "string" ? value : key;
    };
  }, [tRaw]);

  return { locale, t: t as Translator, tRaw: tLoose, tValue };
}
