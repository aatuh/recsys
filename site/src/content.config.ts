import { defineCollection } from "astro:content";
import { glob } from "astro/loaders";
import { z } from "astro/zod";

const blog = defineCollection({
  loader: glob({ pattern: "**/*.md", base: "./src/content/blog" }),
  schema: z.object({
    title: z.string(),
    description: z.string(),
    language: z.enum(["en", "fi"]),
    pubDate: z.coerce.date(),
    updatedDate: z.coerce.date().optional(),
    translationKey: z.string(),
    tags: z.array(z.string()).default([]),
  }),
});

export const collections = { blog };
