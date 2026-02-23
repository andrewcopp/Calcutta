import { useMemo } from 'react';
import { Link } from 'react-router-dom';
import { Bar, BarChart, CartesianGrid, Cell, ResponsiveContainer, Tooltip, XAxis, YAxis } from 'recharts';
import { darkTooltipContentStyle, darkTooltipLabelStyle, darkTooltipItemStyle, darkBarCursor } from './chartTheme';

interface EarnBar {
  key: string;
  label: string;
  yourPoints: number;
  fill: string;
}

const earnBars: EarnBar[] = [
  { key: 'favorite', label: 'Favorite', yourPoints: 21, fill: '#2563EB' },
  { key: 'contender', label: 'Contender', yourPoints: 62.5, fill: '#16A34A' },
  { key: 'value', label: 'Value', yourPoints: 38, fill: '#F59E0B' },
  { key: 'darkhorse', label: 'Dark Horse', yourPoints: 17.5, fill: '#7C3AED' },
];

interface EarnSectionProps {
  revealClass: string;
}

export function EarnSection({ revealClass }: EarnSectionProps) {
  const bars = useMemo(() => earnBars, []);

  return (
    <section id="earn" className="py-20 sm:py-28">
      <div className={`transition-all duration-700 ease-out ${revealClass}`}>
        <div className="text-xs font-semibold tracking-wider text-blue-200">EARN</div>
        <div className="mt-3 flex flex-col lg:flex-row lg:items-end lg:justify-between gap-6">
          <h2 className="text-4xl sm:text-5xl font-bold tracking-tight">Multiple paths to points</h2>
          <div className="text-sm text-white/70 max-w-md">
            Four archetypes can contribute in different ways — a small slice of a favorite, or a bigger slice of the
            right dark horse.
          </div>
        </div>

        <div className="mt-10">
          <div className="h-72 sm:h-80">
            <ResponsiveContainer width="100%" height="100%">
              <BarChart data={bars} margin={{ top: 8, right: 20, left: 0, bottom: 12 }}>
                <CartesianGrid stroke="rgba(255,255,255,0.12)" strokeDasharray="3 3" />
                <XAxis
                  dataKey="label"
                  tick={{ fontSize: 12, fill: 'rgba(255,255,255,0.75)' }}
                  axisLine={{ stroke: 'rgba(255,255,255,0.25)' }}
                />
                <YAxis
                  tick={{ fontSize: 12, fill: 'rgba(255,255,255,0.75)' }}
                  axisLine={{ stroke: 'rgba(255,255,255,0.25)' }}
                />
                <Tooltip
                  formatter={(value: number) => [`${value.toFixed(1)} points`, 'Contribution']}
                  contentStyle={darkTooltipContentStyle}
                  labelStyle={darkTooltipLabelStyle}
                  itemStyle={darkTooltipItemStyle}
                  cursor={darkBarCursor}
                />
                <Bar dataKey="yourPoints" radius={[12, 12, 0, 0]}>
                  {bars.map((b) => (
                    <Cell key={b.key} fill={b.fill} />
                  ))}
                </Bar>
              </BarChart>
            </ResponsiveContainer>
          </div>

          <div className="mt-6 flex justify-center">
            <Link to="/rules" className="text-white/80 hover:text-white underline underline-offset-4">
              See the full scoring + strategy walkthrough
            </Link>
          </div>
        </div>
      </div>
    </section>
  );
}
