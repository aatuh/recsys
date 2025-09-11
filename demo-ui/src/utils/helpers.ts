/**
 * Utility functions for the RecSys demo UI application.
 */

export function randChoice<T>(arr: T[]): T {
  return arr[Math.floor(Math.random() * arr.length)]!;
}

export function randInt(min: number, max: number): number {
  return min + Math.floor(Math.random() * (max - min + 1));
}

export function id(prefix: string, n: number): string {
  return `${prefix}-${String(n).padStart(4, "0")}`;
}

export function iso(dt: Date): string {
  return dt.toISOString();
}

export const now = () => new Date();

export function daysAgo(d: number): Date {
  const dt = new Date();
  dt.setDate(dt.getDate() - d);
  return dt;
}

/**
 * Select a random item from an array based on weighted probabilities
 */
export function weightedChoice<T>(items: T[], weights: number[]): T {
  if (items.length !== weights.length) {
    throw new Error("Items and weights arrays must have the same length");
  }

  if (items.length === 0) {
    throw new Error("Items array cannot be empty");
  }

  const totalWeight = weights.reduce((sum, weight) => sum + weight, 0);
  let random = Math.random() * totalWeight;

  for (let i = 0; i < items.length; i++) {
    random -= weights[i]!;
    if (random <= 0) {
      return items[i]!;
    }
  }

  // Fallback to last item (shouldn't happen with proper weights)
  return items[items.length - 1]!;
}

/**
 * Select multiple random items from an array based on weighted probabilities
 */
export function weightedChoices<T>(
  items: T[],
  weights: number[],
  count: number
): T[] {
  const result: T[] = [];
  const remainingItems = [...items];
  const remainingWeights = [...weights];

  for (let i = 0; i < count && remainingItems.length > 0; i++) {
    const selected = weightedChoice(remainingItems, remainingWeights);
    result.push(selected);

    // Remove selected item and its weight
    const index = remainingItems.indexOf(selected);
    remainingItems.splice(index, 1);
    remainingWeights.splice(index, 1);
  }

  return result;
}

/**
 * Generate a random boolean with given probability
 */
export function randomBoolean(probability: number): boolean {
  return Math.random() < probability;
}
