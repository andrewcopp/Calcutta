import { useEffect, useMemo, useState } from 'react';
import { Cell, Pie, PieChart, ResponsiveContainer, Tooltip } from 'recharts';
import { Card } from '../../components/ui/Card';

interface OwnershipScenario {
  key: string;
  title: string;
  subtitle: string;
  seed: number;
  credits: number;
  totalCredits: number;
  fill: string;
}

const ownershipScenarios: OwnershipScenario[] = [
  {
    key: 'top-seed-small',
    title: 'Top seed (1\u20133): tiny ownership',
    subtitle: 'Lots of investors \u2192 your slice is smaller',
    seed: 1,
    credits: 5,
    totalCredits: 250,
    fill: '#2563EB',
  },
  {
    key: 'mid-seed-medium',
    title: 'Middle seed: balanced ownership',
    subtitle: 'Moderate investors \u2192 a meaningful slice',
    seed: 6,
    credits: 15,
    totalCredits: 120,
    fill: '#16A34A',
  },
  {
    key: 'lower-seed-large',
    title: 'Lower seed (10\u201316): larger ownership',
    subtitle: 'Few investors \u2192 a bigger slice (but higher risk)',
    seed: 12,
    credits: 20,
    totalCredits: 55,
    fill: '#7C3AED',
  },
];

interface OwnCardProps {
  prefersReducedMotion: boolean;
}

export function OwnCard({ prefersReducedMotion }: OwnCardProps) {
  const scenarios = useMemo(() => ownershipScenarios, []);

  const [scenarioIndex, setScenarioIndex] = useState(0);
  const activeScenario = scenarios[scenarioIndex];

  useEffect(() => {
    if (prefersReducedMotion) return;

    const id = window.setInterval(() => {
      setScenarioIndex((i) => (i + 1) % scenarios.length);
    }, 6500);

    return () => window.clearInterval(id);
  }, [scenarios.length, prefersReducedMotion]);

  return (
    <Card className="shadow-lg">
      <div className="flex items-center justify-between gap-4 mb-4">
        <div>
          <h2 className="text-2xl font-semibold text-foreground">Own</h2>
          <p className="text-muted-foreground">Your ownership is proportional to total credits invested by everyone.</p>
        </div>
        <div className="flex items-center gap-2">
          <button
            type="button"
            onClick={() => setScenarioIndex((i) => (i - 1 + scenarios.length) % scenarios.length)}
            className="h-9 w-9 rounded-full bg-card shadow-sm ring-1 ring-gray-200 text-foreground hover:bg-accent"
            aria-label="Previous scenario"
          >
            ←
          </button>
          <button
            type="button"
            onClick={() => setScenarioIndex((i) => (i + 1) % scenarios.length)}
            className="h-9 w-9 rounded-full bg-card shadow-sm ring-1 ring-gray-200 text-foreground hover:bg-accent"
            aria-label="Next scenario"
          >
            →
          </button>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6 items-center">
        <div className="h-72">
          <ResponsiveContainer width="100%" height="100%">
            <PieChart>
              <Pie
                data={[
                  {
                    name: 'Your share',
                    value: (activeScenario.credits / activeScenario.totalCredits) * 100,
                    fill: activeScenario.fill,
                  },
                  {
                    name: 'Everyone else',
                    value: 100 - (activeScenario.credits / activeScenario.totalCredits) * 100,
                    fill: '#E5E7EB',
                  },
                ]}
                cx="50%"
                cy="50%"
                innerRadius={58}
                outerRadius={94}
                paddingAngle={2}
                dataKey="value"
                isAnimationActive={false}
              >
                <Cell fill={activeScenario.fill} />
                <Cell fill="#E5E7EB" />
              </Pie>
              <Tooltip formatter={(value: number) => [`${value.toFixed(2)}%`, 'Ownership']} />
            </PieChart>
          </ResponsiveContainer>
        </div>

        <div>
          <div className="text-lg font-semibold text-foreground">{activeScenario.title}</div>
          <div className="text-sm text-muted-foreground">{activeScenario.subtitle}</div>
          <div className="mt-4 grid grid-cols-3 gap-3">
            <div className="rounded-lg bg-accent p-3">
              <div className="text-xs font-semibold text-muted-foreground">Seed</div>
              <div className="mt-1 text-lg font-bold text-foreground">{activeScenario.seed}</div>
            </div>
            <div className="rounded-lg bg-accent p-3">
              <div className="text-xs font-semibold text-muted-foreground">You invest</div>
              <div className="mt-1 text-lg font-bold text-foreground">{activeScenario.credits}</div>
            </div>
            <div className="rounded-lg bg-accent p-3">
              <div className="text-xs font-semibold text-muted-foreground">Total</div>
              <div className="mt-1 text-lg font-bold text-foreground">{activeScenario.totalCredits}</div>
            </div>
          </div>
          <div className="mt-4 rounded-lg bg-primary/10 p-3">
            <div className="text-xs font-semibold text-primary">Your ownership</div>
            <div className="mt-1 text-2xl font-bold text-blue-900">
              {((activeScenario.credits / activeScenario.totalCredits) * 100).toFixed(2)}%
            </div>
          </div>
          <div className="mt-4 text-sm text-muted-foreground">
            Top seeds are popular (lots of investors), so your slice is often small. Lower seeds can be larger slices—if
            they win, it matters.
          </div>
        </div>
      </div>

      <div className="mt-5 flex items-center justify-between">
        <div className="flex items-center gap-2">
          {scenarios.map((s, idx) => {
            const isActive = idx === scenarioIndex;
            return (
              <button
                key={s.key}
                type="button"
                onClick={() => setScenarioIndex(idx)}
                className={`h-2.5 w-2.5 rounded-full ${isActive ? 'bg-primary' : 'bg-gray-300 hover:bg-gray-400'}`}
                aria-label={`Go to scenario ${idx + 1}`}
              />
            );
          })}
        </div>
        <div className="text-xs text-muted-foreground">
          {scenarioIndex + 1} / {scenarios.length}
        </div>
      </div>
    </Card>
  );
}
