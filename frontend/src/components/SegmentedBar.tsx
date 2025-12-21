import { useMemo, useState } from 'react';

export type SegmentedBarSegment = {
  key: string;
  label: string;
  value: number;
  color: string;
};

export function SegmentedBar({
  barWidthPct,
  segments,
  backgroundColor,
  disabled,
  getTooltipTitle,
  getTooltipValue,
}: {
  barWidthPct: number;
  segments: SegmentedBarSegment[];
  backgroundColor?: string;
  disabled?: boolean;
  getTooltipTitle?: (segment: SegmentedBarSegment) => string;
  getTooltipValue?: (segment: SegmentedBarSegment) => string;
}) {
  const [hover, setHover] = useState<{ key: string; title: string; value: string; x: number; y: number } | null>(null);

  const total = useMemo(() => segments.reduce((sum, s) => sum + s.value, 0), [segments]);

  const normalizedSegments = useMemo(() => segments.filter((s) => s.value > 0), [segments]);

  const safeBarWidthPct = Math.max(0, Math.min(100, barWidthPct));

  return (
    <>
      <div
        className="h-6 w-full rounded overflow-hidden"
        style={{ backgroundColor: disabled ? 'transparent' : backgroundColor || '#E5E7EB' }}
      >
        <div className="h-full flex" style={{ width: `${safeBarWidthPct.toFixed(2)}%` }}>
          {normalizedSegments.map((seg) => {
            const segWidthPct = total > 0 ? (seg.value / total) * 100 : 0;
            const isActive = hover?.key === seg.key;
            const title = getTooltipTitle ? getTooltipTitle(seg) : seg.label;
            const value = getTooltipValue ? getTooltipValue(seg) : seg.value.toFixed(2);

            return (
              <div
                key={seg.key}
                className="h-full"
                style={{
                  width: `${segWidthPct.toFixed(2)}%`,
                  backgroundColor: seg.color,
                  boxSizing: 'border-box',
                  border: isActive ? '2px solid #111827' : '2px solid transparent',
                }}
                onMouseEnter={(e) => {
                  setHover({
                    key: seg.key,
                    title,
                    value,
                    x: e.clientX,
                    y: e.clientY,
                  });
                }}
                onMouseMove={(e) => {
                  setHover((prev) => (prev ? { ...prev, x: e.clientX, y: e.clientY } : prev));
                }}
                onMouseLeave={() => setHover(null)}
              />
            );
          })}
        </div>
      </div>

      {hover && (
        <div
          className="fixed z-50 pointer-events-none rounded bg-gray-900 px-3 py-2 text-xs text-white shadow"
          style={{ left: hover.x + 12, top: hover.y + 12 }}
        >
          <div className="font-medium">{hover.title}</div>
          <div>{hover.value}</div>
        </div>
      )}
    </>
  );
}
