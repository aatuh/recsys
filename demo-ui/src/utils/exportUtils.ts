/**
 * Utility functions for exporting data as JSON
 */

export interface ExportData {
  users?: any[];
  items?: any[];
  events?: any[];
  metadata: {
    namespace: string;
    exportedAt: string;
    totalRecords: number;
    tables: string[];
  };
}

/**
 * Downloads data as a JSON file
 */
export function downloadJsonFile(data: any, filename: string): void {
  const jsonString = JSON.stringify(data, null, 2);
  const blob = new Blob([jsonString], { type: "application/json" });
  const url = URL.createObjectURL(blob);

  const link = document.createElement("a");
  link.href = url;
  link.download = filename;
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);

  // Clean up the URL object
  URL.revokeObjectURL(url);
}

/**
 * Generates a filename for the export based on namespace and selected tables
 */
export function generateExportFilename(
  namespace: string,
  selectedTables: string[]
): string {
  const timestamp = new Date().toISOString().replace(/[:.]/g, "-");
  const tablesStr = selectedTables.sort().join("-");
  return `recsys-export-${namespace}-${tablesStr}-${timestamp}.json`;
}

/**
 * Formats the export data structure
 */
export function formatExportData(
  namespace: string,
  selectedTables: string[],
  data: {
    users?: any[];
    items?: any[];
    events?: any[];
  }
): ExportData {
  const totalRecords =
    (data.users?.length || 0) +
    (data.items?.length || 0) +
    (data.events?.length || 0);

  return {
    ...data,
    metadata: {
      namespace,
      exportedAt: new Date().toISOString(),
      totalRecords,
      tables: selectedTables.sort(),
    },
  };
}
