#!/usr/bin/env node

/* eslint-disable no-console */

/**
 * Sync README files from various directories to the demo-ui public folder
 * This script ensures the documentation in the demo UI is always up-to-date
 */

import fs from "fs";
import path from "path";
import { fileURLToPath } from "url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const PROJECT_ROOT = path.resolve(__dirname, "../..");
const DEMO_UI_PUBLIC = path.resolve(__dirname, "../public");

// Ensure public directory exists
if (!fs.existsSync(DEMO_UI_PUBLIC)) {
  fs.mkdirSync(DEMO_UI_PUBLIC, { recursive: true });
}

// Define README sources and their target names
const readmeSources = [
  {
    source: path.join(PROJECT_ROOT, "README.md"),
    target: "readme-root.md",
    name: "Root Project",
  },
  {
    source: path.join(PROJECT_ROOT, "demo-ui", "README.md"),
    target: "readme-demo-ui.md",
    name: "Demo UI",
  },
  {
    source: path.join(PROJECT_ROOT, "api", "README.md"),
    target: "readme-api.md",
    name: "API",
  },
];

console.log("🔄 Syncing README files...\n");

let syncedCount = 0;
let skippedCount = 0;

readmeSources.forEach(({ source, target, name }) => {
  const targetPath = path.join(DEMO_UI_PUBLIC, target);

  try {
    if (!fs.existsSync(source)) {
      console.log(`⚠️  ${name}: Source file not found at ${source}`);
      skippedCount++;
      return;
    }

    // Read source file
    const content = fs.readFileSync(source, "utf8");

    // Write to target
    fs.writeFileSync(targetPath, content, "utf8");

    console.log(`✅ ${name}: ${source} → ${targetPath}`);
    syncedCount++;
  } catch (error) {
    console.error(`❌ ${name}: Failed to sync - ${error.message}`);
    skippedCount++;
  }
});

console.log(`\n📊 Summary: ${syncedCount} synced, ${skippedCount} skipped`);

if (syncedCount > 0) {
  console.log("🎉 README files synced successfully!");
} else {
  console.log("⚠️  No files were synced.");
  process.exit(1);
}
