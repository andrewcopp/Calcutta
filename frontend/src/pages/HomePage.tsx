import React from 'react';
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

export function HomePage() {
  const investData = [
    { label: 'Favorite', credits: 40, fill: '#2563EB' },
    { label: 'Contender', credits: 25, fill: '#16A34A' },
    { label: 'Value', credits: 20, fill: '#F59E0B' },
    { label: 'Dark Horse', credits: 15, fill: '#7C3AED' },
  ];

  const ownershipData = [
    { name: 'Your share', value: 18, fill: '#2563EB' },
    { name: 'Everyone else', value: 82, fill: '#E5E7EB' },
  ];

  const returnsData = [
    { round: 'R32', points: 50 },
    { round: 'S16', points: 150 },
    { round: 'E8', points: 300 },
    { round: 'F4', points: 500 },
    { round: 'Final', points: 750 },
    { round: 'Title', points: 1050 },
  ];

  return (
    <div className="min-h-screen bg-gradient-to-b from-blue-50 to-white">
      <div className="container mx-auto px-4 py-12">
        <div className="max-w-6xl mx-auto">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-10 items-center mb-12">
            <div>
              <h1 className="text-5xl font-bold text-gray-900 leading-tight mb-5">
                The NCAA Tournament pool that feels like investing
              </h1>
              <p className="text-xl text-gray-700 mb-6">
                Instead of picking 63 games, you build a portfolio. Invest credits in a handful of teams, own a share of their run, and earn returns as they advance.
              </p>
              <div className="flex flex-col sm:flex-row gap-3">
                <Link
                  to="/calcuttas/create"
                  className="inline-flex items-center justify-center bg-blue-600 text-white px-6 py-3 rounded-lg font-semibold hover:bg-blue-700 transition-colors"
                >
                  Create a pool
                </Link>
                <Link
                  to="/calcuttas"
                  className="inline-flex items-center justify-center bg-white text-blue-700 px-6 py-3 rounded-lg font-semibold border border-blue-200 hover:border-blue-300 hover:bg-blue-50 transition-colors"
                >
                  Join a pool
                </Link>
              </div>
              <div className="mt-4 text-sm text-gray-600">
                Want the details?{' '}
                <Link to="/rules" className="text-blue-700 hover:text-blue-900 font-medium">
                  Read how it works
                </Link>
                .
              </div>
            </div>

            <div className="bg-white rounded-xl shadow-lg p-6">
              <div className="flex items-start justify-between gap-4 mb-4">
                <div>
                  <h2 className="text-lg font-semibold text-gray-900">A portfolio, not a bracket</h2>
                  <p className="text-sm text-gray-600">A simple example allocation of 100 credits</p>
                </div>
                <div className="text-xs text-gray-500">Illustrative</div>
              </div>
              <div className="h-56">
                <ResponsiveContainer width="100%" height="100%">
                  <BarChart data={investData} margin={{ top: 6, right: 12, left: 0, bottom: 6 }}>
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis dataKey="label" tick={{ fontSize: 12 }} />
                    <YAxis tick={{ fontSize: 12 }} />
                    <Tooltip formatter={(value: number) => [`${value} credits`, 'Investment']} />
                    <Bar dataKey="credits" radius={[6, 6, 0, 0]}>
                      {investData.map((entry) => (
                        <Cell key={entry.label} fill={entry.fill} />
                      ))}
                    </Bar>
                  </BarChart>
                </ResponsiveContainer>
              </div>
            </div>
          </div>

          <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 mb-12">
            <div className="bg-white rounded-xl shadow-lg p-6">
              <div className="mb-4">
                <div className="text-sm font-semibold text-blue-700">Step 1</div>
                <h2 className="text-2xl font-bold text-gray-900">Invest</h2>
                <p className="text-gray-600 mt-2">
                  You get 100 credits to allocate across teams. Concentrate on favorites for safer points or take positions on long-shots for a bigger share.
                </p>
              </div>
              <div className="rounded-lg bg-blue-50 p-4 text-sm text-blue-900">
                Typical format: invest in 3–10 teams, max 50 credits on any single team.
              </div>
            </div>

            <div className="bg-white rounded-xl shadow-lg p-6">
              <div className="mb-4">
                <div className="text-sm font-semibold text-blue-700">Step 2</div>
                <h2 className="text-2xl font-bold text-gray-900">Own</h2>
                <p className="text-gray-600 mt-2">
                  Your ownership is proportional to total credits invested by everyone. Every new credit slightly dilutes everyone’s share.
                </p>
              </div>
              <div className="h-56">
                <ResponsiveContainer width="100%" height="100%">
                  <PieChart>
                    <Pie
                      data={ownershipData}
                      cx="50%"
                      cy="50%"
                      innerRadius={52}
                      outerRadius={84}
                      paddingAngle={2}
                      dataKey="value"
                      isAnimationActive={false}
                    >
                      {ownershipData.map((entry) => (
                        <Cell key={entry.name} fill={entry.fill} />
                      ))}
                    </Pie>
                    <Tooltip formatter={(value: number) => [`${value.toFixed(0)}%`, 'Ownership']} />
                  </PieChart>
                </ResponsiveContainer>
              </div>
              <div className="text-sm text-gray-600 mt-2">
                Example: invest 10 credits in a team with 100 total credits invested → you own 10%.
              </div>
            </div>

            <div className="bg-white rounded-xl shadow-lg p-6">
              <div className="mb-4">
                <div className="text-sm font-semibold text-blue-700">Step 3</div>
                <h2 className="text-2xl font-bold text-gray-900">Earn</h2>
                <p className="text-gray-600 mt-2">
                  Teams generate points as they advance. Your score is the points earned multiplied by your ownership share.
                </p>
              </div>
              <div className="h-56">
                <ResponsiveContainer width="100%" height="100%">
                  <BarChart data={returnsData} margin={{ top: 6, right: 12, left: 0, bottom: 6 }}>
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis dataKey="round" tick={{ fontSize: 12 }} />
                    <YAxis tick={{ fontSize: 12 }} />
                    <Tooltip formatter={(value: number) => [`${value} pts`, 'Total points']} />
                    <Bar dataKey="points" fill="#16A34A" radius={[6, 6, 0, 0]} />
                  </BarChart>
                </ResponsiveContainer>
              </div>
              <div className="text-sm text-gray-600 mt-2">
                Deeper runs matter a lot: a small slice of a champion can beat 100% of a team that wins once.
              </div>
            </div>
          </div>

          <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 mb-12">
            <div className="lg:col-span-2 bg-white rounded-xl shadow-lg p-8">
              <h2 className="text-2xl font-bold text-gray-900 mb-3">Why it’s more fun than a traditional bracket</h2>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-gray-700">
                <div className="bg-gray-50 rounded-lg p-4">
                  <div className="font-semibold text-gray-900 mb-1">Fewer picks, more strategy</div>
                  <div className="text-sm">You’re making a portfolio decision, not trying to be perfect across 63 games.</div>
                </div>
                <div className="bg-gray-50 rounded-lg p-4">
                  <div className="font-semibold text-gray-900 mb-1">Root for more outcomes</div>
                  <div className="text-sm">You care about ownership percentage and paths, not just “my champion is alive.”</div>
                </div>
                <div className="bg-gray-50 rounded-lg p-4">
                  <div className="font-semibold text-gray-900 mb-1">Underdogs actually matter</div>
                  <div className="text-sm">A cheap team can be a huge swing if you own a big share and they win a game or two.</div>
                </div>
                <div className="bg-gray-50 rounded-lg p-4">
                  <div className="font-semibold text-gray-900 mb-1">Markets create drama</div>
                  <div className="text-sm">As the field invests, ownership gets diluted—there’s real tradeoff between “good” and “good value.”</div>
                </div>
              </div>
            </div>

            <div className="bg-white rounded-xl shadow-lg p-8">
              <h2 className="text-2xl font-bold text-gray-900 mb-3">FAQ</h2>
              <div className="space-y-4 text-sm text-gray-700">
                <div>
                  <div className="font-semibold text-gray-900">How many teams do I pick?</div>
                  <div>Usually 3–10 teams. Enough to diversify, not enough to be boring.</div>
                </div>
                <div>
                  <div className="font-semibold text-gray-900">Do I need to know every matchup?</div>
                  <div>No. You’re optimizing a small set of positions and upside, not predicting every game.</div>
                </div>
                <div>
                  <div className="font-semibold text-gray-900">Where are the full rules?</div>
                  <div>
                    <Link to="/rules" className="text-blue-700 hover:text-blue-900 font-medium">
                      See “How It Works”
                    </Link>
                    .
                  </div>
                </div>
              </div>
            </div>
          </div>

          <div className="bg-white rounded-xl shadow-lg p-10 text-center">
            <h2 className="text-3xl font-bold text-gray-900 mb-3">Ready to build your portfolio?</h2>
            <p className="text-gray-600 mb-6">
              Create a new pool for your group, or join an existing one.
            </p>
            <div className="flex flex-col sm:flex-row gap-3 justify-center">
              <Link
                to="/calcuttas/create"
                className="inline-flex items-center justify-center bg-blue-600 text-white px-8 py-3 rounded-lg font-semibold hover:bg-blue-700 transition-colors"
              >
                Create a pool
              </Link>
              <Link
                to="/calcuttas"
                className="inline-flex items-center justify-center bg-gray-900 text-white px-8 py-3 rounded-lg font-semibold hover:bg-gray-800 transition-colors"
              >
                Join a pool
              </Link>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}