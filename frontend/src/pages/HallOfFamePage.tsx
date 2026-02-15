import React, { useMemo, useState } from 'react';
import { Link } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { type ColumnDef } from '@tanstack/react-table';
import { apiClient } from '../api/apiClient';
import { queryKeys } from '../queryKeys';
import { Alert } from '../components/ui/Alert';
import { Breadcrumb } from '../components/ui/Breadcrumb';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { DataTable } from '../components/ui/DataTable';
import { LoadingState } from '../components/ui/LoadingState';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { Tabs, TabsList, TabsTrigger, TabsContent } from '../components/ui/Tabs';
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

const formatDollarsFromCents = (cents?: number) => {
  if (!cents) return '$0';
  const abs = Math.abs(cents);
  const dollars = Math.floor(abs / 100);
  const remainder = abs % 100;
  const sign = cents < 0 ? '-' : '';
  if (remainder === 0) return `${sign}$${dollars}`;
  return `${sign}$${dollars}.${remainder.toString().padStart(2, '0')}`;
};

const bestTeamsColumns: ColumnDef<BestTeam, unknown>[] = [
  { id: 'rank', header: 'Rank', cell: ({ row }) => row.index + 1, enableSorting: false },
  { accessorKey: 'tournamentYear', header: 'Year' },
  { accessorKey: 'seed', header: 'Seed' },
  { accessorKey: 'schoolName', header: 'Team' },
  { accessorKey: 'teamPoints', header: 'Points', cell: ({ row }) => row.original.teamPoints.toFixed(0) },
  { accessorKey: 'totalBid', header: 'Total Investment', cell: ({ row }) => `${row.original.totalBid.toFixed(2)} pts` },
  { accessorKey: 'rawROI', header: 'Raw ROI', cell: ({ row }) => row.original.rawROI.toFixed(3) },
  {
    accessorKey: 'normalizedROI',
    header: 'Normalized ROI',
    cell: ({ row }) => {
      const val = row.original.normalizedROI;
      return (
        <span className={`font-semibold ${val > 1.0 ? 'text-green-600' : val < 1.0 ? 'text-red-600' : 'text-gray-500'}`}>
          {val.toFixed(3)}
        </span>
      );
    },
  },
];

const bestInvestmentsColumns: ColumnDef<InvestmentLeaderboardRow, unknown>[] = [
  { id: 'rank', header: 'Rank', cell: ({ row }) => row.index + 1, enableSorting: false },
  { accessorKey: 'entryName', header: 'Entry' },
  { accessorKey: 'tournamentYear', header: 'Year' },
  { accessorKey: 'seed', header: 'Seed' },
  { accessorKey: 'schoolName', header: 'Team' },
  { accessorKey: 'investment', header: 'Investment', cell: ({ row }) => `${row.original.investment.toFixed(2)} pts` },
  { accessorKey: 'ownershipPercentage', header: 'Ownership', cell: ({ row }) => `${(row.original.ownershipPercentage * 100).toFixed(2)}%` },
  { accessorKey: 'rawReturns', header: 'Raw Returns', cell: ({ row }) => row.original.rawReturns.toFixed(2) },
  {
    accessorKey: 'normalizedReturns',
    header: 'Normalized Returns',
    cell: ({ row }) => {
      const val = row.original.normalizedReturns;
      return (
        <span className={`font-semibold ${val > 1.0 ? 'text-green-600' : val < 1.0 ? 'text-red-600' : 'text-gray-500'}`}>
          {val.toFixed(3)}
        </span>
      );
    },
  },
];

const bestEntriesColumns: ColumnDef<EntryLeaderboardRow, unknown>[] = [
  { id: 'rank', header: 'Rank', cell: ({ row }) => row.index + 1, enableSorting: false },
  { accessorKey: 'entryName', header: 'Entry' },
  { accessorKey: 'tournamentYear', header: 'Year' },
  { accessorKey: 'totalReturns', header: 'Total Returns', cell: ({ row }) => row.original.totalReturns.toFixed(2) },
  { accessorKey: 'totalParticipants', header: 'Total Participants' },
  { accessorKey: 'averageReturns', header: 'Average Returns', cell: ({ row }) => row.original.averageReturns.toFixed(2) },
  {
    accessorKey: 'normalizedReturns',
    header: 'Normalized Returns',
    cell: ({ row }) => {
      const val = row.original.normalizedReturns;
      return (
        <span className={`font-semibold ${val > 1.0 ? 'text-green-600' : val < 1.0 ? 'text-red-600' : 'text-gray-500'}`}>
          {val.toFixed(3)}
        </span>
      );
    },
  },
];

const bestCareersColumns: ColumnDef<CareerLeaderboardRow, unknown>[] = [
  { id: 'rank', header: 'Rank', cell: ({ row }) => row.index + 1, enableSorting: false },
  {
    accessorKey: 'entryName',
    header: 'Name',
    cell: ({ row }) => (
      <span className={row.original.activeInLatestCalcutta ? 'font-bold' : 'font-medium'}>
        {row.original.entryName}
      </span>
    ),
  },
  { accessorKey: 'years', header: 'Years' },
  { accessorKey: 'bestFinish', header: 'Best Finish' },
  { accessorKey: 'wins', header: 'Wins' },
  { accessorKey: 'podiums', header: 'Podiums' },
  { accessorKey: 'inTheMoneys', header: 'Payouts' },
  { accessorKey: 'top10s', header: 'Top 10s' },
  {
    accessorKey: 'careerEarningsCents',
    header: 'Career Earnings',
    cell: ({ row }) => formatDollarsFromCents(row.original.careerEarningsCents),
  },
];

