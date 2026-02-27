import { PageContainer, PageHeader } from '../components/ui/Page';
import { Alert } from '../components/ui/Alert';
import { ErrorState } from '../components/ui/ErrorState';
import { BiddingSkeleton } from '../components/skeletons/BiddingSkeleton';
import { Button } from '../components/ui/Button';
import { BudgetTracker } from '../components/Bidding/BudgetTracker';
import { BidSlotRow } from '../components/Bidding/BidSlotRow';
import { Breadcrumb } from '../components/ui/Breadcrumb';
import { useInvesting } from '../hooks/useInvesting';

export function InvestingPage() {
  const {
    poolId,
    isCreating,
    investingQuery,
    activeMutation,
    pool,
    BUDGET,
    MIN_INVESTMENT,
    MAX_INVESTMENT,
    MIN_TEAMS,
    MAX_TEAMS,
    slots,
    teamOptions,
    usedTeamIds,
    teams,
    budgetRemaining,
    teamCount,
    validationErrors,
    isValid,
    handleSlotSelect,
    handleSlotClear,
    handleSlotSearchChange,
    handleSlotInvestmentChange,
    handleSubmit,
    handleCancel,
  } = useInvesting();

  if (!poolId) {
    return (
      <PageContainer>
        <Alert variant="error">Missing required parameters</Alert>
      </PageContainer>
    );
  }

  if (investingQuery.isLoading) {
    return (
      <PageContainer>
        <BiddingSkeleton />
      </PageContainer>
    );
  }

  if (investingQuery.isError) {
    return (
      <PageContainer>
        <ErrorState error={investingQuery.error} onRetry={() => investingQuery.refetch()} />
      </PageContainer>
    );
  }

  return (
    <PageContainer>
      <Breadcrumb
        items={[
          { label: 'My Pools', href: '/pools' },
          { label: pool?.name ?? 'Pool', href: `/pools/${poolId}` },
          { label: 'Invest' },
        ]}
      />

      <PageHeader
        title="Place Your Investments"
        subtitle={`Budget: ${BUDGET} credits | Teams: ${MIN_TEAMS}-${MAX_TEAMS} | Max per team: ${MAX_INVESTMENT} credits`}
        actions={
          <div className="flex gap-2">
            <Button variant="secondary" onClick={handleCancel}>Cancel</Button>
            <Button
              onClick={handleSubmit}
              disabled={!isValid || activeMutation.isPending}
              loading={activeMutation.isPending}
              title={!isValid && validationErrors.length > 0 ? validationErrors[0] : undefined}
            >
              {activeMutation.isPending ? 'Saving...' : isCreating ? 'Create Portfolio' : 'Save Investments'}
            </Button>
          </div>
        }
      />

      {activeMutation.isError && (
        <Alert variant="error" className="mb-4">
          {activeMutation.error instanceof Error ? activeMutation.error.message : 'Failed to save investments'}
        </Alert>
      )}

      <BudgetTracker
        budgetRemaining={budgetRemaining}
        totalBudget={BUDGET}
        teamCount={teamCount}
        minTeams={MIN_TEAMS}
        maxTeams={MAX_TEAMS}
        isValid={isValid}
        validationErrors={validationErrors}
      />

      <div className="bg-card shadow-sm rounded-lg border border-border overflow-hidden">
        <div className="px-4 sm:px-6 py-4 border-b border-border">
          <h2 className="text-lg font-semibold text-foreground">Your Portfolio</h2>
          <p className="text-sm text-muted-foreground mt-1">
            Search and select up to {MAX_TEAMS} teams, then set your investment amounts.
          </p>
        </div>

        <div className="px-4 sm:px-6 py-3 divide-y divide-gray-100">
          {slots.map((slot, index) => (
            <BidSlotRow
              key={index}
              slotIndex={index}
              slot={slot}
              teamOptions={teamOptions}
              usedTeamIds={usedTeamIds}
              teams={teams}
              maxBidPoints={MAX_INVESTMENT}
              minBid={MIN_INVESTMENT}
              onSelect={handleSlotSelect}
              onClear={handleSlotClear}
              onSearchChange={handleSlotSearchChange}
              onBidChange={handleSlotInvestmentChange}
              isOptional={index >= MIN_TEAMS}
            />
          ))}
        </div>
      </div>
    </PageContainer>
  );
}
