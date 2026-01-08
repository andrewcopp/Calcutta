import React, { useMemo, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Link, useNavigate, useParams } from 'react-router-dom';
import { Alert } from '../components/ui/Alert';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { Input } from '../components/ui/Input';
import { LoadingState } from '../components/ui/LoadingState';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { Select } from '../components/ui/Select';
import { calcuttaService } from '../services/calcuttaService';
import { analyticsService } from '../services/analyticsService';
import type { CalcuttaEntry } from '../types/calcutta';

type CohortDetailResponse = {
  cohort: {
    id: string;
    name: string;
    advancement_algorithm: { id: string; name: string };
    investment_algorithm: { id: string; name: string };
    optimizer_key: string;
    starting_state_key: string;
    excluded_entry_name?: string | null;
  };
  items: {
    scenario_id: string;
    calcutta_id: string;
    calcutta_name: string;
    tournament_name: string;
    season: string;
    picks: string;
    entry_created_at?: string | null;
    scenario_created_at: string;
    strategy_generation_run_id?: string | null;
  }[];
};

type CreateSandboxExecutionResponse = {
  executionId: string;
  evaluationCount: number;
};

type GenerateLabEntriesResponse = {
	created: number;
	skipped: number;
	failed: number;
	failures: { scenario_id: string; calcutta_id: string; message: string }[];
};

