import { diffChars } from "diff";

export interface DiffPart {
  value: string;
  type: "added" | "removed" | "unchanged";
}

export interface DiffStats {
  added: number;
  removed: number;
  unchanged: number;
}

export interface DiffResult {
  parts: DiffPart[];
  stats: DiffStats;
}

export function computeDiff(oldText: string, newText: string): DiffResult {
  const changes = diffChars(oldText, newText);

  const stats: DiffStats = { added: 0, removed: 0, unchanged: 0 };
  const parts: DiffPart[] = changes.map((change) => {
    const charCount = [...change.value].length; // rune-safe length
    if (change.added) {
      stats.added += charCount;
      return { value: change.value, type: "added" as const };
    }
    if (change.removed) {
      stats.removed += charCount;
      return { value: change.value, type: "removed" as const };
    }
    stats.unchanged += charCount;
    return { value: change.value, type: "unchanged" as const };
  });

  return { parts, stats };
}
