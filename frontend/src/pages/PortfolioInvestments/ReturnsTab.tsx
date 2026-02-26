import { Investment, OwnershipDetail } from '../../schemas/pool';
import type { TournamentTeam } from '../../schemas/tournament';
import { School } from '../../schemas/school';
import { Select } from '../../components/ui/Select';
import { SegmentedBar } from '../../components/SegmentedBar';

export function ReturnsTab({
  portfolioId,
  returnsShowAllTeams,
  setReturnsShowAllTeams,
  sortBy,
  setSortBy,
  tournamentTeams,
  allOwnershipDetails,
  teams,
  schools,
  getOwnershipDetailData,
}: {
  portfolioId: string;
  returnsShowAllTeams: boolean;
  setReturnsShowAllTeams: (value: boolean) => void;
  sortBy: 'points' | 'ownership' | 'credits';
  setSortBy: (value: 'points' | 'ownership' | 'credits') => void;
  tournamentTeams: TournamentTeam[];
  allOwnershipDetails: OwnershipDetail[];
  teams: Investment[];
  schools: School[];
  getOwnershipDetailData: (teamId: string) => OwnershipDetail | undefined;
}) {
  return (
    <>
      <div className="mb-4 flex items-center justify-between gap-4">
        <label className="flex items-center gap-2 text-sm text-foreground">
          <input
            type="checkbox"
            checked={returnsShowAllTeams}
            onChange={(e) => setReturnsShowAllTeams(e.target.checked)}
            className="rounded border-border"
          />
          Show All Teams
        </label>
        <label className="text-sm text-muted-foreground">
          Sort by
          <Select
            className="ml-2 w-auto"
            value={sortBy}
            onChange={(e) => setSortBy(e.target.value as 'points' | 'ownership' | 'credits')}
          >
            <option value="points">Returns</option>
            <option value="ownership">Ownership</option>
            <option value="credits">Investment</option>
          </Select>
        </label>
      </div>

      <div className="bg-card rounded-lg shadow p-6">
        <div className="overflow-x-auto">
          <table className="min-w-full table-fixed border-separate border-spacing-y-2">
            <thead>
              <tr className="text-left text-xs uppercase tracking-wide text-muted-foreground">
                <th className="px-2 py-2 w-14">Seed</th>
                <th className="px-2 py-2 w-20">Region</th>
                <th className="px-2 py-2 w-44">Team</th>
                <th className="px-2 py-2"></th>
                <th className="px-2 py-2 w-28 text-right">Returns</th>
                <th className="px-2 py-2 w-32 text-right">Total Returns</th>
              </tr>
            </thead>
            <tbody>
              {(() => {
                const globalMax = tournamentTeams.reduce((max, tt) => {
                  const teamOwnershipDetails = allOwnershipDetails.filter((pt) => pt.teamId === tt.id);
                  const totalActualReturns = teamOwnershipDetails.reduce((sum, pt) => sum + pt.actualReturns, 0);
                  const totalExpectedReturns = teamOwnershipDetails.reduce((sum, pt) => sum + pt.expectedReturns, 0);
                  const eliminated = tt.isEliminated === true;
                  const totalPossibleReturns = eliminated
                    ? totalActualReturns
                    : Math.max(totalExpectedReturns, totalActualReturns);
                  return Math.max(max, totalActualReturns, totalPossibleReturns);
                }, 0);

                let teamsToShow = returnsShowAllTeams
                  ? tournamentTeams.map((tt) => {
                      const existingTeam = teams.find((t) => t.teamId === tt.id);
                      if (existingTeam) return existingTeam;
                      const schoolMap = new Map(schools.map((s) => [s.id, s]));
                      return {
                        id: `synthetic-${tt.id}`,
                        portfolioId: portfolioId,
                        teamId: tt.id,
                        credits: 0,
                        createdAt: new Date().toISOString(),
                        updatedAt: new Date().toISOString(),
                        team: { ...tt, school: schoolMap.get(tt.schoolId) },
                      } as Investment;
                    })
                  : teams;

                teamsToShow = teamsToShow.slice().sort((a, b) => {
                  const detailA = getOwnershipDetailData(a.teamId);
                  const detailB = getOwnershipDetailData(b.teamId);

                  const pointsA = detailA?.actualReturns || 0;
                  const pointsB = detailB?.actualReturns || 0;
                  const ownershipA = detailA?.ownershipPercentage || 0;
                  const ownershipB = detailB?.ownershipPercentage || 0;
                  const creditsA = a.credits;
                  const creditsB = b.credits;

                  if (sortBy === 'points') {
                    if (pointsB !== pointsA) return pointsB - pointsA;
                    if (ownershipB !== ownershipA) return ownershipB - ownershipA;
                    return creditsB - creditsA;
                  }

                  if (sortBy === 'ownership') {
                    if (ownershipB !== ownershipA) return ownershipB - ownershipA;
                    if (pointsB !== pointsA) return pointsB - pointsA;
                    return creditsB - creditsA;
                  }

                  if (creditsB !== creditsA) return creditsB - creditsA;
                  if (pointsB !== pointsA) return pointsB - pointsA;
                  return ownershipB - ownershipA;
                });

                return teamsToShow.map((team) => {
                  const ownershipDetail = getOwnershipDetailData(team.teamId);
                  const tournamentTeam = tournamentTeams.find((tt) => tt.id === team.teamId);
                  const teamOwnershipDetails = allOwnershipDetails.filter((pt) => pt.teamId === team.teamId);

                  const totalActualReturns = teamOwnershipDetails.reduce((sum, pt) => sum + pt.actualReturns, 0);

                  const userOwnership = ownershipDetail?.ownershipPercentage ?? 0;
                  const userActualReturns = totalActualReturns * userOwnership;
                  const othersActualReturns = totalActualReturns * (1 - userOwnership);

                  return (
                    <tr key={team.id} className="bg-accent">
                      <td className="px-2 py-3 font-medium text-foreground rounded-l-md whitespace-nowrap">
                        {team.team?.seed ?? '\u2014'}
                      </td>
                      <td className="px-2 py-3 text-foreground whitespace-nowrap">{tournamentTeam?.region || '\u2014'}</td>
                      <td className="px-2 py-3 text-foreground font-medium whitespace-nowrap truncate">
                        {team.team?.school?.name || 'Unknown School'}
                      </td>
                      <td className="px-2 py-3">
                        <SegmentedBar
                          barWidthPct={((userActualReturns + othersActualReturns) / (globalMax || 1)) * 100}
                          segments={[
                            ...(userActualReturns > 0
                              ? [
                                  {
                                    key: `${team.teamId}-portfolio`,
                                    label: 'Your Portfolio',
                                    value: userActualReturns,
                                    color: '#4F46E5',
                                  },
                                ]
                              : []),
                            ...(othersActualReturns > 0
                              ? [
                                  {
                                    key: `${team.teamId}-others`,
                                    label: 'Others',
                                    value: othersActualReturns,
                                    color: '#9CA3AF',
                                  },
                                ]
                              : []),
                          ]}
                          backgroundColor="#F3F4F6"
                          disabled={team.team?.isEliminated === true}
                          getTooltipTitle={(seg) => seg.label}
                          getTooltipValue={(seg) => `${seg.value.toFixed(2)} points`}
                        />
                      </td>
                      <td className="px-2 py-3 text-right font-medium text-foreground whitespace-nowrap">
                        {userActualReturns.toFixed(2)}
                      </td>
                      <td className="px-2 py-3 text-right font-medium text-foreground rounded-r-md whitespace-nowrap">
                        {totalActualReturns.toFixed(2)}
                      </td>
                    </tr>
                  );
                });
              })()}
            </tbody>
          </table>
        </div>
      </div>
    </>
  );
}
