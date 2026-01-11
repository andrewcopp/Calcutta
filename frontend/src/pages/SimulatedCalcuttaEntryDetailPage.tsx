import React, { useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Link, useParams, useSearchParams } from 'react-router-dom';

import { ApiError } from '../api/apiClient';
import { Alert } from '../components/ui/Alert';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { LoadingState } from '../components/ui/LoadingState';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { calcuttaService } from '../services/calcuttaService';
import { simulatedCalcuttasService } from '../services/simulatedCalcuttasService';
import { simulatedEntriesService, type SimulatedEntryListItem } from '../services/simulatedEntriesService';
import { tournamentService } from '../services/tournamentService';
import type { School, TournamentTeam } from '../types/calcutta';

export function SimulatedCalcuttaEntryDetailPage() {
	const { id, entryId } = useParams<{ id: string; entryId: string }>();
	const [searchParams] = useSearchParams();
	const cohortId = searchParams.get('cohortId') || '';

	const showError = (err: unknown) => {
		if (err instanceof ApiError) {
			if (err.status === 403) return 'You do not have permission to view simulated entries (403).';
			return `Request failed (${err.status}): ${err.message}`;
		}
		return err instanceof Error ? err.message : 'Unknown error';
	};

	const formatDateTime = (v: string | null | undefined) => {
		if (!v) return '—';
		const d = new Date(v);
		if (Number.isNaN(d.getTime())) return v;
		return d.toLocaleString();
	};

	const simulatedCalcuttaQuery = useQuery({
		queryKey: ['simulated-calcuttas', 'get', id],
		queryFn: () => simulatedCalcuttasService.get(id || ''),
		enabled: Boolean(id),
	});

	const entriesQuery = useQuery({
		queryKey: ['simulated-entries', 'list', id],
		queryFn: () => simulatedEntriesService.list(id || ''),
		enabled: Boolean(id),
	});

	const tournamentId = useMemo(
		() => simulatedCalcuttaQuery.data?.simulated_calcutta.tournament_id || '',
		[simulatedCalcuttaQuery.data?.simulated_calcutta.tournament_id]
	);

	const tournamentTeamsQuery = useQuery<TournamentTeam[]>({
		queryKey: ['tournaments', tournamentId, 'teams'],
		queryFn: () => tournamentService.getTournamentTeams(tournamentId),
		enabled: Boolean(tournamentId),
	});

	const schoolsQuery = useQuery<School[]>({
		queryKey: ['schools'],
		queryFn: () => calcuttaService.getSchools(),
		enabled: Boolean(tournamentId),
	});

	const entry: SimulatedEntryListItem | null = useMemo(() => {
		const items = entriesQuery.data?.items ?? [];
		return items.find((e) => e.id === (entryId || '')) ?? null;
	}, [entriesQuery.data?.items, entryId]);

	const tournamentTeamsById = useMemo(() => {
		const m = new Map<string, TournamentTeam>();
		for (const t of tournamentTeamsQuery.data ?? []) m.set(t.id, t);
		return m;
	}, [tournamentTeamsQuery.data]);

	const schoolsById = useMemo(() => {
		const m = new Map<string, School>();
		for (const s of schoolsQuery.data ?? []) m.set(s.id, s);
		return m;
	}, [schoolsQuery.data]);

	const rows = useMemo(() => {
		const out = (entry?.teams ?? []).map((t) => {
			const tt = tournamentTeamsById.get(t.team_id);
			const school = tt ? schoolsById.get(tt.schoolId) : undefined;
			return {
				teamId: t.team_id,
				schoolName: school?.name || t.team_id,
				seed: tt?.seed ?? 0,
				region: tt?.region ?? '—',
				bidPoints: t.bid_points,
			};
		});
		out.sort((a, b) => b.bidPoints - a.bidPoints);
		return out;
	}, [entry?.teams, schoolsById, tournamentTeamsById]);

	const backUrl = useMemo(() => {
		if (!id) return '/sandbox/cohorts';
		const q = new URLSearchParams();
		if (cohortId) q.set('cohortId', cohortId);
		const suffix = q.toString() ? `?${q.toString()}` : '';
		return `/sandbox/simulated-calcuttas/${encodeURIComponent(id)}${suffix}`;
	}, [cohortId, id]);

	return (
		<PageContainer className="max-w-none">
			<PageHeader
				title={entry?.display_name || 'Simulated Entry'}
				subtitle={entryId}
				leftActions={
					<Link to={backUrl} className="text-blue-600 hover:text-blue-800">
						← Back to Simulated Calcutta
					</Link>
				}
			/>

			{!id || !entryId ? <Alert variant="error">Missing simulated calcutta entry route params.</Alert> : null}

			{id && simulatedCalcuttaQuery.isLoading ? <LoadingState label="Loading simulated calcutta..." /> : null}
			{id && simulatedCalcuttaQuery.isError ? (
				<Alert variant="error">
					<div className="font-semibold mb-1">Failed to load simulated calcutta</div>
					<div className="mb-3">{showError(simulatedCalcuttaQuery.error)}</div>
					<Button size="sm" onClick={() => simulatedCalcuttaQuery.refetch()}>
						Retry
					</Button>
				</Alert>
			) : null}

			{id && entriesQuery.isLoading ? <LoadingState label="Loading entries..." /> : null}
			{id && entriesQuery.isError ? (
				<Alert variant="error">
					<div className="font-semibold mb-1">Failed to load simulated entries</div>
					<div className="mb-3">{showError(entriesQuery.error)}</div>
					<Button size="sm" onClick={() => entriesQuery.refetch()}>
						Retry
					</Button>
				</Alert>
			) : null}

			{!entriesQuery.isLoading && !entriesQuery.isError && !entry ? (
				<Alert variant="error">Simulated entry not found for this simulated calcutta.</Alert>
			) : null}

			{entry ? (
				<div className="space-y-6">
					<Card>
						<h2 className="text-xl font-semibold mb-4">Entry</h2>
						<div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
							<div>
								<div className="text-gray-500">Display name</div>
								<div className="text-gray-900 font-medium break-words">{entry.display_name}</div>
							</div>
							<div>
								<div className="text-gray-500">Entry ID</div>
								<div className="text-gray-900 break-words">{entry.id}</div>
							</div>
							<div>
								<div className="text-gray-500">Source</div>
								<div className="text-gray-900">{entry.source_kind}</div>
							</div>
							<div>
								<div className="text-gray-500">Created</div>
								<div className="text-gray-900">{formatDateTime(entry.created_at)}</div>
							</div>
						</div>
					</Card>

					<Card>
						<h2 className="text-xl font-semibold mb-4">Teams</h2>

						{tournamentTeamsQuery.isLoading || schoolsQuery.isLoading ? (
							<LoadingState label="Loading teams..." layout="inline" />
						) : null}
						{tournamentTeamsQuery.isError ? (
							<Alert variant="error" className="mt-3">
								<div className="font-semibold mb-1">Failed to load tournament teams</div>
								<div className="mb-3">{showError(tournamentTeamsQuery.error)}</div>
								<Button size="sm" onClick={() => tournamentTeamsQuery.refetch()}>
									Retry
								</Button>
							</Alert>
						) : null}
						{schoolsQuery.isError ? (
							<Alert variant="error" className="mt-3">
								<div className="font-semibold mb-1">Failed to load schools</div>
								<div className="mb-3">{showError(schoolsQuery.error)}</div>
								<Button size="sm" onClick={() => schoolsQuery.refetch()}>
									Retry
								</Button>
							</Alert>
						) : null}

						{!tournamentTeamsQuery.isLoading && !schoolsQuery.isLoading && rows.length === 0 ? (
							<Alert variant="info">No teams found for this entry.</Alert>
						) : null}

						{rows.length > 0 ? (
							<div className="overflow-x-auto">
								<table className="min-w-full divide-y divide-gray-200">
									<thead className="bg-gray-50">
										<tr>
											<th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Team</th>
											<th className="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Seed</th>
											<th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Region</th>
											<th className="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Bid</th>
										</tr>
									</thead>
									<tbody className="bg-white divide-y divide-gray-200">
										{rows.map((r) => (
											<tr key={r.teamId} className="hover:bg-gray-50">
												<td className="px-3 py-2 text-sm text-gray-900 break-words">
													<div className="font-medium">{r.schoolName}</div>
													<div className="text-xs text-gray-500 break-words">{r.teamId}</div>
												</td>
												<td className="px-3 py-2 text-sm text-gray-700 text-right">{r.seed || '—'}</td>
												<td className="px-3 py-2 text-sm text-gray-700">{r.region}</td>
												<td className="px-3 py-2 text-sm text-gray-900 text-right font-medium">{r.bidPoints}</td>
											</tr>
										))}
									</tbody>
								</table>
							</div>
						) : null}
					</Card>
				</div>
			) : null}
		</PageContainer>
	);
}
