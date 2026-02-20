/** Format a date string for display. Optionally includes time. */
export function formatDate(dateStr: string, includeTime = false): string {
  const d = new Date(dateStr);
  const opts: Intl.DateTimeFormatOptions = { month: 'short', day: 'numeric', year: 'numeric' };
  if (includeTime) {
    opts.hour = '2-digit';
    opts.minute = '2-digit';
    opts.timeZoneName = 'short';
  }
  return d.toLocaleDateString('en-US', opts);
}

/** Convert an ISO date string to a value suitable for <input type="datetime-local">. */
export function toDatetimeLocalValue(isoStr: string): string {
  const d = new Date(isoStr);
  const pad = (n: number) => n.toString().padStart(2, '0');
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`;
}

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
