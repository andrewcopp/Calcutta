import React, { useEffect, useMemo, useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { TabsNav } from '../components/TabsNav';
import { Card } from '../components/ui/Card';
import { PageContainer, PageHeader } from '../components/ui/Page';

import { ModelsTab } from './Lab/ModelsTab';
import { EntriesTab } from './Lab/EntriesTab';
import { EvaluationsTab } from './Lab/EvaluationsTab';

type TabType = 'models' | 'entries' | 'evaluations';

export function LabPage() {
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();

  const [activeTab, setActiveTab] = useState<TabType>(() => {
    const tab = searchParams.get('tab');
    if (tab === 'entries') return 'entries';
    if (tab === 'evaluations') return 'evaluations';
    return 'models';
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
        { id: 'models' as const, label: 'Models' },
        { id: 'entries' as const, label: 'Entries' },
        { id: 'evaluations' as const, label: 'Evaluations' },
      ] as const,
    []
  );

  return (
    <PageContainer>
      <PageHeader
        title="Lab"
        subtitle="Experiment with investment models and evaluate their performance via simulation."
      />

      <Card className="mb-6">
        <TabsNav tabs={tabs} activeTab={activeTab} onTabChange={setActiveTab} />
      </Card>

      {activeTab === 'models' ? <ModelsTab /> : null}
      {activeTab === 'entries' ? <EntriesTab /> : null}
      {activeTab === 'evaluations' ? <EvaluationsTab /> : null}
    </PageContainer>
  );
}
