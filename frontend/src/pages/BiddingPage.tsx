import { Link } from 'react-router-dom';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { Alert } from '../components/ui/Alert';
import { ErrorState } from '../components/ui/ErrorState';
import { BiddingSkeleton } from '../components/skeletons/BiddingSkeleton';
import { Button } from '../components/ui/Button';
import { BudgetTracker } from '../components/Bidding/BudgetTracker';
import { TeamBidRow } from '../components/Bidding/TeamBidRow';
import { PortfolioSummary } from '../components/Bidding/PortfolioSummary';
import { SeedFilter } from '../components/Bidding/SeedFilter';
import { BidConfirmModal } from '../components/Bidding/BidConfirmModal';
import { Breadcrumb } from '../components/ui/Breadcrumb';
import { cn } from '../lib/cn';
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
    bidsByTeamId,
    budgetRemaining,
    teamCount,
    validationErrors,
    isValid,
    seedFilter,
    setSeedFilter,
    unbidOnly,
    setUnbidOnly,
    collapsedRegions,
    teamsByRegion,
    portfolioSummary,
    handleBidChange,
    handleSubmit,
    handleConfirm,
    matchesSeedFilter,
    toggleRegion,
    showConfirmModal,
    setShowConfirmModal,
  } = useBidding();

  const MIN_TEAMS = calcutta?.minTeams ?? 3;
  const MAX_TEAMS = calcutta?.maxTeams ?? 10;

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
        subtitle={
          <>
            {`Budget: ${BUDGET} pts | Teams: ${MIN_TEAMS}-${MAX_TEAMS} | Max per team: ${MAX_BID} pts`}
            {' '}
            <Link to="/rules" className="text-blue-600 hover:text-blue-800 underline text-sm">Learn the rules</Link>
          </>
        }
        actions={
          <div className="flex gap-2">
            <Link to={`/calcuttas/${calcuttaId}/entries/${entryId}`}>
              <Button variant="secondary">Cancel</Button>
            </Link>
            <Button onClick={handleSubmit} disabled={!isValid || updateEntryMutation.isPending} loading={updateEntryMutation.isPending} title={!isValid && validationErrors.length > 0 ? validationErrors[0] : undefined}>
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

      <PortfolioSummary portfolioSummary={portfolioSummary} />

      <SeedFilter
        seedFilter={seedFilter}
        onSeedFilterChange={setSeedFilter}
        unbidOnly={unbidOnly}
        onUnbidOnlyChange={setUnbidOnly}
      />

      <div className="bg-white shadow rounded-lg overflow-hidden">
        <div className="px-4 sm:px-6 py-4 border-b border-gray-200">
          <h2 className="text-lg font-semibold text-gray-900">Tournament Teams</h2>
          <p className="text-sm text-gray-600 mt-1">
            Select teams and enter your bid amounts. Leave bid at 0 to skip a team.
          </p>
        </div>

        <div className="px-4 sm:px-6 py-2">
          <div className="hidden sm:grid grid-cols-12 gap-4 py-2 text-sm font-medium text-gray-700 border-b-2 border-gray-300">
            <div className="col-span-5">School</div>
            <div className="col-span-2 text-center">Seed</div>
            <div className="col-span-2 text-center">Region</div>
            <div className="col-span-3">Bid Amount</div>
          </div>

          {Array.from(teamsByRegion.entries()).map(([region, teams]) => {
            const isCollapsed = collapsedRegions.has(region);
            const filteredTeams = teams.filter((team) => {
              if (!matchesSeedFilter(team.seed)) return false;
              if (unbidOnly && bidsByTeamId[team.id]) return false;
              return true;
            });

            if (filteredTeams.length === 0 && (seedFilter !== 'all' || unbidOnly)) return null;

            const regionBidCount = teams.filter((t) => bidsByTeamId[t.id]).length;

            return (
              <div key={region}>
                <button
                  type="button"
                  onClick={() => toggleRegion(region)}
                  className="w-full flex items-center justify-between py-3 px-2 bg-gray-50 hover:bg-gray-100 transition-colors border-b border-gray-200 mt-1"
                >
                  <div className="flex items-center gap-2">
                    <svg
                      className={cn('h-4 w-4 text-gray-500 transition-transform', !isCollapsed && 'rotate-90')}
                      fill="none"
                      viewBox="0 0 24 24"
                      strokeWidth="2"
                      stroke="currentColor"
                    >
                      <path strokeLinecap="round" strokeLinejoin="round" d="M8.25 4.5l7.5 7.5-7.5 7.5" />
                    </svg>
                    <span className="text-sm font-semibold text-gray-800">{region}</span>
                  </div>
                  {regionBidCount > 0 && (
                    <span className="text-xs text-blue-600 font-medium">{regionBidCount} bid{regionBidCount !== 1 ? 's' : ''}</span>
                  )}
                </button>
                {!isCollapsed && filteredTeams.map((team) => {
                  const bid = bidsByTeamId[team.id] || 0;
                  const validationError =
                    bid > 0 ? (bid < MIN_BID ? `Min ${MIN_BID} pts` : bid > MAX_BID ? `Max ${MAX_BID} pts` : undefined) : undefined;

                  return (
                    <TeamBidRow
                      key={team.id}
                      teamId={team.id}
                      schoolName={team.school?.name || 'Unknown'}
                      seed={team.seed}
                      region={team.region}
                      bidAmount={bid}
                      maxBid={MAX_BID}
                      onBidChange={handleBidChange}
                      validationError={validationError}
                    />
                  );
                })}
              </div>
            );
          })}
        </div>
      </div>

      <BidConfirmModal
        open={showConfirmModal}
        onClose={() => setShowConfirmModal(false)}
        onConfirm={handleConfirm}
        isPending={updateEntryMutation.isPending}
        portfolioSummary={portfolioSummary}
        totalBudget={BUDGET}
        budgetRemaining={budgetRemaining}
      />
    </PageContainer>
  );
}
