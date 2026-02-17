import React, { useMemo, useState } from 'react';
import { Link } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { type ColumnDef } from '@tanstack/react-table';
import { queryKeys } from '../queryKeys';
import { hallOfFameService } from '../services/hallOfFameService';
import { Breadcrumb } from '../components/ui/Breadcrumb';
import { Card } from '../components/ui/Card';
import { DataTable } from '../components/ui/DataTable';
import { ErrorState } from '../components/ui/ErrorState';
import { LoadingState } from '../components/ui/LoadingState';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { Tabs, TabsList, TabsTrigger, TabsContent } from '../components/ui/Tabs';
import {
  BestTeam,
  CareerLeaderboardRow,
  EntryLeaderboardRow,
  InvestmentLeaderboardRow,
} from '../types/hallOfFame';
import { formatDollarsFromCents } from '../utils/format';

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

  const bestTeamsQuery = useQuery({
    queryKey: queryKeys.hallOfFame.bestTeams(200),
    queryFn: () => hallOfFameService.getBestTeams(200),
  });

  const bestInvestmentsQuery = useQuery({
    queryKey: queryKeys.hallOfFame.bestInvestments(200),
    queryFn: () => hallOfFameService.getBestInvestments(200),
  });

  const bestEntriesQuery = useQuery({
    queryKey: queryKeys.hallOfFame.bestEntries(200),
    queryFn: () => hallOfFameService.getBestEntries(200),
  });

  const bestCareersQuery = useQuery({
    queryKey: queryKeys.hallOfFame.bestCareers(200),
    queryFn: () => hallOfFameService.getBestCareers(200),
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
            <ErrorState
              error={bestTeamsQuery.error instanceof Error ? bestTeamsQuery.error.message : 'Failed to load best teams'}
              onRetry={() => bestTeamsQuery.refetch()}
            />
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
            <ErrorState
              error={bestInvestmentsQuery.error instanceof Error ? bestInvestmentsQuery.error.message : 'Failed to load best investments'}
              onRetry={() => bestInvestmentsQuery.refetch()}
            />
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
            <ErrorState
              error={bestEntriesQuery.error instanceof Error ? bestEntriesQuery.error.message : 'Failed to load best entries'}
              onRetry={() => bestEntriesQuery.refetch()}
            />
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
            <ErrorState
              error={bestCareersQuery.error instanceof Error ? bestCareersQuery.error.message : 'Failed to load best careers'}
              onRetry={() => bestCareersQuery.refetch()}
            />
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
