import { describe, it, expect } from 'vitest';
import type { Investment, OwnershipSummary, OwnershipDetail } from '../schemas/pool';
import type { TournamentTeam } from '../schemas/tournament';
import { makeInvestment, makeOwnershipSummary, makeOwnershipDetail, makeTournamentTeam } from '../test/factories';

// ---------------------------------------------------------------------------
// The hook wraps useCallback/useMemo over three computations:
//   1. getOwnershipDetailData  -- find an ownership detail by teamId for the user
//   2. getInvestorRanking      -- rank the user among all investors of a team
//   3. ownershipTeamsData      -- filter and sort teams for the ownership tab
//
// We extract each computation into a pure function that mirrors the hook
// logic so we can test in a node environment without React rendering.
// ---------------------------------------------------------------------------

interface InvestorRanking {
  rank: number;
  total: number;
}

function computeGetOwnershipDetailData(
  teamId: string,
  ownershipSummaries: OwnershipSummary[],
  allOwnershipDetails: OwnershipDetail[],
): OwnershipDetail | undefined {
  const currentOwnershipSummaryId = ownershipSummaries[0]?.id;
  if (!currentOwnershipSummaryId) return undefined;
  return allOwnershipDetails.find((pt) => pt.teamId === teamId && pt.portfolioId === currentOwnershipSummaryId);
}

function computeGetInvestorRanking(
  teamId: string,
  ownershipSummaries: OwnershipSummary[],
  allOwnershipDetails: OwnershipDetail[],
): InvestorRanking {
  const allInvestors = allOwnershipDetails.filter((pt) => pt.teamId === teamId);
  const sortedInvestors = [...allInvestors].sort((a, b) => b.ownershipPercentage - a.ownershipPercentage);

  const userOwnershipSummary = ownershipSummaries[0];
  const userRank = userOwnershipSummary ? sortedInvestors.findIndex((pt) => pt.portfolioId === userOwnershipSummary.id) + 1 : 0;

  return {
    rank: userRank,
    total: allInvestors.length,
  };
}

function computeOwnershipTeamsData({
  activeTab,
  portfolioId,
  teams,
  schools,
  tournamentTeams,
  ownershipSummaries,
  allOwnershipDetails,
  ownershipShowAllTeams,
  sortBy,
}: {
  activeTab: string;
  portfolioId: string | undefined;
  teams: Investment[];
  schools: { id: string; name: string }[];
  tournamentTeams: TournamentTeam[];
  ownershipSummaries: OwnershipSummary[];
  allOwnershipDetails: OwnershipDetail[];
  ownershipShowAllTeams: boolean;
  sortBy: 'points' | 'ownership' | 'credits';
}): Investment[] {
  if (activeTab !== 'ownerships') return [];
  if (!portfolioId) return [];

  const getOwnershipDetailData = (teamId: string): OwnershipDetail | undefined => {
    const currentOwnershipSummaryId = ownershipSummaries[0]?.id;
    if (!currentOwnershipSummaryId) return undefined;
    return allOwnershipDetails.find((pt) => pt.teamId === teamId && pt.portfolioId === currentOwnershipSummaryId);
  };

  let teamsToShow: Investment[];

  if (ownershipShowAllTeams) {
    const schoolMap = new Map(schools.map((s) => [s.id, s]));
    teamsToShow = tournamentTeams.map((tt) => {
      const existingTeam = teams.find((t) => t.teamId === tt.id);
      if (existingTeam) return existingTeam;

      return {
        id: `synthetic-${tt.id}`,
        portfolioId: portfolioId,
        teamId: tt.id,
        credits: 0,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
        team: {
          ...tt,
          school: schoolMap.get(tt.schoolId),
        },
      } as Investment;
    });
  } else {
    teamsToShow = teams.filter((team) => {
      const ownershipDetail = getOwnershipDetailData(team.teamId);
      return ownershipDetail && ownershipDetail.ownershipPercentage > 0;
    });
  }

  teamsToShow = teamsToShow.slice().sort((a, b) => {
    const detailA = getOwnershipDetailData(a.teamId);
    const detailB = getOwnershipDetailData(b.teamId);

    const pointsA = detailA?.actualReturns || 0;
    const pointsB = detailB?.actualReturns || 0;
    const ownershipA = detailA?.ownershipPercentage || 0;
    const ownershipB = detailB?.ownershipPercentage || 0;
    const creditsA = a.credits;
    const creditsB = b.credits;

    if (sortBy === 'points') {
      if (pointsB !== pointsA) return pointsB - pointsA;
      if (ownershipB !== ownershipA) return ownershipB - ownershipA;
      return creditsB - creditsA;
    }

    if (sortBy === 'ownership') {
      if (ownershipB !== ownershipA) return ownershipB - ownershipA;
      if (pointsB !== pointsA) return pointsB - pointsA;
      return creditsB - creditsA;
    }

    if (creditsB !== creditsA) return creditsB - creditsA;
    if (pointsB !== pointsA) return pointsB - pointsA;
    return ownershipB - ownershipA;
  });

  return teamsToShow;
}

