import { useMemo } from 'react';
import type { PoolDashboard, Portfolio, FinalFourOutcome } from '../../schemas/pool';
import type { School } from '../../schemas/school';
import type { TournamentTeam } from '../../schemas/tournament';
import { formatDollarsFromCents } from '../../utils/format';
import { Card } from '../../components/ui/Card';

interface FinalFourTabProps {
  portfolios: Portfolio[];
  dashboard: PoolDashboard;
  schools: School[];
  tournamentTeams: TournamentTeam[];
}

export function FinalFourTab({ portfolios, dashboard, schools, tournamentTeams }: FinalFourTabProps) {
  const outcomes = dashboard.finalFourOutcomes;

  const schoolNameById = useMemo(() => new Map(schools.map((s) => [s.id, s.name])), [schools]);

  const teamSchoolId = useMemo(() => {
    const m = new Map<string, string>();
    for (const tt of tournamentTeams) {
      m.set(tt.id, tt.schoolId);
    }
    return m;
  }, [tournamentTeams]);

  const resolveTeamName = (teamId: string, schoolId: string) => {
    return schoolNameById.get(schoolId) || schoolNameById.get(teamSchoolId.get(teamId) || '') || 'Unknown';
  };

  const portfolioNameById = useMemo(() => new Map(portfolios.map((e) => [e.id, e.name])), [portfolios]);

  if (!outcomes || outcomes.length === 0) {
    return (
      <Card padding="default">
        <p className="text-muted-foreground text-center py-8">
          Final Four scenarios will appear once all regional champions are determined.
        </p>
      </Card>
    );
  }

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
      {outcomes.map((outcome, index) => (
        <OutcomeCard
          key={index}
          outcome={outcome}
          resolveTeamName={resolveTeamName}
          portfolioNameById={portfolioNameById}
        />
      ))}
    </div>
  );
}

interface OutcomeCardProps {
  outcome: FinalFourOutcome;
  resolveTeamName: (teamId: string, schoolId: string) => string;
  portfolioNameById: Map<string, string>;
}

function OutcomeCard({ outcome, resolveTeamName, portfolioNameById }: OutcomeCardProps) {
  const championName = resolveTeamName(outcome.champion.teamId, outcome.champion.schoolId);
  const runnerUpName = resolveTeamName(outcome.runnerUp.teamId, outcome.runnerUp.schoolId);
  const semi1Name = resolveTeamName(outcome.semifinal1Winner.teamId, outcome.semifinal1Winner.schoolId);
  const semi2Name = resolveTeamName(outcome.semifinal2Winner.teamId, outcome.semifinal2Winner.schoolId);

  const displayPortfolios = outcome.entries.slice(0, 10);

  return (
    <Card padding="default" className="overflow-hidden">
      <div>
        <h3 className="text-lg font-semibold">
          {championName} defeats {runnerUpName}
        </h3>
        <p className="text-sm text-muted-foreground">
          {semi1Name} &middot; {semi2Name}
        </p>
      </div>

      <div className="mt-3 space-y-1">
        {displayPortfolios.map((standing) => {
          const name = portfolioNameById.get(standing.portfolioId) || standing.portfolioId;
          const isInTheMoney = standing.inTheMoney;
          const payoutText = standing.payoutCents ? ` (${formatDollarsFromCents(standing.payoutCents)})` : '';

          return (
            <div
              key={standing.portfolioId}
              className={`flex justify-between items-center px-2 py-1 rounded text-sm ${
                isInTheMoney ? 'bg-accent/50 font-medium' : ''
              }`}
            >
              <span className="truncate mr-2">
                {standing.finishPosition}. {name}
                {isInTheMoney && <span className="text-muted-foreground">{payoutText}</span>}
              </span>
              <span className="tabular-nums shrink-0">{standing.totalReturns.toFixed(2)} pts</span>
            </div>
          );
        })}
      </div>
    </Card>
  );
}
