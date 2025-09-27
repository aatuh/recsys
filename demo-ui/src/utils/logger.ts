// Legacy logger wrapper for backward compatibility.
// New code should use the DI container: import { getLogger } from "../di"

import { getLogger } from "../di";

// Re-export the logger from the DI container for backward compatibility
export const logger = getLogger();
