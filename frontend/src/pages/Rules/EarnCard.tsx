import { useMemo } from 'react';
import {
  Bar,
  BarChart,
  CartesianGrid,
  Cell,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts';
import { Card } from '../../components/ui/Card';

interface SimulatedTeam {
  key: string;
  team: string;
  seed: number;
  ownershipPct: number;
  path: string;
  teamPoints: number;
  fill: string;
}

const simulatedEntry: SimulatedTeam[] = [
  {
    key: 'champion-slice',
    team: 'Top seed favorite',
    seed: 1,
    ownershipPct: 2,
    path: 'Wins the title',
    teamPoints: 1050,
    fill: '#2563EB',
  },
  {
    key: 'deep-run',
    team: 'Contender',
    seed: 6,
    ownershipPct: 12.5,
    path: 'Reaches the Elite 8',
    teamPoints: 500,
    fill: '#16A34A',
  },
  {
    key: 'one-win',
    team: 'Sleeper',
    seed: 12,
    ownershipPct: 35,
    path: 'Wins one game',
    teamPoints: 50,
    fill: '#7C3AED',
  },
];

export function EarnCard() {
  const simulatedEntryWithScore = useMemo(() => {
    return simulatedEntry.map((t) => ({
      ...t,
      yourPoints: (t.teamPoints * t.ownershipPct) / 100,
    }));
  }, []);

  const simulatedTotal = useMemo(() => {
    return simulatedEntryWithScore.reduce((sum, t) => sum + t.yourPoints, 0);
  }, [simulatedEntryWithScore]);

  return (
    <Card className="shadow-lg">
      <div className="flex items-center justify-between gap-4 mb-4">
        <div>
          <h2 className="text-2xl font-semibold text-gray-900">Earn</h2>
          <p className="text-gray-600">Your points are team points × your ownership percentage.</p>
        </div>
        <div className="rounded-full bg-gray-900 px-3 py-1 text-sm font-semibold text-white">
          Example total: {simulatedTotal.toFixed(1)} points
        </div>
      </div>

      <div className="overflow-x-auto">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Team</th>
              <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Seed</th>
              <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Path</th>
              <th className="px-4 py-2 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Ownership</th>
              <th className="px-4 py-2 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Team Points</th>
              <th className="px-4 py-2 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Your Points</th>
            </tr>
          </thead>
          <tbody className="bg-white divide-y divide-gray-200">
            {simulatedEntryWithScore.map((t) => (
              <tr key={t.key}>
                <td className="px-4 py-2 whitespace-nowrap text-sm font-medium text-gray-900">{t.team}</td>
                <td className="px-4 py-2 whitespace-nowrap text-sm text-gray-600">{t.seed}</td>
                <td className="px-4 py-2 whitespace-nowrap text-sm text-gray-600">{t.path}</td>
                <td className="px-4 py-2 whitespace-nowrap text-sm text-right text-gray-600">{t.ownershipPct}%</td>
                <td className="px-4 py-2 whitespace-nowrap text-sm text-right text-gray-600">{t.teamPoints}</td>
                <td className="px-4 py-2 whitespace-nowrap text-sm text-right font-semibold text-gray-900">{t.yourPoints.toFixed(1)}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      <div className="mt-6">
        <h3 className="text-sm font-semibold text-gray-900 mb-2">Contribution to your score</h3>
        <div className="h-64">
          <ResponsiveContainer width="100%" height="100%">
            <BarChart data={simulatedEntryWithScore} margin={{ top: 6, right: 12, left: 0, bottom: 6 }}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="team" tick={{ fontSize: 12 }} />
              <YAxis tick={{ fontSize: 12 }} />
              <Tooltip formatter={(value: number) => [`${value.toFixed(1)} points`, 'Your points']} />
              <Bar dataKey="yourPoints" radius={[8, 8, 0, 0]}>
                {simulatedEntryWithScore.map((t) => (
                  <Cell key={t.key} fill={t.fill} />
                ))}
              </Bar>
            </BarChart>
          </ResponsiveContainer>
        </div>
      </div>
    </Card>
  );
}
