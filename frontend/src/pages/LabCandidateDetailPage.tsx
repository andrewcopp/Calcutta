import React, { useMemo, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Link, useParams, useSearchParams } from 'react-router-dom';

import { ApiError } from '../api/apiClient';
import { Alert } from '../components/ui/Alert';
import { Breadcrumb } from '../components/ui/Breadcrumb';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { LoadingState } from '../components/ui/LoadingState';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { analyticsService } from '../services/analyticsService';
import { calcuttaService } from '../services/calcuttaService';
import { tournamentService } from '../services/tournamentService';
import type { Calcutta, School, Tournament, TournamentTeam } from '../types/calcutta';

type CandidateTeam = {
  team_id: string;
  bid_points: number;
};

type CandidateDetailResponse = {
  candidate_id: string;
  display_name: string;
  source_kind: string;
  source_entry_artifact_id?: string | null;
  calcutta_id: string;
  tournament_id: string;
  strategy_generation_run_id: string;
  market_share_run_id: string;
  market_share_artifact_id: string;
  advancement_run_id: string;
  optimizer_key: string;
  starting_state_key: string;
  excluded_entry_name?: string | null;
  git_sha?: string | null;
  teams: CandidateTeam[];
};

type TeamPredictedReturns = {
	team_id: string;
	expected_value: number;
};

type TeamPredictedMarketShare = {
	team_id: string;
	predicted_share: number;
};

type SortKey =
	| 'schoolName'
	| 'seed'
	| 'region'
	| 'predictedPoints'
	| 'predictedInvestment'
	| 'predictedROI'
	| 'ourBid'
	| 'ourROI';

