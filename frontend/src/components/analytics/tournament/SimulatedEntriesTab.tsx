import React from 'react';
import { useQuery } from '@tanstack/react-query';
import { analyticsService } from '../../../services/analyticsService';

interface TeamSimulatedEntry {
  team_id: string;
  school_name: string;
  seed: number;
  region: string;
  expected_points: number;
  expected_market: number;
  expected_roi: number;
  our_bid: number;
  our_roi: number;
}

// Simulated Entries Tab Component
export function SimulatedEntriesTab({ calcuttaId }: { calcuttaId: string | null }) {
  const [sortColumn, setSortColumn] = React.useState<keyof TeamSimulatedEntry>('seed');
  const [sortDirection, setSortDirection] = React.useState<'asc' | 'desc'>('asc');

  const { data: simulatedEntry, isLoading } = useQuery<{ teams: TeamSimulatedEntry[] } | null>({
    queryKey: ['analytics', 'simulated-entry', calcuttaId],
    queryFn: async () => {
      if (!calcuttaId) return null;
      return analyticsService.getCalcuttaSimulatedEntry<{ teams: TeamSimulatedEntry[] }>(calcuttaId);
    },
    enabled: !!calcuttaId,
  });

  const formatPoints = (points: number) => points.toFixed(1);
  const formatROI = (roi: number) => roi.toFixed(2);

  if (!calcuttaId) {
    return <div className="text-gray-500">Select a calcutta above to view simulated entries.</div>;
  }

  const handleSort = (column: keyof TeamSimulatedEntry) => {
    if (sortColumn === column) {
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc');
    } else {
      setSortColumn(column);
      setSortDirection('asc');
    }
  };

  const sortedTeams = React.useMemo(() => {
    if (!simulatedEntry?.teams) return [];

    return [...simulatedEntry.teams].sort((a, b) => {
      const aVal = a[sortColumn];
      const bVal = b[sortColumn];

      if (typeof aVal === 'string' && typeof bVal === 'string') {
        return sortDirection === 'asc' ? aVal.localeCompare(bVal) : bVal.localeCompare(aVal);
      }

      if (typeof aVal === 'number' && typeof bVal === 'number') {
        return sortDirection === 'asc' ? aVal - bVal : bVal - aVal;
      }

      return 0;
    });
  }, [simulatedEntry?.teams, sortColumn, sortDirection]);

  const SortIcon = ({ column }: { column: keyof TeamSimulatedEntry }) => {
    if (sortColumn !== column) {
      return <span className="ml-1 text-gray-400">⇅</span>;
    }
    return sortDirection === 'asc' ? <span className="ml-1">↑</span> : <span className="ml-1">↓</span>;
  };

  return (
    <div>
      <h2 className="text-xl font-semibold mb-4">Simulated Entry</h2>
      <p className="text-gray-600 mb-6">
        Detailed investment report showing expected performance, market predictions, and ROI analysis for all teams.
      </p>

      {isLoading ? (
        <div className="text-gray-500">Loading simulated entry data...</div>
      ) : simulatedEntry?.teams ? (
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th
                  className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider sticky left-0 bg-gray-50 cursor-pointer hover:bg-gray-100"
                  onClick={() => handleSort('school_name')}
                >
                  Team <SortIcon column="school_name" />
                </th>
                <th
                  className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider cursor-pointer hover:bg-gray-100"
                  onClick={() => handleSort('seed')}
                >
                  Seed <SortIcon column="seed" />
                </th>
                <th
                  className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider cursor-pointer hover:bg-gray-100"
                  onClick={() => handleSort('region')}
                >
                  Region <SortIcon column="region" />
                </th>
                <th
                  className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider cursor-pointer hover:bg-gray-100"
                  onClick={() => handleSort('expected_points')}
                >
                  Exp Pts <SortIcon column="expected_points" />
                </th>
                <th
                  className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider cursor-pointer hover:bg-gray-100"
                  onClick={() => handleSort('expected_market')}
                >
                  Exp Mkt <SortIcon column="expected_market" />
                </th>
                <th
                  className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider cursor-pointer hover:bg-gray-100"
                  onClick={() => handleSort('expected_roi')}
                >
                  Exp ROI <SortIcon column="expected_roi" />
                </th>
                <th
                  className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider bg-blue-50 cursor-pointer hover:bg-blue-100"
                  onClick={() => handleSort('our_bid')}
                >
                  Our Bid <SortIcon column="our_bid" />
                </th>
                <th
                  className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider bg-blue-50 cursor-pointer hover:bg-blue-100"
                  onClick={() => handleSort('our_roi')}
                >
                  Our ROI <SortIcon column="our_roi" />
                </th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {sortedTeams.map((team) => (
                <tr key={team.team_id} className="hover:bg-gray-50">
                  <td className="px-4 py-3 text-sm font-medium text-gray-900 sticky left-0 bg-white">{team.school_name}</td>
                  <td className="px-4 py-3 text-sm text-center text-gray-700">{team.seed}</td>
                  <td className="px-4 py-3 text-sm text-center text-gray-600">{team.region}</td>
                  <td className="px-4 py-3 text-sm text-center text-gray-700">{formatPoints(team.expected_points)}</td>
                  <td className="px-4 py-3 text-sm text-center text-gray-700">{formatPoints(team.expected_market)}</td>
                  <td className="px-4 py-3 text-sm text-center text-gray-700">{formatROI(team.expected_roi)}</td>
                  <td className="px-4 py-3 text-sm text-center font-semibold text-blue-700 bg-blue-50">
                    {team.our_bid > 0 ? formatPoints(team.our_bid) : '-'}
                  </td>
                  <td className="px-4 py-3 text-sm text-center font-semibold text-blue-700 bg-blue-50">
                    {team.our_bid > 0 ? formatROI(team.our_roi) : '-'}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      ) : (
        <div className="text-gray-500">No simulated entry data available for this tournament.</div>
      )}

      {simulatedEntry?.teams && (
        <div className="mt-4 text-sm text-gray-600">
          <p className="mb-2">Coming soon:</p>
          <ul className="text-sm text-gray-600 list-disc list-inside space-y-1">
            <li>Portfolio optimization (Our Bid column will show recommended allocations)</li>
            <li>Actual market data integration</li>
            <li>ROI degradation analysis</li>
          </ul>
        </div>
      )}
    </div>
  );
}
