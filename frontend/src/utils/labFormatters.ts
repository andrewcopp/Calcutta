/**
 * Shared formatting utilities for Lab pages and components.
 *
 * These cover the three most commonly duplicated helpers:
 *   - formatPayoutX  -- render a payout multiplier (e.g. "1.234x")
 *   - formatPct      -- render a probability/percentage (e.g. "12.3%")
 *   - getPayoutColor -- Tailwind class string for payout-based coloring
 */

/** Format a payout multiplier value. Returns "-" for null/undefined. */
export function formatPayoutX(val?: number | null, decimals = 3): string {
  if (val == null) return '-';
  return `${val.toFixed(decimals)}x`;
}

/** Format a 0-1 probability as a percentage string. Returns "-" for null/undefined. */
export function formatPct(val?: number | null, decimals = 1): string {
  if (val == null) return '-';
  return `${(val * 100).toFixed(decimals)}%`;
}

/** Return Tailwind color classes based on a payout multiplier value. */
export function getPayoutColor(payout?: number | null): string {
  if (payout == null) return 'text-gray-400';
  if (payout >= 1.2) return 'text-green-700 font-bold';
  if (payout >= 0.9) return 'text-yellow-700';
  return 'text-red-700';
}
