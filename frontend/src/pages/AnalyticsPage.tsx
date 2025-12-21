import React, { useState } from 'react';
import { Link } from 'react-router-dom';
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
  LineChart,
  Line,
} from 'recharts';
import { AnalyticsResponse } from '../types/analytics';
import { useQuery } from '@tanstack/react-query';
import { apiClient } from '../api/apiClient';
import { queryKeys } from '../queryKeys';

const COLORS = ['#0088FE', '#00C49F', '#FFBB28', '#FF8042', '#8884d8', '#82ca9d', '#ffc658', '#ff7c7c'];
const REGION_COLORS: Record<string, string> = {
  'East': '#0088FE',
  'West': '#00C49F',
  'South': '#FFBB28',
  'Midwest': '#FF8042',
};

export const AnalyticsPage: React.FC = () => {
  const [activeTab, setActiveTab] = useState<'seeds' | 'regions' | 'teams' | 'variance'>('seeds');

  const analyticsQuery = useQuery({
    queryKey: queryKeys.analytics.all(),
    staleTime: 30_000,
    queryFn: () => apiClient.get<AnalyticsResponse>('/analytics'),
  });

  if (analyticsQuery.isLoading) {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="text-center">Loading analytics...</div>
      </div>
    );
  }

  if (analyticsQuery.isError) {
    const message = analyticsQuery.error instanceof Error ? analyticsQuery.error.message : 'An error occurred';
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded">
          Error: {message}
        </div>
      </div>
    );
  }

  const analytics = analyticsQuery.data;

  if (!analytics) {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="text-center">No analytics data available</div>
      </div>
    );
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-6">
        <Link to="/admin" className="text-blue-600 hover:text-blue-800 mb-4 inline-block">
          ← Back to Admin Console
        </Link>
        <h1 className="text-3xl font-bold">Calcutta Analytics</h1>
        <p className="text-gray-600 mt-2">
          Historical analysis across all calcuttas to identify trends and patterns
        </p>
      </div>

      <div className="bg-white rounded-lg shadow-md p-6 mb-6">
        <h2 className="text-xl font-semibold mb-4">Overall Summary</h2>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div className="bg-blue-50 p-4 rounded">
            <div className="text-sm text-gray-600">Total Points Scored</div>
            <div className="text-2xl font-bold text-blue-600">
              {analytics.totalPoints.toFixed(2)}
            </div>
          </div>
          <div className="bg-green-50 p-4 rounded">
            <div className="text-sm text-gray-600">Total Investment</div>
            <div className="text-2xl font-bold text-green-600">
              ${analytics.totalInvestment.toFixed(2)}
            </div>
          </div>
          <div className="bg-purple-50 p-4 rounded">
            <div className="text-sm text-gray-600">Baseline ROI</div>
            <div className="text-2xl font-bold text-purple-600">
              {analytics.baselineROI.toFixed(3)}
            </div>
            <div className="text-xs text-gray-500 mt-1">
              Points per dollar (raw)
            </div>
          </div>
        </div>
        <div className="mt-4 p-3 bg-gray-50 rounded text-sm text-gray-700">
          <strong>ROI Explanation:</strong> ROI values are normalized where 1.0 = average performance. 
          Values &gt;1.0 indicate over-performance (better return than average), 
          while &lt;1.0 indicates under-performance (worse return than average).
        </div>
      </div>

      <div className="mb-6">
        <div className="border-b border-gray-200">
          <nav className="-mb-px flex space-x-8">
            <button
              onClick={() => setActiveTab('seeds')}
              className={`${
                activeTab === 'seeds'
                  ? 'border-blue-500 text-blue-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              } whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm`}
            >
              Seed Analytics
            </button>
            <button
              onClick={() => setActiveTab('regions')}
              className={`${
                activeTab === 'regions'
                  ? 'border-blue-500 text-blue-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              } whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm`}
            >
              Region Analytics
            </button>
            <button
              onClick={() => setActiveTab('teams')}
              className={`${
                activeTab === 'teams'
                  ? 'border-blue-500 text-blue-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              } whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm`}
            >
              Team Analytics
            </button>
            <button
              onClick={() => setActiveTab('variance')}
              className={`${
                activeTab === 'variance'
                  ? 'border-blue-500 text-blue-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              } whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm`}
            >
              Variance Analysis
            </button>
          </nav>
        </div>
      </div>

      {activeTab === 'seeds' && analytics.seedAnalytics && (
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
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Seed
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Teams
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Total Points
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Points %
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Total Investment
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Investment %
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Avg Points
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Avg Investment
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      ROI
                    </th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {analytics.seedAnalytics.map((seed) => (
                    <tr key={seed.seed}>
                      <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                        {seed.seed}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        {seed.teamCount}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        {seed.totalPoints.toFixed(2)}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        {seed.pointsPercentage.toFixed(2)}%
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        ${seed.totalInvestment.toFixed(2)}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        {seed.investmentPercentage.toFixed(2)}%
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        {seed.averagePoints.toFixed(2)}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        ${seed.averageInvestment.toFixed(2)}
                      </td>
                      <td className={`px-6 py-4 whitespace-nowrap text-sm font-semibold ${
                        seed.roi > 1.0 ? 'text-green-600' : seed.roi < 1.0 ? 'text-red-600' : 'text-gray-500'
                      }`}>
                        {seed.roi.toFixed(3)}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      )}

      {activeTab === 'regions' && analytics.regionAnalytics && (
        <div className="space-y-6">
          <div className="bg-white rounded-lg shadow-md p-6">
            <h2 className="text-xl font-semibold mb-4">Points Distribution by Region</h2>
            <p className="text-sm text-gray-600 mb-4">
              Expected: ~25% per region (control group analysis)
            </p>
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
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Region
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Teams
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Total Points
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Points %
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Total Investment
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Investment %
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      ROI
                    </th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {analytics.regionAnalytics.map((region) => (
                    <tr key={region.region}>
                      <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                        {region.region}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        {region.teamCount}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        {region.totalPoints.toFixed(2)}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        {region.pointsPercentage.toFixed(2)}%
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        ${region.totalInvestment.toFixed(2)}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        {region.investmentPercentage.toFixed(2)}%
                      </td>
                      <td className={`px-6 py-4 whitespace-nowrap text-sm font-semibold ${
                        region.roi > 1.0 ? 'text-green-600' : region.roi < 1.0 ? 'text-red-600' : 'text-gray-500'
                      }`}>
                        {region.roi.toFixed(3)}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      )}

      {activeTab === 'teams' && analytics.teamAnalytics && (
        <div className="space-y-6">
          <div className="bg-white rounded-lg shadow-md p-6">
            <h2 className="text-xl font-semibold mb-4">Team Analytics</h2>
            <p className="text-sm text-gray-600 mb-4">
              Top teams by total points scored. Note: This data is not yet normalized by seed - teams with better seeds will naturally score more points.
            </p>
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-gray-200">
                <thead className="bg-gray-50">
                  <tr>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      School
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Appearances
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Avg Seed
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Total Points
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Avg Points
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Total Investment
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Avg Investment
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      ROI
                    </th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {analytics.teamAnalytics.slice(0, 50).map((team) => (
                    <tr key={team.schoolId}>
                      <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                        {team.schoolName}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        {team.appearances}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        {team.averageSeed.toFixed(1)}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        {team.totalPoints.toFixed(2)}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        {team.averagePoints.toFixed(2)}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        ${team.totalInvestment.toFixed(2)}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        ${team.averageInvestment.toFixed(2)}
                      </td>
                      <td className={`px-6 py-4 whitespace-nowrap text-sm font-semibold ${
                        team.roi > 1.0 ? 'text-green-600' : team.roi < 1.0 ? 'text-red-600' : 'text-gray-500'
                      }`}>
                        {team.roi.toFixed(3)}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>

          <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4">
            <h3 className="text-sm font-semibold text-yellow-800 mb-2">Future Enhancement</h3>
            <p className="text-sm text-yellow-700">
              Team analytics will be enhanced in the future to normalize for seed position. This will help identify teams that consistently over-perform or under-perform relative to their seed, revealing potential biases in bidding behavior.
            </p>
          </div>
        </div>
      )}

      {activeTab === 'variance' && analytics.seedVarianceAnalytics && (
        <div className="space-y-6">
          <div className="bg-white rounded-lg shadow-md p-6">
            <h2 className="text-xl font-semibold mb-4">Investment Variance by Seed</h2>
            <p className="text-sm text-gray-600 mb-4">
              This analysis reveals which seeds have the most variance in investment patterns. High variance ratios indicate "ugly duckling" scenarios where some teams at a seed are hot while others are cold.
            </p>
            <ResponsiveContainer width="100%" height={400}>
              <BarChart data={analytics.seedVarianceAnalytics}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="seed" label={{ value: 'Seed', position: 'insideBottom', offset: -5 }} />
                <YAxis label={{ value: 'Standard Deviation ($)', angle: -90, position: 'insideLeft' }} />
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
              Variance Ratio = (Investment CV) / (Points CV). Ratios &gt; 1.0 indicate investment varies more than performance, suggesting "ugly duckling" opportunities.
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
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Seed
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Teams
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Avg Investment
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Investment StdDev
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Investment CV
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Avg Points
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Points StdDev
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Points CV
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Variance Ratio
                    </th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {analytics.seedVarianceAnalytics.map((variance) => (
                    <tr key={variance.seed} className={variance.varianceRatio > 1.5 ? 'bg-yellow-50' : ''}>
                      <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                        {variance.seed}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        {variance.teamCount}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        ${variance.investmentMean.toFixed(2)}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        ${variance.investmentStdDev.toFixed(2)}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        {variance.investmentCV.toFixed(3)}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        {variance.pointsMean.toFixed(2)}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        {variance.pointsStdDev.toFixed(2)}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        {variance.pointsCV.toFixed(3)}
                      </td>
                      <td className={`px-6 py-4 whitespace-nowrap text-sm font-semibold ${
                        variance.varianceRatio > 1.5 ? 'text-red-600' : variance.varianceRatio > 1.0 ? 'text-yellow-600' : 'text-green-600'
                      }`}>
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
              <li><strong>Standard Deviation (StdDev):</strong> Measures spread of values. Higher = more variance.</li>
              <li><strong>Coefficient of Variation (CV):</strong> StdDev / Mean. Allows comparison across different scales.</li>
              <li><strong>Variance Ratio:</strong> Investment CV / Points CV. Shows if investment varies more than performance.</li>
              <li><strong>Interpretation:</strong> Ratio &gt; 1.5 (highlighted) = Strong "ugly duckling" effect. Some teams at this seed are hot, others cold.</li>
            </ul>
          </div>
        </div>
      )}
    </div>
  );
};

export default AnalyticsPage;
