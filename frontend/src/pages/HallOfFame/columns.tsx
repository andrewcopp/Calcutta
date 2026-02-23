import { type ColumnDef } from '@tanstack/react-table';
import { BestTeam, CareerLeaderboardRow, EntryLeaderboardRow, InvestmentLeaderboardRow } from '../../schemas/hallOfFame';
import { formatDollarsFromCents } from '../../utils/format';
import { NormalizedValue } from './NormalizedValue';

export const bestTeamsColumns: ColumnDef<BestTeam, unknown>[] = [
  { id: 'rank', header: 'Rank', cell: ({ row }) => row.index + 1, enableSorting: false },
  { accessorKey: 'tournamentYear', header: 'Year' },
  { accessorKey: 'seed', header: 'Seed' },
  { accessorKey: 'schoolName', header: 'Team' },
  {
    accessorKey: 'teamPoints',
    header: 'Points',
    cell: ({ row }) => row.original.teamPoints.toFixed(0),
  },
  {
    accessorKey: 'totalBid',
    header: 'Total Investment',
    cell: ({ row }) => `${row.original.totalBid.toFixed(2)} credits`,
  },
  {
    accessorKey: 'rawROI',
    header: 'Raw ROI',
    cell: ({ row }) => row.original.rawROI.toFixed(3),
  },
  {
    accessorKey: 'normalizedROI',
    header: 'Normalized ROI',
    cell: ({ row }) => <NormalizedValue value={row.original.normalizedROI} />,
  },
];

export const bestInvestmentsColumns: ColumnDef<InvestmentLeaderboardRow, unknown>[] = [
  { id: 'rank', header: 'Rank', cell: ({ row }) => row.index + 1, enableSorting: false },
  { accessorKey: 'entryName', header: 'Entry' },
  { accessorKey: 'tournamentYear', header: 'Year' },
  { accessorKey: 'seed', header: 'Seed' },
  { accessorKey: 'schoolName', header: 'Team' },
  {
    accessorKey: 'investment',
    header: 'Investment',
    cell: ({ row }) => `${row.original.investment.toFixed(2)} credits`,
  },
  {
    accessorKey: 'ownershipPercentage',
    header: 'Ownership',
    cell: ({ row }) => `${(row.original.ownershipPercentage * 100).toFixed(2)}%`,
  },
  {
    accessorKey: 'rawReturns',
    header: 'Raw Returns',
    cell: ({ row }) => row.original.rawReturns.toFixed(2),
  },
  {
    accessorKey: 'normalizedReturns',
    header: 'Normalized Returns',
    cell: ({ row }) => <NormalizedValue value={row.original.normalizedReturns} />,
  },
];

export const bestEntriesColumns: ColumnDef<EntryLeaderboardRow, unknown>[] = [
  { id: 'rank', header: 'Rank', cell: ({ row }) => row.index + 1, enableSorting: false },
  { accessorKey: 'entryName', header: 'Entry' },
  { accessorKey: 'tournamentYear', header: 'Year' },
  {
    accessorKey: 'totalReturns',
    header: 'Total Returns',
    cell: ({ row }) => row.original.totalReturns.toFixed(2),
  },
  { accessorKey: 'totalParticipants', header: 'Total Participants' },
  {
    accessorKey: 'averageReturns',
    header: 'Average Returns',
    cell: ({ row }) => row.original.averageReturns.toFixed(2),
  },
  {
    accessorKey: 'normalizedReturns',
    header: 'Normalized Returns',
    cell: ({ row }) => <NormalizedValue value={row.original.normalizedReturns} />,
  },
];

export const bestCareersColumns: ColumnDef<CareerLeaderboardRow, unknown>[] = [
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
