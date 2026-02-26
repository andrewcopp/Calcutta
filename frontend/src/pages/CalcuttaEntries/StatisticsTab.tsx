import { Link } from 'react-router-dom';
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';

type StatisticsTabProps = {
  calcuttaId: string;
  totalEntries: number;
  totalInvestment: number;
  totalReturns: number;
  averageReturn: number;
  returnsStdDev: number;
  seedInvestmentData: { seed: number; totalInvestment: number }[];
  teamROIData: {
    teamId: string;
    seed: number;
    region: string;
    teamName: string;
    points: number;
    investment: number;
    roi: number;
  }[];
};

export function StatisticsTab({
  calcuttaId,
  totalEntries,
  totalInvestment,
  totalReturns,
  averageReturn,
  returnsStdDev,
  seedInvestmentData,
  teamROIData,
}: StatisticsTabProps) {
  return (
    <>
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6 mb-8">
        <div className="bg-card rounded-lg shadow p-6">
          <h2 className="text-xl font-semibold mb-4">Tournament Statistics</h2>
          <p className="text-muted-foreground">Total Entries: {totalEntries}</p>
          <p className="text-muted-foreground">Total Investment: {totalInvestment.toFixed(2)} credits</p>
          <p className="text-muted-foreground">Total Returns: {totalReturns.toFixed(2)}</p>
          <p className="text-muted-foreground">Average Return: {averageReturn.toFixed(2)}</p>
          <p className="text-muted-foreground">Std Dev for Returns: {returnsStdDev.toFixed(2)}</p>
        </div>

        <div className="bg-card rounded-lg shadow p-6">
          <h2 className="text-xl font-semibold mb-4">Investment by Seed</h2>
          <div className="h-64">
            <ResponsiveContainer width="100%" height="100%">
              <BarChart data={seedInvestmentData}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="seed" label={{ value: 'Seed', position: 'insideBottom', offset: -5 }} />
                <YAxis label={{ value: 'Total Investment (credits)', angle: -90, position: 'insideLeft' }} />
                <Tooltip formatter={(value: number) => [`${value.toFixed(2)} credits`, 'Total Investment']} />
                <Bar dataKey="totalInvestment" fill="#4F46E5" />
              </BarChart>
            </ResponsiveContainer>
          </div>
          <div className="mt-4 text-center">
            <Link to={`/pools/${calcuttaId}/teams`} className="text-primary hover:text-primary font-medium">
              View All Teams →
            </Link>
          </div>
        </div>
      </div>

      <div className="bg-card rounded-lg shadow p-6 mb-8">
        <h2 className="text-xl font-semibold mb-4">Team ROI - Tournament Heroes</h2>
        <p className="text-sm text-muted-foreground mb-4">
          Return on investment: points scored per credit invested. ROI = Points / (Investment + 1)
        </p>
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-border">
            <thead className="bg-accent">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                  Rank
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                  Seed
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                  Region
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                  Team
                </th>
                <th className="px-6 py-3 text-right text-xs font-medium text-muted-foreground uppercase tracking-wider">
                  Points
                </th>
                <th className="px-6 py-3 text-right text-xs font-medium text-muted-foreground uppercase tracking-wider">
                  Investment
                </th>
                <th className="px-6 py-3 text-right text-xs font-medium text-muted-foreground uppercase tracking-wider">
                  ROI
                </th>
              </tr>
            </thead>
            <tbody className="bg-card divide-y divide-border">
              {teamROIData.map((team, index) => (
                <tr key={team.teamId} className={index < 3 ? 'bg-success/10' : ''}>
                  <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-foreground">{index + 1}</td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-muted-foreground">{team.seed}</td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-muted-foreground">{team.region}</td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-foreground">{team.teamName}</td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-right text-muted-foreground">
                    {team.points.toFixed(2)}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-right text-muted-foreground">
                    {team.investment.toFixed(2)} credits
                  </td>
                  <td
                    className={`px-6 py-4 whitespace-nowrap text-sm text-right font-semibold ${
                      index < 3 ? 'text-success' : 'text-foreground'
                    }`}
                  >
                    {team.roi.toFixed(3)}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </>
  );
}
