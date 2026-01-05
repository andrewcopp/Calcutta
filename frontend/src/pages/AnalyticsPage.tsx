import React, { useState } from 'react';
import { Link } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { queryKeys } from '../queryKeys';
import { analyticsService } from '../services/analyticsService';
import { Alert } from '../components/ui/Alert';
import { LoadingState } from '../components/ui/LoadingState';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { AnalyticsSnapshotExportCard } from '../components/analytics/AnalyticsSnapshotExportCard';
import { AnalyticsSummaryCard } from '../components/analytics/AnalyticsSummaryCard';
import { SeedAnalyticsTab } from '../components/analytics/SeedAnalyticsTab';
import { RegionAnalyticsTab } from '../components/analytics/RegionAnalyticsTab';
import { TeamAnalyticsTab } from '../components/analytics/TeamAnalyticsTab';
import { VarianceAnalyticsTab } from '../components/analytics/VarianceAnalyticsTab';

export const AnalyticsPage: React.FC = () => {
  const [activeTab, setActiveTab] = useState<'seeds' | 'regions' | 'teams' | 'variance'>('seeds');

  const analyticsQuery = useQuery({
    queryKey: queryKeys.analytics.all(),
    staleTime: 30_000,
    queryFn: analyticsService.getAnalytics,
  });

  const seedInvestmentDistributionQuery = useQuery({
    queryKey: queryKeys.analytics.seedInvestmentDistribution(),
    staleTime: 30_000,
    queryFn: analyticsService.getSeedInvestmentDistribution,
  });

  if (analyticsQuery.isLoading) {
    return (
      <PageContainer>
        <LoadingState label="Loading analytics..." />
      </PageContainer>
    );
  }

  if (analyticsQuery.isError) {
    const message = analyticsQuery.error instanceof Error ? analyticsQuery.error.message : 'An error occurred';
    return (
      <PageContainer>
        <Alert variant="error">Error: {message}</Alert>
      </PageContainer>
    );
  }

  const analytics = analyticsQuery.data;

  if (!analytics) {
    return (
      <PageContainer>
        <Alert variant="info">No analytics data available</Alert>
      </PageContainer>
    );
  }

  const seedInvestmentDistribution = seedInvestmentDistributionQuery.data;

  return (
    <PageContainer>
      <PageHeader
        title="Calcutta Analytics"
        subtitle="Historical analysis across all calcuttas to identify trends and patterns"
        actions={
          <Link to="/admin" className="text-blue-600 hover:text-blue-800">
            ← Back to Admin Console
          </Link>
        }
      />

      <AnalyticsSnapshotExportCard />
      <AnalyticsSummaryCard analytics={analytics} />

      <div className="mb-6">
        <div className="border-b border-gray-200">
          <nav className="-mb-px flex space-x-8">
            <button
              onClick={() => setActiveTab('seeds')}
              className={`${
                activeTab === 'seeds'
                  ? 'border-blue-500 text-blue-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              } whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm`}
            >
              Seed Analytics
            </button>
            <button
              onClick={() => setActiveTab('regions')}
              className={`${
                activeTab === 'regions'
                  ? 'border-blue-500 text-blue-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              } whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm`}
            >
              Region Analytics
            </button>
            <button
              onClick={() => setActiveTab('teams')}
              className={`${
                activeTab === 'teams'
                  ? 'border-blue-500 text-blue-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              } whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm`}
            >
              Team Analytics
            </button>
            <button
              onClick={() => setActiveTab('variance')}
              className={`${
                activeTab === 'variance'
                  ? 'border-blue-500 text-blue-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              } whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm`}
            >
              Variance Analysis
            </button>
          </nav>
        </div>
      </div>

      {activeTab === 'seeds' && <SeedAnalyticsTab analytics={analytics} />}

      {activeTab === 'regions' && <RegionAnalyticsTab analytics={analytics} />}

      {activeTab === 'teams' && <TeamAnalyticsTab analytics={analytics} />}

      {activeTab === 'variance' && (
        <VarianceAnalyticsTab analytics={analytics} seedInvestmentDistribution={seedInvestmentDistribution} />
      )}
    </PageContainer>
  );
};

export default AnalyticsPage;
