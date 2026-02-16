import { useMemo, useState } from 'react';
import { CalcuttaEntry, CalcuttaPortfolio, CalcuttaPortfolioTeam, TournamentTeam } from '../../types/calcutta';
import { School } from '../../types/school';
import { SegmentedBar } from '../../components/SegmentedBar';
import { getEntryColorById } from '../../utils/entryColors';

export const ReturnsTab: React.FC<{
  entries: CalcuttaEntry[];
  schools: School[];
  tournamentTeams: TournamentTeam[];
  allCalcuttaPortfolios: (CalcuttaPortfolio & { entryName?: string })[];
  allCalcuttaPortfolioTeams: CalcuttaPortfolioTeam[];
}> = ({ entries, schools, tournamentTeams, allCalcuttaPortfolios, allCalcuttaPortfolioTeams }) => {
  const [returnsSortBy, setReturnsSortBy] = useState<'points' | 'seed' | 'region' | 'team'>('points');

  const entryNameById = useMemo(() => new Map(entries.map((entry) => [entry.id, entry.name])), [entries]);
  const schoolNameById = useMemo(() => new Map(schools.map((school) => [school.id, school.name])), [schools]);

  const entryColorById = useMemo(() => getEntryColorById(entries), [entries]);

  const portfolioEntryIdById = useMemo(() => {
    const map = new Map<string, string>();
    allCalcuttaPortfolios.forEach((p) => {
      map.set(p.id, p.entryId);
    });
    return map;
  }, [allCalcuttaPortfolios]);

  const returnsRows = useMemo(() => {
    const byTeam = new Map<
      string,
      {
        teamId: string;
        seed: number | undefined;
        region: string;
        teamName: string;
        eliminated: boolean;
        pointsSegments: { entryId: string; entryName: string; amount: number }[];
        possibleSegments: { entryId: string; entryName: string; amount: number }[];
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
        eliminated: tt.eliminated === true,
        pointsSegments: [],
        possibleSegments: [],
        totalPoints: 0,
        totalPossible: 0,
      });
    }

    const teamEntryAgg = new Map<string, Map<string, { actual: number; expected: number }>>();
    for (const pt of allCalcuttaPortfolioTeams) {
      const entryId = portfolioEntryIdById.get(pt.portfolioId);
      if (!entryId) continue;
      if (!teamEntryAgg.has(pt.teamId)) teamEntryAgg.set(pt.teamId, new Map());
      const byEntry = teamEntryAgg.get(pt.teamId)!;
      const current = byEntry.get(entryId) || { actual: 0, expected: 0 };
      current.actual += pt.actualPoints || 0;
      current.expected += pt.expectedPoints || 0;
      byEntry.set(entryId, current);
    }

    for (const [teamId, byEntry] of teamEntryAgg.entries()) {
      const row = byTeam.get(teamId);
      if (!row) continue;

      const pointsSegments: { entryId: string; entryName: string; amount: number }[] = [];
      const possibleSegments: { entryId: string; entryName: string; amount: number }[] = [];

      for (const [entryId, agg] of byEntry.entries()) {
        const entryName = entryNameById.get(entryId) || 'Unknown Entry';
        const actual = agg.actual;
        const possible = row.eliminated ? actual : Math.max(agg.expected, actual);

        if (actual > 0) {
          pointsSegments.push({ entryId, entryName, amount: actual });
        }
        if (possible > 0) {
          possibleSegments.push({ entryId, entryName, amount: possible });
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
  }, [allCalcuttaPortfolioTeams, entryNameById, portfolioEntryIdById, schoolNameById, tournamentTeams]);

  const returnsSortedRows = useMemo(() => {
    const rows = returnsRows.rows.slice();
    const seedA = (seed: number | undefined) => seed ?? 999;
    const regionA = (region: string) => region || '';
    const teamA = (name: string) => name || '';

    rows.sort((a, b) => {
      if (returnsSortBy === 'points') {
        if (b.totalPoints !== a.totalPoints) return b.totalPoints - a.totalPoints;
        if (b.totalPossible !== a.totalPossible) return b.totalPossible - a.totalPossible;
        const seedDiff = seedA(a.seed) - seedA(b.seed);
        if (seedDiff !== 0) return seedDiff;
        return teamA(a.teamName).localeCompare(teamA(b.teamName));
      }

      if (returnsSortBy === 'seed') {
        const seedDiff = seedA(a.seed) - seedA(b.seed);
        if (seedDiff !== 0) return seedDiff;
        const regionDiff = regionA(a.region).localeCompare(regionA(b.region));
        if (regionDiff !== 0) return regionDiff;
        return teamA(a.teamName).localeCompare(teamA(b.teamName));
      }

      if (returnsSortBy === 'region') {
        const regionDiff = regionA(a.region).localeCompare(regionA(b.region));
        if (regionDiff !== 0) return regionDiff;
        const seedDiff = seedA(a.seed) - seedA(b.seed);
        if (seedDiff !== 0) return seedDiff;
        return teamA(a.teamName).localeCompare(teamA(b.teamName));
      }

      const nameDiff = teamA(a.teamName).localeCompare(teamA(b.teamName));
      if (nameDiff !== 0) return nameDiff;
      return seedA(a.seed) - seedA(b.seed);
    });

    return rows;
  }, [returnsRows.rows, returnsSortBy]);

  return (
    <div className="bg-white rounded-lg shadow p-6">
      <div className="mb-4 flex items-center justify-end">
        <label className="text-sm text-gray-600">
          Sort by
          <select
            className="ml-2 rounded-md border border-gray-300 bg-white px-3 py-2 text-sm text-gray-900 shadow-sm"
            value={returnsSortBy}
            onChange={(e) => setReturnsSortBy(e.target.value as 'points' | 'seed' | 'region' | 'team')}
          >
            <option value="points">Points</option>
            <option value="seed">Seed</option>
            <option value="region">Region</option>
            <option value="team">Team</option>
          </select>
        </label>
      </div>
      <div className="mt-2 overflow-x-auto">
        <table className="min-w-full table-fixed border-separate border-spacing-y-2">
          <thead>
            <tr className="text-left text-xs uppercase tracking-wide text-gray-500">
              <th className="px-2 py-2 w-14">Seed</th>
              <th className="px-2 py-2 w-20">Region</th>
              <th className="px-2 py-2 w-44">Team</th>
              <th className="px-2 py-2"></th>
              <th className="px-2 py-2 w-32 text-right">Total Points</th>
            </tr>
          </thead>
          <tbody>
            {returnsSortedRows.map((row) => {
              const pointsWidthPct = returnsRows.maxTotal > 0 ? (row.totalPoints / returnsRows.maxTotal) * 100 : 0;

              return (
                <tr key={row.teamId} className="bg-gray-50">
                  <td className="px-2 py-3 font-medium text-gray-900 rounded-l-md whitespace-nowrap">{row.seed ?? '—'}</td>
                  <td className="px-2 py-3 text-gray-700 whitespace-nowrap">{row.region}</td>
                  <td className="px-2 py-3 text-gray-900 font-medium whitespace-nowrap truncate">{row.teamName}</td>
                  <td className="px-2 py-3">
                    <SegmentedBar
                      barWidthPct={pointsWidthPct}
                      segments={row.pointsSegments.map((seg, idx) => ({
                        key: `${row.teamId}-points-${seg.entryId}-${idx}`,
                        label: seg.entryName,
                        value: seg.amount,
                        color: entryColorById.get(seg.entryId) || '#94A3B8',
                      }))}
                      backgroundColor="#F3F4F6"
                      disabled={row.eliminated}
                      getTooltipTitle={(seg) => seg.label}
                      getTooltipValue={(seg) => seg.value.toFixed(2)}
                    />
                  </td>
                  <td className="px-2 py-3 text-right font-medium text-gray-900 rounded-r-md whitespace-nowrap">{row.totalPoints.toFixed(2)}</td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
    </div>
  );
};
