import React, { useMemo, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Link, useNavigate, useParams, useSearchParams } from 'react-router-dom';

import { ApiError } from '../api/apiClient';
import { Alert } from '../components/ui/Alert';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { LoadingState } from '../components/ui/LoadingState';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { Select } from '../components/ui/Select';
import { calcuttaService } from '../services/calcuttaService';
import { simulationRunsService, type SimulationRun } from '../services/simulationRunsService';
import { simulationRunBatchesService } from '../services/simulationRunBatchesService';
import { cohortsService } from '../services/cohortsService';
import { syntheticCalcuttasService, type SyntheticCalcuttaListItem } from '../services/syntheticCalcuttasService';
import { syntheticEntriesService, type SyntheticEntryListItem } from '../services/syntheticEntriesService';
import { entryRunsService, type EntryRunListItem } from '../services/entryRunsService';
import type { Calcutta } from '../types/calcutta';

export function SandboxCohortDetailPage() {
  const navigate = useNavigate();
  const { cohortId } = useParams<{ cohortId?: string }>();
  const queryClient = useQueryClient();
  const [searchParams, setSearchParams] = useSearchParams();
  const selectedExecutionId = searchParams.get('executionId') || '';
  const selectedSyntheticCalcuttaId = searchParams.get('syntheticCalcuttaId') || '';

  const [newSyntheticEntryName, setNewSyntheticEntryName] = useState<string>('');
  const [sourceCalcuttaId, setSourceCalcuttaId] = useState<string>('');
  const [selectedEntryRunId, setSelectedEntryRunId] = useState<string>('');
  const [importSyntheticEntryName, setImportSyntheticEntryName] = useState<string>('');

  const { data: calcuttas = [] } = useQuery<Calcutta[]>({
    queryKey: ['calcuttas', 'all'],
    queryFn: calcuttaService.getAllCalcuttas,
  });

  const calcuttaNameById = useMemo(() => {
    const m = new Map<string, string>();
    for (const c of calcuttas) {
      m.set(c.id, c.name);
    }
    return m;
  }, [calcuttas]);

  const showError = (err: unknown) => {
    if (err instanceof ApiError) {
      if (err.status === 403) {
        return 'You do not have permission to view cohorts (403).';
      }
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

  const formatFloat = (v: number | null | undefined, digits: number) => {
    if (v == null || Number.isNaN(v)) return '—';
    return v.toFixed(digits);
  };

  const seasonFromCalcuttaName = (name: string | null | undefined) => {
    if (!name) return '—';
    const m = name.match(/\b(19|20)\d{2}\b/);
    return m ? m[0] : '—';
  };

  const effectiveCohortId = cohortId || '';

  const cohortQuery = useQuery({
    queryKey: ['cohorts', 'get', effectiveCohortId],
    queryFn: () => cohortsService.get(effectiveCohortId),
    enabled: Boolean(effectiveCohortId),
  });

  const cohortTitle = cohortQuery.data?.name ? `${cohortQuery.data.name}` : effectiveCohortId ? `Cohort ${effectiveCohortId}` : 'Cohort';

  const syntheticCalcuttasQuery = useQuery({
    queryKey: ['synthetic-calcuttas', 'list', effectiveCohortId],
    queryFn: () => syntheticCalcuttasService.list({ cohortId: effectiveCohortId, limit: 200, offset: 0 }),
    enabled: Boolean(effectiveCohortId),
  });

  const syntheticCalcuttas: SyntheticCalcuttaListItem[] = useMemo(
    () => syntheticCalcuttasQuery.data?.items ?? [],
    [syntheticCalcuttasQuery.data?.items]
  );

  const effectiveSourceCalcuttaId = useMemo(() => {
    if (sourceCalcuttaId) return sourceCalcuttaId;
    return calcuttas.length > 0 ? calcuttas[0].id : '';
  }, [calcuttas, sourceCalcuttaId]);

  const createSyntheticCalcuttaMutation = useMutation({
    mutationFn: async () => {
      if (!effectiveCohortId) throw new Error('Missing cohort ID');
      const src = effectiveSourceCalcuttaId;
      if (!src) throw new Error('Missing source calcutta');
      return syntheticCalcuttasService.create({ cohortId: effectiveCohortId, sourceCalcuttaId: src });
    },
    onSuccess: async (res) => {
      await queryClient.invalidateQueries({ queryKey: ['synthetic-calcuttas', 'list', effectiveCohortId] });
      const nextParams: Record<string, string> = {};
      if (selectedExecutionId) nextParams.executionId = selectedExecutionId;
      if (res?.id) nextParams.syntheticCalcuttaId = res.id;
      setSearchParams(nextParams, { replace: true });
    },
  });

  const effectiveSyntheticCalcuttaId = useMemo(() => {
    if (selectedSyntheticCalcuttaId) return selectedSyntheticCalcuttaId;
    return syntheticCalcuttas.length > 0 ? syntheticCalcuttas[0].id : '';
  }, [selectedSyntheticCalcuttaId, syntheticCalcuttas]);

  const effectiveSyntheticCalcutta = useMemo(() => {
    if (!effectiveSyntheticCalcuttaId) return null;
    return syntheticCalcuttas.find((sc) => sc.id === effectiveSyntheticCalcuttaId) ?? null;
  }, [effectiveSyntheticCalcuttaId, syntheticCalcuttas]);

  const entryRunsQuery = useQuery({
    queryKey: ['entry-runs', 'list', effectiveSyntheticCalcutta?.calcutta_id ?? ''],
    queryFn: () => entryRunsService.list({ calcuttaId: effectiveSyntheticCalcutta?.calcutta_id ?? '', limit: 200, offset: 0 }),
    enabled: Boolean(effectiveSyntheticCalcutta?.calcutta_id),
  });

  const entryRuns: EntryRunListItem[] = entryRunsQuery.data?.items ?? [];

  const syntheticEntriesQuery = useQuery({
    queryKey: ['synthetic-entries', 'list', effectiveSyntheticCalcuttaId],
    queryFn: () => syntheticEntriesService.list(effectiveSyntheticCalcuttaId),
    enabled: Boolean(effectiveSyntheticCalcuttaId),
  });

  const syntheticEntries: SyntheticEntryListItem[] = syntheticEntriesQuery.data?.items ?? [];

  const createSyntheticEntryMutation = useMutation({
    mutationFn: async () => {
      const name = newSyntheticEntryName.trim();
      return syntheticEntriesService.create(effectiveSyntheticCalcuttaId, { displayName: name, teams: [] });
    },
    onSuccess: async () => {
      setNewSyntheticEntryName('');
      await queryClient.invalidateQueries({ queryKey: ['synthetic-entries', 'list', effectiveSyntheticCalcuttaId] });
    },
  });

  const renameSyntheticEntryMutation = useMutation({
    mutationFn: async (args: { id: string; displayName: string }) => {
      return syntheticEntriesService.patch(args.id, { displayName: args.displayName });
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['synthetic-entries', 'list', effectiveSyntheticCalcuttaId] });
    },
  });

  const deleteSyntheticEntryMutation = useMutation({
    mutationFn: async (id: string) => syntheticEntriesService.delete(id),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['synthetic-entries', 'list', effectiveSyntheticCalcuttaId] });
    },
  });

  const importSyntheticEntryMutation = useMutation({
    mutationFn: async () => {
      const entryRunId = selectedEntryRunId.trim();
      if (!entryRunId) throw new Error('Missing entry run');
      const displayName = importSyntheticEntryName.trim();

      const metrics = await entryRunsService.getArtifact(entryRunId, 'metrics');
      return syntheticEntriesService.importFromEntryArtifact(effectiveSyntheticCalcuttaId, {
        entryArtifactId: metrics.id,
        displayName: displayName.length > 0 ? displayName : undefined,
      });
    },
    onSuccess: async () => {
      setImportSyntheticEntryName('');
      await queryClient.invalidateQueries({ queryKey: ['synthetic-entries', 'list', effectiveSyntheticCalcuttaId] });
    },
  });

  const executionsQuery = useQuery({
    queryKey: ['simulation-run-batches', 'list', effectiveCohortId],
    queryFn: () => simulationRunBatchesService.list({ cohortId: effectiveCohortId, limit: 200, offset: 0 }),
    enabled: Boolean(effectiveCohortId),
  });

  const executions = useMemo(() => executionsQuery.data?.items ?? [], [executionsQuery.data?.items]);

  const effectiveExecutionId = useMemo(() => {
    if (selectedExecutionId) return selectedExecutionId;
    if (cohortQuery.data?.latest_execution_id) return cohortQuery.data.latest_execution_id;
    return executions.length > 0 ? executions[0].id : '';
  }, [executions, selectedExecutionId, cohortQuery.data?.latest_execution_id]);

  const evaluationsQuery = useQuery({
    queryKey: ['simulation-runs', 'list', 'simulation-run-batch', effectiveExecutionId],
    queryFn: () =>
      simulationRunsService.list({
        cohortId: effectiveCohortId,
        simulationBatchId: effectiveExecutionId || undefined,
        limit: 200,
        offset: 0,
      }),
    enabled: Boolean(effectiveCohortId && effectiveExecutionId),
  });

  const evals: SimulationRun[] = evaluationsQuery.data?.items ?? [];

  return (
    <PageContainer className="max-w-none">
      <PageHeader
        title="Sandbox"
        subtitle={cohortTitle}
        leftActions={
          <Link to="/sandbox/cohorts" className="text-blue-600 hover:text-blue-800">
            ← Back to Cohorts
          </Link>
        }
      />

      {!effectiveCohortId ? <Alert variant="error">Missing cohort ID.</Alert> : null}

      {effectiveCohortId && cohortQuery.isLoading ? <LoadingState label="Loading cohort..." /> : null}
      {effectiveCohortId && cohortQuery.isError ? (
        <Alert variant="error">
          <div className="font-semibold mb-1">Failed to load cohort</div>
          <div className="mb-3">{showError(cohortQuery.error)}</div>
          <Button size="sm" onClick={() => cohortQuery.refetch()}>
            Retry
          </Button>
        </Alert>
      ) : null}

      {cohortQuery.data ? (
        <div className="space-y-6">
          <Card>
            <h2 className="text-xl font-semibold mb-4">Cohort</h2>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
              <div>
                <div className="text-gray-500">Name</div>
                <div className="text-gray-900 font-medium">{cohortQuery.data.name || cohortQuery.data.id}</div>
              </div>
              <div>
                <div className="text-gray-500">Optimizer</div>
                <div className="text-gray-900">{cohortQuery.data.optimizer_key}</div>
              </div>
              <div>
                <div className="text-gray-500">nSims</div>
                <div className="text-gray-900">{cohortQuery.data.n_sims}</div>
              </div>
              <div>
                <div className="text-gray-500">Seed</div>
                <div className="text-gray-900">{cohortQuery.data.seed}</div>
              </div>
              <div>
                <div className="text-gray-500">Latest simulation run batch</div>
                <div className="text-gray-900">
                  {cohortQuery.data.latest_execution_id
                    ? `${cohortQuery.data.latest_execution_status ?? '—'} · ${cohortQuery.data.latest_execution_id.slice(0, 8)}`
                    : '—'}
                </div>
              </div>
              <div>
                <div className="text-gray-500">Updated</div>
                <div className="text-gray-900">{formatDateTime(cohortQuery.data.updated_at)}</div>
              </div>
            </div>
          </Card>

          <Card>
            <div className="flex items-end justify-between gap-4 mb-4">
              <h2 className="text-xl font-semibold">Synthetic Calcuttas</h2>

              <div className="flex items-center gap-3">
                <div className="text-sm text-gray-500 whitespace-nowrap">Source Calcutta</div>
                <Select value={effectiveSourceCalcuttaId} onChange={(e) => setSourceCalcuttaId(e.target.value)} disabled={calcuttas.length === 0}>
                  {calcuttas.map((c) => (
                    <option key={c.id} value={c.id}>
                      {seasonFromCalcuttaName(c.name)} · {c.name}
                    </option>
                  ))}
                </Select>
                <Button
                  size="sm"
                  disabled={createSyntheticCalcuttaMutation.isPending || !effectiveCohortId || !effectiveSourceCalcuttaId}
                  onClick={() => createSyntheticCalcuttaMutation.mutate()}
                >
                  Create
                </Button>
              </div>
            </div>

            {syntheticCalcuttasQuery.isLoading ? <LoadingState label="Loading synthetic calcuttas..." layout="inline" /> : null}
            {syntheticCalcuttasQuery.isError ? (
              <Alert variant="error" className="mt-3">
                <div className="font-semibold mb-1">Failed to load synthetic calcuttas</div>
                <div className="mb-3">{showError(syntheticCalcuttasQuery.error)}</div>
                <Button size="sm" onClick={() => syntheticCalcuttasQuery.refetch()}>
                  Retry
                </Button>
              </Alert>
            ) : null}

            {!syntheticCalcuttasQuery.isLoading && !syntheticCalcuttasQuery.isError && syntheticCalcuttas.length === 0 ? (
              <Alert variant="info" className="mt-3">
                No synthetic calcuttas found for this cohort.
              </Alert>
            ) : null}

            {createSyntheticCalcuttaMutation.isError ? (
              <Alert variant="error" className="mt-3">
                {showError(createSyntheticCalcuttaMutation.error)}
              </Alert>
            ) : null}

            {!syntheticCalcuttasQuery.isLoading && !syntheticCalcuttasQuery.isError && syntheticCalcuttas.length > 0 ? (
              <div className="overflow-x-auto">
                <table className="min-w-full divide-y divide-gray-200">
                  <thead className="bg-gray-50">
                    <tr>
                      <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Calcutta</th>
                      <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Highlighted Entry</th>
                      <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Created</th>
                      <th className="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Open</th>
                    </tr>
                  </thead>
                  <tbody className="bg-white divide-y divide-gray-200">
                    {syntheticCalcuttas.map((sc) => {
                      const calcuttaName = calcuttaNameById.get(sc.calcutta_id) || sc.calcutta_id;
                      const highlighted = sc.focus_entry_name || '—';
                      const href = `/sandbox/synthetic-calcuttas/${encodeURIComponent(sc.id)}?cohortId=${encodeURIComponent(effectiveCohortId)}`;
                      return (
                        <tr key={sc.id} className="hover:bg-gray-50">
                          <td className="px-3 py-2 text-sm text-gray-900" title={calcuttaName}>
                            <div className="font-medium">{seasonFromCalcuttaName(calcuttaName)}</div>
                            <div className="text-xs text-gray-600">{calcuttaName}</div>
                          </td>
                          <td className="px-3 py-2 text-sm text-gray-700">{highlighted}</td>
                          <td className="px-3 py-2 text-sm text-gray-700">{formatDateTime(sc.created_at)}</td>
                          <td className="px-3 py-2 text-sm text-right">
                            <Link to={href} className="text-blue-600 hover:text-blue-800">
                              View
                            </Link>
                          </td>
                        </tr>
                      );
                    })}
                  </tbody>
                </table>
              </div>
            ) : null}
          </Card>

          <Card>
            <div className="flex items-end justify-between gap-4">
              <h2 className="text-xl font-semibold">Synthetic Entries</h2>

              <div className="flex items-center gap-3">
                <div className="text-sm text-gray-500 whitespace-nowrap">Synthetic Calcutta</div>
                <Select
                  value={effectiveSyntheticCalcuttaId}
                  onChange={(e) => {
                    const next = e.target.value;
                    const nextParams: Record<string, string> = {};
                    if (selectedExecutionId) nextParams.executionId = selectedExecutionId;
                    if (next) nextParams.syntheticCalcuttaId = next;
                    setSearchParams(nextParams, { replace: true });
                  }}
                  disabled={syntheticCalcuttas.length === 0}
                >
                  {syntheticCalcuttas.map((sc) => {
                    const calcuttaName = calcuttaNameById.get(sc.calcutta_id) || sc.calcutta_id;
                    return (
                      <option key={sc.id} value={sc.id}>
                        {seasonFromCalcuttaName(calcuttaName)} · {sc.id.slice(0, 8)}
                      </option>
                    );
                  })}
                </Select>
              </div>
            </div>

            {syntheticEntriesQuery.isLoading ? <LoadingState label="Loading synthetic entries..." layout="inline" className="mt-3" /> : null}
            {syntheticEntriesQuery.isError ? (
              <Alert variant="error" className="mt-3">
                <div className="font-semibold mb-1">Failed to load synthetic entries</div>
                <div className="mb-3">{showError(syntheticEntriesQuery.error)}</div>
                <Button size="sm" onClick={() => syntheticEntriesQuery.refetch()}>
                  Retry
                </Button>
              </Alert>
            ) : null}

            {effectiveSyntheticCalcuttaId && !syntheticEntriesQuery.isLoading && !syntheticEntriesQuery.isError ? (
              <div className="mt-3">
                <div className="flex items-center gap-2">
                  <input
                    value={newSyntheticEntryName}
                    onChange={(e) => setNewSyntheticEntryName(e.target.value)}
                    className="flex-1 border border-gray-300 rounded px-3 py-2 text-sm"
                    placeholder="New synthetic entry name"
                  />
                  <Button
                    size="sm"
                    disabled={
                      createSyntheticEntryMutation.isPending ||
                      !effectiveSyntheticCalcuttaId ||
                      newSyntheticEntryName.trim().length === 0
                    }
                    onClick={() => createSyntheticEntryMutation.mutate()}
                  >
                    Create
                  </Button>
                </div>

                <div className="mt-3">
                  <div className="text-sm text-gray-500 mb-1">Import from Entry Run</div>
                  <div className="flex items-center gap-2">
                    <Select
                      value={selectedEntryRunId}
                      onChange={(e) => setSelectedEntryRunId(e.target.value)}
                      disabled={entryRunsQuery.isLoading || entryRuns.length === 0}
                    >
                      <option value="">Select an entry run…</option>
                      {entryRuns.map((r) => (
                        <option key={r.id} value={r.id}>
                          {(r.name || 'Entry Run') + ' · ' + r.id.slice(0, 8)}
                        </option>
                      ))}
                    </Select>
                    {selectedEntryRunId.trim().length > 0 ? (
                      <Link
                        to={`/lab/entry-runs/${encodeURIComponent(selectedEntryRunId.trim())}`}
                        className="text-blue-600 hover:text-blue-800 text-sm whitespace-nowrap"
                      >
                        View run
                      </Link>
                    ) : null}
                    <input
                      value={importSyntheticEntryName}
                      onChange={(e) => setImportSyntheticEntryName(e.target.value)}
                      className="flex-1 border border-gray-300 rounded px-3 py-2 text-sm"
                      placeholder="Optional display name"
                      disabled={!effectiveSyntheticCalcuttaId}
                    />
                    <Button
                      size="sm"
                      disabled={
                        importSyntheticEntryMutation.isPending ||
                        !effectiveSyntheticCalcuttaId ||
                        selectedEntryRunId.trim().length === 0
                      }
                      onClick={() => importSyntheticEntryMutation.mutate()}
                    >
                      Import
                    </Button>
                  </div>

                  {entryRunsQuery.isError ? (
                    <Alert variant="error" className="mt-2">
                      Failed to load entry runs: {showError(entryRunsQuery.error)}
                    </Alert>
                  ) : null}

                  {importSyntheticEntryMutation.isError ? (
                    <Alert variant="error" className="mt-2">
                      {showError(importSyntheticEntryMutation.error)}
                    </Alert>
                  ) : null}
                </div>

                {createSyntheticEntryMutation.isError ? (
                  <Alert variant="error" className="mt-2">
                    {showError(createSyntheticEntryMutation.error)}
                  </Alert>
                ) : null}

                {syntheticEntries.length === 0 ? (
                  <Alert variant="info" className="mt-3">
                    No synthetic entries found for this synthetic calcutta.
                  </Alert>
                ) : (
                  <div className="overflow-x-auto mt-3">
                    <table className="min-w-full divide-y divide-gray-200">
                      <thead className="bg-gray-50">
                        <tr>
                          <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Entry</th>
                          <th className="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Teams</th>
                          <th className="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Actions</th>
                        </tr>
                      </thead>
                      <tbody className="bg-white divide-y divide-gray-200">
                        {syntheticEntries.map((e) => (
                          <tr key={e.id} className="hover:bg-gray-50">
                            <td className="px-3 py-2 text-sm text-gray-900">
                              <div className="font-medium">{e.display_name}</div>
                              <div className="text-xs text-gray-500">{e.is_synthetic ? 'synthetic' : 'imported'} · {e.id.slice(0, 8)}</div>
                            </td>
                            <td className="px-3 py-2 text-sm text-right text-gray-700">{e.teams?.length ?? 0}</td>
                            <td className="px-3 py-2 text-sm text-right">
                              <div className="flex justify-end gap-2">
                                <Button
                                  size="sm"
                                  variant="secondary"
                                  disabled={renameSyntheticEntryMutation.isPending}
                                  onClick={() => {
                                    const next = window.prompt('Rename synthetic entry', e.display_name);
                                    if (!next) return;
                                    const trimmed = next.trim();
                                    if (!trimmed) return;
                                    renameSyntheticEntryMutation.mutate({ id: e.id, displayName: trimmed });
                                  }}
                                >
                                  Rename
                                </Button>
                                <Button
                                  size="sm"
                                  variant="secondary"
                                  className="bg-red-600 text-white hover:bg-red-700 focus-visible:ring-red-500"
                                  disabled={deleteSyntheticEntryMutation.isPending || !e.is_synthetic}
                                  onClick={() => {
                                    if (!e.is_synthetic) return;
                                    if (!window.confirm(`Delete synthetic entry "${e.display_name}"?`)) return;
                                    deleteSyntheticEntryMutation.mutate(e.id);
                                  }}
                                >
                                  Delete
                                </Button>
                              </div>
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                )}

                {renameSyntheticEntryMutation.isError ? (
                  <Alert variant="error" className="mt-2">
                    {showError(renameSyntheticEntryMutation.error)}
                  </Alert>
                ) : null}
                {deleteSyntheticEntryMutation.isError ? (
                  <Alert variant="error" className="mt-2">
                    {showError(deleteSyntheticEntryMutation.error)}
                  </Alert>
                ) : null}
              </div>
            ) : null}
          </Card>

          <Card>
            <div className="flex items-center justify-between gap-4">
              <h2 className="text-xl font-semibold">Simulation Run Batches</h2>
              <div className="flex items-center gap-3">
                <div className="text-sm text-gray-500 whitespace-nowrap">Run batch</div>
                <Select
                  value={effectiveExecutionId}
                  onChange={(e) => {
                    const next = e.target.value;
                    const nextParams: Record<string, string> = {};
                    if (next) nextParams.executionId = next;
                    if (selectedSyntheticCalcuttaId) nextParams.syntheticCalcuttaId = selectedSyntheticCalcuttaId;
                    setSearchParams(nextParams, { replace: true });
                  }}
                  disabled={executions.length === 0}
                >
                  {executions.map((ex) => (
                    <option key={ex.id} value={ex.id}>
                      {ex.status} · {ex.created_at.slice(0, 10)} · {ex.id.slice(0, 8)}
                    </option>
                  ))}
                </Select>
              </div>
            </div>

            {executionsQuery.isLoading ? <LoadingState label="Loading run batches..." layout="inline" className="mt-3" /> : null}
            {executionsQuery.isError ? (
              <Alert variant="error" className="mt-3">
                <div className="font-semibold mb-1">Failed to load run batches</div>
                <div className="mb-3">{showError(executionsQuery.error)}</div>
                <Button size="sm" onClick={() => executionsQuery.refetch()}>
                  Retry
                </Button>
              </Alert>
            ) : null}

            {!executionsQuery.isLoading && !executionsQuery.isError && executions.length === 0 ? (
              <Alert variant="info" className="mt-3">
                No simulation run batches found yet. Create one from the pipeline (or use the legacy Sandbox page to trigger individual runs).
              </Alert>
            ) : null}

            {effectiveExecutionId ? (
              <div className="mt-3 text-sm text-gray-600">
                Showing simulation runs for run batch <code className="text-gray-800">{effectiveExecutionId}</code>
              </div>
            ) : null}
          </Card>

          <Card>
            <h2 className="text-xl font-semibold mb-4">Simulation Runs</h2>

            {evaluationsQuery.isLoading ? <LoadingState label="Loading simulation runs..." layout="inline" /> : null}
            {evaluationsQuery.isError ? (
              <Alert variant="error" className="mt-3">
                <div className="font-semibold mb-1">Failed to load simulation runs</div>
                <div className="mb-3">{showError(evaluationsQuery.error)}</div>
                <Button size="sm" onClick={() => evaluationsQuery.refetch()}>
                  Retry
                </Button>
              </Alert>
            ) : null}

            {!evaluationsQuery.isLoading && !evaluationsQuery.isError && evals.length === 0 ? (
              <Alert variant="info" className="mt-3">
                No simulation runs found for this run batch.
              </Alert>
            ) : null}

            {!evaluationsQuery.isLoading && !evaluationsQuery.isError && evals.length > 0 ? (
              <div className="overflow-x-auto">
                <table className="min-w-full divide-y divide-gray-200">
                  <thead className="bg-gray-50">
                    <tr>
                      <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Season</th>
                      <th className="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Rank</th>
                      <th className="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Mean Norm Payout</th>
                      <th className="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">P(Top 1)</th>
                      <th className="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">P(In Money)</th>
                      <th className="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Finish</th>
                    </tr>
                  </thead>
                  <tbody className="bg-white divide-y divide-gray-200">
                    {evals.map((it) => {
                      const detailUrl = `/sandbox/evaluations/${encodeURIComponent(it.id)}?cohortId=${encodeURIComponent(
                        effectiveCohortId
                      )}${effectiveExecutionId ? `&executionId=${encodeURIComponent(effectiveExecutionId)}` : ''}`;

                      const calcuttaName = calcuttaNameById.get(it.calcutta_id) || it.calcutta_id;
                      const season = seasonFromCalcuttaName(calcuttaName);
                      const finish =
                        it.realized_finish_position != null
                          ? `${it.realized_finish_position}${it.realized_is_tied ? ' (tied)' : ''}`
                          : '—';

                      return (
                        <tr
                          key={it.id}
                          className="hover:bg-gray-50 cursor-pointer focus-visible:outline focus-visible:outline-2 focus-visible:outline-blue-500"
                          role="link"
                          tabIndex={0}
                          onClick={() => navigate(detailUrl)}
                          onKeyDown={(e) => {
                            if (e.key === 'Enter' || e.key === ' ') {
                              e.preventDefault();
                              navigate(detailUrl);
                            }
                          }}
                          aria-label={`Open simulation run ${it.id}`}
                        >
                          <td className="px-3 py-2 text-sm text-gray-900" title={calcuttaName}>
                            <div className="font-medium">{season}</div>
                            <div className="text-xs text-gray-500">
                              {it.optimizer_key} · n={it.n_sims} · seed={it.seed}
                            </div>
                          </td>

                          <td className="px-3 py-2 text-sm text-right text-gray-700">{it.our_rank ?? '—'}</td>
                          <td className="px-3 py-2 text-sm text-right text-gray-700">{formatFloat(it.our_mean_normalized_payout, 4)}</td>
                          <td className="px-3 py-2 text-sm text-right text-gray-700">{formatFloat(it.our_p_top1, 4)}</td>
                          <td className="px-3 py-2 text-sm text-right text-gray-700">{formatFloat(it.our_p_in_money, 4)}</td>
                          <td className="px-3 py-2 text-sm text-right text-gray-700">{finish}</td>
                        </tr>
                      );
                    })}
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
