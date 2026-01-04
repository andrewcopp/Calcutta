import React, { useMemo, useState } from 'react';
import { Link } from 'react-router-dom';

type SimulationBatch = {
  id: string;
  season: number;
  totalSimulations: number;
  createdAt: string;
};

export function SimulationsPage() {
  const seasons = useMemo(() => {
    const currentYear = new Date().getFullYear();
    return [currentYear, currentYear - 1, currentYear - 2, currentYear - 3, currentYear - 4].filter((y, i, arr) =>
      arr.indexOf(y) === i
    );
  }, []);

  const [season, setSeason] = useState<number>(seasons[0] ?? new Date().getFullYear());
  const [numSimulations, setNumSimulations] = useState<string>('5000');
  const [batchesBySeason, setBatchesBySeason] = useState<Record<number, SimulationBatch[]>>(() => {
    const now = new Date();
    const makeBatch = (s: number, index: number, total: number): SimulationBatch => ({
      id: `${s}-${index}`,
      season: s,
      totalSimulations: total,
      createdAt: new Date(now.getTime() - index * 86_400_000).toISOString(),
    });

    const seed: Record<number, SimulationBatch[]> = {};
    for (const s of seasons) {
      seed[s] = [makeBatch(s, 1, 5000), makeBatch(s, 2, 10000)];
    }
    return seed;
  });

  const batches = batchesBySeason[season] ?? [];

  const parsedNumSimulations = useMemo(() => {
    const n = Number(numSimulations);
    if (!Number.isFinite(n)) return null;
    const floored = Math.floor(n);
    if (floored <= 0) return null;
    return floored;
  }, [numSimulations]);

  const handleGenerate = () => {
    if (parsedNumSimulations == null) return;

    setBatchesBySeason((prev) => {
      const existing = prev[season] ?? [];
      const nextIndex = existing.length + 1;
      const nextBatch: SimulationBatch = {
        id: `${season}-${nextIndex}`,
        season,
        totalSimulations: parsedNumSimulations,
        createdAt: new Date().toISOString(),
      };

      return {
        ...prev,
        [season]: [nextBatch, ...existing],
      };
    });
  };

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-8">
        <h1 className="text-3xl font-bold mb-2">Simulations</h1>
        <p className="text-gray-600">Generate and manage batches of simulated tournaments.</p>
      </div>

      <div className="bg-blue-50 border border-blue-200 rounded-lg p-4 mb-6">
        <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-2">
          <div>
            <div className="text-sm font-medium text-blue-900">New: Runs viewer</div>
            <div className="text-sm text-blue-800">Browse recent pipeline runs (read-only): runs → rankings → entry portfolio.</div>
          </div>
          <div className="flex gap-2">
            <Link
              to={`/runs/${season}`}
              className="px-3 py-2 bg-blue-600 text-white rounded text-sm hover:bg-blue-700"
            >
              Open runs for {season}
            </Link>
            <Link
              to={`/runs/${new Date().getFullYear()}`}
              className="px-3 py-2 border border-blue-300 text-blue-900 rounded text-sm hover:bg-blue-100"
            >
              Latest
            </Link>
          </div>
        </div>
      </div>

      <div className="bg-white rounded-lg shadow p-6 mb-6">
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4 items-end">
          <div>
            <label htmlFor="season" className="block text-sm font-medium text-gray-700 mb-1">
              Season
            </label>
            <select
              id="season"
              value={season}
              onChange={(e) => setSeason(Number(e.target.value))}
              className="w-full border border-gray-300 rounded px-3 py-2"
            >
              {seasons.map((y) => (
                <option key={y} value={y}>
                  {y}
                </option>
              ))}
            </select>
          </div>

          <div>
            <label htmlFor="numSimulations" className="block text-sm font-medium text-gray-700 mb-1">
              Number of simulations
            </label>
            <input
              id="numSimulations"
              type="number"
              min={1}
              step={1}
              value={numSimulations}
              onChange={(e) => setNumSimulations(e.target.value)}
              className="w-full border border-gray-300 rounded px-3 py-2"
            />
          </div>

          <div className="md:col-span-2">
            <button
              type="button"
              onClick={handleGenerate}
              disabled={parsedNumSimulations == null}
              className="w-full md:w-auto px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50"
            >
              Generate
            </button>
            <div className="mt-2 text-sm text-gray-600">
              This is UI-only for now (no async workers wired up yet).
            </div>
          </div>
        </div>
      </div>

      <div className="bg-white rounded-lg shadow p-6">
        <h2 className="text-xl font-semibold mb-4">Simulation Batches</h2>

        {batches.length === 0 ? (
          <div className="text-gray-500">No simulation batches found for {season}.</div>
        ) : (
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Batch</th>
                  <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">Simulations</th>
                  <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">Created</th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {batches.map((b) => (
                  <tr key={b.id} className="hover:bg-gray-50">
                    <td className="px-4 py-3 text-sm font-medium text-gray-900">{b.id}</td>
                    <td className="px-4 py-3 text-sm text-center text-gray-700">{b.totalSimulations.toLocaleString()}</td>
                    <td className="px-4 py-3 text-sm text-center text-gray-600">
                      {new Date(b.createdAt).toLocaleString()}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
}
