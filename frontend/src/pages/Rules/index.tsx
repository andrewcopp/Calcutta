import { Link } from 'react-router-dom';
import { Card } from '../../components/ui/Card';
import { PageContainer, PageHeader } from '../../components/ui/Page';
import { useReducedMotion } from '../../hooks/useReducedMotion';
import { InvestCard } from './InvestCard';
import { OwnCard } from './OwnCard';
import { EarnCard } from './EarnCard';
import { ScoringCard } from './ScoringCard';

export function RulesPage() {
  const prefersReducedMotion = useReducedMotion();

  return (
    <PageContainer>
      <div className="max-w-4xl mx-auto">
        <PageHeader
          title={<span className="text-4xl font-bold text-gray-900">How Calcutta Works</span>}
          subtitle="This is a friendly game. Calcutta helps your group track ownership and points — it does not facilitate gambling or real-money winnings."
          actions={
            <Link to="/" className="text-blue-600 hover:text-blue-800">
              ← Back to Home
            </Link>
          }
        />

        <div className="space-y-8">
          <Card className="shadow-lg">
            <h2 className="text-2xl font-semibold text-gray-900 mb-4">What is a Calcutta?</h2>
            <p className="text-gray-600">
              Instead of picking 63 games, you invest credits in a small portfolio of teams. You own a percentage of each team, and your score is based on how far those teams advance.
            </p>
          </Card>

          <InvestCard />
          <OwnCard prefersReducedMotion={prefersReducedMotion} />
          <EarnCard />
          <ScoringCard />
        </div>
      </div>
    </PageContainer>
  );
}
