import React from 'react';
import { AnalyticsResponse } from '../../types/analytics';

export function TeamAnalyticsTab({ analytics }: { analytics: AnalyticsResponse }) {
  if (!analytics.teamAnalytics) {
    return null;
  }

  return (
    <div className="space-y-6">
      <div className="bg-white rounded-lg shadow-md p-6">
        <h2 className="text-xl font-semibold mb-4">Team Analytics</h2>
        <p className="text-sm text-gray-600 mb-4">
          Top teams by total points scored. Note: This data is not yet normalized by seed - teams with better seeds will
          naturally score more points.
        </p>
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">School</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Appearances
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Avg Seed</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Total Points
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Avg Points</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Total Investment
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Avg Investment
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">ROI</th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {analytics.teamAnalytics.slice(0, 50).map((team) => (
                <tr key={team.schoolId}>
                  <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">{team.schoolName}</td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{team.appearances}</td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{team.averageSeed.toFixed(1)}</td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{team.totalPoints.toFixed(2)}</td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{team.averagePoints.toFixed(2)}</td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">${team.totalInvestment.toFixed(2)}</td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    ${team.averageInvestment.toFixed(2)}
                  </td>
                  <td
                    className={`px-6 py-4 whitespace-nowrap text-sm font-semibold ${
                      team.roi > 1.0 ? 'text-green-600' : team.roi < 1.0 ? 'text-red-600' : 'text-gray-500'
                    }`}
                  >
                    {team.roi.toFixed(3)}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4">
        <h3 className="text-sm font-semibold text-yellow-800 mb-2">Future Enhancement</h3>
        <p className="text-sm text-yellow-700">
          Team analytics will be enhanced in the future to normalize for seed position. This will help identify teams
          that consistently over-perform or under-perform relative to their seed, revealing potential biases in bidding
          behavior.
        </p>
      </div>
    </div>
  );
}
