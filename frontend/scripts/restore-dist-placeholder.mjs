import { mkdirSync, writeFileSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const placeholderContent =
  "frontend build placeholder for `go test`.\n" +
  "This file keeps the embed target present in fresh checkouts.\n";

const scriptDir = dirname(fileURLToPath(import.meta.url));
const distDir = join(scriptDir, "..", "dist");
const placeholderPath = join(distDir, "placeholder.txt");

mkdirSync(distDir, { recursive: true });
writeFileSync(placeholderPath, placeholderContent, "utf8");
