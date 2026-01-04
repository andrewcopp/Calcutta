import { AnalyticsResponse } from '../../types/analytics';

export function AnalyticsSummaryCard({ analytics }: { analytics: AnalyticsResponse }) {
  return (
    <div className="bg-white rounded-lg shadow-md p-6 mb-6">
      <h2 className="text-xl font-semibold mb-4">Overall Summary</h2>
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <div className="bg-blue-50 p-4 rounded">
          <div className="text-sm text-gray-600">Total Points Scored</div>
          <div className="text-2xl font-bold text-blue-600">{analytics.totalPoints.toFixed(2)}</div>
        </div>
        <div className="bg-green-50 p-4 rounded">
          <div className="text-sm text-gray-600">Total Investment</div>
          <div className="text-2xl font-bold text-green-600">${analytics.totalInvestment.toFixed(2)}</div>
        </div>
        <div className="bg-purple-50 p-4 rounded">
          <div className="text-sm text-gray-600">Baseline ROI</div>
          <div className="text-2xl font-bold text-purple-600">{analytics.baselineROI.toFixed(3)}</div>
          <div className="text-xs text-gray-500 mt-1">Points per dollar (raw)</div>
        </div>
      </div>
      <div className="mt-4 p-3 bg-gray-50 rounded text-sm text-gray-700">
        <strong>ROI Explanation:</strong> ROI values are normalized where 1.0 = average performance. Values &gt;1.0 indicate
        over-performance (better return than average), while &lt;1.0 indicates under-performance (worse return than average).
      </div>
    </div>
  );
}