export function LabCandidateDetailPage() {
  const { candidateId } = useParams<{ candidateId: string }>();
	const [searchParams] = useSearchParams();

	// Read filter context from URL params
	const gameOutcomesAlgorithmId = searchParams.get('gameOutcomesAlgorithmId') || '';
	const marketShareAlgorithmId = searchParams.get('marketShareAlgorithmId') || '';
	const optimizerKey = searchParams.get('optimizerKey') || '';
	const startingStateKey = searchParams.get('startingStateKey') || '';
	const excludedEntryName = searchParams.get('excludedEntryName') || '';

	// Build back links preserving filter context
	const candidatesBackLink = useMemo(() => {
		const params = new URLSearchParams();
		if (startingStateKey) params.set('startingStateKey', startingStateKey);
		if (excludedEntryName) params.set('excludedEntryName', excludedEntryName);
		const suffix = params.toString() ? `?${params.toString()}` : '';
		return `/lab/candidates${suffix}`;
	}, [startingStateKey, excludedEntryName]);

	const cohortBackLink = useMemo(() => {
		if (!gameOutcomesAlgorithmId || !marketShareAlgorithmId || !optimizerKey) {
			return null; // No cohort context available
		}
		const params = new URLSearchParams();
		params.set('gameOutcomesAlgorithmId', gameOutcomesAlgorithmId);
		params.set('marketShareAlgorithmId', marketShareAlgorithmId);
		params.set('optimizerKey', optimizerKey);
		if (startingStateKey) params.set('startingStateKey', startingStateKey);
		if (excludedEntryName) params.set('excludedEntryName', excludedEntryName);
		return `/lab/candidates/cohorts?${params.toString()}`;
	}, [gameOutcomesAlgorithmId, marketShareAlgorithmId, optimizerKey, startingStateKey, excludedEntryName]);

	const [sortKey, setSortKey] = useState<SortKey>('predictedROI');
	const [sortDir, setSortDir] = useState<'asc' | 'desc'>('desc');

  const showError = (err: unknown) => {
    if (err instanceof ApiError) {
      if (err.status === 403) return 'You do not have permission to view Lab candidates (403).';
      return `Request failed (${err.status}): ${err.message}`;
    }
    return err instanceof Error ? err.message : 'Unknown error';
  };

  const candidateQuery = useQuery<CandidateDetailResponse | null>({
    queryKey: ['lab', 'candidates', 'get', candidateId],
    queryFn: async () => {
      if (!candidateId) return null;
      return analyticsService.getLabCandidate<CandidateDetailResponse>(candidateId);
    },
    enabled: Boolean(candidateId),
  });

  const teams = useMemo(() => candidateQuery.data?.teams ?? [], [candidateQuery.data?.teams]);

  const calcuttaId = useMemo(() => candidateQuery.data?.calcutta_id || '', [candidateQuery.data?.calcutta_id]);
  const gameOutcomeRunId = useMemo(() => candidateQuery.data?.advancement_run_id || '', [candidateQuery.data?.advancement_run_id]);
  const marketShareRunId = useMemo(() => candidateQuery.data?.market_share_run_id || '', [candidateQuery.data?.market_share_run_id]);

  const tournamentId = useMemo(() => candidateQuery.data?.tournament_id || '', [candidateQuery.data?.tournament_id]);

  const calcuttaQuery = useQuery<Calcutta | null>({
    queryKey: ['calcuttas', 'get', calcuttaId],
    queryFn: async () => {
      if (!calcuttaId) return null;
      return calcuttaService.getCalcutta(calcuttaId);
    },
    enabled: Boolean(calcuttaId),
  });

  const tournamentQuery = useQuery<Tournament | null>({
    queryKey: ['tournaments', 'get', tournamentId],
    queryFn: async () => {
      if (!tournamentId) return null;
      return tournamentService.getTournament(tournamentId);
    },
    enabled: Boolean(tournamentId),
  });

  const tournamentTeamsQuery = useQuery<TournamentTeam[]>({
    queryKey: ['tournaments', tournamentId, 'teams'],
    queryFn: async () => tournamentService.getTournamentTeams(tournamentId),
    enabled: Boolean(tournamentId),
  });

  const schoolsQuery = useQuery<School[]>({
    queryKey: ['schools'],
    queryFn: async () => calcuttaService.getSchools(),
    enabled: Boolean(tournamentId),
  });

  const tournamentTeamsById = useMemo(() => {
    const map = new Map<string, TournamentTeam>();
    for (const t of tournamentTeamsQuery.data ?? []) {
      map.set(t.id, t);
    }
    return map;
  }, [tournamentTeamsQuery.data]);

  const schoolsById = useMemo(() => {
    const map = new Map<string, School>();
    for (const s of schoolsQuery.data ?? []) {
      map.set(s.id, s);
    }
    return map;
  }, [schoolsQuery.data]);

  const predictedReturnsQuery = useQuery<{ teams: TeamPredictedReturns[] } | null>({
    queryKey: ['analytics', 'predicted-returns', calcuttaId, gameOutcomeRunId],
    queryFn: async () => {
      if (!calcuttaId || !gameOutcomeRunId) return null;
      return analyticsService.getCalcuttaPredictedReturns<{ teams: TeamPredictedReturns[] }>({
        calcuttaId,
        gameOutcomeRunId,
      });
    },
    enabled: Boolean(calcuttaId) && Boolean(gameOutcomeRunId),
  });

  const predictedMarketShareQuery = useQuery<{ teams: TeamPredictedMarketShare[] } | null>({
    queryKey: ['analytics', 'predicted-market-share', calcuttaId, marketShareRunId, gameOutcomeRunId],
    queryFn: async () => {
      if (!calcuttaId || !marketShareRunId) return null;
      return analyticsService.getCalcuttaPredictedMarketShare<{ teams: TeamPredictedMarketShare[] }>({
        calcuttaId,
        marketShareRunId,
        gameOutcomeRunId: gameOutcomeRunId || undefined,
      });
    },
    enabled: Boolean(calcuttaId) && Boolean(marketShareRunId),
  });

  const ourBidByTeamId = useMemo(() => {
    const map = new Map<string, number>();
    for (const t of teams) {
      map.set(t.team_id, t.bid_points);
    }
    return map;
  }, [teams]);

  const predictedPointsByTeamId = useMemo(() => {
    const map = new Map<string, number>();
    for (const t of predictedReturnsQuery.data?.teams ?? []) {
      map.set(t.team_id, t.expected_value);
    }
    return map;
  }, [predictedReturnsQuery.data?.teams]);

  const predictedShareByTeamId = useMemo(() => {
    const map = new Map<string, number>();
    for (const t of predictedMarketShareQuery.data?.teams ?? []) {
      map.set(t.team_id, t.predicted_share);
    }
    return map;
  }, [predictedMarketShareQuery.data?.teams]);

	const baseRows = useMemo(() => {
    const pointsPerEntry = 100;
    const nEntries = 47;

    const base = (tournamentTeamsQuery.data ?? []).map((tt) => {
      const school = schoolsById.get(tt.schoolId);
      const predictedPoints = predictedPointsByTeamId.get(tt.id) ?? 0;
      const predictedShare = predictedShareByTeamId.get(tt.id) ?? 0;
      const predictedInvestment = predictedShare * pointsPerEntry * nEntries;
      const predictedROI = predictedPoints / (predictedInvestment + 1);
      const ourBid = ourBidByTeamId.get(tt.id) ?? 0;
      const ourROI = predictedPoints / (predictedInvestment + ourBid);

      return {
        teamId: tt.id,
        schoolName: school?.name || tt.id,
        seed: tt.seed,
        region: tt.region,
        predictedPoints,
        predictedInvestment,
        predictedROI,
        ourBid,
        ourROI,
      };
    });

    return base;
  }, [tournamentTeamsQuery.data, schoolsById, predictedPointsByTeamId, predictedShareByTeamId, ourBidByTeamId]);

	const rows = useMemo(() => {
		const out = [...baseRows];

		const dir = sortDir === 'asc' ? 1 : -1;
		const cmpNum = (a: number, b: number) => {
			const aa = Number.isFinite(a) ? a : -Infinity;
			const bb = Number.isFinite(b) ? b : -Infinity;
			if (aa < bb) return -1 * dir;
			if (aa > bb) return 1 * dir;
			return 0;
		};
		const cmpStr = (a: string, b: string) => {
			const aa = (a || '').toLowerCase();
			const bb = (b || '').toLowerCase();
			if (aa < bb) return -1 * dir;
			if (aa > bb) return 1 * dir;
			return 0;
		};

		out.sort((a, b) => {
			switch (sortKey) {
				case 'schoolName':
					return cmpStr(a.schoolName, b.schoolName);
				case 'region':
					return cmpStr(a.region, b.region);
				case 'seed':
					return cmpNum(a.seed, b.seed);
				case 'predictedPoints':
					return cmpNum(a.predictedPoints, b.predictedPoints);
				case 'predictedInvestment':
					return cmpNum(a.predictedInvestment, b.predictedInvestment);
				case 'predictedROI':
					return cmpNum(a.predictedROI, b.predictedROI);
				case 'ourBid':
					return cmpNum(a.ourBid, b.ourBid);
				case 'ourROI':
					return cmpNum(a.ourROI, b.ourROI);
				default:
					return 0;
			}
		});

		return out;
	}, [baseRows, sortDir, sortKey]);

	const defaultSortDir = (k: SortKey): 'asc' | 'desc' => {
		if (k === 'schoolName' || k === 'region') return 'asc';
		if (k === 'seed') return 'asc';
		return 'desc';
	};

	const setSort = (k: SortKey) => {
		if (k === sortKey) {
			setSortDir((d) => (d === 'asc' ? 'desc' : 'asc'));
			return;
		}
		setSortKey(k);
		setSortDir(defaultSortDir(k));
	};

	const sortLabel = (k: SortKey) => {
		if (sortKey !== k) return '';
		return sortDir === 'asc' ? ' ↑' : ' ↓';
	};

  const formatPoints = (n: number) => {
    if (!Number.isFinite(n)) return '—';
    return n.toFixed(1);
  };

  const formatROI = (n: number) => {
    if (!Number.isFinite(n)) return '—';
    return n.toFixed(2);
  };

  return (
    <PageContainer className="max-w-none">
      <Breadcrumb
        items={[
          { label: 'Lab', href: '/lab' },
          { label: 'Candidates', href: candidatesBackLink },
          ...(cohortBackLink ? [{ label: 'Cohort', href: cohortBackLink }] : []),
          { label: candidateQuery.data?.display_name || 'Candidate' },
        ]}
      />
      <PageHeader
        title="Candidate"
        subtitle={candidateQuery.data?.display_name || candidateId}
      />

      {!candidateId ? <Alert variant="error">Missing candidateId.</Alert> : null}

      {candidateId && candidateQuery.isLoading ? <LoadingState label="Loading candidate..." /> : null}

      {candidateId && candidateQuery.isError ? (
        <Alert variant="error">
          <div className="font-semibold mb-1">Failed to load candidate</div>
          <div className="mb-3">{showError(candidateQuery.error)}</div>
          <Button size="sm" onClick={() => candidateQuery.refetch()}>
            Retry
          </Button>
        </Alert>
      ) : null}

      {candidateQuery.data ? (
        <div className="space-y-6">
          <Card>
            <h2 className="text-xl font-semibold mb-4">Provenance</h2>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
              <div>
                <div className="text-gray-500">Display name</div>
                <div className="text-gray-900 font-medium break-words">{candidateQuery.data.display_name}</div>
              </div>
              <div>
                <div className="text-gray-500">Candidate ID</div>
                <div className="text-gray-900 break-words">{candidateQuery.data.candidate_id}</div>
              </div>
              <div>
                <div className="text-gray-500">Calcutta</div>
					<div className="text-gray-900 break-words">
						{calcuttaQuery.data ? (
							<>
								<div className="font-medium break-words">{calcuttaQuery.data.name}</div>
								<div className="text-xs text-gray-500 break-words">{calcuttaQuery.data.id}</div>
							</>
						) : (
							candidateQuery.data.calcutta_id || '—'
						)}
					</div>
              </div>
              <div>
                <div className="text-gray-500">Tournament</div>
					<div className="text-gray-900 break-words">
						{tournamentQuery.data ? (
							<>
								<div className="font-medium break-words">{tournamentQuery.data.name}</div>
								<div className="text-xs text-gray-500 break-words">{tournamentQuery.data.id}</div>
							</>
						) : (
							candidateQuery.data.tournament_id || '—'
						)}
					</div>
              </div>
              <div>
                <div className="text-gray-500">Optimizer</div>
                <div className="text-gray-900">{candidateQuery.data.optimizer_key}</div>
              </div>
              <div>
                <div className="text-gray-500">Starting state</div>
                <div className="text-gray-900">{candidateQuery.data.starting_state_key}</div>
              </div>
              <div>
                <div className="text-gray-500">Excluded entry</div>
                <div className="text-gray-900 break-words">{candidateQuery.data.excluded_entry_name || '—'}</div>
              </div>
              <div>
                <div className="text-gray-500">Strategy generation run</div>
                <div className="text-gray-900 break-words">
                  {candidateQuery.data.strategy_generation_run_id ? (
                    <Link
                      to={`/lab/entry-runs/${encodeURIComponent(candidateQuery.data.strategy_generation_run_id)}`}
                      className="text-blue-600 hover:text-blue-800"
                    >
                      {candidateQuery.data.strategy_generation_run_id}
                    </Link>
                  ) : (
                    '—'
                  )}
                </div>
              </div>
              <div>
                <div className="text-gray-500">Metrics artifact</div>
                <div className="text-gray-900 break-words">
                  {candidateQuery.data.source_entry_artifact_id ? (
                    <Link
                      to={`/lab/entry-artifacts/${encodeURIComponent(candidateQuery.data.source_entry_artifact_id)}`}
                      className="text-blue-600 hover:text-blue-800"
                    >
                      {candidateQuery.data.source_entry_artifact_id}
                    </Link>
                  ) : (
                    '—'
                  )}
                </div>
              </div>
            </div>
          </Card>

          <Card>
            <h2 className="text-xl font-semibold mb-4">Bids</h2>
            {!tournamentId ? <Alert variant="info">Tournament teams not available yet.</Alert> : null}
            {tournamentTeamsQuery.isLoading || schoolsQuery.isLoading ? (
              <LoadingState label="Loading tournament teams..." layout="inline" />
            ) : null}
            {predictedReturnsQuery.isLoading || predictedMarketShareQuery.isLoading ? (
              <LoadingState label="Loading predictions..." layout="inline" />
            ) : null}
            {!tournamentTeamsQuery.isLoading && !tournamentTeamsQuery.isError && rows.length === 0 ? (
              <Alert variant="info">No tournament teams found.</Alert>
            ) : null}
            {rows.length > 0 ? (
              <div className="overflow-x-auto">
                <table className="min-w-full divide-y divide-gray-200">
                  <thead className="bg-gray-50">
                    <tr>
						<th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
							<button type="button" className="hover:text-gray-900" onClick={() => setSort('schoolName')}>
								School{sortLabel('schoolName')}
							</button>
						</th>
						<th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
							<button type="button" className="hover:text-gray-900" onClick={() => setSort('seed')}>
								Seed{sortLabel('seed')}
							</button>
						</th>
						<th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
							<button type="button" className="hover:text-gray-900" onClick={() => setSort('region')}>
								Region{sortLabel('region')}
							</button>
						</th>
						<th className="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
							<button type="button" className="hover:text-gray-900" onClick={() => setSort('predictedPoints')}>
								Predicted Points{sortLabel('predictedPoints')}
							</button>
						</th>
						<th className="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
							<button type="button" className="hover:text-gray-900" onClick={() => setSort('predictedInvestment')}>
								Predicted Investment{sortLabel('predictedInvestment')}
							</button>
						</th>
						<th className="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
							<button type="button" className="hover:text-gray-900" onClick={() => setSort('predictedROI')}>
								Predicted ROI{sortLabel('predictedROI')}
							</button>
						</th>
						<th className="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
							<button type="button" className="hover:text-gray-900" onClick={() => setSort('ourBid')}>
								Our Bid{sortLabel('ourBid')}
							</button>
						</th>
						<th className="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
							<button type="button" className="hover:text-gray-900" onClick={() => setSort('ourROI')}>
								Our ROI{sortLabel('ourROI')}
							</button>
						</th>
                    </tr>
                  </thead>
                  <tbody className="bg-white divide-y divide-gray-200">
                    {rows.map((r) => (
                      <tr key={r.teamId} className={r.ourBid > 0 ? 'bg-amber-50 hover:bg-amber-100' : 'hover:bg-gray-50'}>
                        <td className="px-3 py-2 text-sm text-gray-700 break-words">{r.schoolName}</td>
                        <td className="px-3 py-2 text-sm text-gray-700">{r.seed}</td>
                        <td className="px-3 py-2 text-sm text-gray-700">{r.region}</td>
                        <td className="px-3 py-2 text-sm text-gray-900 text-right font-medium">{formatPoints(r.predictedPoints)}</td>
                        <td className="px-3 py-2 text-sm text-gray-900 text-right font-medium">{formatPoints(r.predictedInvestment)}</td>
                        <td className="px-3 py-2 text-sm text-gray-900 text-right font-medium">{formatROI(r.predictedROI)}</td>
                        <td className="px-3 py-2 text-sm text-gray-900 text-right font-medium">{r.ourBid}</td>
                        <td className="px-3 py-2 text-sm text-gray-900 text-right font-medium">{formatROI(r.ourROI)}</td>
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
