import { useMemo } from 'react';
import { OwnershipPieChart } from '../../components/charts/OwnershipPieChart';
import { LoadingState } from '../../components/ui/LoadingState';
import { Select } from '../../components/ui/Select';
import { CalcuttaEntryTeam, CalcuttaPortfolio, CalcuttaPortfolioTeam } from '../../schemas/calcutta';

export function OwnershipsTab({
  ownershipShowAllTeams,
  setOwnershipShowAllTeams,
  sortBy,
  setSortBy,
  ownershipLoading,
  ownershipTeamsData,
  getPortfolioTeamData,
  getInvestorRanking,
  allCalcuttaPortfolioTeams,
  allCalcuttaPortfolios,
  portfolios,
}: {
  ownershipShowAllTeams: boolean;
  setOwnershipShowAllTeams: (value: boolean) => void;
  sortBy: 'points' | 'ownership' | 'bidPoints';
  setSortBy: (value: 'points' | 'ownership' | 'bidPoints') => void;
  ownershipLoading: boolean;
  ownershipTeamsData: CalcuttaEntryTeam[];
  getPortfolioTeamData: (teamId: string) => CalcuttaPortfolioTeam | undefined;
  getInvestorRanking: (teamId: string) => { rank: number; total: number };
  allCalcuttaPortfolioTeams: CalcuttaPortfolioTeam[];
  allCalcuttaPortfolios: (CalcuttaPortfolio & { entryName?: string })[];
  portfolios: CalcuttaPortfolio[];
}) {
  const currentPortfolioId = portfolios[0]?.id;

  const portfolioNameById = useMemo(() => {
    const map = new Map<string, string>();
    allCalcuttaPortfolios.forEach((p) => {
      map.set(p.id, p.entryName || `Portfolio ${p.id.slice(0, 4)}`);
    });
    return map;
  }, [allCalcuttaPortfolios]);

  return (
    <>
      <div className="mb-4 flex items-center justify-between gap-4">
        <label className="flex items-center gap-2 text-sm text-foreground">
          <input
            type="checkbox"
            checked={ownershipShowAllTeams}
            onChange={(e) => setOwnershipShowAllTeams(e.target.checked)}
            className="rounded border-border"
          />
          Show All Teams
        </label>
        <label className="text-sm text-muted-foreground">
          Sort by
          <Select
            className="ml-2 w-auto"
            value={sortBy}
            onChange={(e) => setSortBy(e.target.value as 'points' | 'ownership' | 'bidPoints')}
          >
            <option value="points">Points</option>
            <option value="ownership">Ownership</option>
            <option value="bidPoints">Bid Amount</option>
          </Select>
        </label>
      </div>

      {ownershipLoading ? (
        <div className="bg-card rounded-lg shadow p-6">
          <LoadingState label="Loading ownerships…" />
        </div>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
          {ownershipTeamsData.map((team) => {
            const portfolioTeam = getPortfolioTeamData(team.teamId);
            const investorRanking = getInvestorRanking(team.teamId);
            const teamPortfolioTeams = allCalcuttaPortfolioTeams.filter((pt) => pt.teamId === team.teamId);
            const ownershipPct = portfolioTeam ? portfolioTeam.ownershipPercentage * 100 : undefined;

            const CURRENT_FILL = '#1F2937';
            const OTHER_FILL = '#CBD5E1';

            const pieData = teamPortfolioTeams
              .filter((pt) => pt.ownershipPercentage > 0)
              .sort((a, b) => b.ownershipPercentage - a.ownershipPercentage)
              .map((pt) => {
                const name = portfolioNameById.get(pt.portfolioId) || `Portfolio ${pt.portfolioId.slice(0, 4)}`;
                const isCurrent = pt.portfolioId === currentPortfolioId;
                return {
                  key: `${team.teamId}-${pt.portfolioId}`,
                  name,
                  value: pt.ownershipPercentage * 100,
                  fill: isCurrent ? CURRENT_FILL : OTHER_FILL,
                };
              });

            const topOwners: { name: string; pct: number | null }[] = [...teamPortfolioTeams]
              .filter((pt) => pt.ownershipPercentage > 0)
              .sort((a, b) => b.ownershipPercentage - a.ownershipPercentage)
              .slice(0, 3)
              .map((pt) => {
                const portfolio = allCalcuttaPortfolios.find((p) => p.id === pt.portfolioId);
                const name = portfolio?.entryName || `Portfolio ${pt.portfolioId.slice(0, 4)}`;
                return {
                  name,
                  pct: pt.ownershipPercentage * 100,
                };
              });

            while (topOwners.length < 3) {
              topOwners.push({ name: '--', pct: null });
            }

            return (
              <div key={team.id} className="bg-card rounded-lg shadow p-4 flex flex-col">
                <div className="flex items-start justify-between gap-3">
                  <div className="min-w-0">
                    <h2 className="text-lg font-semibold leading-snug truncate">
                      {team.team?.school?.name || 'Unknown School'}
                    </h2>
                    <div className="text-xs text-muted-foreground">
                      {team.team?.region ?? '?'} - {team.team?.seed ?? '?'} seed
                    </div>
                    <div className="mt-1 text-sm text-muted-foreground">
                      Investor Rank: {investorRanking.rank} / {investorRanking.total}
                    </div>
                  </div>
                  <div className="text-right">
                    {ownershipPct !== undefined && (
                      <div className="text-sm text-muted-foreground">
                        Ownership
                        <div className="text-base font-semibold text-foreground">{ownershipPct.toFixed(2)}%</div>
                      </div>
                    )}
                  </div>
                </div>

                <div className="mt-4 flex justify-center">
                  <OwnershipPieChart data={pieData} sizePx={220} emptyLabel="No ownership" />
                </div>

                <div className="mt-4">
                  <div className="text-sm font-medium text-foreground">Top Shareholders</div>
                  <div className="mt-2 space-y-2">
                    {topOwners.map((owner, idx) => (
                      <div key={idx} className="flex items-center justify-between gap-3 text-sm">
                        <div className="min-w-0 truncate text-foreground flex items-center gap-2">
                          <div className="w-4 shrink-0 text-muted-foreground">{idx + 1}</div>
                          <div className="truncate">{owner.name}</div>
                        </div>
                        <div className="font-medium text-foreground">
                          {owner.pct === null ? '--' : `${owner.pct.toFixed(2)}%`}
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      )}
    </>
  );
}
