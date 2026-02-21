export function NormalizedValue({ value }: { value: number }) {
  const colorClass =
    value > 1.0 ? 'text-green-600' : value < 1.0 ? 'text-red-600' : 'text-gray-500';
  return <span className={`font-semibold ${colorClass}`}>{value.toFixed(3)}</span>;
}
