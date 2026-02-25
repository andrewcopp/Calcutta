import { useMemo } from 'react';
import type { CalcuttaDashboard, CalcuttaEntry, FinalFourOutcome } from '../../schemas/calcutta';
import type { School } from '../../schemas/school';
import type { TournamentTeam } from '../../schemas/tournament';
import { formatDollarsFromCents } from '../../utils/format';
import { Card } from '../../components/ui/Card';

interface FinalFourTabProps {
  entries: CalcuttaEntry[];
  dashboard: CalcuttaDashboard;
  schools: School[];
  tournamentTeams: TournamentTeam[];
}

export function FinalFourTab({ entries, dashboard, schools, tournamentTeams }: FinalFourTabProps) {
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

  const entryNameById = useMemo(() => new Map(entries.map((e) => [e.id, e.name])), [entries]);

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
          entryNameById={entryNameById}
        />
      ))}
    </div>
  );
}

interface OutcomeCardProps {
  outcome: FinalFourOutcome;
  resolveTeamName: (teamId: string, schoolId: string) => string;
  entryNameById: Map<string, string>;
}

function OutcomeCard({ outcome, resolveTeamName, entryNameById }: OutcomeCardProps) {
  const championName = resolveTeamName(outcome.champion.teamId, outcome.champion.schoolId);
  const runnerUpName = resolveTeamName(outcome.runnerUp.teamId, outcome.runnerUp.schoolId);
  const semi1Name = resolveTeamName(outcome.semifinal1Winner.teamId, outcome.semifinal1Winner.schoolId);
  const semi2Name = resolveTeamName(outcome.semifinal2Winner.teamId, outcome.semifinal2Winner.schoolId);

  const displayEntries = outcome.entries.slice(0, 10);

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
        {displayEntries.map((standing) => {
          const name = entryNameById.get(standing.entryId) || standing.entryId;
          const isInTheMoney = standing.inTheMoney;
          const payoutText = standing.payoutCents ? ` (${formatDollarsFromCents(standing.payoutCents)})` : '';

          return (
            <div
              key={standing.entryId}
              className={`flex justify-between items-center px-2 py-1 rounded text-sm ${
                isInTheMoney ? 'bg-accent/50 font-medium' : ''
              }`}
            >
              <span className="truncate mr-2">
                {standing.finishPosition}. {name}
                {isInTheMoney && <span className="text-muted-foreground">{payoutText}</span>}
              </span>
              <span className="tabular-nums shrink-0">{standing.totalPoints.toFixed(2)} pts</span>
            </div>
          );
        })}
      </div>
    </Card>
  );
}
