import { useMemo, useState } from 'react';
import { Link } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { queryKeys } from '../queryKeys';
import { hallOfFameService } from '../services/hallOfFameService';
import { Breadcrumb } from '../components/ui/Breadcrumb';
import { Card } from '../components/ui/Card';
import { DataTable } from '../components/ui/DataTable';
import { ErrorState } from '../components/ui/ErrorState';
import { LoadingState } from '../components/ui/LoadingState';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { Tabs, TabsList, TabsTrigger, TabsContent } from '../components/ui/Tabs';
import {
  bestTeamsColumns,
  bestInvestmentsColumns,
  bestEntriesColumns,
  bestCareersColumns,
} from './HallOfFame/columns';

export function HallOfFamePage() {
  const [activeTab, setActiveTab] = useState('bestTeams');
  const [hideInactiveCareers, setHideInactiveCareers] = useState<boolean>(false);

  const bestTeamsQuery = useQuery({
    queryKey: queryKeys.hallOfFame.bestTeams(200),
    queryFn: () => hallOfFameService.getBestTeams(200),
    enabled: activeTab === 'bestTeams',
  });

  const bestInvestmentsQuery = useQuery({
    queryKey: queryKeys.hallOfFame.bestInvestments(200),
    queryFn: () => hallOfFameService.getBestInvestments(200),
    enabled: activeTab === 'bestInvestments',
  });

  const bestEntriesQuery = useQuery({
    queryKey: queryKeys.hallOfFame.bestEntries(200),
    queryFn: () => hallOfFameService.getBestEntries(200),
    enabled: activeTab === 'bestEntries',
  });

  const bestCareersQuery = useQuery({
    queryKey: queryKeys.hallOfFame.bestCareers(200),
    queryFn: () => hallOfFameService.getBestCareers(200),
    enabled: activeTab === 'bestCareers',
  });

  const filteredCareers = useMemo(() => {
    if (!bestCareersQuery.data) return [];
    return hideInactiveCareers
      ? bestCareersQuery.data.careers.filter((c) => c.activeInLatestCalcutta)
      : bestCareersQuery.data.careers;
  }, [bestCareersQuery.data, hideInactiveCareers]);

  return (
    <PageContainer>
      <Breadcrumb
        items={[
          { label: 'Admin', href: '/admin' },
          { label: 'Hall of Fame' },
        ]}
      />
      <PageHeader
        title="Hall of Fame"
        subtitle="Leaderboards across all pools, adjusted for inflation so every year is comparable."
        actions={
          <Link to="/admin" className="text-blue-600 hover:text-blue-800">
            Back to Admin Console
          </Link>
        }
      />

      <Tabs value={activeTab} onValueChange={setActiveTab}>
        <TabsList>
          <TabsTrigger value="bestTeams">Best Teams</TabsTrigger>
          <TabsTrigger value="bestInvestments">Best Investments</TabsTrigger>
          <TabsTrigger value="bestEntries">Best Portfolios</TabsTrigger>
          <TabsTrigger value="bestCareers">Best Careers</TabsTrigger>
        </TabsList>

      <TabsContent value="bestTeams">
        <Card>
          <h2 className="text-xl font-semibold mb-2">Best Teams</h2>
          <p className="text-sm text-gray-600 mb-4">
            Which teams performed best relative to expectations? Performance adjusted for inflation, because a 12-seed making the Sweet 16 is more impressive than a 1-seed doing it.
          </p>

          {bestTeamsQuery.isLoading && <LoadingState label="Loading best teams..." layout="inline" />}

          {bestTeamsQuery.isError && (
            <ErrorState
              error={bestTeamsQuery.error instanceof Error ? bestTeamsQuery.error.message : 'Failed to load best teams'}
              onRetry={() => bestTeamsQuery.refetch()}
            />
          )}

          {bestTeamsQuery.data && (
            <DataTable
              columns={bestTeamsColumns}
              data={bestTeamsQuery.data.teams}
              pagination
              pageSize={25}
            />
          )}
        </Card>
      </TabsContent>

      <TabsContent value="bestInvestments">
        <Card>
          <h2 className="text-xl font-semibold mb-2">Best Investments</h2>
          <p className="text-sm text-gray-600 mb-4">
            Which individual picks returned the most? Returns adjusted for inflation so every pool is on equal footing.
          </p>

          {bestInvestmentsQuery.isLoading && <LoadingState label="Loading best investments..." layout="inline" />}

          {bestInvestmentsQuery.isError && (
            <ErrorState
              error={bestInvestmentsQuery.error instanceof Error ? bestInvestmentsQuery.error.message : 'Failed to load best investments'}
              onRetry={() => bestInvestmentsQuery.refetch()}
            />
          )}

          {bestInvestmentsQuery.data && (
            <DataTable
              columns={bestInvestmentsColumns}
              data={bestInvestmentsQuery.data.investments}
              pagination
              pageSize={25}
            />
          )}
        </Card>
      </TabsContent>

      <TabsContent value="bestEntries">
        <Card>
          <h2 className="text-xl font-semibold mb-2">Best Portfolios</h2>
          <p className="text-sm text-gray-600 mb-4">
            Which portfolios returned the most overall? Adjusted for inflation so results across years are comparable.
          </p>

          {bestEntriesQuery.isLoading && <LoadingState label="Loading best entries..." layout="inline" />}

          {bestEntriesQuery.isError && (
            <ErrorState
              error={bestEntriesQuery.error instanceof Error ? bestEntriesQuery.error.message : 'Failed to load best entries'}
              onRetry={() => bestEntriesQuery.refetch()}
            />
          )}

          {bestEntriesQuery.data && (
            <DataTable
              columns={bestEntriesColumns}
              data={bestEntriesQuery.data.entries}
              pagination
              pageSize={25}
            />
          )}
        </Card>
      </TabsContent>

      <TabsContent value="bestCareers">
        <Card>
          <h2 className="text-xl font-semibold mb-2">Best Careers</h2>
          <p className="text-sm text-gray-600 mb-4">
            Who's been the most consistent over the years? Ranked by average annual performance, with tie breaks on wins, podiums, payouts, and top-10 finishes.
          </p>

          <div className="mb-4 flex items-center gap-2">
            <input
              id="hideInactiveCareers"
              type="checkbox"
              checked={hideInactiveCareers}
              onChange={(e) => setHideInactiveCareers(e.target.checked)}
              className="h-4 w-4"
            />
            <label htmlFor="hideInactiveCareers" className="text-sm text-gray-700">
              Hide inactive careers
            </label>
          </div>

          {bestCareersQuery.isLoading && <LoadingState label="Loading best careers..." layout="inline" />}

          {bestCareersQuery.isError && (
            <ErrorState
              error={bestCareersQuery.error instanceof Error ? bestCareersQuery.error.message : 'Failed to load best careers'}
              onRetry={() => bestCareersQuery.refetch()}
            />
          )}

          {bestCareersQuery.data && (
            <DataTable
              columns={bestCareersColumns}
              data={filteredCareers}
              pagination
              pageSize={25}
            />
          )}
        </Card>
      </TabsContent>
      </Tabs>
    </PageContainer>
  );
}
