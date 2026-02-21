import { useMemo } from 'react';
import { BarChart, Bar, XAxis, YAxis, Tooltip, ResponsiveContainer, Cell } from 'recharts';

import { Card } from '../../../components/ui/Card';

interface SeedAllocationChartProps {
  investedRows: Array<{ seed: number; ourInvestment: number }>;
  totalOurInvestment: number;
}

export function SeedAllocationChart({ investedRows, totalOurInvestment }: SeedAllocationChartProps) {
  const seedAllocationData = useMemo(() => {
    const groups = [
      { name: '1-4', seeds: [1, 2, 3, 4], color: '#1d4ed8' },
      { name: '5-8', seeds: [5, 6, 7, 8], color: '#3b82f6' },
      { name: '9-12', seeds: [9, 10, 11, 12], color: '#93c5fd' },
      { name: '13-16', seeds: [13, 14, 15, 16], color: '#dbeafe' },
    ];
    return groups.map(({ name, seeds, color }) => {
      const amount = investedRows
        .filter((r) => seeds.includes(r.seed))
        .reduce((sum, r) => sum + r.ourInvestment, 0);
      return { name, amount, color, pct: totalOurInvestment > 0 ? (amount / totalOurInvestment * 100) : 0 };
    });
  }, [investedRows, totalOurInvestment]);

  if (investedRows.length === 0 || totalOurInvestment === 0) return null;

  return (
    <Card>
      <h2 className="text-lg font-semibold mb-3">Seed Allocation</h2>
      <p className="text-sm text-gray-500 mb-3">Budget distribution across seed groups.</p>
      <ResponsiveContainer width="100%" height={160}>
        <BarChart data={seedAllocationData} layout="vertical">
          <XAxis type="number" tickFormatter={(v: number) => `${v} pts`} fontSize={12} />
          <YAxis type="category" dataKey="name" width={50} fontSize={12} />
          <Tooltip
            formatter={(value: number, _name: string, props: { payload?: { pct?: number } }) =>
              [`${value} pts (${props.payload?.pct?.toFixed(0) ?? 0}%)`, 'Investment']
            }
          />
          <Bar dataKey="amount" radius={[0, 4, 4, 0]}>
            {seedAllocationData.map((entry) => (
              <Cell key={entry.name} fill={entry.color} />
            ))}
          </Bar>
        </BarChart>
      </ResponsiveContainer>
    </Card>
  );
}
