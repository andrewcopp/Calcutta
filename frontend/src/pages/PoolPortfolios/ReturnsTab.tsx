import { useMemo, useState } from 'react';
import { Portfolio, OwnershipSummary, OwnershipDetail } from '../../schemas/pool';
import type { TournamentTeam } from '../../schemas/tournament';
import { School } from '../../schemas/school';
import { Select } from '../../components/ui/Select';
import { SegmentedBar } from '../../components/SegmentedBar';
import { getEntryColorById } from '../../utils/entryColors';
import { createTeamSortComparator } from '../../utils/teamSort';

export function ReturnsTab({
  portfolios,
  schools,
  tournamentTeams,
  allOwnershipSummaries,
  allOwnershipDetails,
}: {
  portfolios: Portfolio[];
  schools: School[];
  tournamentTeams: TournamentTeam[];
  allOwnershipSummaries: (OwnershipSummary & { portfolioName?: string })[];
  allOwnershipDetails: OwnershipDetail[];
}) {
  const [returnsSortBy, setReturnsSortBy] = useState<'points' | 'seed' | 'region' | 'team'>('points');

  const portfolioNameById = useMemo(() => new Map(portfolios.map((p) => [p.id, p.name])), [portfolios]);
  const schoolNameById = useMemo(() => new Map(schools.map((school) => [school.id, school.name])), [schools]);

  const portfolioColorById = useMemo(() => getEntryColorById(portfolios), [portfolios]);

  const ownershipSummaryPortfolioIdById = useMemo(() => {
    const map = new Map<string, string>();
    allOwnershipSummaries.forEach((p) => {
      map.set(p.id, p.portfolioId);
    });
    return map;
  }, [allOwnershipSummaries]);

  const returnsRows = useMemo(() => {
    const byTeam = new Map<
      string,
      {
        teamId: string;
        seed: number | undefined;
        region: string;
        teamName: string;
        isEliminated: boolean;
        pointsSegments: { portfolioId: string; portfolioName: string; amount: number }[];
        possibleSegments: { portfolioId: string; portfolioName: string; amount: number }[];
        totalPoints: number;
        totalPossible: number;
      }
    >();

    for (const tt of tournamentTeams) {
      byTeam.set(tt.id, {
        teamId: tt.id,
        seed: tt.seed,
        region: tt.region,
        teamName: schoolNameById.get(tt.schoolId) || 'Unknown School',
        isEliminated: tt.isEliminated === true,
        pointsSegments: [],
        possibleSegments: [],
        totalPoints: 0,
        totalPossible: 0,
      });
    }

    const teamPortfolioAgg = new Map<string, Map<string, { actual: number; expected: number }>>();
    for (const pt of allOwnershipDetails) {
      const portfolioId = ownershipSummaryPortfolioIdById.get(pt.ownershipSummaryId);
      if (!portfolioId) continue;
      if (!teamPortfolioAgg.has(pt.teamId)) teamPortfolioAgg.set(pt.teamId, new Map());
      const byPortfolio = teamPortfolioAgg.get(pt.teamId)!;
      const current = byPortfolio.get(portfolioId) || { actual: 0, expected: 0 };
      current.actual += pt.actualReturns;
      current.expected += pt.expectedReturns;
      byPortfolio.set(portfolioId, current);
    }

    for (const [teamId, byPortfolio] of teamPortfolioAgg.entries()) {
      const row = byTeam.get(teamId);
      if (!row) continue;

      const pointsSegments: { portfolioId: string; portfolioName: string; amount: number }[] = [];
      const possibleSegments: { portfolioId: string; portfolioName: string; amount: number }[] = [];

      for (const [portfolioId, agg] of byPortfolio.entries()) {
        const portfolioName = portfolioNameById.get(portfolioId) || 'Unknown Portfolio';
        const actual = agg.actual;
        const possible = row.isEliminated ? actual : Math.max(agg.expected, actual);

        if (actual > 0) {
          pointsSegments.push({ portfolioId, portfolioName, amount: actual });
        }
        if (possible > 0) {
          possibleSegments.push({ portfolioId, portfolioName, amount: possible });
        }
      }

      pointsSegments.sort((a, b) => b.amount - a.amount);
      possibleSegments.sort((a, b) => b.amount - a.amount);

      row.pointsSegments = pointsSegments;
      row.possibleSegments = possibleSegments;
      row.totalPoints = pointsSegments.reduce((sum, s) => sum + s.amount, 0);
      row.totalPossible = possibleSegments.reduce((sum, s) => sum + s.amount, 0);
    }

    const rows = Array.from(byTeam.values()).sort((a, b) => {
      const seedA = a.seed ?? 999;
      const seedB = b.seed ?? 999;
      if (seedA !== seedB) return seedA - seedB;
      const regionDiff = (a.region || '').localeCompare(b.region || '');
      if (regionDiff !== 0) return regionDiff;
      return (a.teamName || '').localeCompare(b.teamName || '');
    });

    const maxTotal = rows.reduce((max, r) => Math.max(max, r.totalPossible, r.totalPoints), 0);

    return { rows, maxTotal };
  }, [allOwnershipDetails, portfolioNameById, ownershipSummaryPortfolioIdById, schoolNameById, tournamentTeams]);

  const returnsSortedRows = useMemo(() => {
    const rows = returnsRows.rows.slice();

    if (returnsSortBy === 'points') {
      rows.sort((a, b) => {
        if (b.totalPoints !== a.totalPoints) return b.totalPoints - a.totalPoints;
        if (b.totalPossible !== a.totalPossible) return b.totalPossible - a.totalPossible;
        const seedDiff = (a.seed ?? 999) - (b.seed ?? 999);
        if (seedDiff !== 0) return seedDiff;
        return (a.teamName || '').localeCompare(b.teamName || '');
      });
    } else {
      rows.sort(createTeamSortComparator(returnsSortBy));
    }

    return rows;
  }, [returnsRows.rows, returnsSortBy]);

  return (
    <div className="bg-card rounded-lg shadow p-6">
      <div className="mb-4 flex items-center justify-end">
        <label className="text-sm text-muted-foreground">
          Sort by
          <Select
            className="ml-2 w-auto"
            value={returnsSortBy}
            onChange={(e) => setReturnsSortBy(e.target.value as 'points' | 'seed' | 'region' | 'team')}
          >
            <option value="points">Returns</option>
            <option value="seed">Seed</option>
            <option value="region">Region</option>
            <option value="team">Team</option>
          </Select>
        </label>
      </div>
      <div className="mt-2 overflow-x-auto">
        <table className="min-w-full table-fixed border-separate border-spacing-y-2">
          <thead>
            <tr className="text-left text-xs uppercase tracking-wide text-muted-foreground">
              <th className="px-2 py-2 w-14">Seed</th>
              <th className="px-2 py-2 w-20">Region</th>
              <th className="px-2 py-2 w-44">Team</th>
              <th className="px-2 py-2"></th>
              <th className="px-2 py-2 w-32 text-right">Total Returns</th>
            </tr>
          </thead>
          <tbody>
            {returnsSortedRows.map((row) => {
              const pointsWidthPct = returnsRows.maxTotal > 0 ? (row.totalPoints / returnsRows.maxTotal) * 100 : 0;

              return (
                <tr key={row.teamId} className="bg-accent">
                  <td className="px-2 py-3 font-medium text-foreground rounded-l-md whitespace-nowrap">
                    {row.seed ?? '\u2014'}
                  </td>
                  <td className="px-2 py-3 text-foreground whitespace-nowrap">{row.region}</td>
                  <td className="px-2 py-3 text-foreground font-medium whitespace-nowrap truncate">{row.teamName}</td>
                  <td className="px-2 py-3">
                    <SegmentedBar
                      barWidthPct={pointsWidthPct}
                      segments={row.pointsSegments.map((seg, idx) => ({
                        key: `${row.teamId}-points-${seg.portfolioId}-${idx}`,
                        label: seg.portfolioName,
                        value: seg.amount,
                        color: portfolioColorById.get(seg.portfolioId) || '#94A3B8',
                      }))}
                      backgroundColor="#F3F4F6"
                      disabled={row.isEliminated}
                      getTooltipTitle={(seg) => seg.label}
                      getTooltipValue={(seg) => seg.value.toFixed(2)}
                    />
                  </td>
                  <td className="px-2 py-3 text-right font-medium text-foreground rounded-r-md whitespace-nowrap">
                    {row.totalPoints.toFixed(2)}
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
