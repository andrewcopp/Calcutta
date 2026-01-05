import React, { useMemo, useState } from 'react';
import { Link } from 'react-router-dom';
import { Card } from '../components/ui/Card';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { Select } from '../components/ui/Select';

type EvaluationRowId = 'return-algorithm' | 'investment-algorithm' | 'entry-optimizer';

type EvaluationRow = {
  id: EvaluationRowId;
  label: string;
  description: string;
};

export function EvaluationsPage() {
  const returnAlgorithms = useMemo(() => ['KenPom Ratings', 'Elo Blend', 'Vegas Implied'], []);
  const investmentAlgorithms = useMemo(() => ['Ridge Regression 5 Features', 'Ridge Regression 10 Features', 'Heuristic Baseline'], []);
  const entryOptimizers = useMemo(() => ['MINLP', 'Greedy', 'Simulated Annealing'], []);

  const [selectedReturnAlgorithm, setSelectedReturnAlgorithm] = useState<string>(returnAlgorithms[0] ?? 'KenPom Ratings');
  const [selectedInvestmentAlgorithm, setSelectedInvestmentAlgorithm] = useState<string>(investmentAlgorithms[0] ?? 'Ridge Regression 5 Features');
  const [selectedEntryOptimizer, setSelectedEntryOptimizer] = useState<string>(entryOptimizers[0] ?? 'MINLP');

  const rows = useMemo<EvaluationRow[]>(
    () => [
      {
        id: 'return-algorithm',
        label: 'Return Algorithm',
        description: 'How well return models perform across season simulations.',
      },
      {
        id: 'investment-algorithm',
        label: 'Investment Algorithm',
        description: 'How well market/investment models perform across season simulations.',
      },
      {
        id: 'entry-optimizer',
        label: 'Entry Optimizer',
        description: 'How well portfolio construction strategies perform across season simulations.',
      },
    ],
    []
  );

  const handleRun = () => {
    // UI-only placeholder
  };

  const queryString = useMemo(() => {
    const params = new URLSearchParams();
    params.set('return', selectedReturnAlgorithm);
    params.set('investment', selectedInvestmentAlgorithm);
    params.set('optimizer', selectedEntryOptimizer);
    return params.toString();
  }, [selectedReturnAlgorithm, selectedInvestmentAlgorithm, selectedEntryOptimizer]);

  return (
    <PageContainer>
      <PageHeader title="Evaluations" subtitle="Compare modeling/optimization components across season simulations." />

      <Card className="mb-6">
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4 items-end">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Return Algorithm</label>
            <Select
              value={selectedReturnAlgorithm}
              onChange={(e) => setSelectedReturnAlgorithm(e.target.value)}
              className="w-full"
            >
              {returnAlgorithms.map((a) => (
                <option key={a} value={a}>
                  {a}
                </option>
              ))}
            </Select>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Investment Algorithm</label>
            <Select
              value={selectedInvestmentAlgorithm}
              onChange={(e) => setSelectedInvestmentAlgorithm(e.target.value)}
              className="w-full"
            >
              {investmentAlgorithms.map((a) => (
                <option key={a} value={a}>
                  {a}
                </option>
              ))}
            </Select>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Entry Optimizer</label>
            <Select
              value={selectedEntryOptimizer}
              onChange={(e) => setSelectedEntryOptimizer(e.target.value)}
              className="w-full"
            >
              {entryOptimizers.map((a) => (
                <option key={a} value={a}>
                  {a}
                </option>
              ))}
            </Select>
          </div>

          <div>
            <button
              type="button"
              onClick={handleRun}
              className="w-full px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50"
              disabled
            >
              Run
            </button>
            <div className="mt-2 text-sm text-gray-600">Not wired up yet.</div>
          </div>
        </div>
      </Card>

      <Card>
        <h2 className="text-xl font-semibold mb-4">Evaluation Types</h2>
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Evaluation</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Description</th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {rows.map((row) => (
                <tr key={row.id} className="hover:bg-gray-50">
                  <td className="px-4 py-3 text-sm font-medium text-blue-700">
                    <Link to={`/evaluations/${row.id}?${queryString}`} className="hover:underline">
                      {row.label}
                    </Link>
                  </td>
                  <td className="px-4 py-3 text-sm text-gray-700">{row.description}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </Card>
    </PageContainer>
  );
}
