/**
 * Formats a value in cents to a dollar string (e.g., 1050 -> "$10.50").
 * Handles negative values, zero, and undefined.
 * Omits decimal places when the amount is a whole number of dollars.
 */
export function formatDollarsFromCents(cents?: number): string {
  if (cents == null) return '$0';
  const abs = Math.abs(cents);
  const dollars = Math.floor(abs / 100);
  const remainder = abs % 100;
  const sign = cents < 0 ? '-' : '';
  if (remainder === 0) return `${sign}$${dollars}`;
  return `${sign}$${dollars}.${remainder.toString().padStart(2, '0')}`;
}
