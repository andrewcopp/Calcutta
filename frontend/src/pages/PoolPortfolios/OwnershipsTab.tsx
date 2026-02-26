import { useMemo, useState } from 'react';
import { Portfolio, Investment, OwnershipSummary, OwnershipDetail } from '../../schemas/pool';
import type { TournamentTeam } from '../../schemas/tournament';
import { School } from '../../schemas/school';
import { Select } from '../../components/ui/Select';
import { getEntryColorById } from '../../utils/entryColors';
import { OwnershipPieChart } from '../../components/charts/OwnershipPieChart';
import { LoadingState } from '../../components/ui/LoadingState';
import { createTeamSortComparator } from '../../utils/teamSort';

interface OwnershipSlice {
  name: string;
  value: number;
  portfolioId: string;
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

export function OwnershipsTab({
  portfolios,
  schools,
  tournamentTeams,
  allInvestments,
  allOwnershipSummaries,
  allOwnershipDetails,
  isFetching,
}: {
  portfolios: Portfolio[];
  schools: School[];
  tournamentTeams: TournamentTeam[];
  allInvestments: Investment[];
  allOwnershipSummaries: (OwnershipSummary & { portfolioName?: string })[];
  allOwnershipDetails: OwnershipDetail[];
  isFetching: boolean;
}) {
  const [ownershipSortBy, setOwnershipSortBy] = useState<'seed' | 'region' | 'team' | 'investment'>('seed');

  const portfolioNameById = useMemo(() => new Map(portfolios.map((p) => [p.id, p.name])), [portfolios]);
  const schoolNameById = useMemo(() => new Map(schools.map((school) => [school.id, school.name])), [schools]);

  const portfolioColorById = useMemo(() => getEntryColorById(portfolios), [portfolios]);

  const ownershipTeamsData = useMemo(() => {
    const byTeamSpend = new Map<string, number>();
    for (const investment of allInvestments) {
      const amount = investment.credits;
      if (amount <= 0) continue;
      byTeamSpend.set(investment.teamId, (byTeamSpend.get(investment.teamId) || 0) + amount);
    }

    const cards: OwnershipTeamCard[] = tournamentTeams.map((team) => {
      const teamId = team.id;
      const spend = byTeamSpend.get(teamId) || 0;
      const teamOwnershipDetails = allOwnershipDetails.filter(
        (pt) => pt.teamId === teamId && pt.ownershipPercentage > 0,
      );

      const slices: OwnershipSlice[] = teamOwnershipDetails
        .slice()
        .sort((a, b) => b.ownershipPercentage - a.ownershipPercentage)
        .map((pt) => {
          const ownershipSummary = allOwnershipSummaries.find((p) => p.id === pt.ownershipSummaryId);
          const portfolioId = ownershipSummary?.portfolioId || '';
          const portfolioName =
            ownershipSummary?.portfolioName || (portfolioId ? portfolioNameById.get(portfolioId) : undefined) || 'Unknown Portfolio';
          return {
            name: portfolioName,
            value: pt.ownershipPercentage * 100,
            portfolioId,
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
  }, [allOwnershipSummaries, allOwnershipDetails, allInvestments, portfolioNameById, schoolNameById, tournamentTeams]);

  const ownershipSortedTeams = useMemo(() => {
    const cards = ownershipTeamsData.slice();

    if (ownershipSortBy === 'investment') {
      cards.sort((a, b) => {
        if (b.totalSpend !== a.totalSpend) return b.totalSpend - a.totalSpend;
        const seedDiff = a.seed - b.seed;
        if (seedDiff !== 0) return seedDiff;
        return (a.teamName || '').localeCompare(b.teamName || '');
      });
    } else {
      cards.sort(createTeamSortComparator(ownershipSortBy));
    }

    return cards;
  }, [ownershipSortBy, ownershipTeamsData]);

  const ownershipLoading = isFetching;

  return (
    <>
      <div className="mb-4 flex items-center justify-end">
        <label className="text-sm text-muted-foreground">
          Sort by
          <Select
            className="ml-2 w-auto"
            value={ownershipSortBy}
            onChange={(e) => setOwnershipSortBy(e.target.value as 'seed' | 'region' | 'team' | 'investment')}
          >
            <option value="seed">Seed</option>
            <option value="region">Region</option>
            <option value="team">Team</option>
            <option value="investment">Investment</option>
          </Select>
        </label>
      </div>

      {ownershipLoading ? (
        <div className="bg-card rounded-lg shadow p-6">
          <LoadingState label="Loading ownership..." />
        </div>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
          {ownershipSortedTeams.map((team) => (
            <div key={team.teamId} className="bg-card rounded-lg shadow p-4 flex flex-col">
              <div className="flex items-start justify-between gap-3">
                <div className="min-w-0">
                  <h2 className="text-lg font-semibold leading-snug truncate">{team.teamName}</h2>
                  <div className="mt-1 text-sm text-muted-foreground">
                    {team.seed} Seed - {team.region}
                  </div>
                </div>
                <div className="text-right">
                  <div className="text-sm text-muted-foreground whitespace-nowrap">
                    Total Investment
                    <div className="text-base font-semibold text-foreground">{team.totalSpend.toFixed(2)} credits</div>
                  </div>
                </div>
              </div>

              <div className="mt-4 flex justify-center">
                <OwnershipPieChart
                  data={team.slices.map((slice) => ({
                    key: `${team.teamId}-${slice.portfolioId}-${slice.name}`,
                    name: slice.name,
                    value: slice.value,
                    fill: portfolioColorById.get(slice.portfolioId) || '#94A3B8',
                  }))}
                  sizePx={220}
                  emptyLabel="No ownership"
                />
              </div>

              <div className="mt-4">
                <div className="text-sm font-medium text-foreground">Top Shareholders</div>
                <div className="mt-2 space-y-2">
                  {Array.from({ length: 3 }).map((_, idx) => {
                    const owner = team.topOwners[idx];
                    const name = owner?.name ?? '--';
                    const pct = owner ? `${owner.value.toFixed(2)}%` : '--';
                    return (
                      <div key={idx} className="flex items-center justify-between gap-3 text-sm">
                        <div className="min-w-0 truncate text-foreground flex items-center gap-2">
                          <div className="w-4 shrink-0 text-muted-foreground">{idx + 1}</div>
                          <div className="truncate">{name}</div>
                        </div>
                        <div className="font-medium text-foreground">{pct}</div>
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
}
