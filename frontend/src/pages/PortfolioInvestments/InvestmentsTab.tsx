import { useMemo } from 'react';
import { Investment } from '../../schemas/pool';
import type { TournamentTeam } from '../../schemas/tournament';
import { School } from '../../schemas/school';
import { Select } from '../../components/ui/Select';
import { SegmentedBar } from '../../components/SegmentedBar';
import { createTeamSortComparator } from '../../utils/teamSort';

export function InvestmentsTab({
  portfolioId,
  tournamentTeams,
  allInvestments,
  schools,
  investmentsSortBy,
  setInvestmentsSortBy,
  showAllTeams,
  setShowAllTeams,
}: {
  portfolioId: string;
  tournamentTeams: TournamentTeam[];
  allInvestments: Investment[];
  schools: School[];
  investmentsSortBy: 'total' | 'seed' | 'region' | 'team';
  setInvestmentsSortBy: (value: 'total' | 'seed' | 'region' | 'team') => void;
  showAllTeams: boolean;
  setShowAllTeams: (value: boolean) => void;
}) {
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
        portfolioInvestment: number;
        otherInvestments: { portfolioId: string; portfolioName: string; amount: number }[];
      }
    >();

    for (const tt of tournamentTeams) {
      byTeam.set(tt.id, {
        teamId: tt.id,
        seed: tt.seed,
        region: tt.region,
        teamName: schoolNameById.get(tt.schoolId) || 'Unknown School',
        totalInvestment: 0,
        portfolioInvestment: 0,
        otherInvestments: [],
      });
    }

    for (const investment of allInvestments) {
      const amount = investment.credits;
      if (amount <= 0) continue;

      const row = byTeam.get(investment.teamId);
      if (!row) continue;

      row.totalInvestment += amount;
      if (investment.portfolioId === portfolioId) {
        row.portfolioInvestment += amount;
      } else {
        row.otherInvestments.push({
          portfolioId: investment.portfolioId,
          portfolioName: 'Other',
          amount,
        });
      }
    }

    let rows = Array.from(byTeam.values());

    if (!showAllTeams) {
      rows = rows.filter((row) => row.portfolioInvestment > 0);
    }

    const maxTotal = rows.reduce((max, row) => Math.max(max, row.totalInvestment), 0);

    return { rows, maxTotal };
  }, [allInvestments, portfolioId, schoolNameById, showAllTeams, tournamentTeams]);

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

  const PORTFOLIO_COLOR = '#2563EB';
  const OTHER_COLOR = '#cfd6df';

  return (
    <div className="bg-card rounded-lg shadow p-6">
      <div className="mb-4 flex items-center justify-between gap-4">
        <label className="flex items-center gap-2 text-sm text-foreground">
          <input
            type="checkbox"
            checked={showAllTeams}
            onChange={(e) => setShowAllTeams(e.target.checked)}
            className="rounded border-border"
          />
          Show All Teams
        </label>
        <label className="text-sm text-muted-foreground">
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
            <tr className="text-left text-xs uppercase tracking-wide text-muted-foreground">
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
              const barWidthPct =
                investmentRows.maxTotal > 0 ? (row.totalInvestment / investmentRows.maxTotal) * 100 : 0;
              const otherTotal = row.otherInvestments.reduce((sum, inv) => sum + inv.amount, 0);

              return (
                <tr key={row.teamId} className="bg-accent">
                  <td className="px-2 py-3 font-medium text-foreground rounded-l-md whitespace-nowrap">
                    {row.seed ?? '\u2014'}
                  </td>
                  <td className="px-2 py-3 text-foreground whitespace-nowrap">{row.region}</td>
                  <td className="px-2 py-3 text-foreground font-medium whitespace-nowrap truncate">{row.teamName}</td>
                  <td className="px-2 py-3">
                    <SegmentedBar
                      barWidthPct={barWidthPct}
                      segments={[
                        ...(row.portfolioInvestment > 0
                          ? [
                              {
                                key: `${row.teamId}-portfolio`,
                                label: 'Your Portfolio',
                                value: row.portfolioInvestment,
                                color: PORTFOLIO_COLOR,
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
                      getTooltipValue={(seg) => `${seg.value.toFixed(2)} credits`}
                    />
                  </td>
                  <td className="px-2 py-3 text-right font-medium text-foreground whitespace-nowrap">
                    {row.portfolioInvestment.toFixed(2)} credits
                  </td>
                  <td className="px-2 py-3 text-right font-medium text-foreground rounded-r-md whitespace-nowrap">
                    {row.totalInvestment.toFixed(2)} credits
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
    </div>
  );
}
