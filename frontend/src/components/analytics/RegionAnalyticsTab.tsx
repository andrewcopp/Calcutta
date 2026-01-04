import React from 'react';
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
  PieChart,
  Pie,
  Cell,
} from 'recharts';
import { AnalyticsResponse } from '../../types/analytics';

const REGION_COLORS: Record<string, string> = {
  East: '#0088FE',
  West: '#00C49F',
  South: '#FFBB28',
  Midwest: '#FF8042',
};

export function RegionAnalyticsTab({ analytics }: { analytics: AnalyticsResponse }) {
  if (!analytics.regionAnalytics) {
    return null;
  }

  return (
    <div className="space-y-6">
      <div className="bg-white rounded-lg shadow-md p-6">
        <h2 className="text-xl font-semibold mb-4">Points Distribution by Region</h2>
        <p className="text-sm text-gray-600 mb-4">Expected: ~25% per region (control group analysis)</p>
        <ResponsiveContainer width="100%" height={400}>
          <PieChart>
            <Pie
              data={analytics.regionAnalytics}
              cx="50%"
              cy="50%"
              labelLine={false}
              label={({ region, pointsPercentage }) => `${region}: ${pointsPercentage.toFixed(1)}%`}
              outerRadius={120}
              fill="#8884d8"
              dataKey="pointsPercentage"
            >
              {analytics.regionAnalytics.map((entry) => (
                <Cell key={`cell-${entry.region}`} fill={REGION_COLORS[entry.region] || '#999999'} />
              ))}
            </Pie>
            <Tooltip formatter={(value: number) => `${value.toFixed(2)}%`} />
          </PieChart>
        </ResponsiveContainer>
      </div>

      <div className="bg-white rounded-lg shadow-md p-6">
        <h2 className="text-xl font-semibold mb-4">Investment Distribution by Region</h2>
        <ResponsiveContainer width="100%" height={400}>
          <PieChart>
            <Pie
              data={analytics.regionAnalytics}
              cx="50%"
              cy="50%"
              labelLine={false}
              label={({ region, investmentPercentage }) => `${region}: ${investmentPercentage.toFixed(1)}%`}
              outerRadius={120}
              fill="#8884d8"
              dataKey="investmentPercentage"
            >
              {analytics.regionAnalytics.map((entry) => (
                <Cell key={`cell-${entry.region}`} fill={REGION_COLORS[entry.region] || '#999999'} />
              ))}
            </Pie>
            <Tooltip formatter={(value: number) => `${value.toFixed(2)}%`} />
          </PieChart>
        </ResponsiveContainer>
      </div>

      <div className="bg-white rounded-lg shadow-md p-6">
        <h2 className="text-xl font-semibold mb-4">Points vs Investment by Region</h2>
        <ResponsiveContainer width="100%" height={300}>
          <BarChart data={analytics.regionAnalytics}>
            <CartesianGrid strokeDasharray="3 3" />
            <XAxis dataKey="region" />
            <YAxis label={{ value: 'Percentage (%)', angle: -90, position: 'insideLeft' }} />
            <Tooltip formatter={(value: number) => `${value.toFixed(2)}%`} />
            <Legend />
            <Bar dataKey="pointsPercentage" fill="#0088FE" name="Points %" />
            <Bar dataKey="investmentPercentage" fill="#00C49F" name="Investment %" />
          </BarChart>
        </ResponsiveContainer>
      </div>

      <div className="bg-white rounded-lg shadow-md p-6">
        <h2 className="text-xl font-semibold mb-4">Region Analytics Table</h2>
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Region</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Teams</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Total Points</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Points %</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Total Investment
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Investment %
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">ROI</th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {analytics.regionAnalytics.map((region) => (
                <tr key={region.region}>
                  <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">{region.region}</td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{region.teamCount}</td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{region.totalPoints.toFixed(2)}</td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{region.pointsPercentage.toFixed(2)}%</td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">${region.totalInvestment.toFixed(2)}</td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    {region.investmentPercentage.toFixed(2)}%
                  </td>
                  <td
                    className={`px-6 py-4 whitespace-nowrap text-sm font-semibold ${
                      region.roi > 1.0 ? 'text-green-600' : region.roi < 1.0 ? 'text-red-600' : 'text-gray-500'
                    }`}
                  >
                    {region.roi.toFixed(3)}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
