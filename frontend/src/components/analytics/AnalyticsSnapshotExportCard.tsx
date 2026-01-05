import React, { useMemo, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { queryKeys } from '../../queryKeys';
import { analyticsService } from '../../services/analyticsService';
import { apiClient } from '../../api/apiClient';
import { Tournament } from '../../types/tournament';
import { Calcutta } from '../../types/calcutta';

export function AnalyticsSnapshotExportCard() {
  const [exportTournamentId, setExportTournamentId] = useState<string>('');
  const [exportCalcuttaId, setExportCalcuttaId] = useState<string>('');
  const [exportBusy, setExportBusy] = useState(false);
  const [exportError, setExportError] = useState<string | null>(null);

  const tournamentsQuery = useQuery({
    queryKey: queryKeys.tournaments.all(),
    staleTime: 30_000,
    queryFn: () => apiClient.get<Tournament[]>('/tournaments'),
  });

  const calcuttasQuery = useQuery({
    queryKey: queryKeys.calcuttas.all(),
    staleTime: 30_000,
    queryFn: () => apiClient.get<Calcutta[]>('/calcuttas'),
  });

  const tournaments = tournamentsQuery.data ?? [];
  const calcuttas = useMemo(() => calcuttasQuery.data ?? [], [calcuttasQuery.data]);

  const filteredCalcuttas = useMemo(() => {
    return exportTournamentId ? calcuttas.filter((c) => c.tournamentId === exportTournamentId) : calcuttas;
  }, [calcuttas, exportTournamentId]);

  const downloadSnapshot = async () => {
    setExportError(null);
    setExportBusy(true);
    try {
      if (!exportTournamentId) {
        throw new Error('Please select a tournament');
      }
      if (!exportCalcuttaId) {
        throw new Error('Please select a calcutta');
      }

      const { blob, filename } = await analyticsService.exportAnalyticsSnapshot(exportTournamentId, exportCalcuttaId);

      const objectUrl = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = objectUrl;
      a.download = filename;
      document.body.appendChild(a);
      a.click();
      a.remove();
      window.URL.revokeObjectURL(objectUrl);
    } catch (e) {
      setExportError(e instanceof Error ? e.message : String(e));
    } finally {
      setExportBusy(false);
    }
  };

  return (
    <div className="bg-white rounded-lg shadow-md p-6 mb-6">
      <h2 className="text-xl font-semibold mb-2">Export Analytics Snapshot</h2>
      <p className="text-gray-600 mb-4">Download a zip containing CSV tables and a manifest for offline Python analysis.</p>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-4 items-end">
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">Tournament</label>
          <select
            value={exportTournamentId}
            onChange={(e) => {
              setExportTournamentId(e.target.value);
              setExportCalcuttaId('');
            }}
            className="w-full border border-gray-300 rounded px-3 py-2"
            disabled={exportBusy}
          >
            <option value="">Select tournament</option>
            {tournaments.map((t) => (
              <option key={t.id} value={t.id}>
                {t.name}
              </option>
            ))}
          </select>
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">Calcutta</label>
          <select
            value={exportCalcuttaId}
            onChange={(e) => setExportCalcuttaId(e.target.value)}
            className="w-full border border-gray-300 rounded px-3 py-2"
            disabled={exportBusy}
          >
            <option value="">Select calcutta</option>
            {filteredCalcuttas.map((c) => (
              <option key={c.id} value={c.id}>
                {c.name}
              </option>
            ))}
          </select>
        </div>

        <div>
          <button
            onClick={downloadSnapshot}
            disabled={exportBusy}
            className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50"
          >
            Download snapshot (.zip)
          </button>
        </div>
      </div>

      {exportError && <div className="mt-4 text-red-600">{exportError}</div>}
    </div>
  );
}
