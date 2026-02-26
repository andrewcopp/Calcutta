import { useMemo, useState } from 'react';
import { Portfolio, Investment } from '../../schemas/pool';
import type { TournamentTeam } from '../../schemas/tournament';
import { School } from '../../schemas/school';
import { Select } from '../../components/ui/Select';
import { SegmentedBar } from '../../components/SegmentedBar';
import { getEntryColorById } from '../../utils/entryColors';
import { createTeamSortComparator } from '../../utils/teamSort';

export function InvestmentsTab({
  portfolios,
  schools,
  tournamentTeams,
  allInvestments,
}: {
  portfolios: Portfolio[];
  schools: School[];
  tournamentTeams: TournamentTeam[];
  allInvestments: Investment[];
}) {
  const [investmentSortBy, setInvestmentSortBy] = useState<'total' | 'seed' | 'region' | 'team'>('total');

  const portfolioNameById = useMemo(() => new Map(portfolios.map((p) => [p.id, p.name])), [portfolios]);
  const tournamentTeamById = useMemo(() => new Map(tournamentTeams.map((team) => [team.id, team])), [tournamentTeams]);
  const schoolNameById = useMemo(() => new Map(schools.map((school) => [school.id, school.name])), [schools]);

  const portfolioColorById = useMemo(() => getEntryColorById(portfolios), [portfolios]);

  const investmentRows = useMemo(() => {
    const byTeam = new Map<
      string,
      {
        teamId: string;
        seed: number | undefined;
        region: string;
        teamName: string;
        totalInvestment: number;
        segments: { portfolioId: string; portfolioName: string; amount: number }[];
      }
    >();

    for (const tournamentTeam of tournamentTeams) {
      const schoolName = schoolNameById.get(tournamentTeam.schoolId) || 'Unknown School';
      byTeam.set(tournamentTeam.id, {
        teamId: tournamentTeam.id,
        seed: tournamentTeam.seed,
        region: tournamentTeam.region,
        teamName: schoolName,
        totalInvestment: 0,
        segments: [],
      });
    }

    for (const investment of allInvestments) {
      const amount = investment.credits;
      if (amount <= 0) continue;

      const teamId = investment.teamId;
      const row = byTeam.get(teamId);
      if (!row) {
        const tournamentTeam = tournamentTeamById.get(teamId);
        const seed = tournamentTeam?.seed;
        const region = tournamentTeam?.region || 'Unknown';
        const schoolName = tournamentTeam ? schoolNameById.get(tournamentTeam.schoolId) : undefined;

        byTeam.set(teamId, {
          teamId,
          seed,
          region,
          teamName: schoolName || 'Unknown School',
          totalInvestment: amount,
          segments: [
            {
              portfolioId: investment.portfolioId,
              portfolioName: portfolioNameById.get(investment.portfolioId) || 'Unknown Portfolio',
              amount,
            },
          ],
        });
        continue;
      }

      row.totalInvestment += amount;
      row.segments.push({
        portfolioId: investment.portfolioId,
        portfolioName: portfolioNameById.get(investment.portfolioId) || 'Unknown Portfolio',
        amount,
      });
    }

    const rows = Array.from(byTeam.values())
      .map((row) => ({
        ...row,
        segments: row.segments
          .slice()
          .sort((a, b) => b.amount - a.amount)
          .filter((seg) => seg.amount > 0),
      }))
      .sort((a, b) => {
        if (b.totalInvestment !== a.totalInvestment) return b.totalInvestment - a.totalInvestment;
        const seedA = a.seed ?? 999;
        const seedB = b.seed ?? 999;
        return seedA - seedB;
      });

    const maxTotal = rows.reduce((max, row) => Math.max(max, row.totalInvestment), 0);

    return { rows, maxTotal };
  }, [allInvestments, portfolioNameById, schoolNameById, tournamentTeams, tournamentTeamById]);

  const investmentSortedRows = useMemo(() => {
    const rows = investmentRows.rows.slice();

    if (investmentSortBy === 'total') {
      rows.sort((a, b) => {
        if (b.totalInvestment !== a.totalInvestment) return b.totalInvestment - a.totalInvestment;
        return (a.seed ?? 999) - (b.seed ?? 999);
      });
    } else {
      rows.sort(createTeamSortComparator(investmentSortBy));
    }

    return rows;
  }, [investmentRows.rows, investmentSortBy]);

  return (
    <div className="bg-card rounded-lg shadow p-6">
      <div className="mb-4 flex items-center justify-end">
        <label className="text-sm text-muted-foreground">
          Sort by
          <Select
            className="ml-2 w-auto"
            value={investmentSortBy}
            onChange={(e) => setInvestmentSortBy(e.target.value as 'total' | 'seed' | 'region' | 'team')}
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
              <th className="px-2 py-2 w-24 text-right">Total Investment</th>
            </tr>
          </thead>
          <tbody>
            {investmentSortedRows.map((row) => {
              const barWidthPct =
                investmentRows.maxTotal > 0 ? (row.totalInvestment / investmentRows.maxTotal) * 100 : 0;
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
                      segments={row.segments.map((seg, idx) => ({
                        key: `${row.teamId}-${seg.portfolioId}-${idx}`,
                        label: seg.portfolioName,
                        value: seg.amount,
                        color: portfolioColorById.get(seg.portfolioId) || '#94A3B8',
                      }))}
                      backgroundColor="#E5E7EB"
                      getTooltipTitle={(seg) => seg.label}
                      getTooltipValue={(seg) => `${seg.value.toFixed(2)} credits`}
                    />
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
