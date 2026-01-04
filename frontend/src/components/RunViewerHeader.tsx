import React from 'react';
import { Link } from 'react-router-dom';

type RunViewerTab = 'rankings' | 'returns' | 'investments';

export function RunViewerHeader({
  year,
  runId,
  runName,
  activeTab,
}: {
  year: number;
  runId: string;
  runName?: string | null;
  activeTab: RunViewerTab;
}) {
  const encodedRunId = encodeURIComponent(runId);

  const displayName = (runName ?? '').trim() || runId;

  const tabClass = (tab: RunViewerTab) =>
    `px-3 py-2 border rounded text-sm transition-colors ${
      activeTab === tab ? 'bg-blue-600 text-white border-blue-600' : 'text-gray-700 hover:bg-gray-50 border-gray-200'
    }`;

  return (
    <div className="mb-8">
      <div className="flex items-center justify-between">
        <Link to={`/runs/${year}`} className="text-blue-600 hover:text-blue-800">
          ← Back to Runs
        </Link>

        <div className="flex gap-2">
          <Link to={`/runs/${year}/${encodedRunId}`} className={tabClass('rankings')}>
            Rankings
          </Link>
          <Link to={`/runs/${year}/${encodedRunId}/returns`} className={tabClass('returns')}>
            Returns
          </Link>
          <Link to={`/runs/${year}/${encodedRunId}/investments`} className={tabClass('investments')}>
            Investments
          </Link>
        </div>
      </div>

      <div className="mt-6">
        <div className="text-sm text-gray-600">{year}</div>
        <h1 className="text-3xl font-bold mt-1">{displayName}</h1>
        {displayName !== runId && <div className="text-sm text-gray-500 mt-1">{runId}</div>}
      </div>
    </div>
  );
}
