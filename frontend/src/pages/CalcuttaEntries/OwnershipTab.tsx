import { useMemo, useState } from 'react';
import {
  CalcuttaEntry,
  CalcuttaEntryTeam,
  CalcuttaPortfolio,
  CalcuttaPortfolioTeam,
  TournamentTeam,
} from '../../types/calcutta';
import { School } from '../../types/school';
import { getEntryColorById } from '../../utils/entryColors';
import { OwnershipPieChart } from '../../components/charts/OwnershipPieChart';
import { LoadingState } from '../../components/ui/LoadingState';

interface OwnershipSlice {
  name: string;
  value: number;
  entryId: string;
}

interface OwnershipTeamCard {
  teamId: string;
  seed: number;
  region: string;
  teamName: string;
  totalSpend: number;
  slices: OwnershipSlice[];
  topOwners: OwnershipSlice[];
}

export const OwnershipTab: React.FC<{
  entries: CalcuttaEntry[];
  schools: School[];
  tournamentTeams: TournamentTeam[];
  allEntryTeams: CalcuttaEntryTeam[];
  allCalcuttaPortfolios: (CalcuttaPortfolio & { entryName?: string })[];
  allCalcuttaPortfolioTeams: CalcuttaPortfolioTeam[];
  isFetching: boolean;
}> = ({
  entries,
  schools,
  tournamentTeams,
  allEntryTeams,
  allCalcuttaPortfolios,
  allCalcuttaPortfolioTeams,
  isFetching,
}) => {
  const [ownershipSortBy, setOwnershipSortBy] = useState<'seed' | 'region' | 'team' | 'investment'>('seed');

  const entryNameById = useMemo(() => new Map(entries.map((entry) => [entry.id, entry.name])), [entries]);
  const schoolNameById = useMemo(() => new Map(schools.map((school) => [school.id, school.name])), [schools]);

  const entryColorById = useMemo(() => getEntryColorById(entries), [entries]);

  const ownershipTeamsData = useMemo(() => {
    const byTeamSpend = new Map<string, number>();
    for (const entryTeam of allEntryTeams) {
      const amount = entryTeam.bid || 0;
      if (amount <= 0) continue;
      byTeamSpend.set(entryTeam.teamId, (byTeamSpend.get(entryTeam.teamId) || 0) + amount);
    }

    const cards: OwnershipTeamCard[] = tournamentTeams.map((team) => {
      const teamId = team.id;
      const spend = byTeamSpend.get(teamId) || 0;
      const teamPortfolioTeams = allCalcuttaPortfolioTeams.filter((pt) => pt.teamId === teamId && pt.ownershipPercentage > 0);

      const slices: OwnershipSlice[] = teamPortfolioTeams
        .slice()
        .sort((a, b) => b.ownershipPercentage - a.ownershipPercentage)
        .map((pt) => {
          const portfolio = allCalcuttaPortfolios.find((p) => p.id === pt.portfolioId);
          const entryId = portfolio?.entryId || '';
          const entryName = portfolio?.entryName || (entryId ? entryNameById.get(entryId) : undefined) || 'Unknown Entry';
          return {
            name: entryName,
            value: pt.ownershipPercentage * 100,
            entryId,
          };
        });

      return {
        teamId,
        seed: team.seed,
        region: team.region,
        teamName: schoolNameById.get(team.schoolId) || 'Unknown School',
        totalSpend: spend,
        slices,
        topOwners: slices.slice(0, 3),
      };
    });

    return cards;
  }, [allCalcuttaPortfolios, allCalcuttaPortfolioTeams, allEntryTeams, entryNameById, schoolNameById, tournamentTeams]);

  const ownershipSortedTeams = useMemo(() => {
    const cards = ownershipTeamsData.slice();

    const seedA = (seed: number) => seed;
    const regionA = (region: string) => region || '';
    const teamA = (name: string) => name || '';

    cards.sort((a, b) => {
      if (ownershipSortBy === 'investment') {
        if (b.totalSpend !== a.totalSpend) return b.totalSpend - a.totalSpend;
        const seedDiff = seedA(a.seed) - seedA(b.seed);
        if (seedDiff !== 0) return seedDiff;
        return teamA(a.teamName).localeCompare(teamA(b.teamName));
      }

      if (ownershipSortBy === 'seed') {
        const seedDiff = seedA(a.seed) - seedA(b.seed);
        if (seedDiff !== 0) return seedDiff;
        const regionDiff = regionA(a.region).localeCompare(regionA(b.region));
        if (regionDiff !== 0) return regionDiff;
        return teamA(a.teamName).localeCompare(teamA(b.teamName));
      }

      if (ownershipSortBy === 'region') {
        const regionDiff = regionA(a.region).localeCompare(regionA(b.region));
        if (regionDiff !== 0) return regionDiff;
        const seedDiff = seedA(a.seed) - seedA(b.seed);
        if (seedDiff !== 0) return seedDiff;
        return teamA(a.teamName).localeCompare(teamA(b.teamName));
      }

      const nameDiff = teamA(a.teamName).localeCompare(teamA(b.teamName));
      if (nameDiff !== 0) return nameDiff;
      return seedA(a.seed) - seedA(b.seed);
    });

    return cards;
  }, [ownershipSortBy, ownershipTeamsData]);

  const ownershipLoading = isFetching;

  return (
    <>
      <div className="mb-4 flex items-center justify-end">
        <label className="text-sm text-gray-600">
          Sort by
          <select
            className="ml-2 rounded-md border border-gray-300 bg-white px-3 py-2 text-sm text-gray-900 shadow-sm"
            value={ownershipSortBy}
            onChange={(e) => setOwnershipSortBy(e.target.value as 'seed' | 'region' | 'team' | 'investment')}
          >
            <option value="seed">Seed</option>
            <option value="region">Region</option>
            <option value="team">Team</option>
            <option value="investment">Investment</option>
          </select>
        </label>
      </div>

      {ownershipLoading ? (
        <div className="bg-white rounded-lg shadow p-6">
          <LoadingState label="Loading ownership…" />
        </div>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
          {ownershipSortedTeams.map((team) => (
            <div key={team.teamId} className="bg-white rounded-lg shadow p-4 flex flex-col">
              <div className="flex items-start justify-between gap-3">
                <div className="min-w-0">
                  <h2 className="text-lg font-semibold leading-snug truncate">{team.teamName}</h2>
                  <div className="mt-1 text-sm text-gray-600">
                    {team.seed} Seed - {team.region}
                  </div>
                </div>
                <div className="text-right">
                  <div className="text-sm text-gray-600 whitespace-nowrap">
                    Total Investment
                    <div className="text-base font-semibold text-gray-900">{team.totalSpend.toFixed(2)} pts</div>
                  </div>
                </div>
              </div>

              <div className="mt-4 flex justify-center">
                <OwnershipPieChart
                  data={team.slices.map((slice) => ({
                    key: `${team.teamId}-${slice.entryId}-${slice.name}`,
                    name: slice.name,
                    value: slice.value,
                    fill: entryColorById.get(slice.entryId) || '#94A3B8',
                  }))}
                  sizePx={220}
                  emptyLabel="No ownership"
                />
              </div>

              <div className="mt-4">
                <div className="text-sm font-medium text-gray-900">Top Shareholders</div>
                <div className="mt-2 space-y-2">
                  {Array.from({ length: 3 }).map((_, idx) => {
                    const owner = team.topOwners[idx];
                    const name = owner?.name ?? '--';
                    const pct = owner ? `${owner.value.toFixed(2)}%` : '--';
                    return (
                      <div key={idx} className="flex items-center justify-between gap-3 text-sm">
                        <div className="min-w-0 truncate text-gray-700 flex items-center gap-2">
                          <div className="w-4 shrink-0 text-gray-500">{idx + 1}</div>
                          <div className="truncate">{name}</div>
                        </div>
                        <div className="font-medium text-gray-900">{pct}</div>
                      </div>
                    );
                  })}
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </>
  );
};
