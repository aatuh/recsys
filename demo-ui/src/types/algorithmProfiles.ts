import type { types_Overrides } from "../lib/api-client";

export interface AlgorithmProfile {
  id: string;
  name: string;
  description: string;
  overrides: types_Overrides;
}

export const ALGORITHM_PROFILES: AlgorithmProfile[] = [
  {
    id: "trending-now",
    name: "Trending-Now",
    description:
      "Very reactive, newsy feel. Use when you want clicks to reshuffle quickly.",
    overrides: {
      popularity_halflife_days: 3,
      covis_window_days: 14,
      popularity_fanout: 1000,
      mmr_lambda: 0,
      brand_cap: 0,
      category_cap: 0,
      rule_exclude_purchased: false,
      purchased_window_days: undefined,
      profile_window_days: 7,
      profile_boost: 0.15,
      profile_top_n: 32,
      blend_alpha: 1.0,
      blend_beta: 0.0,
      blend_gamma: 0.0,
    },
  },
  {
    id: "personalized-pop",
    name: "Personalized-Pop",
    description:
      "Let the user's trail matter a lot. Interact (views/clicks/ATC) and refresh to see ranking shift.",
    overrides: {
      popularity_halflife_days: 10,
      covis_window_days: 21,
      popularity_fanout: 1000,
      mmr_lambda: 0.1,
      brand_cap: 0,
      category_cap: 0,
      rule_exclude_purchased: true,
      purchased_window_days: 365,
      profile_window_days: 14,
      profile_boost: 0.6,
      profile_top_n: 48,
      blend_alpha: 1.0,
      blend_beta: 0.0,
      blend_gamma: 0.0,
    },
  },
  {
    id: "covis-discovery",
    name: "Co-Vis Discovery",
    description:
      "People who did X also did Y. Great after you create a few co-occurrence events in the demo.",
    overrides: {
      popularity_halflife_days: 14,
      covis_window_days: 30,
      popularity_fanout: 1500,
      mmr_lambda: 0.2,
      brand_cap: 0,
      category_cap: 0,
      rule_exclude_purchased: true,
      purchased_window_days: undefined,
      profile_window_days: 14,
      profile_boost: 0.25,
      profile_top_n: 32,
      blend_alpha: 0.5,
      blend_beta: 0.5,
      blend_gamma: 0.0,
    },
  },
  {
    id: "semantic-similarity",
    name: "Semantic Similarity",
    description:
      "If you have embeddings wired. Puts strong weight on content similarity; pop still anchors results.",
    overrides: {
      popularity_halflife_days: 21,
      covis_window_days: 30,
      popularity_fanout: 2000,
      mmr_lambda: 0.25,
      brand_cap: 0,
      category_cap: 0,
      rule_exclude_purchased: true,
      purchased_window_days: undefined,
      profile_window_days: 21,
      profile_boost: 0.35,
      profile_top_n: 64,
      blend_alpha: 0.3,
      blend_beta: 0.2,
      blend_gamma: 0.5,
    },
  },
  {
    id: "diverse-grid",
    name: "Diverse Grid / Safe Browsing",
    description:
      "Avoids spammy dominance and shows variety; great for catalog demos.",
    overrides: {
      popularity_halflife_days: 14,
      covis_window_days: 30,
      popularity_fanout: 2000,
      mmr_lambda: 0.35,
      brand_cap: 3,
      category_cap: 4,
      rule_exclude_purchased: true,
      purchased_window_days: 180,
      profile_window_days: 21,
      profile_boost: 0.25,
      profile_top_n: 48,
      blend_alpha: 0.6,
      blend_beta: 0.2,
      blend_gamma: 0.2,
    },
  },
  {
    id: "production-balanced",
    name: "Production-ish Balanced Baseline",
    description: "Stable, mixed signals, gentle diversity, no weird surprises.",
    overrides: {
      popularity_halflife_days: 14,
      covis_window_days: 30,
      popularity_fanout: 3000,
      mmr_lambda: 0.25,
      brand_cap: 4,
      category_cap: 6,
      rule_exclude_purchased: true,
      purchased_window_days: 365,
      profile_window_days: 30,
      profile_boost: 0.2,
      profile_top_n: 64,
      blend_alpha: 0.6,
      blend_beta: 0.25,
      blend_gamma: 0.15,
    },
  },
];
