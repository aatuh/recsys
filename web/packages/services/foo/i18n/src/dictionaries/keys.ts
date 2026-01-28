import type { Dictionary } from "./types";

type AppendPrefix<Prefix extends string, Key extends string> = Prefix extends "" ? Key : `${Prefix}.${Key}`;

type StringLeafPaths<T, Prefix extends string = ""> = T extends string
  ? Prefix
  : T extends Array<infer U>
    ? StringLeafPaths<U, AppendPrefix<Prefix, `${number}`>>
    : T extends object
      ? {
          [K in keyof T]-?: StringLeafPaths<T[K], AppendPrefix<Prefix, K & string>>;
        }[keyof T]
      : never;

export type DictionaryKey = StringLeafPaths<Dictionary>;
