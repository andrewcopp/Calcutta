import React, { useMemo, useState } from 'react';
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
  ScatterChart,
  Scatter,
} from 'recharts';
import { AnalyticsResponse, SeedInvestmentDistributionResponse } from '../types/analytics';
import { useQuery } from '@tanstack/react-query';
import { apiClient } from '../api/apiClient';
import { queryKeys } from '../queryKeys';
import { Tournament } from '../types/tournament';
import { Calcutta } from '../types/calcutta';

const COLORS = ['#0088FE', '#00C49F', '#FFBB28', '#FF8042', '#8884d8', '#82ca9d', '#ffc658', '#ff7c7c'];
const REGION_COLORS: Record<string, string> = {
  'East': '#0088FE',
  'West': '#00C49F',
  'South': '#FFBB28',
  'Midwest': '#FF8042',
};

export const AnalyticsPage: React.FC = () => {
  const API_URL = useMemo(() => import.meta.env.VITE_API_URL || import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080', []);

  const [activeTab, setActiveTab] = useState<'seeds' | 'regions' | 'teams' | 'variance'>('seeds');

  const [exportTournamentId, setExportTournamentId] = useState<string>('');
  const [exportCalcuttaId, setExportCalcuttaId] = useState<string>('');
  const [exportBusy, setExportBusy] = useState(false);
  const [exportError, setExportError] = useState<string | null>(null);

  const tournamentsQuery = useQuery({
    queryKey: queryKeys.tournaments.all(),
    staleTime: 30_000,
    queryFn: () => apiClient.get<Tournament[]>('/tournaments'),
  });

  const calcuttasQuery = useQuery({
    queryKey: queryKeys.calcuttas.all(),
    staleTime: 30_000,
    queryFn: () => apiClient.get<Calcutta[]>('/calcuttas'),
  });

  const analyticsQuery = useQuery({
    queryKey: queryKeys.analytics.all(),
    staleTime: 30_000,
    queryFn: () => apiClient.get<AnalyticsResponse>('/analytics'),
  });

  const seedInvestmentDistributionQuery = useQuery({
    queryKey: queryKeys.analytics.seedInvestmentDistribution(),
    staleTime: 30_000,
    queryFn: () => apiClient.get<SeedInvestmentDistributionResponse>('/analytics/seed-investment-distribution'),
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

  const seedInvestmentDistribution = seedInvestmentDistributionQuery.data;

  const tournaments = tournamentsQuery.data ?? [];
  const calcuttas = calcuttasQuery.data ?? [];
  const filteredCalcuttas = exportTournamentId ? calcuttas.filter((c) => c.tournamentId === exportTournamentId) : calcuttas;

  const downloadSnapshot = async () => {
    setExportError(null);
    setExportBusy(true);
    try {
      if (!exportTournamentId) {
        throw new Error('Please select a tournament');
      }
      if (!exportCalcuttaId) {
        throw new Error('Please select a calcutta');
      }

      const url = new URL(`${API_URL}/api/admin/analytics/export`);
      url.searchParams.set('tournamentId', exportTournamentId);
      url.searchParams.set('calcuttaId', exportCalcuttaId);

      const res = await apiClient.fetch(url.toString(), { credentials: 'include' });
      if (!res.ok) {
        const txt = await res.text().catch(() => '');
        throw new Error(txt || `Export failed (${res.status})`);
      }

      const blob = await res.blob();
      const cd = res.headers.get('content-disposition') || '';
      const match = /filename="([^"]+)"/i.exec(cd);
      const filename = match?.[1] || 'analytics-snapshot.zip';

      const objectUrl = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = objectUrl;
      a.download = filename;
      document.body.appendChild(a);
      a.click();
      a.remove();
      window.URL.revokeObjectURL(objectUrl);
    } catch (e) {
      setExportError(e instanceof Error ? e.message : String(e));
    } finally {
      setExportBusy(false);
    }
  };

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
        <h2 className="text-xl font-semibold mb-2">Export Analytics Snapshot</h2>
        <p className="text-gray-600 mb-4">
          Download a zip containing CSV tables and a manifest for offline Python analysis.
        </p>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-4 items-end">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Tournament</label>
            <select
              value={exportTournamentId}
              onChange={(e) => {
                setExportTournamentId(e.target.value);
                setExportCalcuttaId('');
              }}
              className="w-full border border-gray-300 rounded px-3 py-2"
              disabled={exportBusy}
            >
              <option value="">Select tournament</option>
              {tournaments.map((t) => (
                <option key={t.id} value={t.id}>
                  {t.name}
                </option>
              ))}
            </select>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Calcutta</label>
            <select
              value={exportCalcuttaId}
              onChange={(e) => setExportCalcuttaId(e.target.value)}
              className="w-full border border-gray-300 rounded px-3 py-2"
              disabled={exportBusy}
            >
              <option value="">Select calcutta</option>
              {filteredCalcuttas.map((c) => (
                <option key={c.id} value={c.id}>
                  {c.name}
                </option>
              ))}
            </select>
          </div>

          <div>
            <button
              onClick={downloadSnapshot}
              disabled={exportBusy}
              className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50"
            >
              Download snapshot (.zip)
            </button>
          </div>
        </div>

        {exportError && <div className="mt-4 text-red-600">{exportError}</div>}
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
          {seedInvestmentDistribution && (
            <div className="bg-white rounded-lg shadow-md p-6">
              <h2 className="text-xl font-semibold mb-4">Investment Distribution by Seed (Normalized)</h2>
              <p className="text-sm text-gray-600 mb-4">
                Each dot is a team-season (e.g. 2021 Duke) showing total investment (sum of all bids on that team) normalized by total bids in that calcutta. Teams with neither a bye nor a win are excluded.
              </p>

              <ResponsiveContainer width="100%" height={420}>
                <ScatterChart>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="seed" type="number" domain={[1, 16]} allowDecimals={false} />
                  <YAxis dataKey="normalizedBid" type="number" />
                  <Tooltip
                    formatter={(value: number, name: string, props: any) => {
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
                  <Scatter
                    name="Normalized Bid"
                    data={seedInvestmentDistribution.points}
                    fill="#0088FE"
                    opacity={0.35}
                  />
                </ScatterChart>
              </ResponsiveContainer>

              <div className="mt-6">
                <h3 className="text-lg font-semibold mb-3">Per-Seed Summary (Box-Plot Stats)</h3>
                <div className="overflow-x-auto">
                  <table className="min-w-full divide-y divide-gray-200">
                    <thead className="bg-gray-50">
                      <tr>
                        <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Seed</th>
                        <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Count</th>
                        <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Mean</th>
                        <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">StdDev</th>
                        <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Min</th>
                        <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Q1</th>
                        <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Median</th>
                        <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Q3</th>
                        <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Max</th>
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
              This analysis reveals which seeds have the most variance in investment patterns. High variance ratios indicate "ugly duckling" scenarios where some teams at a seed are hot while others are cold.
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
                        {variance.investmentMean.toFixed(4)}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        {variance.investmentStdDev.toFixed(4)}
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
