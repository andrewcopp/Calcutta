import { useMemo } from 'react';
import { CalcuttaEntryTeam, TournamentTeam } from '../../types/calcutta';
import { School } from '../../types/school';
import { Select } from '../../components/ui/Select';
import { SegmentedBar } from '../../components/SegmentedBar';
import { createTeamSortComparator } from '../../utils/teamSort';
export const InvestmentsTab: React.FC<{
  entryId: string;
  tournamentTeams: TournamentTeam[];
  allEntryTeams: CalcuttaEntryTeam[];
  schools: School[];
  investmentsSortBy: 'total' | 'seed' | 'region' | 'team';
  setInvestmentsSortBy: (value: 'total' | 'seed' | 'region' | 'team') => void;
  showAllTeams: boolean;
  setShowAllTeams: (value: boolean) => void;
}> = ({
  entryId,
  tournamentTeams,
  allEntryTeams,
  schools,
  investmentsSortBy,
  setInvestmentsSortBy,
  showAllTeams,
  setShowAllTeams,
}) => {
  const schoolNameById = useMemo(() => new Map(schools.map((school) => [school.id, school.name])), [schools]);

  const investmentRows = useMemo(() => {
    const byTeam = new Map<
      string,
      {
        teamId: string;
        seed: number | undefined;
        region: string;
        teamName: string;
        totalInvestment: number;
        entryInvestment: number;
        otherInvestments: { entryId: string; entryName: string; amount: number }[];
      }
    >();

    for (const tt of tournamentTeams) {
      byTeam.set(tt.id, {
        teamId: tt.id,
        seed: tt.seed,
        region: tt.region,
        teamName: schoolNameById.get(tt.schoolId) || 'Unknown School',
        totalInvestment: 0,
        entryInvestment: 0,
        otherInvestments: [],
      });
    }

    for (const entryTeam of allEntryTeams) {
      const amount = entryTeam.bid;
      if (amount <= 0) continue;

      const row = byTeam.get(entryTeam.teamId);
      if (!row) continue;

      row.totalInvestment += amount;
      if (entryTeam.entryId === entryId) {
        row.entryInvestment += amount;
      } else {
        row.otherInvestments.push({
          entryId: entryTeam.entryId,
          entryName: 'Other',
          amount,
        });
      }
    }

    let rows = Array.from(byTeam.values());

    if (!showAllTeams) {
      rows = rows.filter((row) => row.entryInvestment > 0);
    }

    const maxTotal = rows.reduce((max, row) => Math.max(max, row.totalInvestment), 0);

    return { rows, maxTotal };
  }, [allEntryTeams, entryId, schoolNameById, showAllTeams, tournamentTeams]);

  const sortedRows = useMemo(() => {
    const rows = investmentRows.rows.slice();

    if (investmentsSortBy === 'total') {
      rows.sort((a, b) => {
        if (b.totalInvestment !== a.totalInvestment) return b.totalInvestment - a.totalInvestment;
        return (a.seed ?? 999) - (b.seed ?? 999);
      });
    } else {
      rows.sort(createTeamSortComparator(investmentsSortBy));
    }

    return rows;
  }, [investmentRows.rows, investmentsSortBy]);

  const ENTRY_COLOR = '#2563EB';
  const OTHER_COLOR = '#cfd6df';

  return (
    <div className="bg-white rounded-lg shadow p-6">
      <div className="mb-4 flex items-center justify-between gap-4">
        <label className="flex items-center gap-2 text-sm text-gray-700">
          <input
            type="checkbox"
            checked={showAllTeams}
            onChange={(e) => setShowAllTeams(e.target.checked)}
            className="rounded border-gray-300"
          />
          Show All Teams
        </label>
        <label className="text-sm text-gray-600">
          Sort by
          <Select
            className="ml-2 w-auto"
            value={investmentsSortBy}
            onChange={(e) => setInvestmentsSortBy(e.target.value as 'total' | 'seed' | 'region' | 'team')}
          >
            <option value="total">Total</option>
            <option value="seed">Seed</option>
            <option value="region">Region</option>
            <option value="team">Team</option>
          </Select>
        </label>
      </div>

      <div className="mt-4 overflow-x-auto">
        <table className="min-w-full table-fixed border-separate border-spacing-y-2">
          <thead>
            <tr className="text-left text-xs uppercase tracking-wide text-gray-500">
              <th className="px-2 py-2 w-14">Seed</th>
              <th className="px-2 py-2 w-20">Region</th>
              <th className="px-2 py-2 w-44">Team</th>
              <th className="px-2 py-2"></th>
              <th className="px-2 py-2 w-24 text-right">Investment</th>
              <th className="px-2 py-2 w-32 text-right">Total Investment</th>
            </tr>
          </thead>
          <tbody>
            {sortedRows.map((row) => {
              const barWidthPct = investmentRows.maxTotal > 0 ? (row.totalInvestment / investmentRows.maxTotal) * 100 : 0;
              const otherTotal = row.otherInvestments.reduce((sum, inv) => sum + inv.amount, 0);

              return (
                <tr key={row.teamId} className="bg-gray-50">
                  <td className="px-2 py-3 font-medium text-gray-900 rounded-l-md whitespace-nowrap">{row.seed ?? '—'}</td>
                  <td className="px-2 py-3 text-gray-700 whitespace-nowrap">{row.region}</td>
                  <td className="px-2 py-3 text-gray-900 font-medium whitespace-nowrap truncate">{row.teamName}</td>
                  <td className="px-2 py-3">
                    <SegmentedBar
                      barWidthPct={barWidthPct}
                      segments={[
                        ...(row.entryInvestment > 0
                          ? [
                              {
                                key: `${row.teamId}-entry`,
                                label: 'Your Entry',
                                value: row.entryInvestment,
                                color: ENTRY_COLOR,
                              },
                            ]
                          : []),
                        ...(otherTotal > 0
                          ? [
                              {
                                key: `${row.teamId}-others`,
                                label: 'Others',
                                value: otherTotal,
                                color: OTHER_COLOR,
                              },
                            ]
                          : []),
                      ]}
                      backgroundColor="#E5E7EB"
                      getTooltipTitle={(seg) => seg.label}
                      getTooltipValue={(seg) => `${seg.value.toFixed(2)} pts`}
                    />
                  </td>
                  <td className="px-2 py-3 text-right font-medium text-gray-900 whitespace-nowrap">{row.entryInvestment.toFixed(2)} pts</td>
                  <td className="px-2 py-3 text-right font-medium text-gray-900 rounded-r-md whitespace-nowrap">{row.totalInvestment.toFixed(2)} pts</td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
    </div>
  );
};
