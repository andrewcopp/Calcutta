import React, { useEffect, useMemo, useState } from 'react';
import { Link } from 'react-router-dom';
import {
  Bar,
  BarChart,
  CartesianGrid,
  Cell,
  Pie,
  PieChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts';

export function RulesPage() {
  const prefersReducedMotion = useMemo(() => {
    return (
      typeof window !== 'undefined' &&
      typeof window.matchMedia === 'function' &&
      window.matchMedia('(prefers-reduced-motion: reduce)').matches
    );
  }, []);

  const investData = useMemo(
    () => [
      { label: 'Top seed', credits: 35, fill: '#2563EB' },
      { label: 'Contender', credits: 25, fill: '#16A34A' },
      { label: 'Value', credits: 20, fill: '#F59E0B' },
      { label: 'Sleeper', credits: 20, fill: '#7C3AED' },
    ],
    [],
  );

  const ownershipScenarios = useMemo(
    () => [
      {
        key: 'top-seed-small',
        title: 'Top seed (1–3): tiny ownership',
        subtitle: 'Lots of investors → your slice is smaller',
        seed: 1,
        credits: 5,
        totalCredits: 250,
        fill: '#2563EB',
      },
      {
        key: 'mid-seed-medium',
        title: 'Middle seed: balanced ownership',
        subtitle: 'Moderate investors → a meaningful slice',
        seed: 6,
        credits: 15,
        totalCredits: 120,
        fill: '#16A34A',
      },
      {
        key: 'lower-seed-large',
        title: 'Lower seed (10–16): larger ownership',
        subtitle: 'Few investors → a bigger slice (but higher risk)',
        seed: 12,
        credits: 20,
        totalCredits: 55,
        fill: '#7C3AED',
      },
    ],
    [],
  );

  const [scenarioIndex, setScenarioIndex] = useState(0);
  const activeScenario = ownershipScenarios[scenarioIndex];

  useEffect(() => {
    if (prefersReducedMotion) return;

    const id = window.setInterval(() => {
      setScenarioIndex((i) => (i + 1) % ownershipScenarios.length);
    }, 6500);

    return () => window.clearInterval(id);
  }, [ownershipScenarios.length, prefersReducedMotion]);

  const simulatedEntry = useMemo(
    () => [
      {
        key: 'champion-slice',
        team: 'Top seed favorite',
        seed: 1,
        ownershipPct: 2,
        path: 'Wins the title',
        teamPoints: 1050,
        fill: '#2563EB',
      },
      {
        key: 'deep-run',
        team: 'Contender',
        seed: 6,
        ownershipPct: 12.5,
        path: 'Reaches the Elite 8',
        teamPoints: 500,
        fill: '#16A34A',
      },
      {
        key: 'one-win',
        team: 'Sleeper',
        seed: 12,
        ownershipPct: 35,
        path: 'Wins one game',
        teamPoints: 50,
        fill: '#7C3AED',
      },
    ],
    [],
  );

  const simulatedEntryWithScore = useMemo(() => {
    return simulatedEntry.map((t) => ({
      ...t,
      yourPoints: (t.teamPoints * t.ownershipPct) / 100,
    }));
  }, [simulatedEntry]);

  const simulatedTotal = useMemo(() => {
    return simulatedEntryWithScore.reduce((sum, t) => sum + t.yourPoints, 0);
  }, [simulatedEntryWithScore]);

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-6">
        <Link to="/" className="text-blue-600 hover:text-blue-800">← Back to Home</Link>
      </div>

      <div className="max-w-4xl mx-auto">
        <h1 className="text-4xl font-bold text-gray-900 mb-2">How Calcutta Works</h1>
        <p className="text-gray-600 mb-8">
          This is a friendly game. Calcutta helps your group track ownership and points — it does not facilitate gambling or real-money winnings.
        </p>

        <div className="space-y-8">
          <section className="bg-white rounded-lg shadow-lg p-6">
            <h2 className="text-2xl font-semibold text-gray-900 mb-4">What is a Calcutta?</h2>
            <p className="text-gray-600">
              Instead of picking 63 games, you invest credits in a small portfolio of teams. You own a percentage of each team, and your score is based on how far those teams advance.
            </p>
          </section>

          <section className="bg-white rounded-lg shadow-lg p-6">
            <div className="flex items-center justify-between gap-4 mb-4">
              <div>
                <h2 className="text-2xl font-semibold text-gray-900">Invest</h2>
                <p className="text-gray-600">Allocate 100 credits across a handful of teams.</p>
              </div>
              <div className="text-xs text-gray-500">Illustrative</div>
            </div>
            <div className="h-72">
              <ResponsiveContainer width="100%" height="100%">
                <BarChart data={investData} margin={{ top: 6, right: 12, left: 0, bottom: 6 }}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="label" tick={{ fontSize: 12 }} />
                  <YAxis tick={{ fontSize: 12 }} />
                  <Tooltip formatter={(value: number) => [`${value} credits`, 'Investment']} />
                  <Bar dataKey="credits" radius={[8, 8, 0, 0]}>
                    {investData.map((entry) => (
                      <Cell key={entry.label} fill={entry.fill} />
                    ))}
                  </Bar>
                </BarChart>
              </ResponsiveContainer>
            </div>
            <div className="mt-4 text-sm text-gray-600">
              Typical constraints: invest in 3–10 teams, max 50 credits on any single team.
            </div>
          </section>

          <section className="bg-white rounded-lg shadow-lg p-6">
            <div className="flex items-center justify-between gap-4 mb-4">
              <div>
                <h2 className="text-2xl font-semibold text-gray-900">Own</h2>
                <p className="text-gray-600">Your ownership is proportional to total credits invested by everyone.</p>
              </div>
              <div className="flex items-center gap-2">
                <button
                  type="button"
                  onClick={() => setScenarioIndex((i) => (i - 1 + ownershipScenarios.length) % ownershipScenarios.length)}
                  className="h-9 w-9 rounded-full bg-white shadow-sm ring-1 ring-gray-200 text-gray-700 hover:bg-gray-50"
                  aria-label="Previous scenario"
                >
                  ←
                </button>
                <button
                  type="button"
                  onClick={() => setScenarioIndex((i) => (i + 1) % ownershipScenarios.length)}
                  className="h-9 w-9 rounded-full bg-white shadow-sm ring-1 ring-gray-200 text-gray-700 hover:bg-gray-50"
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
                <div className="text-lg font-semibold text-gray-900">{activeScenario.title}</div>
                <div className="text-sm text-gray-600">{activeScenario.subtitle}</div>
                <div className="mt-4 grid grid-cols-3 gap-3">
                  <div className="rounded-lg bg-gray-50 p-3">
                    <div className="text-xs font-semibold text-gray-500">Seed</div>
                    <div className="mt-1 text-lg font-bold text-gray-900">{activeScenario.seed}</div>
                  </div>
                  <div className="rounded-lg bg-gray-50 p-3">
                    <div className="text-xs font-semibold text-gray-500">You invest</div>
                    <div className="mt-1 text-lg font-bold text-gray-900">{activeScenario.credits}</div>
                  </div>
                  <div className="rounded-lg bg-gray-50 p-3">
                    <div className="text-xs font-semibold text-gray-500">Total</div>
                    <div className="mt-1 text-lg font-bold text-gray-900">{activeScenario.totalCredits}</div>
                  </div>
                </div>
                <div className="mt-4 rounded-lg bg-blue-50 p-3">
                  <div className="text-xs font-semibold text-blue-700">Your ownership</div>
                  <div className="mt-1 text-2xl font-bold text-blue-900">
                    {((activeScenario.credits / activeScenario.totalCredits) * 100).toFixed(2)}%
                  </div>
                </div>
                <div className="mt-4 text-sm text-gray-600">
                  Top seeds are popular (lots of investors), so your slice is often small. Lower seeds can be larger slices—if they win, it matters.
                </div>
              </div>
            </div>

            <div className="mt-5 flex items-center justify-between">
              <div className="flex items-center gap-2">
                {ownershipScenarios.map((s, idx) => {
                  const isActive = idx === scenarioIndex;
                  return (
                    <button
                      key={s.key}
                      type="button"
                      onClick={() => setScenarioIndex(idx)}
                      className={`h-2.5 w-2.5 rounded-full ${isActive ? 'bg-blue-600' : 'bg-gray-300 hover:bg-gray-400'}`}
                      aria-label={`Go to scenario ${idx + 1}`}
                    />
                  );
                })}
              </div>
              <div className="text-xs text-gray-500">
                {scenarioIndex + 1} / {ownershipScenarios.length}
              </div>
            </div>
          </section>

          <section className="bg-white rounded-lg shadow-lg p-6">
            <div className="flex items-center justify-between gap-4 mb-4">
              <div>
                <h2 className="text-2xl font-semibold text-gray-900">Earn</h2>
                <p className="text-gray-600">Your points are team points × your ownership percentage.</p>
              </div>
              <div className="rounded-full bg-gray-900 px-3 py-1 text-sm font-semibold text-white">
                Example total: {simulatedTotal.toFixed(1)} pts
              </div>
            </div>

            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-gray-200">
                <thead className="bg-gray-50">
                  <tr>
                    <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Team</th>
                    <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Seed</th>
                    <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Path</th>
                    <th className="px-4 py-2 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Ownership</th>
                    <th className="px-4 py-2 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Team Points</th>
                    <th className="px-4 py-2 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Your Points</th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {simulatedEntryWithScore.map((t) => (
                    <tr key={t.key}>
                      <td className="px-4 py-2 whitespace-nowrap text-sm font-medium text-gray-900">{t.team}</td>
                      <td className="px-4 py-2 whitespace-nowrap text-sm text-gray-600">{t.seed}</td>
                      <td className="px-4 py-2 whitespace-nowrap text-sm text-gray-600">{t.path}</td>
                      <td className="px-4 py-2 whitespace-nowrap text-sm text-right text-gray-600">{t.ownershipPct}%</td>
                      <td className="px-4 py-2 whitespace-nowrap text-sm text-right text-gray-600">{t.teamPoints}</td>
                      <td className="px-4 py-2 whitespace-nowrap text-sm text-right font-semibold text-gray-900">{t.yourPoints.toFixed(1)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>

            <div className="mt-6">
              <h3 className="text-sm font-semibold text-gray-900 mb-2">Contribution to your score</h3>
              <div className="h-64">
                <ResponsiveContainer width="100%" height="100%">
                  <BarChart data={simulatedEntryWithScore} margin={{ top: 6, right: 12, left: 0, bottom: 6 }}>
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis dataKey="team" tick={{ fontSize: 12 }} />
                    <YAxis tick={{ fontSize: 12 }} />
                    <Tooltip formatter={(value: number) => [`${value.toFixed(1)} pts`, 'Your points']} />
                    <Bar dataKey="yourPoints" radius={[8, 8, 0, 0]}>
                      {simulatedEntryWithScore.map((t) => (
                        <Cell key={t.key} fill={t.fill} />
                      ))}
                    </Bar>
                  </BarChart>
                </ResponsiveContainer>
              </div>
            </div>
          </section>

          <section className="bg-white rounded-lg shadow-lg p-6">
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
                Your score is the sum of each team’s points multiplied by your ownership percentage.
              </p>
            </div>
          </section>
        </div>
      </div>
    </div>
  );
}