export const HallOfFamePage: React.FC = () => {
  const [hideInactiveCareers, setHideInactiveCareers] = useState<boolean>(false);

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

  const filteredCareers = useMemo(() => {
    if (!bestCareersQuery.data) return [];
    return hideInactiveCareers
      ? bestCareersQuery.data.careers.filter((c) => c.activeInLatestCalcutta)
      : bestCareersQuery.data.careers;
  }, [bestCareersQuery.data, hideInactiveCareers]);

  return (
    <PageContainer>
      <Breadcrumb
        items={[
          { label: 'Admin', href: '/admin' },
          { label: 'Hall of Fame' },
        ]}
      />
      <PageHeader
        title="Hall of Fame"
        subtitle="Leaderboards across all calcuttas (normalized for year-to-year comparisons)."
        actions={
          <Link to="/admin" className="text-blue-600 hover:text-blue-800">
            Back to Admin Console
          </Link>
        }
      />

      <Tabs defaultValue="bestTeams">
        <TabsList>
          <TabsTrigger value="bestTeams">Best Teams</TabsTrigger>
          <TabsTrigger value="bestInvestments">Best Investments</TabsTrigger>
          <TabsTrigger value="bestEntries">Best Entries</TabsTrigger>
          <TabsTrigger value="bestCareers">Best Careers</TabsTrigger>
        </TabsList>

      <TabsContent value="bestTeams">
        <Card>
          <h2 className="text-xl font-semibold mb-2">Best Teams</h2>
          <p className="text-sm text-gray-600 mb-4">
            Teams ranked by normalized ROI where 1.0 = average performance within that Calcutta (levels the playing field across seeds). Yes, we call it
            "Adjusted for Inflation".
          </p>

          {bestTeamsQuery.isLoading && <LoadingState label="Loading best teams..." layout="inline" />}

          {bestTeamsQuery.isError && (
            <Alert variant="error" className="mt-3">
              <div className="font-semibold mb-1">Failed to load best teams</div>
              <div className="mb-3">{bestTeamsQuery.error instanceof Error ? bestTeamsQuery.error.message : 'An error occurred'}</div>
              <Button size="sm" onClick={() => bestTeamsQuery.refetch()}>
                Retry
              </Button>
            </Alert>
          )}

          {bestTeamsQuery.data && (
            <DataTable
              columns={bestTeamsColumns}
              data={bestTeamsQuery.data.teams}
              pagination
              pageSize={25}
            />
          )}
        </Card>
      </TabsContent>

      <TabsContent value="bestInvestments">
        <Card>
          <h2 className="text-xl font-semibold mb-2">Best Investments</h2>
          <p className="text-sm text-gray-600 mb-4">
            Individual picks ranked by normalized returns: Raw Returns / (Total Returns / Total Investment), where Total Investment = $100 × participants. Yes, we call it "Adjusted for Inflation".
          </p>

          {bestInvestmentsQuery.isLoading && <LoadingState label="Loading best investments..." layout="inline" />}

          {bestInvestmentsQuery.isError && (
            <Alert variant="error" className="mt-3">
              <div className="font-semibold mb-1">Failed to load best investments</div>
              <div className="mb-3">{bestInvestmentsQuery.error instanceof Error ? bestInvestmentsQuery.error.message : 'An error occurred'}</div>
              <Button size="sm" onClick={() => bestInvestmentsQuery.refetch()}>
                Retry
              </Button>
            </Alert>
          )}

          {bestInvestmentsQuery.data && (
            <DataTable
              columns={bestInvestmentsColumns}
              data={bestInvestmentsQuery.data.investments}
              pagination
              pageSize={25}
            />
          )}
        </Card>
      </TabsContent>

      <TabsContent value="bestEntries">
        <Card>
          <h2 className="text-xl font-semibold mb-2">Best Entries</h2>
          <p className="text-sm text-gray-600 mb-4">
            Entries ranked by normalized returns: Entry Total Returns / (Calcutta Total Returns / Calcutta Total Investment), where Calcutta Total Investment = $100 × participants.
          </p>

          {bestEntriesQuery.isLoading && <LoadingState label="Loading best entries..." layout="inline" />}

          {bestEntriesQuery.isError && (
            <Alert variant="error" className="mt-3">
              <div className="font-semibold mb-1">Failed to load best entries</div>
              <div className="mb-3">{bestEntriesQuery.error instanceof Error ? bestEntriesQuery.error.message : 'An error occurred'}</div>
              <Button size="sm" onClick={() => bestEntriesQuery.refetch()}>
                Retry
              </Button>
            </Alert>
          )}

          {bestEntriesQuery.data && (
            <DataTable
              columns={bestEntriesColumns}
              data={bestEntriesQuery.data.entries}
              pagination
              pageSize={25}
            />
          )}
        </Card>
      </TabsContent>

      <TabsContent value="bestCareers">
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
            <Alert variant="error" className="mt-3">
              <div className="font-semibold mb-1">Failed to load best careers</div>
              <div className="mb-3">{bestCareersQuery.error instanceof Error ? bestCareersQuery.error.message : 'An error occurred'}</div>
              <Button size="sm" onClick={() => bestCareersQuery.refetch()}>
                Retry
              </Button>
            </Alert>
          )}

          {bestCareersQuery.data && (
            <DataTable
              columns={bestCareersColumns}
              data={filteredCareers}
              pagination
              pageSize={25}
            />
          )}
        </Card>
      </TabsContent>
      </Tabs>
    </PageContainer>
  );
};

export default HallOfFamePage;