export function LabEntriesCohortDetailPage() {
  const { cohortId } = useParams<{ cohortId: string }>();
  const navigate = useNavigate();

	const [nSimsText, setNSimsText] = useState<string>('');
	const [excludedEntryId, setExcludedEntryId] = useState<string>('');
	const [runInSandboxError, setRunInSandboxError] = useState<string | null>(null);
	const [generateError, setGenerateError] = useState<string | null>(null);
	const [generateResult, setGenerateResult] = useState<GenerateLabEntriesResponse | null>(null);
	const [isGenerating, setIsGenerating] = useState<boolean>(false);

  const detailQuery = useQuery<CohortDetailResponse | null>({
    queryKey: ['lab', 'entries', 'cohort', cohortId],
    queryFn: async () => {
      if (!cohortId) return null;
      return analyticsService.getLabEntriesCohortDetail<CohortDetailResponse>(cohortId);
    },
    enabled: Boolean(cohortId),
  });

  const items = detailQuery.data?.items ?? [];

  const sorted = useMemo(() => {
    return items
      .slice()
      .sort((a, b) => {
        // Prefer season desc; fallback to created_at desc.
        const ay = Number(a.season);
        const by = Number(b.season);
        if (Number.isFinite(ay) && Number.isFinite(by) && by !== ay) return by - ay;
        return (b.entry_created_at ?? b.scenario_created_at).localeCompare(a.entry_created_at ?? a.scenario_created_at);
      });
  }, [items]);

	const latestCalcuttaId = sorted.length > 0 ? sorted[0].calcutta_id : '';

	const entriesQuery = useQuery<CalcuttaEntry[]>({
		queryKey: ['calcuttas', latestCalcuttaId, 'entries'],
		queryFn: async () => {
			if (!latestCalcuttaId) return [];
			return calcuttaService.getCalcuttaEntries(latestCalcuttaId);
		},
		enabled: Boolean(latestCalcuttaId),
	});

	const calcuttaEntries = entriesQuery.data ?? [];

	const canRunInSandbox = useMemo(() => {
		return items.some((it) => Boolean(it.strategy_generation_run_id));
	}, [items]);
  const cohort = detailQuery.data?.cohort ?? null;

  const fmtDateTime = (iso?: string | null) => {
    if (!iso) return '—';
    const d = new Date(iso);
    if (Number.isNaN(d.getTime())) return '—';
    return d.toLocaleString();
  };

  return (
    <PageContainer className="max-w-none">
      <PageHeader
        title="Entries"
        subtitle={cohort ? cohort.name : cohortId}
        leftActions={
          <Link to="/lab/entries" className="text-blue-600 hover:text-blue-800">
            ← Back to Entries
          </Link>
        }
        actions={
			<div className="flex items-center gap-2">
				<Button
					size="sm"
					variant="secondary"
					disabled={!cohortId || detailQuery.isLoading || isGenerating}
					onClick={async () => {
						if (!cohortId) return;
						setGenerateError(null);
						setGenerateResult(null);
						setIsGenerating(true);
						try {
							const res = await analyticsService.generateLabEntriesForCohort<GenerateLabEntriesResponse>(cohortId);
							setGenerateResult(res);
							await detailQuery.refetch();
						} catch (err) {
							setGenerateError(err instanceof Error ? err.message : 'Failed to generate entries');
						} finally {
							setIsGenerating(false);
						}
					}}
				>
					{isGenerating ? 'Generating…' : 'Generate Entries'}
				</Button>

				<div className="w-32">
					<Input
						type="number"
						min={0}
						placeholder="nSims"
						value={nSimsText}
						onChange={(e) => setNSimsText(e.target.value)}
						disabled={!cohortId || detailQuery.isLoading}
					/>
				</div>
				<div className="w-64">
					<Select
						value={excludedEntryId}
						onChange={(e) => setExcludedEntryId(e.target.value)}
						disabled={!cohortId || detailQuery.isLoading || entriesQuery.isLoading || calcuttaEntries.length === 0}
					>
						<option value="">Exclude entry (none)</option>
						{calcuttaEntries.map((e) => (
							<option key={e.id} value={e.id}>
								{e.name}
							</option>
						))}
					</Select>
				</div>
				<Button
					size="sm"
					disabled={!cohortId || detailQuery.isLoading || !canRunInSandbox}
					onClick={async () => {
						if (!cohortId) return;
						setRunInSandboxError(null);
						const trimmed = nSimsText.trim();
						let nSims: number | undefined;
						if (trimmed !== '') {
							if (!/^\d+$/.test(trimmed)) {
								setRunInSandboxError('nSims must be a non-negative integer');
								return;
							}
							nSims = Number(trimmed);
						}

						try {
							const res = await analyticsService.createLabCohortSandboxExecution<CreateSandboxExecutionResponse>(cohortId, {
								nSims,
								excludedEntryId: excludedEntryId || undefined,
							});
							navigate(`/sandbox/cohorts/${encodeURIComponent(cohortId)}?executionId=${encodeURIComponent(res.executionId)}`);
						} catch (err) {
							setRunInSandboxError(err instanceof Error ? err.message : 'Failed to run sandbox execution');
						}
					}}
				>
					Run in Sandbox
				</Button>
			</div>
        }
      />

			{runInSandboxError ? (
				<Alert variant="error" className="mb-4">
					{runInSandboxError}
				</Alert>
			) : null}

      <div className="space-y-6">
        <Card>
          <h2 className="text-xl font-semibold mb-2">Algorithm Combo</h2>
          {detailQuery.isLoading ? <LoadingState label="Loading cohort..." layout="inline" /> : null}

					{generateError ? (
						<Alert variant="error" className="mt-3">
							<div className="font-semibold mb-1">Failed to generate entries</div>
							<div>{generateError}</div>
						</Alert>
					) : null}

					{generateResult ? (
						<Alert variant={generateResult.failed > 0 ? 'warning' : 'success'} className="mt-3">
							<div className="font-semibold mb-1">Generate Entries</div>
							<div>
								created={generateResult.created} | skipped={generateResult.skipped} | failed={generateResult.failed}
							</div>
						</Alert>
					) : null}

          {cohort ? (
            <div className="text-sm text-gray-700">
              <div>
                <span className="font-medium text-gray-900">Advancement:</span> {cohort.advancement_algorithm.name}
              </div>
              <div>
                <span className="font-medium text-gray-900">Investment:</span> {cohort.investment_algorithm.name}
              </div>
              <div>
                <span className="font-medium text-gray-900">Optimizer:</span> {cohort.optimizer_key}
              </div>
              <div className="text-gray-600 mt-1">
                starting_state_key={cohort.starting_state_key}
                {cohort.excluded_entry_name ? ` | excluded_entry_name=${cohort.excluded_entry_name}` : ''}
              </div>
            </div>
          ) : null}

          {detailQuery.isError ? (
            <Alert variant="error" className="mt-3">
              <div className="font-semibold mb-1">Failed to load cohort</div>
              <div className="mb-3">{detailQuery.error instanceof Error ? detailQuery.error.message : 'An error occurred'}</div>
              <Button size="sm" onClick={() => detailQuery.refetch()}>
                Retry
              </Button>
            </Alert>
          ) : null}
        </Card>

        <Card>
          <h2 className="text-xl font-semibold mb-4">Calcuttas</h2>

          {!detailQuery.isLoading && !detailQuery.isError && sorted.length === 0 ? (
            <Alert variant="info">No calcuttas found for this algorithm combo.</Alert>
          ) : null}

          {!detailQuery.isLoading && !detailQuery.isError && sorted.length > 0 ? (
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-gray-200">
                <thead className="bg-gray-50">
                  <tr>
                    <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Season</th>
                    <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Calcutta</th>
                    <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Picks</th>
                    <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Entry Created</th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {sorted.map((row) => (
                    <tr
                      key={row.scenario_id}
                      className="hover:bg-gray-50 cursor-pointer"
                      onClick={() => navigate(`/lab/entries/scenarios/${encodeURIComponent(row.scenario_id)}`)}
                    >
                      <td className="px-3 py-2 text-sm text-gray-700">{row.season}</td>
                      <td className="px-3 py-2 text-sm text-gray-900">
                        <div className="font-medium">{row.calcutta_name}</div>
                        <div className="text-xs text-gray-600">{row.tournament_name}</div>
                      </td>
                      <td className="px-3 py-2 text-sm text-gray-700 font-mono">{row.picks}</td>
                      <td className="px-3 py-2 text-sm text-gray-700">{fmtDateTime(row.entry_created_at)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          ) : null}
        </Card>
      </div>
    </PageContainer>
  );
}
