import React, { useState } from 'react';
import { Link } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { apiClient } from '../api/apiClient';
import { queryKeys } from '../queryKeys';
import { Alert } from '../components/ui/Alert';
import { Card } from '../components/ui/Card';
import { LoadingState } from '../components/ui/LoadingState';
import { PageContainer, PageHeader } from '../components/ui/Page';
import {
  BestTeam,
  BestTeamsResponse,
  CareerLeaderboardResponse,
  CareerLeaderboardRow,
  EntryLeaderboardResponse,
  EntryLeaderboardRow,
  InvestmentLeaderboardResponse,
  InvestmentLeaderboardRow,
} from '../types/hallOfFame';

export const HallOfFamePage: React.FC = () => {
  const [activeTab, setActiveTab] = useState<'bestTeams' | 'bestInvestments' | 'bestEntries' | 'bestCareers'>('bestTeams');
  const [hideInactiveCareers, setHideInactiveCareers] = useState<boolean>(false);

  const formatDollarsFromCents = (cents?: number) => {
    if (!cents) return '$0';
    const abs = Math.abs(cents);
    const dollars = Math.floor(abs / 100);
    const remainder = abs % 100;
    const sign = cents < 0 ? '-' : '';
    if (remainder === 0) return `${sign}$${dollars}`;
    return `${sign}$${dollars}.${remainder.toString().padStart(2, '0')}`;
  };

  const bestTeamsQuery = useQuery<BestTeamsResponse, Error>({
    queryKey: queryKeys.hallOfFame.bestTeams(200),
    staleTime: 30_000,
    queryFn: () => apiClient.get<BestTeamsResponse>('/hall-of-fame/best-teams?limit=200'),
  });

  const bestInvestmentsQuery = useQuery<InvestmentLeaderboardResponse, Error>({
    queryKey: queryKeys.hallOfFame.bestInvestments(200),
    staleTime: 30_000,
    queryFn: () => apiClient.get<InvestmentLeaderboardResponse>('/hall-of-fame/best-investments?limit=200'),
  });

  const bestEntriesQuery = useQuery<EntryLeaderboardResponse, Error>({
    queryKey: queryKeys.hallOfFame.bestEntries(200),
    staleTime: 30_000,
    queryFn: () => apiClient.get<EntryLeaderboardResponse>('/hall-of-fame/best-entries?limit=200'),
  });

  const bestCareersQuery = useQuery<CareerLeaderboardResponse, Error>({
    queryKey: queryKeys.hallOfFame.bestCareers(200),
    staleTime: 30_000,
    queryFn: () => apiClient.get<CareerLeaderboardResponse>('/hall-of-fame/best-careers?limit=200'),
  });

  return (
    <PageContainer>
      <PageHeader
        title="Hall of Fame"
        subtitle="Leaderboards across all calcuttas (normalized for year-to-year comparisons)."
        actions={
          <Link to="/admin" className="text-blue-600 hover:text-blue-800">
            Back to Admin Console
          </Link>
        }
      />

      <div className="mb-6">
        <div className="border-b border-gray-200">
          <nav className="-mb-px flex space-x-8">
            <button
              onClick={() => setActiveTab('bestTeams')}
              className={`${
                activeTab === 'bestTeams'
                  ? 'border-blue-500 text-blue-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              } whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm`}
            >
              Best Teams
            </button>

            <button
              onClick={() => setActiveTab('bestInvestments')}
              className={`${
                activeTab === 'bestInvestments'
                  ? 'border-blue-500 text-blue-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              } whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm`}
            >
              Best Investments
            </button>

            <button
              onClick={() => setActiveTab('bestEntries')}
              className={`${
                activeTab === 'bestEntries'
                  ? 'border-blue-500 text-blue-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              } whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm`}
            >
              Best Entries
            </button>

            <button
              onClick={() => setActiveTab('bestCareers')}
              className={`${
                activeTab === 'bestCareers'
                  ? 'border-blue-500 text-blue-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              } whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm`}
            >
              Best Careers
            </button>
          </nav>
        </div>
      </div>

      {activeTab === 'bestTeams' && (
        <Card>
          <h2 className="text-xl font-semibold mb-2">Best Teams</h2>
          <p className="text-sm text-gray-600 mb-4">
            Teams ranked by normalized ROI where 1.0 = average performance within that Calcutta (levels the playing field across seeds). Yes, we call it
            "Adjusted for Inflation".
          </p>

          {bestTeamsQuery.isLoading && <LoadingState label="Loading best teams..." layout="inline" />}

          {bestTeamsQuery.isError && (
            <Alert variant="error">Error: {bestTeamsQuery.error instanceof Error ? bestTeamsQuery.error.message : 'An error occurred'}</Alert>
          )}

          {bestTeamsQuery.data && (
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-gray-200">
                <thead className="bg-gray-50">
                  <tr>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Rank</th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Year</th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Seed</th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Team</th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Points</th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Total Investment</th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Raw ROI</th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Normalized ROI</th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {bestTeamsQuery.data.teams.map((team: BestTeam, idx: number) => (
                    <tr key={`${team.calcuttaId}-${team.teamId}`}>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{idx + 1}</td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{team.tournamentYear || ''}</td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{team.seed}</td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">{team.schoolName}</td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{team.teamPoints.toFixed(0)}</td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">${team.totalBid.toFixed(2)}</td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{team.rawROI.toFixed(3)}</td>
                      <td
                        className={`px-6 py-4 whitespace-nowrap text-sm font-semibold ${
                          team.normalizedROI > 1.0 ? 'text-green-600' : team.normalizedROI < 1.0 ? 'text-red-600' : 'text-gray-500'
                        }`}
                      >
                        {team.normalizedROI.toFixed(3)}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </Card>
      )}

      {activeTab === 'bestInvestments' && (
        <Card>
          <h2 className="text-xl font-semibold mb-2">Best Investments</h2>
          <p className="text-sm text-gray-600 mb-4">
            Individual picks ranked by normalized returns: Raw Returns / (Total Returns / Total Investment), where Total Investment = $100 × participants. Yes, we call it "Adjusted for Inflation".
          </p>

          {bestInvestmentsQuery.isLoading && <LoadingState label="Loading best investments..." layout="inline" />}

          {bestInvestmentsQuery.isError && (
            <Alert variant="error">Error: {bestInvestmentsQuery.error instanceof Error ? bestInvestmentsQuery.error.message : 'An error occurred'}</Alert>
          )}

          {bestInvestmentsQuery.data && (
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-gray-200">
                <thead className="bg-gray-50">
                  <tr>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Rank</th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Entry</th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Year</th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Seed</th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Team</th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Investment</th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Ownership</th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Raw Returns</th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Normalized Returns</th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {bestInvestmentsQuery.data.investments.map((inv: InvestmentLeaderboardRow, idx: number) => (
                    <tr key={`${inv.calcuttaId}-${inv.entryId}-${inv.teamId}-${idx}`}>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{idx + 1}</td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{inv.entryName}</td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{inv.tournamentYear || ''}</td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{inv.seed}</td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">{inv.schoolName}</td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">${inv.investment.toFixed(2)}</td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{(inv.ownershipPercentage * 100).toFixed(2)}%</td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{inv.rawReturns.toFixed(2)}</td>
                      <td
                        className={`px-6 py-4 whitespace-nowrap text-sm font-semibold ${
                          inv.normalizedReturns > 1.0
                            ? 'text-green-600'
                            : inv.normalizedReturns < 1.0
                              ? 'text-red-600'
                              : 'text-gray-500'
                        }`}
                      >
                        {inv.normalizedReturns.toFixed(3)}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </Card>
      )}

      {activeTab === 'bestEntries' && (
        <Card>
          <h2 className="text-xl font-semibold mb-2">Best Entries</h2>
          <p className="text-sm text-gray-600 mb-4">
            Entries ranked by normalized returns: Entry Total Returns / (Calcutta Total Returns / Calcutta Total Investment), where Calcutta Total Investment = $100 × participants.
          </p>

          {bestEntriesQuery.isLoading && <LoadingState label="Loading best entries..." layout="inline" />}

          {bestEntriesQuery.isError && (
            <Alert variant="error">Error: {bestEntriesQuery.error instanceof Error ? bestEntriesQuery.error.message : 'An error occurred'}</Alert>
          )}

          {bestEntriesQuery.data && (
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-gray-200">
                <thead className="bg-gray-50">
                  <tr>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Rank</th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Entry</th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Year</th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Total Returns</th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Total Participants</th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Average Returns</th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Normalized Returns</th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {bestEntriesQuery.data.entries.map((e: EntryLeaderboardRow, idx: number) => (
                    <tr key={`${e.calcuttaId}-${e.entryId}`}>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{idx + 1}</td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">{e.entryName}</td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{e.tournamentYear || ''}</td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{e.totalReturns.toFixed(2)}</td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{e.totalParticipants}</td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{e.averageReturns.toFixed(2)}</td>
                      <td
                        className={`px-6 py-4 whitespace-nowrap text-sm font-semibold ${
                          e.normalizedReturns > 1.0 ? 'text-green-600' : e.normalizedReturns < 1.0 ? 'text-red-600' : 'text-gray-500'
                        }`}
                      >
                        {e.normalizedReturns.toFixed(3)}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </Card>
      )}

      {activeTab === 'bestCareers' && (
        <Card>
          <h2 className="text-xl font-semibold mb-2">Best Careers</h2>
          <p className="text-sm text-gray-600 mb-4">
            Careers ranked by average winnings per year (not shown). Tie breaks: wins, podiums, payouts, then top 10s.
          </p>

          <div className="mb-4 flex items-center gap-2">
            <input
              id="hideInactiveCareers"
              type="checkbox"
              checked={hideInactiveCareers}
              onChange={(e) => setHideInactiveCareers(e.target.checked)}
              className="h-4 w-4"
            />
            <label htmlFor="hideInactiveCareers" className="text-sm text-gray-700">
              Hide inactive careers
            </label>
          </div>

          {bestCareersQuery.isLoading && <LoadingState label="Loading best careers..." layout="inline" />}

          {bestCareersQuery.isError && (
            <Alert variant="error">Error: {bestCareersQuery.error instanceof Error ? bestCareersQuery.error.message : 'An error occurred'}</Alert>
          )}

          {bestCareersQuery.data && (
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-gray-200">
                <thead className="bg-gray-50">
                  <tr>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Rank</th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Name</th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Years</th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Best Finish</th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Wins</th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Podiums</th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Payouts</th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Top 10s</th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Career Earnings</th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {(hideInactiveCareers
                    ? bestCareersQuery.data.careers.filter((c: CareerLeaderboardRow) => c.activeInLatestCalcutta)
                    : bestCareersQuery.data.careers
                  ).map((c: CareerLeaderboardRow, idx: number) => (
                    <tr key={`${c.entryName}-${idx}`}>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{idx + 1}</td>
                      <td className={`px-6 py-4 whitespace-nowrap text-sm text-gray-900 ${c.activeInLatestCalcutta ? 'font-bold' : 'font-medium'}`}>{c.entryName}</td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{c.years}</td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{c.bestFinish}</td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{c.wins}</td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{c.podiums}</td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{c.inTheMoneys}</td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{c.top10s}</td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{formatDollarsFromCents(c.careerEarningsCents)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </Card>
      )}
    </PageContainer>
  );
};

export default HallOfFamePage;
