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
          { label: 'Calcuttas', href: '/calcuttas' },
          { label: calcutta?.name ?? 'Pool', href: `/calcuttas/${calcuttaId}` },
          { label: 'Bid' },
        ]}
      />

      <PageHeader
        title="Place Your Bids"
        subtitle={`Budget: ${BUDGET} pts | Teams: ${MIN_TEAMS}-${MAX_TEAMS} | Max per team: ${MAX_BID} pts`}
        actions={
          <div className="flex gap-2">
            <Link to={`/calcuttas/${calcuttaId}`}>
              <Button variant="secondary">Cancel</Button>
            </Link>
            <Button onClick={handleSubmit} disabled={!isValid || updateEntryMutation.isPending} loading={updateEntryMutation.isPending} title={!isValid && validationErrors.length > 0 ? validationErrors[0] : undefined}>
              {updateEntryMutation.isPending ? 'Submitting...' : 'Submit Entry'}
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

      <div className="bg-white shadow rounded-lg overflow-hidden">
        <div className="px-4 sm:px-6 py-4 border-b border-gray-200">
          <h2 className="text-lg font-semibold text-gray-900">Your Roster</h2>
          <p className="text-sm text-gray-600 mt-1">
            Search and select up to {MAX_TEAMS} teams, then set your bid amounts.
          </p>
        </div>

        <div className="px-4 sm:px-6 py-2 divide-y divide-gray-200">
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
