import { useParams } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { calcuttaService } from '../services/calcuttaService';
import { queryKeys } from '../queryKeys';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { Breadcrumb } from '../components/ui/Breadcrumb';
import { Alert } from '../components/ui/Alert';
import { ErrorState } from '../components/ui/ErrorState';
import { SettingsSkeleton } from '../components/skeletons/SettingsSkeleton';
import { useFlashMessage } from '../hooks/useFlashMessage';
import { SettingsForm } from './CalcuttaSettings/SettingsForm';
import { PayoutsForm } from './CalcuttaSettings/PayoutsForm';

export function CalcuttaSettingsPage() {
  const { calcuttaId } = useParams<{ calcuttaId: string }>();
  const [successMessage, flash] = useFlashMessage();

  const calcuttaQuery = useQuery({
    queryKey: queryKeys.calcuttas.settings(calcuttaId),
    enabled: Boolean(calcuttaId),
    queryFn: () => calcuttaService.getCalcutta(calcuttaId!),
  });

  if (!calcuttaId) {
    return (
      <PageContainer>
        <Alert variant="error">Missing calcutta ID</Alert>
      </PageContainer>
    );
  }

  if (calcuttaQuery.isLoading) {
    return (
      <PageContainer>
        <SettingsSkeleton />
      </PageContainer>
    );
  }

  if (calcuttaQuery.isError) {
    return (
      <PageContainer>
        <ErrorState error={calcuttaQuery.error} onRetry={() => calcuttaQuery.refetch()} />
      </PageContainer>
    );
  }

  const calcutta = calcuttaQuery.data;

  if (calcutta && !calcutta.abilities?.canEditSettings) {
    return (
      <PageContainer>
        <Alert variant="error">You do not have permission to access settings.</Alert>
      </PageContainer>
    );
  }

  if (!calcutta) return null;

  return (
    <PageContainer>
      <Breadcrumb
        items={[
          { label: 'Calcuttas', href: '/calcuttas' },
          { label: calcutta.name, href: `/calcuttas/${calcuttaId}` },
          { label: 'Settings' },
        ]}
      />

      <PageHeader title="Pool Settings" />

      {successMessage && <Alert variant="success" className="mb-4">{successMessage}</Alert>}

      <SettingsForm
        calcuttaId={calcuttaId}
        calcutta={calcutta}
        onSuccess={() => flash('Settings saved successfully.')}
      />

      <PageHeader title="Payout Structure" className="mt-8" />

      <PayoutsForm
        calcuttaId={calcuttaId}
        onSuccess={() => flash('Payouts saved successfully.')}
      />
    </PageContainer>
  );
}
