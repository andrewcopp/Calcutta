import React from 'react';
import { useNavigate } from 'react-router-dom';

import { Alert } from '../components/ui/Alert';
import { Button } from '../components/ui/Button';
import { PageContainer, PageHeader } from '../components/ui/Page';

export function SandboxPage() {
  const navigate = useNavigate();
  return (
    <PageContainer className="max-w-none">
      <PageHeader title="Sandbox" subtitle="This page has been replaced by Cohorts" />

      <Alert variant="info">
        <div className="font-semibold mb-1">Sandbox has moved</div>
        <div className="mb-3">Use the cohort-based sandbox at /sandbox/cohorts.</div>
        <Button size="sm" onClick={() => navigate('/sandbox/cohorts')}>
          Go to Cohorts
        </Button>
      </Alert>

    </PageContainer>
  );
}