// ---------------------------------------------------------------------------
// Default inputs for computeOwnershipTeamsData
// ---------------------------------------------------------------------------

interface OwnershipInput {
  activeTab: string;
  portfolioId: string | undefined;
  teams: Investment[];
  schools: { id: string; name: string }[];
  tournamentTeams: TournamentTeam[];
  ownershipSummaries: OwnershipSummary[];
  allOwnershipDetails: OwnershipDetail[];
  ownershipShowAllTeams: boolean;
  sortBy: 'points' | 'ownership' | 'credits';
}

function makeOwnershipInput(overrides: Partial<OwnershipInput> = {}): OwnershipInput {
  return {
    activeTab: 'ownerships',
    portfolioId: 'p1',
    teams: [],
    schools: [],
    tournamentTeams: [],
    ownershipSummaries: [],
    allOwnershipDetails: [],
    ownershipShowAllTeams: false,
    sortBy: 'ownership',
    ...overrides,
  };
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

describe('usePortfolioOwnershipData (pure transformation)', () => {
  // =========================================================================
  // getOwnershipDetailData
  // =========================================================================

  describe('getOwnershipDetailData', () => {
    it('returns undefined when ownershipSummaries array is empty', () => {
      const ownershipSummaries: OwnershipSummary[] = [];
      const ownershipDetails: OwnershipDetail[] = [];
      const result = computeGetOwnershipDetailData('t1', ownershipSummaries, ownershipDetails);
      expect(result).toBeUndefined();
    });

    it('returns the ownership detail matching the first summary and the given team', () => {
      const ownershipSummaries = [makeOwnershipSummary({ id: 'os1', portfolioId: 'p1' })];
      const ownershipDetails = [
        makeOwnershipDetail({ id: 'od1', portfolioId: 'os1', teamId: 't1', ownershipPercentage: 0.75 }),
      ];
      const result = computeGetOwnershipDetailData('t1', ownershipSummaries, ownershipDetails);
      expect(result?.ownershipPercentage).toBe(0.75);
    });

    it('returns undefined when team is not in the first summary', () => {
      const ownershipSummaries = [makeOwnershipSummary({ id: 'os1', portfolioId: 'p1' })];
      const ownershipDetails = [makeOwnershipDetail({ id: 'od1', portfolioId: 'os2', teamId: 't1' })];
      const result = computeGetOwnershipDetailData('t1', ownershipSummaries, ownershipDetails);
      expect(result).toBeUndefined();
    });

    it('uses only the first summary when multiple summaries exist', () => {
      const ownershipSummaries = [
        makeOwnershipSummary({ id: 'os1', portfolioId: 'p1' }),
        makeOwnershipSummary({ id: 'os2', portfolioId: 'p2' }),
      ];
      const ownershipDetails = [
        makeOwnershipDetail({ id: 'od1', portfolioId: 'os1', teamId: 't1', ownershipPercentage: 0.5 }),
        makeOwnershipDetail({ id: 'od2', portfolioId: 'os2', teamId: 't1', ownershipPercentage: 0.8 }),
      ];
      const result = computeGetOwnershipDetailData('t1', ownershipSummaries, ownershipDetails);
      expect(result?.ownershipPercentage).toBe(0.5);
    });

    it('returns undefined for a team that has no ownership detail record', () => {
      const ownershipSummaries = [makeOwnershipSummary({ id: 'os1', portfolioId: 'p1' })];
      const ownershipDetails = [makeOwnershipDetail({ id: 'od1', portfolioId: 'os1', teamId: 't1' })];
      const result = computeGetOwnershipDetailData('t2', ownershipSummaries, ownershipDetails);
      expect(result).toBeUndefined();
    });
  });

  // =========================================================================
  // getInvestorRanking
  // =========================================================================

  describe('getInvestorRanking', () => {
    it('returns rank 0 and total 0 when no investors own the team', () => {
      const ownershipSummaries = [makeOwnershipSummary({ id: 'os1', portfolioId: 'p1' })];
      const ownershipDetails: OwnershipDetail[] = [];
      const result = computeGetInvestorRanking('t1', ownershipSummaries, ownershipDetails);
      expect(result).toEqual({ rank: 0, total: 0 });
    });

    it('returns rank 0 when ownershipSummaries array is empty', () => {
      const ownershipSummaries: OwnershipSummary[] = [];
      const ownershipDetails = [
        makeOwnershipDetail({ id: 'od1', portfolioId: 'os1', teamId: 't1', ownershipPercentage: 0.5 }),
      ];
      const result = computeGetInvestorRanking('t1', ownershipSummaries, ownershipDetails);
      expect(result.rank).toBe(0);
    });

    it('returns correct total count of investors for a team', () => {
      const ownershipSummaries = [makeOwnershipSummary({ id: 'os1', portfolioId: 'p1' })];
      const ownershipDetails = [
        makeOwnershipDetail({ id: 'od1', portfolioId: 'os1', teamId: 't1', ownershipPercentage: 0.5 }),
        makeOwnershipDetail({ id: 'od2', portfolioId: 'os2', teamId: 't1', ownershipPercentage: 0.3 }),
        makeOwnershipDetail({ id: 'od3', portfolioId: 'os3', teamId: 't1', ownershipPercentage: 0.2 }),
      ];
      const result = computeGetInvestorRanking('t1', ownershipSummaries, ownershipDetails);
      expect(result.total).toBe(3);
    });

    it('ranks the user first when they have the highest ownership', () => {
      const ownershipSummaries = [makeOwnershipSummary({ id: 'os1', portfolioId: 'p1' })];
      const ownershipDetails = [
        makeOwnershipDetail({ id: 'od1', portfolioId: 'os1', teamId: 't1', ownershipPercentage: 0.6 }),
        makeOwnershipDetail({ id: 'od2', portfolioId: 'os2', teamId: 't1', ownershipPercentage: 0.3 }),
        makeOwnershipDetail({ id: 'od3', portfolioId: 'os3', teamId: 't1', ownershipPercentage: 0.1 }),
      ];
      const result = computeGetInvestorRanking('t1', ownershipSummaries, ownershipDetails);
      expect(result.rank).toBe(1);
    });

    it('ranks the user second when another investor has higher ownership', () => {
      const ownershipSummaries = [makeOwnershipSummary({ id: 'os1', portfolioId: 'p1' })];
      const ownershipDetails = [
        makeOwnershipDetail({ id: 'od1', portfolioId: 'os1', teamId: 't1', ownershipPercentage: 0.3 }),
        makeOwnershipDetail({ id: 'od2', portfolioId: 'os2', teamId: 't1', ownershipPercentage: 0.7 }),
      ];
      const result = computeGetInvestorRanking('t1', ownershipSummaries, ownershipDetails);
      expect(result.rank).toBe(2);
    });

    it('ranks the user as sole investor when only one investor exists', () => {
      const ownershipSummaries = [makeOwnershipSummary({ id: 'os1', portfolioId: 'p1' })];
      const ownershipDetails = [
        makeOwnershipDetail({ id: 'od1', portfolioId: 'os1', teamId: 't1', ownershipPercentage: 1.0 }),
      ];
      const result = computeGetInvestorRanking('t1', ownershipSummaries, ownershipDetails);
      expect(result).toEqual({ rank: 1, total: 1 });
    });

    it('only counts investors for the specified team', () => {
      const ownershipSummaries = [makeOwnershipSummary({ id: 'os1', portfolioId: 'p1' })];
      const ownershipDetails = [
        makeOwnershipDetail({ id: 'od1', portfolioId: 'os1', teamId: 't1', ownershipPercentage: 0.5 }),
        makeOwnershipDetail({ id: 'od2', portfolioId: 'os2', teamId: 't1', ownershipPercentage: 0.5 }),
        makeOwnershipDetail({ id: 'od3', portfolioId: 'os3', teamId: 't2', ownershipPercentage: 1.0 }),
      ];
      const result = computeGetInvestorRanking('t1', ownershipSummaries, ownershipDetails);
      expect(result.total).toBe(2);
    });
  });

  // =========================================================================
  // ownershipTeamsData -- early returns
  // =========================================================================

  describe('ownershipTeamsData when not on ownerships tab', () => {
    it('returns empty array when activeTab is not ownerships', () => {
      const input = makeOwnershipInput({ activeTab: 'standings' });
      const result = computeOwnershipTeamsData(input);
      expect(result).toEqual([]);
    });
  });

  describe('ownershipTeamsData when portfolioId is undefined', () => {
    it('returns empty array when portfolioId is undefined', () => {
      const input = makeOwnershipInput({ portfolioId: undefined });
      const result = computeOwnershipTeamsData(input);
      expect(result).toEqual([]);
    });
  });

  // =========================================================================
  // ownershipTeamsData -- filtering (ownershipShowAllTeams = false)
  // =========================================================================

  describe('ownershipTeamsData filtering with ownershipShowAllTeams=false', () => {
    it('includes teams with positive ownership percentage', () => {
      const input = makeOwnershipInput({
        ownershipShowAllTeams: false,
        ownershipSummaries: [makeOwnershipSummary({ id: 'os1', portfolioId: 'p1' })],
        teams: [makeInvestment({ id: 'i1', portfolioId: 'p1', teamId: 't1' })],
        allOwnershipDetails: [
          makeOwnershipDetail({ id: 'od1', portfolioId: 'os1', teamId: 't1', ownershipPercentage: 0.5 }),
        ],
      });
      const result = computeOwnershipTeamsData(input);
      expect(result).toHaveLength(1);
    });

    it('excludes teams with zero ownership percentage', () => {
      const input = makeOwnershipInput({
        ownershipShowAllTeams: false,
        ownershipSummaries: [makeOwnershipSummary({ id: 'os1', portfolioId: 'p1' })],
        teams: [makeInvestment({ id: 'i1', portfolioId: 'p1', teamId: 't1' })],
        allOwnershipDetails: [
          makeOwnershipDetail({ id: 'od1', portfolioId: 'os1', teamId: 't1', ownershipPercentage: 0 }),
        ],
      });
      const result = computeOwnershipTeamsData(input);
      expect(result).toEqual([]);
    });

    it('excludes teams with no ownership detail record', () => {
      const input = makeOwnershipInput({
        ownershipShowAllTeams: false,
        ownershipSummaries: [makeOwnershipSummary({ id: 'os1', portfolioId: 'p1' })],
        teams: [makeInvestment({ id: 'i1', portfolioId: 'p1', teamId: 't1' })],
        allOwnershipDetails: [],
      });
      const result = computeOwnershipTeamsData(input);
      expect(result).toEqual([]);
    });
  });

  // =========================================================================
  // ownershipTeamsData -- show all teams mode
  // =========================================================================

  describe('ownershipTeamsData with ownershipShowAllTeams=true', () => {
    it('returns existing investment when a match exists in tournament teams', () => {
      const existingInvestment = makeInvestment({ id: 'i1', portfolioId: 'p1', teamId: 't1', credits: 25 });
      const input = makeOwnershipInput({
        ownershipShowAllTeams: true,
        teams: [existingInvestment],
        tournamentTeams: [makeTournamentTeam({ id: 't1', schoolId: 's1' })],
        schools: [{ id: 's1', name: 'Duke' }],
      });
      const result = computeOwnershipTeamsData(input);
      expect(result[0].credits).toBe(25);
    });

    it('creates synthetic investment with zero credits for unowned tournament teams', () => {
      const input = makeOwnershipInput({
        ownershipShowAllTeams: true,
        teams: [],
        tournamentTeams: [makeTournamentTeam({ id: 't1', schoolId: 's1' })],
        schools: [{ id: 's1', name: 'Duke' }],
      });
      const result = computeOwnershipTeamsData(input);
      expect(result[0].credits).toBe(0);
    });

    it('assigns synthetic team ID with "synthetic-" prefix', () => {
      const input = makeOwnershipInput({
        ownershipShowAllTeams: true,
        teams: [],
        tournamentTeams: [makeTournamentTeam({ id: 't1', schoolId: 's1' })],
        schools: [{ id: 's1', name: 'Duke' }],
      });
      const result = computeOwnershipTeamsData(input);
      expect(result[0].id).toBe('synthetic-t1');
    });

    it('attaches school info to synthetic teams', () => {
      const input = makeOwnershipInput({
        ownershipShowAllTeams: true,
        teams: [],
        tournamentTeams: [makeTournamentTeam({ id: 't1', schoolId: 's1' })],
        schools: [{ id: 's1', name: 'Duke' }],
      });
      const result = computeOwnershipTeamsData(input);
      expect(result[0].team?.school?.name).toBe('Duke');
    });

    it('includes all tournament teams in the result', () => {
      const input = makeOwnershipInput({
        ownershipShowAllTeams: true,
        teams: [makeInvestment({ id: 'i1', portfolioId: 'p1', teamId: 't1' })],
        tournamentTeams: [
          makeTournamentTeam({ id: 't1', schoolId: 's1' }),
          makeTournamentTeam({ id: 't2', schoolId: 's2' }),
          makeTournamentTeam({ id: 't3', schoolId: 's3' }),
        ],
        schools: [
          { id: 's1', name: 'Duke' },
          { id: 's2', name: 'UNC' },
          { id: 's3', name: 'Kansas' },
        ],
      });
      const result = computeOwnershipTeamsData(input);
      expect(result).toHaveLength(3);
    });
  });

  // =========================================================================
  // ownershipTeamsData -- sorting by ownership
  // =========================================================================

  describe('ownershipTeamsData sorting by ownership', () => {
    it('sorts teams by ownership percentage descending', () => {
      const input = makeOwnershipInput({
        sortBy: 'ownership',
        ownershipShowAllTeams: false,
        ownershipSummaries: [makeOwnershipSummary({ id: 'os1', portfolioId: 'p1' })],
        teams: [
          makeInvestment({ id: 'i1', portfolioId: 'p1', teamId: 't1' }),
          makeInvestment({ id: 'i2', portfolioId: 'p1', teamId: 't2' }),
        ],
        allOwnershipDetails: [
          makeOwnershipDetail({ id: 'od1', portfolioId: 'os1', teamId: 't1', ownershipPercentage: 0.3 }),
          makeOwnershipDetail({ id: 'od2', portfolioId: 'os1', teamId: 't2', ownershipPercentage: 0.7 }),
        ],
      });
      const result = computeOwnershipTeamsData(input);
      expect(result[0].teamId).toBe('t2');
    });

    it('breaks ownership ties by points descending', () => {
      const input = makeOwnershipInput({
        sortBy: 'ownership',
        ownershipShowAllTeams: false,
        ownershipSummaries: [makeOwnershipSummary({ id: 'os1', portfolioId: 'p1' })],
        teams: [
          makeInvestment({ id: 'i1', portfolioId: 'p1', teamId: 't1' }),
          makeInvestment({ id: 'i2', portfolioId: 'p1', teamId: 't2' }),
        ],
        allOwnershipDetails: [
          makeOwnershipDetail({ id: 'od1', portfolioId: 'os1', teamId: 't1', ownershipPercentage: 0.5, actualReturns: 10 }),
          makeOwnershipDetail({ id: 'od2', portfolioId: 'os1', teamId: 't2', ownershipPercentage: 0.5, actualReturns: 30 }),
        ],
      });
      const result = computeOwnershipTeamsData(input);
      expect(result[0].teamId).toBe('t2');
    });
  });

  // =========================================================================
  // ownershipTeamsData -- sorting by points
  // =========================================================================

  describe('ownershipTeamsData sorting by points', () => {
    it('sorts teams by actual returns descending', () => {
      const input = makeOwnershipInput({
        sortBy: 'points',
        ownershipShowAllTeams: false,
        ownershipSummaries: [makeOwnershipSummary({ id: 'os1', portfolioId: 'p1' })],
        teams: [
          makeInvestment({ id: 'i1', portfolioId: 'p1', teamId: 't1' }),
          makeInvestment({ id: 'i2', portfolioId: 'p1', teamId: 't2' }),
        ],
        allOwnershipDetails: [
          makeOwnershipDetail({ id: 'od1', portfolioId: 'os1', teamId: 't1', actualReturns: 50, ownershipPercentage: 0.5 }),
          makeOwnershipDetail({ id: 'od2', portfolioId: 'os1', teamId: 't2', actualReturns: 80, ownershipPercentage: 0.5 }),
        ],
      });
      const result = computeOwnershipTeamsData(input);
      expect(result[0].teamId).toBe('t2');
    });
  });

  // =========================================================================
  // ownershipTeamsData -- sorting by credits
  // =========================================================================

  describe('ownershipTeamsData sorting by credits', () => {
    it('sorts teams by credits descending', () => {
      const input = makeOwnershipInput({
        sortBy: 'credits',
        ownershipShowAllTeams: false,
        ownershipSummaries: [makeOwnershipSummary({ id: 'os1', portfolioId: 'p1' })],
        teams: [
          makeInvestment({ id: 'i1', portfolioId: 'p1', teamId: 't1', credits: 30 }),
          makeInvestment({ id: 'i2', portfolioId: 'p1', teamId: 't2', credits: 10 }),
        ],
        allOwnershipDetails: [
          makeOwnershipDetail({ id: 'od1', portfolioId: 'os1', teamId: 't1', ownershipPercentage: 0.5 }),
          makeOwnershipDetail({ id: 'od2', portfolioId: 'os1', teamId: 't2', ownershipPercentage: 0.5 }),
        ],
      });
      const result = computeOwnershipTeamsData(input);
      expect(result[0].teamId).toBe('t1');
    });
  });

  // =========================================================================
  // ownershipTeamsData -- edge cases
  // =========================================================================

  describe('ownershipTeamsData edge cases', () => {
    it('returns empty array when no teams have ownership and showAll is false', () => {
      const input = makeOwnershipInput({
        ownershipShowAllTeams: false,
        ownershipSummaries: [makeOwnershipSummary({ id: 'os1', portfolioId: 'p1' })],
        teams: [
          makeInvestment({ id: 'i1', portfolioId: 'p1', teamId: 't1' }),
          makeInvestment({ id: 'i2', portfolioId: 'p1', teamId: 't2' }),
        ],
        allOwnershipDetails: [],
      });
      const result = computeOwnershipTeamsData(input);
      expect(result).toEqual([]);
    });

    it('handles a single team without errors', () => {
      const input = makeOwnershipInput({
        ownershipShowAllTeams: false,
        ownershipSummaries: [makeOwnershipSummary({ id: 'os1', portfolioId: 'p1' })],
        teams: [makeInvestment({ id: 'i1', portfolioId: 'p1', teamId: 't1' })],
        allOwnershipDetails: [
          makeOwnershipDetail({ id: 'od1', portfolioId: 'os1', teamId: 't1', ownershipPercentage: 1.0 }),
        ],
      });
      const result = computeOwnershipTeamsData(input);
      expect(result).toHaveLength(1);
    });
  });
});
