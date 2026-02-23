import { Bar, BarChart, CartesianGrid, Cell, ResponsiveContainer, Tooltip, XAxis, YAxis } from 'recharts';
import { darkTooltipContentStyle, darkTooltipLabelStyle, darkTooltipItemStyle, darkBarCursor } from './chartTheme';

const investData = [
  { label: 'Favorite', credits: 40, fill: '#2563EB' },
  { label: 'Contender', credits: 25, fill: '#16A34A' },
  { label: 'Value', credits: 20, fill: '#F59E0B' },
  { label: 'Dark Horse', credits: 15, fill: '#7C3AED' },
];

interface InvestSectionProps {
  revealClass: string;
}

export function InvestSection({ revealClass }: InvestSectionProps) {
  return (
    <section id="invest" className="py-20 sm:py-28">
      <div className={`transition-all duration-700 ease-out ${revealClass}`}>
        <div className="text-xs font-semibold tracking-wider text-blue-200">INVEST</div>
        <div className="mt-3 flex flex-col lg:flex-row lg:items-end lg:justify-between gap-6">
          <h2 className="text-4xl sm:text-5xl font-bold tracking-tight">Allocate 100 credits</h2>
          <div className="text-sm text-white/70 max-w-md">
            Favorite, Contender, Value, Dark Horse — your portfolio is your strategy.
          </div>
        </div>

        <div className="mt-10">
          <div className="h-72 sm:h-80">
            <ResponsiveContainer width="100%" height="100%">
              <BarChart data={investData} margin={{ top: 8, right: 20, left: 0, bottom: 10 }}>
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
                  formatter={(value: number) => [`${value} credits`, 'Investment']}
                  contentStyle={darkTooltipContentStyle}
                  labelStyle={darkTooltipLabelStyle}
                  itemStyle={darkTooltipItemStyle}
                  cursor={darkBarCursor}
                />
                <Bar dataKey="credits" radius={[12, 12, 0, 0]}>
                  {investData.map((entry) => (
                    <Cell key={entry.label} fill={entry.fill} />
                  ))}
                </Bar>
              </BarChart>
            </ResponsiveContainer>
          </div>
        </div>
      </div>
    </section>
  );
}
