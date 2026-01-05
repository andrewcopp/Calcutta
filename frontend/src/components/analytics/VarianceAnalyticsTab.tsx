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
  ScatterChart,
  Scatter,
} from 'recharts';
import { AnalyticsResponse, SeedInvestmentDistributionResponse } from '../../types/analytics';

export function VarianceAnalyticsTab({
  analytics,
  seedInvestmentDistribution,
}: {
  analytics: AnalyticsResponse;
  seedInvestmentDistribution?: SeedInvestmentDistributionResponse;
}) {
  if (!analytics.seedVarianceAnalytics) {
    return null;
  }

  return (
    <div className="space-y-6">
      {seedInvestmentDistribution && (
        <div className="bg-white rounded-lg shadow-md p-6">
          <h2 className="text-xl font-semibold mb-4">Investment Distribution by Seed (Normalized)</h2>
          <p className="text-sm text-gray-600 mb-4">
            Each dot is a team-season (e.g. 2021 Duke) showing total investment (sum of all bids on that team)
            normalized by total bids in that calcutta. Teams with neither a bye nor a win are excluded.
          </p>

          <ResponsiveContainer width="100%" height={420}>
            <ScatterChart>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="seed" type="number" domain={[1, 16]} allowDecimals={false} />
              <YAxis dataKey="normalizedBid" type="number" />
              <Tooltip
                formatter={(value: number, name: string) => {
                  if (name === 'normalizedBid') return [value.toFixed(4), 'Normalized Total Bid'];
                  if (name === 'totalBid') return [`$${value.toFixed(2)}`, 'Total Bid'];
                  if (name === 'calcuttaTotalBid') return [`$${value.toFixed(2)}`, 'Calcutta Total'];
                  if (name === 'schoolName') return [String(value), 'Team'];
                  if (name === 'tournamentYear') return [String(value), 'Year'];
                  return [value, name];
                }}
                labelFormatter={(label: number) => `Seed ${label}`}
              />
              <Legend />
              <Scatter name="Normalized Bid" data={seedInvestmentDistribution.points} fill="#0088FE" opacity={0.35} />
            </ScatterChart>
          </ResponsiveContainer>

          <div className="mt-6">
            <h3 className="text-lg font-semibold mb-3">Per-Seed Summary (Box-Plot Stats)</h3>
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-gray-200">
                <thead className="bg-gray-50">
                  <tr>
                    <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Seed
                    </th>
                    <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Count
                    </th>
                    <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Mean
                    </th>
                    <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      StdDev
                    </th>
                    <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Min
                    </th>
                    <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Q1</th>
                    <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Median
                    </th>
                    <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Q3</th>
                    <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Max
                    </th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {seedInvestmentDistribution.summaries.map((s) => (
                    <tr key={s.seed}>
                      <td className="px-4 py-2 whitespace-nowrap text-sm font-medium text-gray-900">{s.seed}</td>
                      <td className="px-4 py-2 whitespace-nowrap text-sm text-gray-500">{s.count}</td>
                      <td className="px-4 py-2 whitespace-nowrap text-sm text-gray-500">{s.mean.toFixed(4)}</td>
                      <td className="px-4 py-2 whitespace-nowrap text-sm text-gray-500">{s.stdDev.toFixed(4)}</td>
                      <td className="px-4 py-2 whitespace-nowrap text-sm text-gray-500">{s.min.toFixed(4)}</td>
                      <td className="px-4 py-2 whitespace-nowrap text-sm text-gray-500">{s.q1.toFixed(4)}</td>
                      <td className="px-4 py-2 whitespace-nowrap text-sm text-gray-500">{s.median.toFixed(4)}</td>
                      <td className="px-4 py-2 whitespace-nowrap text-sm text-gray-500">{s.q3.toFixed(4)}</td>
                      <td className="px-4 py-2 whitespace-nowrap text-sm text-gray-500">{s.max.toFixed(4)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      )}

      <div className="bg-white rounded-lg shadow-md p-6">
        <h2 className="text-xl font-semibold mb-4">Investment Variance by Seed</h2>
        <p className="text-sm text-gray-600 mb-4">
          This analysis reveals which seeds have the most variance in investment patterns. High variance ratios indicate
          "ugly duckling" scenarios where some teams at a seed are hot while others are cold.
        </p>
        <ResponsiveContainer width="100%" height={400}>
          <BarChart data={analytics.seedVarianceAnalytics}>
            <CartesianGrid strokeDasharray="3 3" />
            <XAxis dataKey="seed" label={{ value: 'Seed', position: 'insideBottom', offset: -5 }} />
            <YAxis label={{ value: 'Standard Deviation (Normalized Bid)', angle: -90, position: 'insideLeft' }} />
            <Tooltip />
            <Legend />
            <Bar dataKey="investmentStdDev" fill="#0088FE" name="Investment StdDev" />
            <Bar dataKey="pointsStdDev" fill="#00C49F" name="Points StdDev" />
          </BarChart>
        </ResponsiveContainer>
      </div>

      <div className="bg-white rounded-lg shadow-md p-6">
        <h2 className="text-xl font-semibold mb-4">Variance Ratio by Seed</h2>
        <p className="text-sm text-gray-600 mb-4">
          Variance Ratio = (Investment CV) / (Points CV). Ratios &gt; 1.0 indicate investment varies more than
          performance, suggesting "ugly duckling" opportunities.
        </p>
        <ResponsiveContainer width="100%" height={300}>
          <LineChart data={analytics.seedVarianceAnalytics}>
            <CartesianGrid strokeDasharray="3 3" />
            <XAxis dataKey="seed" />
            <YAxis />
            <Tooltip />
            <Legend />
            <Line type="monotone" dataKey="varianceRatio" stroke="#8884d8" strokeWidth={2} name="Variance Ratio" />
          </LineChart>
        </ResponsiveContainer>
      </div>

      <div className="bg-white rounded-lg shadow-md p-6">
        <h2 className="text-xl font-semibold mb-4">Detailed Variance Analysis</h2>
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Seed</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Data Points
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Avg Normalized Bid
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Normalized Bid StdDev
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Investment CV
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Avg Points</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Points StdDev
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Points CV</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Variance Ratio
                </th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {analytics.seedVarianceAnalytics.map((variance) => (
                <tr key={variance.seed} className={variance.varianceRatio > 1.5 ? 'bg-yellow-50' : ''}>
                  <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">{variance.seed}</td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{variance.teamCount}</td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{variance.investmentMean.toFixed(4)}</td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{variance.investmentStdDev.toFixed(4)}</td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{variance.investmentCV.toFixed(3)}</td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{variance.pointsMean.toFixed(2)}</td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{variance.pointsStdDev.toFixed(2)}</td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{variance.pointsCV.toFixed(3)}</td>
                  <td
                    className={`px-6 py-4 whitespace-nowrap text-sm font-semibold ${
                      variance.varianceRatio > 1.5
                        ? 'text-red-600'
                        : variance.varianceRatio > 1.0
                          ? 'text-yellow-600'
                          : 'text-green-600'
                    }`}
                  >
                    {variance.varianceRatio.toFixed(3)}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
        <h3 className="text-sm font-semibold text-blue-800 mb-2">Understanding the Metrics</h3>
        <ul className="text-sm text-blue-700 space-y-2">
          <li>
            <strong>Standard Deviation (StdDev):</strong> Measures spread of values. Higher = more variance.
          </li>
          <li>
            <strong>Coefficient of Variation (CV):</strong> StdDev / Mean. Allows comparison across different scales.
          </li>
          <li>
            <strong>Variance Ratio:</strong> Investment CV / Points CV. Shows if investment varies more than performance.
          </li>
          <li>
            <strong>Interpretation:</strong> Ratio &gt; 1.5 (highlighted) = Strong "ugly duckling" effect. Some teams at this
            seed are hot, others cold.
          </li>
        </ul>
      </div>
    </div>
  );
}
