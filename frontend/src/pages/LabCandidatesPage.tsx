import React, { useMemo, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Link, useNavigate, useSearchParams } from 'react-router-dom';

import { ApiError } from '../api/apiClient';
import { TabsNav } from '../components/TabsNav';
import { Alert } from '../components/ui/Alert';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { LoadingState } from '../components/ui/LoadingState';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { Input } from '../components/ui/Input';
import { analyticsService } from '../services/analyticsService';
import { calcuttaService } from '../services/calcuttaService';

type ComboItem = {
	game_outcomes_algorithm_id: string;
	game_outcomes_name: string;
	market_share_algorithm_id: string;
	market_share_name: string;
	optimizer_key: string;
	optimizer: {
		id: string;
		display_name: string;
		schema_version?: string;
		deprecated?: boolean;
	};
	display_name: string;
	existing_candidates: number;
	total_calcuttas: number;
};

type ListCombosResponse = {
	items: ComboItem[];
	count: number;
};

type GenerateResponse = {
	total_calcuttas: number;
	eligible_calcuttas: number;
	created_candidates: number;
	skipped_existing: number;
	skipped_missing_upstream: number;
};

export function LabCandidatesPage() {
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();

	const excludedEntryName = searchParams.get('excludedEntryName') || '';
	const startingStateKey = searchParams.get('startingStateKey') || 'post_first_four';

	const [runBusyKey, setRunBusyKey] = useState<string | null>(null);
	const [runError, setRunError] = useState<string | null>(null);
	const [runResult, setRunResult] = useState<GenerateResponse | null>(null);

	const combosQuery = useQuery<ListCombosResponse>({
		queryKey: ['lab', 'candidates', 'combos', startingStateKey, excludedEntryName],
		queryFn: async () => {
			return analyticsService.listLabCandidateCombos<ListCombosResponse>({
				startingStateKey,
				excludedEntryName: excludedEntryName || undefined,
			});
		},
	});

	const calcuttasQuery = useQuery({
		queryKey: ['calcuttas', 'all'],
		queryFn: async () => {
			return calcuttaService.getAllCalcuttas();
		},
	});

  const showError = (err: unknown) => {
    if (err instanceof ApiError) {
      if (err.status === 403) return 'You do not have permission to view Lab candidates (403).';
      return `Request failed (${err.status}): ${err.message}`;
    }
    return err instanceof Error ? err.message : 'Unknown error';
  };

	const combos = useMemo(() => combosQuery.data?.items ?? [], [combosQuery.data?.items]);

	const runCombo = async (combo: ComboItem) => {
		setRunError(null);
		setRunResult(null);
		const key = `${combo.game_outcomes_algorithm_id}|${combo.market_share_algorithm_id}|${combo.optimizer_key}`;
		setRunBusyKey(key);
		try {
			const res = await analyticsService.generateLabCandidates<GenerateResponse>({
				gameOutcomesAlgorithmId: combo.game_outcomes_algorithm_id,
				marketShareAlgorithmId: combo.market_share_algorithm_id,
				optimizerKey: combo.optimizer_key,
				startingStateKey,
				excludedEntryName: excludedEntryName || undefined,
				displayName: combo.display_name,
			});
			setRunResult(res);
			await combosQuery.refetch();
		} catch (e) {
			const msg = e instanceof Error ? e.message : 'Failed to generate candidates';
			setRunError(msg);
		} finally {
			setRunBusyKey(null);
		}
	};

	const openCohort = (combo: ComboItem) => {
		const q = new URLSearchParams();
		q.set('gameOutcomesAlgorithmId', combo.game_outcomes_algorithm_id);
		q.set('marketShareAlgorithmId', combo.market_share_algorithm_id);
		q.set('optimizerKey', combo.optimizer_key);
		q.set('startingStateKey', startingStateKey);
		if (excludedEntryName) q.set('excludedEntryName', excludedEntryName);
		navigate(`/lab/candidates/cohorts?${q.toString()}`);
	};

	const tabs = useMemo(
		() =>
			[
				{ id: 'advancements' as const, label: 'Advancements' },
				{ id: 'investments' as const, label: 'Investments' },
				{ id: 'candidates' as const, label: 'Candidates' },
			] as const,
		[]
	);

  return (
    <PageContainer className="max-w-none">
      <PageHeader
        title="Candidates"
        subtitle="Candidate combos (algorithm choices) and cohort generation."
        leftActions={
          <Link to="/lab" className="text-blue-600 hover:text-blue-800">
            ← Back to Lab
          </Link>
        }
      />

		<Card className="mb-6">
			<TabsNav
				tabs={tabs}
				activeTab={'candidates'}
				onTabChange={(tab) => {
					if (tab === 'advancements') navigate('/lab?tab=advancements');
					else if (tab === 'investments') navigate('/lab?tab=investments');
					else navigate('/lab/candidates');
				}}
			/>
		</Card>

		<Card className="mb-6">
			<div className="flex items-center justify-between mb-4">
				<h2 className="text-xl font-semibold">Candidate Combos</h2>
				<Button size="sm" onClick={() => combosQuery.refetch()} disabled={combosQuery.isLoading}>
					Refresh
				</Button>
			</div>

			<div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-4">
				<div>
					<div className="text-xs text-gray-500 mb-1">startingStateKey</div>
					<Input
						value={startingStateKey}
						placeholder="post_first_four"
						onChange={(e) => {
							const next = new URLSearchParams(searchParams);
							const v = e.target.value;
							if (v) next.set('startingStateKey', v);
							else next.delete('startingStateKey');
							setSearchParams(next, { replace: true });
						}}
					/>
				</div>
				<div>
					<div className="text-xs text-gray-500 mb-1">excludedEntryName (optional)</div>
					<Input
						value={excludedEntryName}
						placeholder="(blank to use whatever the market_share_run used)"
						onChange={(e) => {
							const next = new URLSearchParams(searchParams);
							const v = e.target.value;
							if (v) next.set('excludedEntryName', v);
							else next.delete('excludedEntryName');
							setSearchParams(next, { replace: true });
						}}
					/>
				</div>
			</div>

			{combosQuery.isLoading ? <LoadingState label="Loading combos..." layout="inline" /> : null}
			{combosQuery.isError ? <Alert variant="error">Failed to load combos</Alert> : null}
			{!combosQuery.isLoading && !combosQuery.isError && combos.length === 0 ? (
				<Alert variant="info">No combos available.</Alert>
			) : null}

			{!combosQuery.isLoading && !combosQuery.isError && combos.length > 0 ? (
				<div className="overflow-x-auto">
					<table className="min-w-full divide-y divide-gray-200">
						<thead className="bg-gray-50">
							<tr>
								<th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Combo</th>
								<th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Coverage</th>
								<th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Actions</th>
							</tr>
						</thead>
						<tbody className="bg-white divide-y divide-gray-200">
							{combos.map((c) => {
								const key = `${c.game_outcomes_algorithm_id}|${c.market_share_algorithm_id}|${c.optimizer_key}`;
								const apiTotal = c.total_calcuttas ?? 0;
								const fallbackTotal = (calcuttasQuery.data?.length ?? runResult?.total_calcuttas ?? 0) || 0;
								const total = apiTotal > 0 ? apiTotal : fallbackTotal;
								const coverage = `${(c.existing_candidates ?? 0) >= total ? total : 0}/${total}`;
								return (
									<tr
										key={key}
										className="hover:bg-gray-50 cursor-pointer"
										onClick={() => openCohort(c)}
									>
										<td className="px-3 py-2 text-sm text-gray-900">
											<div className="font-medium">{c.display_name}</div>
											<div className="text-xs text-gray-500 break-words">
												{c.game_outcomes_name} / {c.market_share_name} / {c.optimizer?.display_name || c.optimizer_key}
											</div>
										</td>
										<td className="px-3 py-2 text-sm text-gray-700">
											<div className="font-medium">{coverage}</div>
										</td>
										<td className="px-3 py-2 text-sm text-gray-700">
											<div className="flex gap-2">
												<Button
													size="sm"
													onClick={(e) => {
													e.stopPropagation();
													runCombo(c);
												}}
													loading={runBusyKey === key}
													disabled={!!runBusyKey}
												>
													Run
												</Button>
												<Button
													size="sm"
													onClick={(e) => {
													e.stopPropagation();
													openCohort(c);
												}}
												>
													View cohort
												</Button>
											</div>
										</td>
									</tr>
								);
							})}
						</tbody>
					</table>
				</div>
			) : null}

			{runError ? (
				<Alert variant="error" className="mt-4">
					{runError}
				</Alert>
			) : null}
			{runResult ? (
				<Alert variant="info" className="mt-4">
					created_candidates={runResult.created_candidates} eligible_calcuttas={runResult.eligible_calcuttas} skipped_existing={runResult.skipped_existing} skipped_missing_upstream={runResult.skipped_missing_upstream}
				</Alert>
			) : null}
		</Card>
    </PageContainer>
  );
}
