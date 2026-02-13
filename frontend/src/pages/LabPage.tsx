import React, { useEffect, useMemo, useState } from 'react';
import { useSearchParams } from 'react-router-dom';

import { PipelineByModel } from './Lab/PipelineByModel';
import { PipelineByCalcutta } from './Lab/PipelineByCalcutta';
import { ModelsTab } from './Lab/ModelsTab';
import { cn } from '../lib/cn';

type TabType = 'by-model' | 'by-calcutta' | 'leaderboard';

export function LabPage() {
  const [searchParams, setSearchParams] = useSearchParams();

  const [activeTab, setActiveTab] = useState<TabType>(() => {
    const tab = searchParams.get('tab');
    if (tab === 'by-calcutta') return 'by-calcutta';
    if (tab === 'leaderboard') return 'leaderboard';
    // Handle legacy tab values
    if (tab === 'models') return 'leaderboard';
    if (tab === 'entries' || tab === 'evaluations') return 'by-model';
    return 'by-model';
  });

  useEffect(() => {
    const next = new URLSearchParams();
    next.set('tab', activeTab);

    if (next.toString() !== searchParams.toString()) {
      setSearchParams(next, { replace: true });
    }
  }, [activeTab, searchParams, setSearchParams]);

  const tabs = useMemo(
    () =>
      [
        { id: 'by-model' as const, label: 'By Model' },
        { id: 'by-calcutta' as const, label: 'By Calcutta' },
        { id: 'leaderboard' as const, label: 'Leaderboard' },
      ] as const,
    []
  );

  return (
    <div className="container mx-auto px-4 py-4">
      {/* Compact inline header */}
      <div className="flex items-center gap-6 mb-4 border-b border-gray-200 pb-3">
        <h1 className="text-xl font-bold text-gray-900">Lab</h1>
        <nav className="flex gap-1" role="tablist">
          {tabs.map((tab) => (
            <button
              key={tab.id}
              type="button"
              role="tab"
              aria-selected={activeTab === tab.id}
              onClick={() => setActiveTab(tab.id)}
              className={cn(
                'px-3 py-1.5 text-sm font-medium rounded-md transition-colors',
                activeTab === tab.id
                  ? 'bg-gray-900 text-white'
                  : 'text-gray-600 hover:text-gray-900 hover:bg-gray-100'
              )}
            >
              {tab.label}
            </button>
          ))}
        </nav>
      </div>

      {activeTab === 'by-model' ? <PipelineByModel /> : null}
      {activeTab === 'by-calcutta' ? <PipelineByCalcutta /> : null}
      {activeTab === 'leaderboard' ? <ModelsTab /> : null}
    </div>
  );
}
