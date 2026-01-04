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
  LineChart,
  Line,
} from 'recharts';
import { AnalyticsResponse } from '../../types/analytics';

export function SeedAnalyticsTab({ analytics }: { analytics: AnalyticsResponse }) {
  if (!analytics.seedAnalytics) {
    return null;
  }

  return (
    <div className="space-y-6">
      <div className="bg-white rounded-lg shadow-md p-6">
        <h2 className="text-xl font-semibold mb-4">Points vs Investment by Seed</h2>
        <p className="text-sm text-gray-600 mb-4">
          Comparing the percentage of total points scored vs percentage of total investment for each seed
        </p>
        <ResponsiveContainer width="100%" height={400}>
          <BarChart data={analytics.seedAnalytics}>
            <CartesianGrid strokeDasharray="3 3" />
            <XAxis dataKey="seed" label={{ value: 'Seed', position: 'insideBottom', offset: -5 }} />
            <YAxis label={{ value: 'Percentage (%)', angle: -90, position: 'insideLeft' }} />
            <Tooltip formatter={(value: number) => `${value.toFixed(2)}%`} />
            <Legend />
            <Bar dataKey="pointsPercentage" fill="#0088FE" name="Points %" />
            <Bar dataKey="investmentPercentage" fill="#00C49F" name="Investment %" />
          </BarChart>
        </ResponsiveContainer>
      </div>

      <div className="bg-white rounded-lg shadow-md p-6">
        <h2 className="text-xl font-semibold mb-4">Average Points per Team by Seed</h2>
        <ResponsiveContainer width="100%" height={300}>
          <LineChart data={analytics.seedAnalytics}>
            <CartesianGrid strokeDasharray="3 3" />
            <XAxis dataKey="seed" />
            <YAxis />
            <Tooltip />
            <Legend />
            <Line type="monotone" dataKey="averagePoints" stroke="#8884d8" name="Avg Points" />
          </LineChart>
        </ResponsiveContainer>
      </div>

      <div className="bg-white rounded-lg shadow-md p-6">
        <h2 className="text-xl font-semibold mb-4">Seed Analytics Table</h2>
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Seed</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Teams</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Total Points</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Points %</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Total Investment
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Investment %
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Avg Points</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Avg Investment
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">ROI</th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {analytics.seedAnalytics.map((seed) => (
                <tr key={seed.seed}>
                  <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">{seed.seed}</td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{seed.teamCount}</td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{seed.totalPoints.toFixed(2)}</td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{seed.pointsPercentage.toFixed(2)}%</td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">${seed.totalInvestment.toFixed(2)}</td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    {seed.investmentPercentage.toFixed(2)}%
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{seed.averagePoints.toFixed(2)}</td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">${seed.averageInvestment.toFixed(2)}</td>
                  <td
                    className={`px-6 py-4 whitespace-nowrap text-sm font-semibold ${
                      seed.roi > 1.0 ? 'text-green-600' : seed.roi < 1.0 ? 'text-red-600' : 'text-gray-500'
                    }`}
                  >
                    {seed.roi.toFixed(3)}
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
