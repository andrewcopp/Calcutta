import { useMemo, useState } from 'react';
import { CalcuttaEntry, CalcuttaEntryTeam, TournamentTeam } from '../../types/calcutta';
import { School } from '../../types/school';
import { Select } from '../../components/ui/Select';
import { SegmentedBar } from '../../components/SegmentedBar';
import { getEntryColorById } from '../../utils/entryColors';

export const InvestmentTab: React.FC<{
  entries: CalcuttaEntry[];
  schools: School[];
  tournamentTeams: TournamentTeam[];
  allEntryTeams: CalcuttaEntryTeam[];
}> = ({ entries, schools, tournamentTeams, allEntryTeams }) => {
  const [investmentSortBy, setInvestmentSortBy] = useState<'total' | 'seed' | 'region' | 'team'>('total');

  const entryNameById = useMemo(() => new Map(entries.map((entry) => [entry.id, entry.name])), [entries]);
  const tournamentTeamById = useMemo(() => new Map(tournamentTeams.map((team) => [team.id, team])), [tournamentTeams]);
  const schoolNameById = useMemo(() => new Map(schools.map((school) => [school.id, school.name])), [schools]);

  const entryColorById = useMemo(() => getEntryColorById(entries), [entries]);

  const investmentRows = useMemo(() => {
    const byTeam = new Map<
      string,
      {
        teamId: string;
        seed: number | undefined;
        region: string;
        teamName: string;
        totalInvestment: number;
        segments: { entryId: string; entryName: string; amount: number }[];
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

    for (const entryTeam of allEntryTeams) {
      const amount = entryTeam.bid;
      if (amount <= 0) continue;

      const teamId = entryTeam.teamId;
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
              entryId: entryTeam.entryId,
              entryName: entryNameById.get(entryTeam.entryId) || 'Unknown Entry',
              amount,
            },
          ],
        });
        continue;
      }

      row.totalInvestment += amount;
      row.segments.push({
        entryId: entryTeam.entryId,
        entryName: entryNameById.get(entryTeam.entryId) || 'Unknown Entry',
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
  }, [allEntryTeams, entryNameById, schoolNameById, tournamentTeams, tournamentTeamById]);

  const investmentSortedRows = useMemo(() => {
    const rows = investmentRows.rows.slice();

    const seedA = (seed: number | undefined) => seed ?? 999;
    const regionA = (region: string) => region || '';
    const teamA = (name: string) => name || '';

    rows.sort((a, b) => {
      if (investmentSortBy === 'total') {
        if (b.totalInvestment !== a.totalInvestment) return b.totalInvestment - a.totalInvestment;
        return seedA(a.seed) - seedA(b.seed);
      }

      if (investmentSortBy === 'seed') {
        const seedDiff = seedA(a.seed) - seedA(b.seed);
        if (seedDiff !== 0) return seedDiff;
        const regionDiff = regionA(a.region).localeCompare(regionA(b.region));
        if (regionDiff !== 0) return regionDiff;
        return teamA(a.teamName).localeCompare(teamA(b.teamName));
      }

      if (investmentSortBy === 'region') {
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
  }, [investmentRows.rows, investmentSortBy]);

  return (
    <div className="bg-white rounded-lg shadow p-6">
      <div className="mb-4 flex items-center justify-end">
        <label className="text-sm text-gray-600">
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
            <tr className="text-left text-xs uppercase tracking-wide text-gray-500">
              <th className="px-2 py-2 w-14">Seed</th>
              <th className="px-2 py-2 w-20">Region</th>
              <th className="px-2 py-2 w-44">Team</th>
              <th className="px-2 py-2"></th>
              <th className="px-2 py-2 w-24 text-right">Total Investment</th>
            </tr>
          </thead>
          <tbody>
            {investmentSortedRows.map((row) => {
              const barWidthPct = investmentRows.maxTotal > 0 ? (row.totalInvestment / investmentRows.maxTotal) * 100 : 0;
              return (
                <tr key={row.teamId} className="bg-gray-50">
                  <td className="px-2 py-3 font-medium text-gray-900 rounded-l-md whitespace-nowrap">{row.seed ?? '—'}</td>
                  <td className="px-2 py-3 text-gray-700 whitespace-nowrap">{row.region}</td>
                  <td className="px-2 py-3 text-gray-900 font-medium whitespace-nowrap truncate">{row.teamName}</td>
                  <td className="px-2 py-3">
                    <SegmentedBar
                      barWidthPct={barWidthPct}
                      segments={row.segments.map((seg, idx) => ({
                        key: `${row.teamId}-${seg.entryId}-${idx}`,
                        label: seg.entryName,
                        value: seg.amount,
                        color: entryColorById.get(seg.entryId) || '#94A3B8',
                      }))}
                      backgroundColor="#E5E7EB"
                      getTooltipTitle={(seg) => seg.label}
                      getTooltipValue={(seg) => `${seg.value.toFixed(2)} pts`}
                    />
                  </td>
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
