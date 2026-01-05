import React from 'react';

import { Card } from '../components/ui/Card';
import { PageContainer, PageHeader } from '../components/ui/Page';

export function SandboxPage() {
  return (
    <PageContainer>
      <PageHeader title="Sandbox" subtitle="Browse historical TestSuite runs and drill into results." />

      <Card>
        <div className="text-gray-700">
          Suite browsing will appear here once suite run persistence and the results endpoints are wired.
        </div>
      </Card>
    </PageContainer>
  );
}
