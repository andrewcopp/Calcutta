export function NormalizedValue({ value }: { value: number }) {
  const colorClass = value > 1.0 ? 'text-success' : value < 1.0 ? 'text-destructive' : 'text-muted-foreground';
  return <span className={`font-semibold ${colorClass}`}>{value.toFixed(3)}</span>;
}
