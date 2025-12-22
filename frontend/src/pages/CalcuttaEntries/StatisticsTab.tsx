import { Link } from 'react-router-dom';
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';

export const StatisticsTab: React.FC<{
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
}> = ({ calcuttaId, totalEntries, totalInvestment, totalReturns, averageReturn, returnsStdDev, seedInvestmentData, teamROIData }) => {
  return (
    <>
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6 mb-8">
        <div className="bg-white rounded-lg shadow p-6">
          <h2 className="text-xl font-semibold mb-4">Tournament Statistics</h2>
          <p className="text-gray-600">Total Entries: {totalEntries}</p>
          <p className="text-gray-600">Total Investment: ${totalInvestment.toFixed(2)}</p>
          <p className="text-gray-600">Total Returns: {totalReturns.toFixed(2)}</p>
          <p className="text-gray-600">Average Return: {averageReturn.toFixed(2)}</p>
          <p className="text-gray-600">Std Dev for Returns: {returnsStdDev.toFixed(2)}</p>
        </div>

        <div className="bg-white rounded-lg shadow p-6">
          <h2 className="text-xl font-semibold mb-4">Investment by Seed</h2>
          <div className="h-64">
            <ResponsiveContainer width="100%" height="100%">
              <BarChart data={seedInvestmentData}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="seed" label={{ value: 'Seed', position: 'insideBottom', offset: -5 }} />
                <YAxis label={{ value: 'Total Investment ($)', angle: -90, position: 'insideLeft' }} />
                <Tooltip formatter={(value: number) => [`$${value.toFixed(2)}`, 'Total Investment']} />
                <Bar dataKey="totalInvestment" fill="#4F46E5" />
              </BarChart>
            </ResponsiveContainer>
          </div>
          <div className="mt-4 text-center">
            <Link to={`/calcuttas/${calcuttaId}/teams`} className="text-blue-600 hover:text-blue-800 font-medium">
              View All Teams →
            </Link>
          </div>
        </div>
      </div>

      <div className="bg-white rounded-lg shadow p-6 mb-8">
        <h2 className="text-xl font-semibold mb-4">Team ROI - Tournament Heroes</h2>
        <p className="text-sm text-gray-600 mb-4">Return on Investment: Points scored per dollar spent. ROI = Points / (Investment + $1)</p>
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Rank</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Seed</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Region</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Team</th>
                <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Points</th>
                <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Investment</th>
                <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">ROI</th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {teamROIData.map((team, index) => (
                <tr key={team.teamId} className={index < 3 ? 'bg-green-50' : ''}>
                  <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">{index + 1}</td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{team.seed}</td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{team.region}</td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">{team.teamName}</td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-right text-gray-500">{team.points.toFixed(2)}</td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-right text-gray-500">${team.investment.toFixed(2)}</td>
                  <td
                    className={`px-6 py-4 whitespace-nowrap text-sm text-right font-semibold ${
                      index < 3 ? 'text-green-600' : 'text-gray-900'
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
};
