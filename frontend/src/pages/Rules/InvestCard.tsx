import { useMemo } from 'react';
import { Bar, BarChart, CartesianGrid, Cell, ResponsiveContainer, Tooltip, XAxis, YAxis } from 'recharts';
import { Card } from '../../components/ui/Card';

interface InvestDatum {
  label: string;
  credits: number;
  fill: string;
}

const investData: InvestDatum[] = [
  { label: 'Top seed', credits: 35, fill: '#2563EB' },
  { label: 'Contender', credits: 25, fill: '#16A34A' },
  { label: 'Value', credits: 20, fill: '#F59E0B' },
  { label: 'Sleeper', credits: 20, fill: '#7C3AED' },
];

export function InvestCard() {
  const data = useMemo(() => investData, []);

  return (
    <Card className="shadow-lg">
      <div className="flex items-center justify-between gap-4 mb-4">
        <div>
          <h2 className="text-2xl font-semibold text-foreground">Invest</h2>
          <p className="text-muted-foreground">Allocate 100 credits across a handful of teams.</p>
        </div>
        <div className="text-xs text-muted-foreground">Illustrative</div>
      </div>
      <div className="h-72">
        <ResponsiveContainer width="100%" height="100%">
          <BarChart data={data} margin={{ top: 6, right: 12, left: 0, bottom: 6 }}>
            <CartesianGrid strokeDasharray="3 3" />
            <XAxis dataKey="label" tick={{ fontSize: 12 }} />
            <YAxis tick={{ fontSize: 12 }} />
            <Tooltip formatter={(value: number) => [`${value} credits`, 'Investment']} />
            <Bar dataKey="credits" radius={[8, 8, 0, 0]}>
              {data.map((entry) => (
                <Cell key={entry.label} fill={entry.fill} />
              ))}
            </Bar>
          </BarChart>
        </ResponsiveContainer>
      </div>
      <div className="mt-4 text-sm text-muted-foreground">
        Typical constraints: invest in 3\u201310 teams, max 50 credits on any single team.
      </div>
    </Card>
  );
}
