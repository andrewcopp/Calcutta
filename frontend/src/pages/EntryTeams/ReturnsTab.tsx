import { CalcuttaEntryTeam, CalcuttaPortfolioTeam } from '../../schemas/calcutta';
import type { TournamentTeam } from '../../schemas/tournament';
import { School } from '../../schemas/school';
import { Select } from '../../components/ui/Select';
import { SegmentedBar } from '../../components/SegmentedBar';

export function ReturnsTab({
  entryId,
  returnsShowAllTeams,
  setReturnsShowAllTeams,
  sortBy,
  setSortBy,
  tournamentTeams,
  allCalcuttaPortfolioTeams,
  teams,
  schools,
  getPortfolioTeamData,
}: {
  entryId: string;
  returnsShowAllTeams: boolean;
  setReturnsShowAllTeams: (value: boolean) => void;
  sortBy: 'points' | 'ownership' | 'bid';
  setSortBy: (value: 'points' | 'ownership' | 'bid') => void;
  tournamentTeams: TournamentTeam[];
  allCalcuttaPortfolioTeams: CalcuttaPortfolioTeam[];
  teams: CalcuttaEntryTeam[];
  schools: School[];
  getPortfolioTeamData: (teamId: string) => CalcuttaPortfolioTeam | undefined;
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
            onChange={(e) => setSortBy(e.target.value as 'points' | 'ownership' | 'bid')}
          >
            <option value="points">Points</option>
            <option value="ownership">Ownership</option>
            <option value="bid">Bid Amount</option>
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
                <th className="px-2 py-2 w-28 text-right">Points</th>
                <th className="px-2 py-2 w-32 text-right">Total Points</th>
              </tr>
            </thead>
            <tbody>
              {(() => {
                const globalMax = tournamentTeams.reduce((max, tt) => {
                  const teamPortfolioTeams = allCalcuttaPortfolioTeams.filter((pt) => pt.teamId === tt.id);
                  const totalActualPoints = teamPortfolioTeams.reduce((sum, pt) => sum + pt.actualPoints, 0);
                  const totalExpectedPoints = teamPortfolioTeams.reduce((sum, pt) => sum + pt.expectedPoints, 0);
                  const eliminated = tt.isEliminated === true;
                  const totalPossiblePoints = eliminated
                    ? totalActualPoints
                    : Math.max(totalExpectedPoints, totalActualPoints);
                  return Math.max(max, totalActualPoints, totalPossiblePoints);
                }, 0);

                let teamsToShow = returnsShowAllTeams
                  ? tournamentTeams.map((tt) => {
                      const existingTeam = teams.find((t) => t.teamId === tt.id);
                      if (existingTeam) return existingTeam;
                      const schoolMap = new Map(schools.map((s) => [s.id, s]));
                      return {
                        id: `synthetic-${tt.id}`,
                        entryId: entryId,
                        teamId: tt.id,
                        bid: 0,
                        createdAt: new Date().toISOString(),
                        updatedAt: new Date().toISOString(),
                        team: { ...tt, school: schoolMap.get(tt.schoolId) },
                      } as CalcuttaEntryTeam;
                    })
                  : teams;

                teamsToShow = teamsToShow.slice().sort((a, b) => {
                  const portfolioTeamA = getPortfolioTeamData(a.teamId);
                  const portfolioTeamB = getPortfolioTeamData(b.teamId);

                  const pointsA = portfolioTeamA?.actualPoints || 0;
                  const pointsB = portfolioTeamB?.actualPoints || 0;
                  const ownershipA = portfolioTeamA?.ownershipPercentage || 0;
                  const ownershipB = portfolioTeamB?.ownershipPercentage || 0;
                  const bidA = a.bid;
                  const bidB = b.bid;

                  if (sortBy === 'points') {
                    if (pointsB !== pointsA) return pointsB - pointsA;
                    if (ownershipB !== ownershipA) return ownershipB - ownershipA;
                    return bidB - bidA;
                  }

                  if (sortBy === 'ownership') {
                    if (ownershipB !== ownershipA) return ownershipB - ownershipA;
                    if (pointsB !== pointsA) return pointsB - pointsA;
                    return bidB - bidA;
                  }

                  if (bidB !== bidA) return bidB - bidA;
                  if (pointsB !== pointsA) return pointsB - pointsA;
                  return ownershipB - ownershipA;
                });

                return teamsToShow.map((team) => {
                  const portfolioTeam = getPortfolioTeamData(team.teamId);
                  const tournamentTeam = tournamentTeams.find((tt) => tt.id === team.teamId);
                  const teamPortfolioTeams = allCalcuttaPortfolioTeams.filter((pt) => pt.teamId === team.teamId);

                  const totalActualPoints = teamPortfolioTeams.reduce((sum, pt) => sum + pt.actualPoints, 0);

                  const userOwnership = portfolioTeam?.ownershipPercentage ?? 0;
                  const userActualPoints = totalActualPoints * userOwnership;
                  const othersActualPoints = totalActualPoints * (1 - userOwnership);

                  return (
                    <tr key={team.id} className="bg-accent">
                      <td className="px-2 py-3 font-medium text-foreground rounded-l-md whitespace-nowrap">
                        {team.team?.seed ?? '—'}
                      </td>
                      <td className="px-2 py-3 text-foreground whitespace-nowrap">{tournamentTeam?.region || '—'}</td>
                      <td className="px-2 py-3 text-foreground font-medium whitespace-nowrap truncate">
                        {team.team?.school?.name || 'Unknown School'}
                      </td>
                      <td className="px-2 py-3">
                        <SegmentedBar
                          barWidthPct={((userActualPoints + othersActualPoints) / (globalMax || 1)) * 100}
                          segments={[
                            ...(userActualPoints > 0
                              ? [
                                  {
                                    key: `${team.teamId}-entry`,
                                    label: 'Your Portfolio',
                                    value: userActualPoints,
                                    color: '#4F46E5',
                                  },
                                ]
                              : []),
                            ...(othersActualPoints > 0
                              ? [
                                  {
                                    key: `${team.teamId}-others`,
                                    label: 'Others',
                                    value: othersActualPoints,
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
                        {userActualPoints.toFixed(2)}
                      </td>
                      <td className="px-2 py-3 text-right font-medium text-foreground rounded-r-md whitespace-nowrap">
                        {totalActualPoints.toFixed(2)}
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
