export const DEFAULT_ENTRY_COLOR_PALETTE = [
  '#2563EB',
  '#DC2626',
  '#059669',
  '#7C3AED',
  '#DB2777',
  '#D97706',
  '#0D9488',
  '#4B5563',
  '#9333EA',
  '#16A34A',
  '#EA580C',
  '#0EA5E9',
] as const;

export function getEntryColorById<T extends { id: string }>(
  entries: readonly T[],
  palette: readonly string[] = DEFAULT_ENTRY_COLOR_PALETTE
): Map<string, string> {
  const map = new Map<string, string>();
  entries.forEach((entry, idx) => {
    map.set(entry.id, palette[idx % palette.length]);
  });
  return map;
}
