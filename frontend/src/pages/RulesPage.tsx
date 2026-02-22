import { Link } from 'react-router-dom';
import { Card } from '../components/ui/Card';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { useReducedMotion } from '../hooks/useReducedMotion';
import { InvestCard } from './Rules/InvestCard';
import { OwnCard } from './Rules/OwnCard';
import { EarnCard } from './Rules/EarnCard';
import { ScoringCard } from './Rules/ScoringCard';

export function RulesPage() {
  const prefersReducedMotion = useReducedMotion();

  return (
    <PageContainer>
      <div className="max-w-4xl mx-auto">
        <PageHeader
          title={<span className="text-4xl font-bold text-gray-900">How It Works</span>}
          subtitle="Free to play. No money changes hands — just bragging rights."
          actions={
            <Link to="/" className="text-blue-600 hover:text-blue-800">
              ← Back to Home
            </Link>
          }
        />

        <div className="space-y-8">
          <Card className="shadow-lg">
            <h2 className="text-2xl font-semibold text-gray-900 mb-4">What is a Calcutta pool?</h2>
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
