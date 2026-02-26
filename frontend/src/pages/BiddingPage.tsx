import { Link } from 'react-router-dom';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { Alert } from '../components/ui/Alert';
import { ErrorState } from '../components/ui/ErrorState';
import { BiddingSkeleton } from '../components/skeletons/BiddingSkeleton';
import { Button } from '../components/ui/Button';
import { BudgetTracker } from '../components/Bidding/BudgetTracker';
import { BidSlotRow } from '../components/Bidding/BidSlotRow';
import { Breadcrumb } from '../components/ui/Breadcrumb';
import { useBidding } from '../hooks/useBidding';

export function BiddingPage() {
  const {
    calcuttaId,
    entryId,
    biddingQuery,
    updateEntryMutation,
    calcutta,
    BUDGET,
    MIN_BID,
    MAX_BID,
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
    handleSlotBidChange,
    handleSubmit,
  } = useBidding();

  if (!calcuttaId || !entryId) {
    return (
      <PageContainer>
        <Alert variant="error">Missing required parameters</Alert>
      </PageContainer>
    );
  }

  if (biddingQuery.isLoading) {
    return (
      <PageContainer>
        <BiddingSkeleton />
      </PageContainer>
    );
  }

  if (biddingQuery.isError) {
    return (
      <PageContainer>
        <ErrorState error={biddingQuery.error} onRetry={() => biddingQuery.refetch()} />
      </PageContainer>
    );
  }

  return (
    <PageContainer>
      <Breadcrumb
        items={[
          { label: 'My Pools', href: '/pools' },
          { label: calcutta?.name ?? 'Pool', href: `/pools/${calcuttaId}` },
          { label: 'Bid' },
        ]}
      />

      <PageHeader
        title="Place Your Bids"
        subtitle={`Budget: ${BUDGET} credits | Teams: ${MIN_TEAMS}-${MAX_TEAMS} | Max per team: ${MAX_BID} credits`}
        actions={
          <div className="flex gap-2">
            <Link to={`/pools/${calcuttaId}`}>
              <Button variant="secondary">Cancel</Button>
            </Link>
            <Button
              onClick={handleSubmit}
              disabled={!isValid || updateEntryMutation.isPending}
              loading={updateEntryMutation.isPending}
              title={!isValid && validationErrors.length > 0 ? validationErrors[0] : undefined}
            >
              {updateEntryMutation.isPending ? 'Saving...' : 'Save Bids'}
            </Button>
          </div>
        }
      />

      {updateEntryMutation.isError && (
        <Alert variant="error" className="mb-4">
          {updateEntryMutation.error instanceof Error ? updateEntryMutation.error.message : 'Failed to save bids'}
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
          <h2 className="text-lg font-semibold text-foreground">Your Roster</h2>
          <p className="text-sm text-muted-foreground mt-1">
            Search and select up to {MAX_TEAMS} teams, then set your bid amounts.
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
              maxBidPoints={MAX_BID}
              minBid={MIN_BID}
              onSelect={handleSlotSelect}
              onClear={handleSlotClear}
              onSearchChange={handleSlotSearchChange}
              onBidChange={handleSlotBidChange}
              isOptional={index >= MIN_TEAMS}
            />
          ))}
        </div>
      </div>
    </PageContainer>
  );
}
