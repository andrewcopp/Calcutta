import React, { useMemo, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Link, useNavigate, useSearchParams } from 'react-router-dom';

import { ApiError } from '../api/apiClient';
import { Alert } from '../components/ui/Alert';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { LoadingState } from '../components/ui/LoadingState';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { analyticsService } from '../services/analyticsService';
import { cohortsService } from '../services/cohortsService';
import { simulatedCalcuttasService } from '../services/simulatedCalcuttasService';
import { simulatedEntriesService } from '../services/simulatedEntriesService';
import { simulationRunsService } from '../services/simulationRunsService';

type CandidateListItem = {
	candidate_id: string;
	display_name: string;
	source_kind: string;
	source_entry_artifact_id?: string | null;
	calcutta_id: string;
	calcutta_name?: string | null;
	tournament_id: string;
	strategy_generation_run_id: string;
	market_share_run_id: string;
	market_share_artifact_id: string;
	advancement_run_id: string;
	optimizer_key: string;
	starting_state_key: string;
	excluded_entry_name?: string | null;
	git_sha?: string | null;
};

type ListCandidatesResponse = {
	items: CandidateListItem[];
};

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

export function LabCandidateCohortPage() {
	const navigate = useNavigate();
	const [searchParams] = useSearchParams();
	const [isEvaluating, setIsEvaluating] = useState(false);
	const [evaluateError, setEvaluateError] = useState<string | null>(null);
	const [evaluateProgress, setEvaluateProgress] = useState<{ current: number; total: number; message: string } | null>(null);

	const gameOutcomesAlgorithmId = searchParams.get('gameOutcomesAlgorithmId') || '';
	const marketShareAlgorithmId = searchParams.get('marketShareAlgorithmId') || '';
	const optimizerKey = searchParams.get('optimizerKey') || '';
	const startingStateKey = searchParams.get('startingStateKey') || 'post_first_four';
	const excludedEntryName = searchParams.get('excludedEntryName') || '';

	const showError = (err: unknown) => {
		if (err instanceof ApiError) {
			if (err.status === 403) return 'You do not have permission to view Lab candidates (403).';
			return `Request failed (${err.status}): ${err.message}`;
		}
		return err instanceof Error ? err.message : 'Unknown error';
	};

	const combosQuery = useQuery<ListCombosResponse>({
		queryKey: ['lab', 'candidates', 'combos', startingStateKey, excludedEntryName],
		queryFn: async () => {
			return analyticsService.listLabCandidateCombos<ListCombosResponse>({
				startingStateKey,
				excludedEntryName: excludedEntryName || undefined,
			});
		},
		enabled: !!gameOutcomesAlgorithmId && !!marketShareAlgorithmId && !!optimizerKey,
	});

	const combo = useMemo(() => {
		const combos = combosQuery.data?.items ?? [];
		return (
			combos.find(
				(c) =>
					c.game_outcomes_algorithm_id === gameOutcomesAlgorithmId &&
					c.market_share_algorithm_id === marketShareAlgorithmId &&
					c.optimizer_key === optimizerKey
			) ?? null
		);
	}, [combosQuery.data?.items, gameOutcomesAlgorithmId, marketShareAlgorithmId, optimizerKey]);

	const listQuery = useQuery<ListCandidatesResponse>({
		queryKey: [
			'lab',
			'candidates',
			'cohort',
			gameOutcomesAlgorithmId,
			marketShareAlgorithmId,
			optimizerKey,
			startingStateKey,
			excludedEntryName,
		],
		enabled: !!gameOutcomesAlgorithmId && !!marketShareAlgorithmId && !!optimizerKey,
		queryFn: async () => {
			return analyticsService.listLabCandidates<ListCandidatesResponse>({
				gameOutcomesAlgorithmId,
				marketShareAlgorithmId,
				optimizerKey,
				startingStateKey,
				excludedEntryName: excludedEntryName || undefined,
				limit: 200,
			});
		},
	});

	const items = useMemo(() => listQuery.data?.items ?? [], [listQuery.data?.items]);

	const handleEvaluateCohort = async () => {
		setEvaluateError(null);
		if (isEvaluating) return;
		if (!gameOutcomesAlgorithmId || !marketShareAlgorithmId || !optimizerKey) {
			setEvaluateError('Missing required cohort parameters.');
			return;
		}
		if (items.length === 0) {
			setEvaluateError('No candidates to evaluate.');
			return;
		}

		const nSimsRaw = window.prompt('Number of simulations (nSims):', '5000');
		if (nSimsRaw == null) return;
		const nSims = Number(nSimsRaw);
		if (!Number.isFinite(nSims) || nSims <= 0) {
			setEvaluateError('nSims must be a positive number.');
			return;
		}

		const excludedRaw = window.prompt('Excluded entry name (optional):', excludedEntryName || '');
		if (excludedRaw == null) return;
		const excluded = excludedRaw.trim();

		const cohortDisplayName = combo?.display_name || `${gameOutcomesAlgorithmId} - ${marketShareAlgorithmId} - ${optimizerKey}`;
		const ts = new Date().toISOString().replace(/[:.]/g, '-');
		const cohortName = `Lab: ${cohortDisplayName} (${ts})`;

		setIsEvaluating(true);
		setEvaluateProgress({ current: 0, total: items.length, message: 'Creating sandbox cohort…' });

		try {
			const cohort = await cohortsService.create({
				name: cohortName,
				description: `Autogenerated from Lab candidate cohort: ${cohortDisplayName}`,
				gameOutcomesAlgorithmId,
				marketShareAlgorithmId,
				optimizerKey,
				nSims,
				seed: 42,
				startingStateKey,
				excludedEntryName: excluded || undefined,
			});

			for (let i = 0; i < items.length; i++) {
				const c = items[i];
				setEvaluateProgress({
					current: i + 1,
					total: items.length,
					message: `Preparing ${c.display_name}…`,
				});

				const createdSim = await simulatedCalcuttasService.createFromCalcutta({
					calcuttaId: c.calcutta_id,
					name: `Lab Candidate: ${c.display_name}`,
					startingStateKey,
					excludedEntryName: excluded || undefined,
					metadata: {
						candidate_id: c.candidate_id,
						cohort_id: cohort.id,
						lab_optimizer_key: c.optimizer_key,
						lab_market_share_run_id: c.market_share_run_id,
						lab_advancement_run_id: c.advancement_run_id,
						lab_git_sha: c.git_sha ?? null,
					},
				});

				await simulatedEntriesService.importCandidate(createdSim.id, {
					candidateId: c.candidate_id,
					displayName: c.display_name,
				});

				await simulationRunsService.create({
					cohortId: cohort.id,
					req: {
						calcuttaId: '',
						simulatedCalcuttaId: createdSim.id,
						optimizerKey,
						marketShareRunId: c.market_share_run_id,
						nSims,
						seed: 42,
						startingStateKey,
						excludedEntryName: excluded || undefined,
					},
				});
			}

			navigate(`/sandbox/cohorts/${encodeURIComponent(cohort.id)}`);
		} catch (err) {
			setEvaluateError(showError(err));
		} finally {
			setIsEvaluating(false);
			setEvaluateProgress(null);
		}
	};

	const backLink = useMemo(() => {
		const q = new URLSearchParams();
		if (startingStateKey) q.set('startingStateKey', startingStateKey);
		if (excludedEntryName) q.set('excludedEntryName', excludedEntryName);
		const suffix = q.toString() ? `?${q.toString()}` : '';
		return `/lab/candidates${suffix}`;
	}, [startingStateKey, excludedEntryName]);

	return (
		<PageContainer className="max-w-none">
			<PageHeader
				title="Candidate Cohort"
				subtitle={combo?.display_name || `${gameOutcomesAlgorithmId} - ${marketShareAlgorithmId} - ${optimizerKey}`}
				leftActions={
					<Link to={backLink} className="text-blue-600 hover:text-blue-800">
						← Back to Candidates
					</Link>
				}
			/>

			{!gameOutcomesAlgorithmId || !marketShareAlgorithmId || !optimizerKey ? (
				<Alert variant="error">Missing required cohort parameters.</Alert>
			) : null}

			<Card>
				<div className="flex items-center justify-between mb-4">
					<h2 className="text-xl font-semibold">Candidates</h2>
					<div className="flex items-center gap-2">
						<Button size="sm" variant="ghost" onClick={handleEvaluateCohort} disabled={isEvaluating || listQuery.isLoading}>
							Evaluate cohort
						</Button>
						<Button size="sm" onClick={() => listQuery.refetch()} disabled={listQuery.isLoading || !listQuery.isFetched || isEvaluating}>
							Refresh
						</Button>
					</div>
				</div>

				{evaluateProgress ? (
					<Alert variant="info" className="mb-3">
						<div className="font-semibold">Evaluating cohort</div>
						<div className="text-sm">
							{evaluateProgress.message} ({evaluateProgress.current}/{evaluateProgress.total})
						</div>
					</Alert>
				) : null}
				{evaluateError ? (
					<Alert variant="error" className="mb-3">
						<div className="font-semibold">Evaluate cohort failed</div>
						<div className="text-sm">{evaluateError}</div>
					</Alert>
				) : null}

				{listQuery.isLoading ? <LoadingState label="Loading cohort candidates..." layout="inline" /> : null}
				{listQuery.isError ? (
					<Alert variant="error" className="mt-3">
						<div className="font-semibold mb-1">Failed to load cohort candidates</div>
						<div className="mb-3">{showError(listQuery.error)}</div>
					</Alert>
				) : null}
				{!listQuery.isLoading && !listQuery.isError && items.length === 0 ? (
					<Alert variant="info">No candidates found for this combo yet.</Alert>
				) : null}

				{!listQuery.isLoading && !listQuery.isError && items.length > 0 ? (
					<div className="overflow-x-auto">
						<table className="min-w-full divide-y divide-gray-200">
							<thead className="bg-gray-50">
								<tr>
									<th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Candidate</th>
									<th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Calcutta</th>
									<th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Starting</th>
								</tr>
							</thead>
							<tbody className="bg-white divide-y divide-gray-200">
								{items.map((c) => (
									<tr
										key={c.candidate_id}
										className="hover:bg-gray-50 cursor-pointer"
										onClick={() => navigate(`/lab/candidates/${encodeURIComponent(c.candidate_id)}`)}
									>
										<td className="px-3 py-2 text-sm text-gray-900">
											<div className="font-medium">{c.display_name}</div>
											<div className="text-xs text-gray-500 break-words">{c.candidate_id}</div>
										</td>
										<td className="px-3 py-2 text-sm text-gray-700 break-words">
											<div className="font-medium">{c.calcutta_name || c.calcutta_id || '—'}</div>
											{c.calcutta_name ? <div className="text-xs text-gray-500 break-words">{c.calcutta_id}</div> : null}
										</td>
										<td className="px-3 py-2 text-sm text-gray-700">{c.starting_state_key || '—'}</td>
									</tr>
								))}
							</tbody>
						</table>
					</div>
				) : null}
			</Card>
		</PageContainer>
	);
}
