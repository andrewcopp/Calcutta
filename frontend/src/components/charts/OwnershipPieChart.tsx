import { useState } from 'react';
import { Cell, Pie, PieChart, ResponsiveContainer, Tooltip } from 'recharts';

export type OwnershipPieSlice = {
  key: string;
  name: string;
  value: number;
  fill: string;
};

export function OwnershipPieChart({
  data,
  sizePx = 220,
  emptyLabel = 'No ownership',
}: {
  data: OwnershipPieSlice[];
  sizePx?: number;
  emptyLabel?: string;
}) {
  const [activeIndex, setActiveIndex] = useState<number | null>(null);

  const DIVIDER_STROKE = '#FFFFFF';
  const HOVER_STROKE = '#111827';

  if (data.length === 0) {
    return (
      <div className="flex h-[220px] w-[220px] items-center justify-center rounded-full bg-gray-100 text-sm text-gray-500">
        {emptyLabel}
      </div>
    );
  }

  const innerRadius = Math.max(20, Math.floor(sizePx * 0.22));
  const outerRadius = Math.max(40, Math.floor(sizePx * 0.42));

  return (
    <div style={{ height: sizePx, width: sizePx }}>
      <ResponsiveContainer width="100%" height="100%">
        <PieChart>
          <Pie
            data={data}
            cx="50%"
            cy="50%"
            innerRadius={innerRadius}
            outerRadius={outerRadius}
            paddingAngle={2}
            dataKey="value"
            isAnimationActive={false}
            onMouseEnter={(_, index) => setActiveIndex(index)}
            onMouseLeave={() => setActiveIndex(null)}
          >
            {data.map((slice, index) => {
              const isActive = activeIndex === index;
              return (
                <Cell
                  key={slice.key}
                  fill={slice.fill}
                  stroke={isActive ? HOVER_STROKE : DIVIDER_STROKE}
                  strokeWidth={2}
                />
              );
            })}
          </Pie>
          <Tooltip formatter={(value: number) => `${value.toFixed(2)}%`} labelFormatter={(label) => label} />
        </PieChart>
      </ResponsiveContainer>
    </div>
  );
}
