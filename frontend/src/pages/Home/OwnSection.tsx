import { useMemo } from 'react';
import {
  Cell,
  Pie,
  PieChart,
  ResponsiveContainer,
  Tooltip,
} from 'recharts';
import {
  darkTooltipContentStyle,
  darkTooltipLabelStyle,
  darkTooltipItemStyle,
} from './chartTheme';

interface OwnershipScenario {
  key: string;
  archetype: string;
  subtitle: string;
  seed: number;
  credits: number;
  totalCredits: number;
  fill: string;
}

const ownershipScenarios: OwnershipScenario[] = [
  {
    key: 'favorite-small',
    archetype: 'Favorite',
    subtitle: 'Top seed (1\u20133), heavily owned',
    seed: 1,
    credits: 5,
    totalCredits: 250,
    fill: '#2563EB',
  },
  {
    key: 'contender-balanced',
    archetype: 'Contender',
    subtitle: 'Strong seed, moderately owned',
    seed: 6,
    credits: 15,
    totalCredits: 120,
    fill: '#16A34A',
  },
  {
    key: 'value-medium',
    archetype: 'Value',
    subtitle: 'Middle seed, lightly owned',
    seed: 9,
    credits: 20,
    totalCredits: 80,
    fill: '#F59E0B',
  },
  {
    key: 'darkhorse-large',
    archetype: 'Dark Horse',
    subtitle: 'Lower seed (10\u201316), lightly owned',
    seed: 12,
    credits: 25,
    totalCredits: 55,
    fill: '#7C3AED',
  },
];

interface OwnSectionProps {
  prefersReducedMotion: boolean;
  revealClass: string;
}

export function OwnSection({ prefersReducedMotion, revealClass }: OwnSectionProps) {
  const duplicatedScenarios = useMemo(
    () => [...ownershipScenarios, ...ownershipScenarios],
    [],
  );

  return (
    <section id="own" className="py-20 sm:py-28">
      <div className={`transition-all duration-700 ease-out ${revealClass}`}>
        <div className="text-xs font-semibold tracking-wider text-blue-200">OWN</div>
        <div className="mt-3 flex flex-col lg:flex-row lg:items-end lg:justify-between gap-6">
          <h2 className="text-4xl sm:text-5xl font-bold tracking-tight">Your slice of each team</h2>
          <div className="text-sm text-white/70 max-w-md">
            Four different paths: favorites dilute, value and dark horses can be bigger shares.
          </div>
        </div>

        <div className="mt-10">
          <div className="home-own-marquee">
            <div
              className={`home-own-track ${prefersReducedMotion ? 'home-own-track--no-anim' : ''}`}
              aria-label="Ownership examples"
            >
              {duplicatedScenarios.map((s, idx) => {
                const yourPct = (s.credits / s.totalCredits) * 100;
                const restPct = 100 - yourPct;

                return (
                  <div
                    key={`${s.key}-${idx}`}
                    className="home-own-card rounded-3xl bg-white/5 ring-1 ring-white/10 backdrop-blur px-6 py-6"
                  >
                    <div className="flex items-start justify-between gap-3">
                      <div className="h-[56px] flex flex-col justify-center overflow-hidden">
                        <div className="text-lg font-semibold leading-snug truncate">{s.archetype}</div>
                        <div className="text-sm text-white/70 leading-snug truncate">{s.subtitle}</div>
                      </div>
                      <div className="text-xs font-semibold text-white/80 rounded-full bg-white/10 ring-1 ring-white/10 px-3 py-1">
                        Seed {s.seed}
                      </div>
                    </div>

                    <div className="mt-5 h-56">
                      <ResponsiveContainer width="100%" height="100%">
                        <PieChart>
                          <Pie
                            data={[
                              { name: 'Your share', value: yourPct, fill: s.fill },
                              { name: 'Everyone else', value: restPct, fill: 'rgba(255,255,255,0.18)' },
                            ]}
                            cx="50%"
                            cy="50%"
                            innerRadius={54}
                            outerRadius={92}
                            paddingAngle={2}
                            dataKey="value"
                            isAnimationActive={false}
                          >
                            <Cell fill={s.fill} />
                            <Cell fill="rgba(255,255,255,0.18)" />
                          </Pie>
                          <Tooltip
                            formatter={(value: number) => [`${value.toFixed(2)}%`, 'Ownership']}
                            contentStyle={darkTooltipContentStyle}
                            labelStyle={darkTooltipLabelStyle}
                            itemStyle={darkTooltipItemStyle}
                          />
                        </PieChart>
                      </ResponsiveContainer>
                    </div>

                    <div className="mt-5 grid grid-cols-3 gap-3">
                      <div className="rounded-2xl bg-white/5 ring-1 ring-white/10 px-4 py-3">
                        <div className="text-xs text-white/60">You</div>
                        <div className="mt-1 text-lg font-bold">{s.credits}</div>
                      </div>
                      <div className="rounded-2xl bg-white/5 ring-1 ring-white/10 px-4 py-3">
                        <div className="text-xs text-white/60">Total</div>
                        <div className="mt-1 text-lg font-bold">{s.totalCredits}</div>
                      </div>
                      <div className="rounded-2xl bg-white/5 ring-1 ring-white/10 px-4 py-3">
                        <div className="text-xs text-white/60">Your %</div>
                        <div className="mt-1 text-lg font-bold" style={{ color: s.fill }}>
                          {yourPct.toFixed(2)}%
                        </div>
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
