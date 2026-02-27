import { Card } from '../../components/ui/Card';
import type { ScoringRule } from '../../schemas/pool';

const WIN_INDEX_LABELS: Record<number, string> = {
  0: 'Bye',
  1: 'First Four win',
  2: 'Round of 64 win',
  3: 'Round of 32 win',
  4: 'Sweet 16 win',
  5: 'Elite 8 win',
  6: 'Final Four win',
  7: 'Championship win',
};

function labelForWinIndex(winIndex: number): string {
  return WIN_INDEX_LABELS[winIndex] ?? `Win #${winIndex}`;
}

interface ScoringCardProps {
  scoringRules?: ScoringRule[];
}

export function ScoringCard({ scoringRules }: ScoringCardProps) {
  return (
    <Card className="shadow-lg">
      <h2 className="text-2xl font-semibold text-foreground mb-4">Scoring</h2>
      <div className="space-y-4">
        <p className="text-muted-foreground">Every win is a dividend. The deeper the run, the bigger the payout.</p>
        {scoringRules && scoringRules.length > 0 ? (
          <ul className="list-disc list-inside space-y-2 text-muted-foreground">
            {[...scoringRules]
              .sort((a, b) => a.winIndex - b.winIndex)
              .map((rule) => (
                <li key={rule.winIndex}>
                  {labelForWinIndex(rule.winIndex)}: {rule.pointsAwarded} pts
                </li>
              ))}
          </ul>
        ) : (
          <p className="text-muted-foreground">
            Points are awarded for each win based on the round. The deeper your teams go, the more you earn.
          </p>
        )}
        <p className="text-muted-foreground mt-4">
          Your returns are each team's points multiplied by your ownership percentage, totaled across your portfolio.
        </p>
      </div>
    </Card>
  );
}
