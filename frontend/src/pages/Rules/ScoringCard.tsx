import { Card } from '../../components/ui/Card';

export function ScoringCard() {
  return (
    <Card className="shadow-lg">
      <h2 className="text-2xl font-semibold text-foreground mb-4">Scoring</h2>
      <div className="space-y-4">
        <p className="text-muted-foreground">Every win is a dividend. The deeper the run, the bigger the payout.</p>
        <ul className="list-disc list-inside space-y-2 text-muted-foreground">
          <li>First round exit: 0 pts</li>
          <li>Round of 32: 50 pts</li>
          <li>Sweet 16: 150 pts</li>
          <li>Elite 8: 300 pts</li>
          <li>Final Four: 500 pts</li>
          <li>Championship game: 750 pts</li>
          <li>Win the title: 1050 pts</li>
        </ul>
        <p className="text-muted-foreground mt-4">
          Your score is each team's points multiplied by your ownership percentage, totaled across your portfolio.
        </p>
      </div>
    </Card>
  );
}
