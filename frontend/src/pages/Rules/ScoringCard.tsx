import { Card } from '../../components/ui/Card';

export function ScoringCard() {
  return (
    <Card className="shadow-lg">
      <h2 className="text-2xl font-semibold text-gray-900 mb-4">Scoring System</h2>
      <div className="space-y-4">
        <p className="text-gray-600">
          Points are awarded based on how far teams advance in the tournament:
        </p>
        <ul className="list-disc list-inside space-y-2 text-gray-600">
          <li>Eliminated in first round (or play-in): 0</li>
          <li>Reach Round of 32: 50</li>
          <li>Reach Sweet 16: 150</li>
          <li>Reach Elite 8: 300</li>
          <li>Reach Final Four: 500</li>
          <li>Reach Championship: 750</li>
          <li>Win the title: 1050</li>
        </ul>
        <p className="text-gray-600 mt-4">
          Your score is the sum of each team's points multiplied by your ownership percentage.
        </p>
      </div>
    </Card>
  );
}
