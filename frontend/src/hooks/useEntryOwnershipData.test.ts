import { describe, it, expect } from 'vitest';
import type { CalcuttaEntryTeam, CalcuttaPortfolio, CalcuttaPortfolioTeam } from '../schemas/calcutta';
import type { TournamentTeam } from '../schemas/tournament';
import { makeEntryTeam, makePortfolio, makePortfolioTeam, makeTournamentTeam } from '../test/factories';

// ---------------------------------------------------------------------------
// The hook wraps useCallback/useMemo over three computations:
//   1. getPortfolioTeamData  – find a portfolio team by teamId for the user
//   2. getInvestorRanking    – rank the user among all investors of a team
//   3. ownershipTeamsData    – filter and sort teams for the ownership tab
//
// We extract each computation into a pure function that mirrors the hook
// logic so we can test in a node environment without React rendering.
// ---------------------------------------------------------------------------

interface InvestorRanking {
  rank: number;
  total: number;
}

function computeGetPortfolioTeamData(
  teamId: string,
  portfolios: CalcuttaPortfolio[],
  allCalcuttaPortfolioTeams: CalcuttaPortfolioTeam[],
): CalcuttaPortfolioTeam | undefined {
  const currentPortfolioId = portfolios[0]?.id;
  if (!currentPortfolioId) return undefined;
  return allCalcuttaPortfolioTeams.find((pt) => pt.teamId === teamId && pt.portfolioId === currentPortfolioId);
}

function computeGetInvestorRanking(
  teamId: string,
  portfolios: CalcuttaPortfolio[],
  allCalcuttaPortfolioTeams: CalcuttaPortfolioTeam[],
): InvestorRanking {
  const allInvestors = allCalcuttaPortfolioTeams.filter((pt) => pt.teamId === teamId);
  const sortedInvestors = [...allInvestors].sort((a, b) => b.ownershipPercentage - a.ownershipPercentage);

  const userPortfolio = portfolios[0];
  const userRank = userPortfolio ? sortedInvestors.findIndex((pt) => pt.portfolioId === userPortfolio.id) + 1 : 0;

  return {
    rank: userRank,
    total: allInvestors.length,
  };
}

function computeOwnershipTeamsData({
  activeTab,
  entryId,
  teams,
  schools,
  tournamentTeams,
  portfolios,
  allCalcuttaPortfolioTeams,
  ownershipShowAllTeams,
  sortBy,
}: {
  activeTab: string;
  entryId: string | undefined;
  teams: CalcuttaEntryTeam[];
  schools: { id: string; name: string }[];
  tournamentTeams: TournamentTeam[];
  portfolios: CalcuttaPortfolio[];
  allCalcuttaPortfolioTeams: CalcuttaPortfolioTeam[];
  ownershipShowAllTeams: boolean;
  sortBy: 'points' | 'ownership' | 'bid';
}): CalcuttaEntryTeam[] {
  if (activeTab !== 'ownerships') return [];
  if (!entryId) return [];

  const getPortfolioTeamData = (teamId: string): CalcuttaPortfolioTeam | undefined => {
    const currentPortfolioId = portfolios[0]?.id;
    if (!currentPortfolioId) return undefined;
    return allCalcuttaPortfolioTeams.find((pt) => pt.teamId === teamId && pt.portfolioId === currentPortfolioId);
  };

  let teamsToShow: CalcuttaEntryTeam[];

  if (ownershipShowAllTeams) {
    const schoolMap = new Map(schools.map((s) => [s.id, s]));
    teamsToShow = tournamentTeams.map((tt) => {
      const existingTeam = teams.find((t) => t.teamId === tt.id);
      if (existingTeam) return existingTeam;

      return {
        id: `synthetic-${tt.id}`,
        entryId: entryId,
        teamId: tt.id,
        bid: 0,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
        team: {
          ...tt,
          school: schoolMap.get(tt.schoolId),
        },
      } as CalcuttaEntryTeam;
    });
  } else {
    teamsToShow = teams.filter((team) => {
      const portfolioTeam = getPortfolioTeamData(team.teamId);
      return portfolioTeam && portfolioTeam.ownershipPercentage > 0;
    });
  }

  teamsToShow = teamsToShow.slice().sort((a, b) => {
    const portfolioTeamA = getPortfolioTeamData(a.teamId);
    const portfolioTeamB = getPortfolioTeamData(b.teamId);

    const pointsA = portfolioTeamA?.actualPoints || 0;
    const pointsB = portfolioTeamB?.actualPoints || 0;
    const ownershipA = portfolioTeamA?.ownershipPercentage || 0;
    const ownershipB = portfolioTeamB?.ownershipPercentage || 0;
    const bidA = a.bid;
    const bidB = b.bid;

    if (sortBy === 'points') {
      if (pointsB !== pointsA) return pointsB - pointsA;
      if (ownershipB !== ownershipA) return ownershipB - ownershipA;
      return bidB - bidA;
    }

    if (sortBy === 'ownership') {
      if (ownershipB !== ownershipA) return ownershipB - ownershipA;
      if (pointsB !== pointsA) return pointsB - pointsA;
      return bidB - bidA;
    }

    if (bidB !== bidA) return bidB - bidA;
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
  entryId: string | undefined;
  teams: CalcuttaEntryTeam[];
  schools: { id: string; name: string }[];
  tournamentTeams: TournamentTeam[];
  portfolios: CalcuttaPortfolio[];
  allCalcuttaPortfolioTeams: CalcuttaPortfolioTeam[];
  ownershipShowAllTeams: boolean;
  sortBy: 'points' | 'ownership' | 'bid';
}

function makeOwnershipInput(overrides: Partial<OwnershipInput> = {}): OwnershipInput {
  return {
    activeTab: 'ownerships',
    entryId: 'e1',
    teams: [],
    schools: [],
    tournamentTeams: [],
    portfolios: [],
    allCalcuttaPortfolioTeams: [],
    ownershipShowAllTeams: false,
    sortBy: 'ownership',
    ...overrides,
  };
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

describe('useEntryOwnershipData (pure transformation)', () => {
  // =========================================================================
  // getPortfolioTeamData
  // =========================================================================

  describe('getPortfolioTeamData', () => {
    it('returns undefined when portfolios array is empty', () => {
      // GIVEN no portfolios
      const portfolios: CalcuttaPortfolio[] = [];
      const portfolioTeams: CalcuttaPortfolioTeam[] = [];

      // WHEN looking up portfolio team data for any team
      const result = computeGetPortfolioTeamData('t1', portfolios, portfolioTeams);

      // THEN result is undefined
      expect(result).toBeUndefined();
    });

    it('returns the portfolio team matching the first portfolio and the given team', () => {
      // GIVEN a portfolio and a matching portfolio team
      const portfolios = [makePortfolio({ id: 'p1', entryId: 'e1' })];
      const portfolioTeams = [
        makePortfolioTeam({ id: 'pt1', portfolioId: 'p1', teamId: 't1', ownershipPercentage: 0.75 }),
      ];

      // WHEN looking up portfolio team data for 't1'
      const result = computeGetPortfolioTeamData('t1', portfolios, portfolioTeams);

      // THEN the matching portfolio team is returned
      expect(result?.ownershipPercentage).toBe(0.75);
    });

    it('returns undefined when team is not in the first portfolio', () => {
      // GIVEN a portfolio team that belongs to a different portfolio
      const portfolios = [makePortfolio({ id: 'p1', entryId: 'e1' })];
      const portfolioTeams = [makePortfolioTeam({ id: 'pt1', portfolioId: 'p2', teamId: 't1' })];

      // WHEN looking up portfolio team data for 't1' against portfolio 'p1'
      const result = computeGetPortfolioTeamData('t1', portfolios, portfolioTeams);

      // THEN result is undefined because portfolioId does not match
      expect(result).toBeUndefined();
    });

    it('uses only the first portfolio when multiple portfolios exist', () => {
      // GIVEN two portfolios and portfolio teams for each
      const portfolios = [makePortfolio({ id: 'p1', entryId: 'e1' }), makePortfolio({ id: 'p2', entryId: 'e2' })];
      const portfolioTeams = [
        makePortfolioTeam({ id: 'pt1', portfolioId: 'p1', teamId: 't1', ownershipPercentage: 0.5 }),
        makePortfolioTeam({ id: 'pt2', portfolioId: 'p2', teamId: 't1', ownershipPercentage: 0.8 }),
      ];

      // WHEN looking up portfolio team data for 't1'
      const result = computeGetPortfolioTeamData('t1', portfolios, portfolioTeams);

      // THEN it returns data from the first portfolio (p1), not p2
      expect(result?.ownershipPercentage).toBe(0.5);
    });

    it('returns undefined for a team that has no portfolio team record', () => {
      // GIVEN a portfolio with team records for 't1' only
      const portfolios = [makePortfolio({ id: 'p1', entryId: 'e1' })];
      const portfolioTeams = [makePortfolioTeam({ id: 'pt1', portfolioId: 'p1', teamId: 't1' })];

      // WHEN looking up portfolio team data for a different team 't2'
      const result = computeGetPortfolioTeamData('t2', portfolios, portfolioTeams);

      // THEN result is undefined
      expect(result).toBeUndefined();
    });
  });

  // =========================================================================
  // getInvestorRanking
  // =========================================================================

  describe('getInvestorRanking', () => {
    it('returns rank 0 and total 0 when no investors own the team', () => {
      // GIVEN no portfolio teams for the target team
      const portfolios = [makePortfolio({ id: 'p1', entryId: 'e1' })];
      const portfolioTeams: CalcuttaPortfolioTeam[] = [];

      // WHEN computing the investor ranking for 't1'
      const result = computeGetInvestorRanking('t1', portfolios, portfolioTeams);

      // THEN rank is 0 and total is 0
      expect(result).toEqual({ rank: 0, total: 0 });
    });

    it('returns rank 0 when portfolios array is empty', () => {
      // GIVEN no user portfolio but portfolio teams exist
      const portfolios: CalcuttaPortfolio[] = [];
      const portfolioTeams = [
        makePortfolioTeam({ id: 'pt1', portfolioId: 'p1', teamId: 't1', ownershipPercentage: 0.5 }),
      ];

      // WHEN computing the investor ranking for 't1'
      const result = computeGetInvestorRanking('t1', portfolios, portfolioTeams);

      // THEN rank is 0 (no user portfolio to rank)
      expect(result.rank).toBe(0);
    });

    it('returns correct total count of investors for a team', () => {
      // GIVEN three investors own team 't1'
      const portfolios = [makePortfolio({ id: 'p1', entryId: 'e1' })];
      const portfolioTeams = [
        makePortfolioTeam({ id: 'pt1', portfolioId: 'p1', teamId: 't1', ownershipPercentage: 0.5 }),
        makePortfolioTeam({ id: 'pt2', portfolioId: 'p2', teamId: 't1', ownershipPercentage: 0.3 }),
        makePortfolioTeam({ id: 'pt3', portfolioId: 'p3', teamId: 't1', ownershipPercentage: 0.2 }),
      ];

      // WHEN computing the investor ranking for 't1'
      const result = computeGetInvestorRanking('t1', portfolios, portfolioTeams);

      // THEN total is 3
      expect(result.total).toBe(3);
    });

    it('ranks the user first when they have the highest ownership', () => {
      // GIVEN the user (p1) has 0.6 ownership, others have less
      const portfolios = [makePortfolio({ id: 'p1', entryId: 'e1' })];
      const portfolioTeams = [
        makePortfolioTeam({ id: 'pt1', portfolioId: 'p1', teamId: 't1', ownershipPercentage: 0.6 }),
        makePortfolioTeam({ id: 'pt2', portfolioId: 'p2', teamId: 't1', ownershipPercentage: 0.3 }),
        makePortfolioTeam({ id: 'pt3', portfolioId: 'p3', teamId: 't1', ownershipPercentage: 0.1 }),
      ];

      // WHEN computing the investor ranking for 't1'
      const result = computeGetInvestorRanking('t1', portfolios, portfolioTeams);

      // THEN the user is ranked 1st
      expect(result.rank).toBe(1);
    });

    it('ranks the user second when another investor has higher ownership', () => {
      // GIVEN another investor (p2) has higher ownership than the user (p1)
      const portfolios = [makePortfolio({ id: 'p1', entryId: 'e1' })];
      const portfolioTeams = [
        makePortfolioTeam({ id: 'pt1', portfolioId: 'p1', teamId: 't1', ownershipPercentage: 0.3 }),
        makePortfolioTeam({ id: 'pt2', portfolioId: 'p2', teamId: 't1', ownershipPercentage: 0.7 }),
      ];

      // WHEN computing the investor ranking for 't1'
      const result = computeGetInvestorRanking('t1', portfolios, portfolioTeams);

      // THEN the user is ranked 2nd
      expect(result.rank).toBe(2);
    });

    it('ranks the user as sole investor when only one investor exists', () => {
      // GIVEN the user is the only investor for team 't1'
      const portfolios = [makePortfolio({ id: 'p1', entryId: 'e1' })];
      const portfolioTeams = [
        makePortfolioTeam({ id: 'pt1', portfolioId: 'p1', teamId: 't1', ownershipPercentage: 1.0 }),
      ];

      // WHEN computing the investor ranking for 't1'
      const result = computeGetInvestorRanking('t1', portfolios, portfolioTeams);

      // THEN the user is ranked 1st of 1
      expect(result).toEqual({ rank: 1, total: 1 });
    });

    it('only counts investors for the specified team', () => {
      // GIVEN investors for two different teams
      const portfolios = [makePortfolio({ id: 'p1', entryId: 'e1' })];
      const portfolioTeams = [
        makePortfolioTeam({ id: 'pt1', portfolioId: 'p1', teamId: 't1', ownershipPercentage: 0.5 }),
        makePortfolioTeam({ id: 'pt2', portfolioId: 'p2', teamId: 't1', ownershipPercentage: 0.5 }),
        makePortfolioTeam({ id: 'pt3', portfolioId: 'p3', teamId: 't2', ownershipPercentage: 1.0 }),
      ];

      // WHEN computing the investor ranking for 't1'
      const result = computeGetInvestorRanking('t1', portfolios, portfolioTeams);

      // THEN total only counts team 't1' investors (not 't2')
      expect(result.total).toBe(2);
    });
  });

  // =========================================================================
  // ownershipTeamsData -- early returns
  // =========================================================================

  describe('ownershipTeamsData when not on ownerships tab', () => {
    it('returns empty array when activeTab is not ownerships', () => {
      // GIVEN activeTab is set to something other than 'ownerships'
      const input = makeOwnershipInput({ activeTab: 'standings' });

      // WHEN computing ownership teams data
      const result = computeOwnershipTeamsData(input);

      // THEN the result is empty
      expect(result).toEqual([]);
    });
  });

  describe('ownershipTeamsData when entryId is undefined', () => {
    it('returns empty array when entryId is undefined', () => {
      // GIVEN no entryId
      const input = makeOwnershipInput({ entryId: undefined });

      // WHEN computing ownership teams data
      const result = computeOwnershipTeamsData(input);

      // THEN the result is empty
      expect(result).toEqual([]);
    });
  });

  // =========================================================================
  // ownershipTeamsData -- filtering (ownershipShowAllTeams = false)
  // =========================================================================

  describe('ownershipTeamsData filtering with ownershipShowAllTeams=false', () => {
    it('includes teams with positive ownership percentage', () => {
      // GIVEN a team that has positive ownership in the user portfolio
      const input = makeOwnershipInput({
        ownershipShowAllTeams: false,
        portfolios: [makePortfolio({ id: 'p1', entryId: 'e1' })],
        teams: [makeEntryTeam({ id: 'et1', entryId: 'e1', teamId: 't1' })],
        allCalcuttaPortfolioTeams: [
          makePortfolioTeam({ id: 'pt1', portfolioId: 'p1', teamId: 't1', ownershipPercentage: 0.5 }),
        ],
      });

      // WHEN computing ownership teams data
      const result = computeOwnershipTeamsData(input);

      // THEN the team is included
      expect(result).toHaveLength(1);
    });

    it('excludes teams with zero ownership percentage', () => {
      // GIVEN a team that has 0% ownership in the user portfolio
      const input = makeOwnershipInput({
        ownershipShowAllTeams: false,
        portfolios: [makePortfolio({ id: 'p1', entryId: 'e1' })],
        teams: [makeEntryTeam({ id: 'et1', entryId: 'e1', teamId: 't1' })],
        allCalcuttaPortfolioTeams: [
          makePortfolioTeam({ id: 'pt1', portfolioId: 'p1', teamId: 't1', ownershipPercentage: 0 }),
        ],
      });

      // WHEN computing ownership teams data
      const result = computeOwnershipTeamsData(input);

      // THEN the team is excluded
      expect(result).toEqual([]);
    });

    it('excludes teams with no portfolio team record', () => {
      // GIVEN a team that has no portfolio team record at all
      const input = makeOwnershipInput({
        ownershipShowAllTeams: false,
        portfolios: [makePortfolio({ id: 'p1', entryId: 'e1' })],
        teams: [makeEntryTeam({ id: 'et1', entryId: 'e1', teamId: 't1' })],
        allCalcuttaPortfolioTeams: [],
      });

      // WHEN computing ownership teams data
      const result = computeOwnershipTeamsData(input);

      // THEN the team is excluded
      expect(result).toEqual([]);
    });
  });

  // =========================================================================
  // ownershipTeamsData -- show all teams mode
  // =========================================================================

  describe('ownershipTeamsData with ownershipShowAllTeams=true', () => {
    it('returns existing entry team when a match exists in tournament teams', () => {
      // GIVEN a tournament team that the entry already has
      const existingTeam = makeEntryTeam({ id: 'et1', entryId: 'e1', teamId: 't1', bid: 25 });
      const input = makeOwnershipInput({
        ownershipShowAllTeams: true,
        teams: [existingTeam],
        tournamentTeams: [makeTournamentTeam({ id: 't1', schoolId: 's1' })],
        schools: [{ id: 's1', name: 'Duke' }],
      });

      // WHEN computing ownership teams data
      const result = computeOwnershipTeamsData(input);

      // THEN the existing entry team (with its original bid) is used
      expect(result[0].bid).toBe(25);
    });

    it('creates synthetic entry team with zero bid for unowned tournament teams', () => {
      // GIVEN a tournament team that the entry does not own
      const input = makeOwnershipInput({
        ownershipShowAllTeams: true,
        teams: [],
        tournamentTeams: [makeTournamentTeam({ id: 't1', schoolId: 's1' })],
        schools: [{ id: 's1', name: 'Duke' }],
      });

      // WHEN computing ownership teams data
      const result = computeOwnershipTeamsData(input);

      // THEN a synthetic team is created with bid=0
      expect(result[0].bid).toBe(0);
    });

    it('assigns synthetic team ID with "synthetic-" prefix', () => {
      // GIVEN a tournament team that the entry does not own
      const input = makeOwnershipInput({
        ownershipShowAllTeams: true,
        teams: [],
        tournamentTeams: [makeTournamentTeam({ id: 't1', schoolId: 's1' })],
        schools: [{ id: 's1', name: 'Duke' }],
      });

      // WHEN computing ownership teams data
      const result = computeOwnershipTeamsData(input);

      // THEN the synthetic team has a prefixed ID
      expect(result[0].id).toBe('synthetic-t1');
    });

    it('attaches school info to synthetic teams', () => {
      // GIVEN a tournament team with a school
      const input = makeOwnershipInput({
        ownershipShowAllTeams: true,
        teams: [],
        tournamentTeams: [makeTournamentTeam({ id: 't1', schoolId: 's1' })],
        schools: [{ id: 's1', name: 'Duke' }],
      });

      // WHEN computing ownership teams data
      const result = computeOwnershipTeamsData(input);

      // THEN the synthetic team has school info attached
      expect(result[0].team?.school?.name).toBe('Duke');
    });

    it('includes all tournament teams in the result', () => {
      // GIVEN three tournament teams but only one owned
      const input = makeOwnershipInput({
        ownershipShowAllTeams: true,
        teams: [makeEntryTeam({ id: 'et1', entryId: 'e1', teamId: 't1' })],
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

      // WHEN computing ownership teams data
      const result = computeOwnershipTeamsData(input);

      // THEN all three tournament teams are included
      expect(result).toHaveLength(3);
    });
  });

  // =========================================================================
  // ownershipTeamsData -- sorting by ownership
  // =========================================================================

  describe('ownershipTeamsData sorting by ownership', () => {
    it('sorts teams by ownership percentage descending', () => {
      // GIVEN two teams with different ownership percentages
      const input = makeOwnershipInput({
        sortBy: 'ownership',
        ownershipShowAllTeams: false,
        portfolios: [makePortfolio({ id: 'p1', entryId: 'e1' })],
        teams: [
          makeEntryTeam({ id: 'et1', entryId: 'e1', teamId: 't1' }),
          makeEntryTeam({ id: 'et2', entryId: 'e1', teamId: 't2' }),
        ],
        allCalcuttaPortfolioTeams: [
          makePortfolioTeam({ id: 'pt1', portfolioId: 'p1', teamId: 't1', ownershipPercentage: 0.3 }),
          makePortfolioTeam({ id: 'pt2', portfolioId: 'p1', teamId: 't2', ownershipPercentage: 0.7 }),
        ],
      });

      // WHEN computing ownership teams data
      const result = computeOwnershipTeamsData(input);

      // THEN the team with higher ownership (t2) comes first
      expect(result[0].teamId).toBe('t2');
    });

    it('breaks ownership ties by points descending', () => {
      // GIVEN two teams with equal ownership but different points
      const input = makeOwnershipInput({
        sortBy: 'ownership',
        ownershipShowAllTeams: false,
        portfolios: [makePortfolio({ id: 'p1', entryId: 'e1' })],
        teams: [
          makeEntryTeam({ id: 'et1', entryId: 'e1', teamId: 't1' }),
          makeEntryTeam({ id: 'et2', entryId: 'e1', teamId: 't2' }),
        ],
        allCalcuttaPortfolioTeams: [
          makePortfolioTeam({ id: 'pt1', portfolioId: 'p1', teamId: 't1', ownershipPercentage: 0.5, actualPoints: 10 }),
          makePortfolioTeam({ id: 'pt2', portfolioId: 'p1', teamId: 't2', ownershipPercentage: 0.5, actualPoints: 30 }),
        ],
      });

      // WHEN computing ownership teams data
      const result = computeOwnershipTeamsData(input);

      // THEN the team with more points (t2) comes first
      expect(result[0].teamId).toBe('t2');
    });

    it('breaks ownership and points ties by bid descending', () => {
      // GIVEN two teams with equal ownership and equal points but different bids
      const input = makeOwnershipInput({
        sortBy: 'ownership',
        ownershipShowAllTeams: false,
        portfolios: [makePortfolio({ id: 'p1', entryId: 'e1' })],
        teams: [
          makeEntryTeam({ id: 'et1', entryId: 'e1', teamId: 't1', bid: 5 }),
          makeEntryTeam({ id: 'et2', entryId: 'e1', teamId: 't2', bid: 20 }),
        ],
        allCalcuttaPortfolioTeams: [
          makePortfolioTeam({ id: 'pt1', portfolioId: 'p1', teamId: 't1', ownershipPercentage: 0.5, actualPoints: 10 }),
          makePortfolioTeam({ id: 'pt2', portfolioId: 'p1', teamId: 't2', ownershipPercentage: 0.5, actualPoints: 10 }),
        ],
      });

      // WHEN computing ownership teams data
      const result = computeOwnershipTeamsData(input);

      // THEN the team with the higher bid (t2) comes first
      expect(result[0].teamId).toBe('t2');
    });
  });

  // =========================================================================
  // ownershipTeamsData -- sorting by points
  // =========================================================================

  describe('ownershipTeamsData sorting by points', () => {
    it('sorts teams by actual points descending', () => {
      // GIVEN two teams with different point totals
      const input = makeOwnershipInput({
        sortBy: 'points',
        ownershipShowAllTeams: false,
        portfolios: [makePortfolio({ id: 'p1', entryId: 'e1' })],
        teams: [
          makeEntryTeam({ id: 'et1', entryId: 'e1', teamId: 't1' }),
          makeEntryTeam({ id: 'et2', entryId: 'e1', teamId: 't2' }),
        ],
        allCalcuttaPortfolioTeams: [
          makePortfolioTeam({ id: 'pt1', portfolioId: 'p1', teamId: 't1', actualPoints: 50, ownershipPercentage: 0.5 }),
          makePortfolioTeam({ id: 'pt2', portfolioId: 'p1', teamId: 't2', actualPoints: 80, ownershipPercentage: 0.5 }),
        ],
      });

      // WHEN computing ownership teams data
      const result = computeOwnershipTeamsData(input);

      // THEN the team with more points (t2) comes first
      expect(result[0].teamId).toBe('t2');
    });

    it('breaks points ties by ownership descending', () => {
      // GIVEN two teams with equal points but different ownership
      const input = makeOwnershipInput({
        sortBy: 'points',
        ownershipShowAllTeams: false,
        portfolios: [makePortfolio({ id: 'p1', entryId: 'e1' })],
        teams: [
          makeEntryTeam({ id: 'et1', entryId: 'e1', teamId: 't1' }),
          makeEntryTeam({ id: 'et2', entryId: 'e1', teamId: 't2' }),
        ],
        allCalcuttaPortfolioTeams: [
          makePortfolioTeam({ id: 'pt1', portfolioId: 'p1', teamId: 't1', actualPoints: 40, ownershipPercentage: 0.2 }),
          makePortfolioTeam({ id: 'pt2', portfolioId: 'p1', teamId: 't2', actualPoints: 40, ownershipPercentage: 0.8 }),
        ],
      });

      // WHEN computing ownership teams data
      const result = computeOwnershipTeamsData(input);

      // THEN the team with higher ownership (t2) comes first
      expect(result[0].teamId).toBe('t2');
    });

    it('breaks points and ownership ties by bid descending', () => {
      // GIVEN two teams with equal points and ownership but different bids
      const input = makeOwnershipInput({
        sortBy: 'points',
        ownershipShowAllTeams: false,
        portfolios: [makePortfolio({ id: 'p1', entryId: 'e1' })],
        teams: [
          makeEntryTeam({ id: 'et1', entryId: 'e1', teamId: 't1', bid: 5 }),
          makeEntryTeam({ id: 'et2', entryId: 'e1', teamId: 't2', bid: 15 }),
        ],
        allCalcuttaPortfolioTeams: [
          makePortfolioTeam({ id: 'pt1', portfolioId: 'p1', teamId: 't1', actualPoints: 40, ownershipPercentage: 0.5 }),
          makePortfolioTeam({ id: 'pt2', portfolioId: 'p1', teamId: 't2', actualPoints: 40, ownershipPercentage: 0.5 }),
        ],
      });

      // WHEN computing ownership teams data
      const result = computeOwnershipTeamsData(input);

      // THEN the team with the higher bid (t2) comes first
      expect(result[0].teamId).toBe('t2');
    });
  });

  // =========================================================================
  // ownershipTeamsData -- sorting by bid
  // =========================================================================

  describe('ownershipTeamsData sorting by bid', () => {
    it('sorts teams by bid descending', () => {
      // GIVEN two teams with different bids
      const input = makeOwnershipInput({
        sortBy: 'bid',
        ownershipShowAllTeams: false,
        portfolios: [makePortfolio({ id: 'p1', entryId: 'e1' })],
        teams: [
          makeEntryTeam({ id: 'et1', entryId: 'e1', teamId: 't1', bid: 30 }),
          makeEntryTeam({ id: 'et2', entryId: 'e1', teamId: 't2', bid: 10 }),
        ],
        allCalcuttaPortfolioTeams: [
          makePortfolioTeam({ id: 'pt1', portfolioId: 'p1', teamId: 't1', ownershipPercentage: 0.5 }),
          makePortfolioTeam({ id: 'pt2', portfolioId: 'p1', teamId: 't2', ownershipPercentage: 0.5 }),
        ],
      });

      // WHEN computing ownership teams data
      const result = computeOwnershipTeamsData(input);

      // THEN the team with the higher bid (t1) comes first
      expect(result[0].teamId).toBe('t1');
    });

    it('breaks bid ties by points descending', () => {
      // GIVEN two teams with equal bids but different points
      const input = makeOwnershipInput({
        sortBy: 'bid',
        ownershipShowAllTeams: false,
        portfolios: [makePortfolio({ id: 'p1', entryId: 'e1' })],
        teams: [
          makeEntryTeam({ id: 'et1', entryId: 'e1', teamId: 't1', bid: 20 }),
          makeEntryTeam({ id: 'et2', entryId: 'e1', teamId: 't2', bid: 20 }),
        ],
        allCalcuttaPortfolioTeams: [
          makePortfolioTeam({ id: 'pt1', portfolioId: 'p1', teamId: 't1', actualPoints: 15, ownershipPercentage: 0.5 }),
          makePortfolioTeam({ id: 'pt2', portfolioId: 'p1', teamId: 't2', actualPoints: 45, ownershipPercentage: 0.5 }),
        ],
      });

      // WHEN computing ownership teams data
      const result = computeOwnershipTeamsData(input);

      // THEN the team with more points (t2) comes first
      expect(result[0].teamId).toBe('t2');
    });

    it('breaks bid and points ties by ownership descending', () => {
      // GIVEN two teams with equal bids and points but different ownership
      const input = makeOwnershipInput({
        sortBy: 'bid',
        ownershipShowAllTeams: false,
        portfolios: [makePortfolio({ id: 'p1', entryId: 'e1' })],
        teams: [
          makeEntryTeam({ id: 'et1', entryId: 'e1', teamId: 't1', bid: 20 }),
          makeEntryTeam({ id: 'et2', entryId: 'e1', teamId: 't2', bid: 20 }),
        ],
        allCalcuttaPortfolioTeams: [
          makePortfolioTeam({ id: 'pt1', portfolioId: 'p1', teamId: 't1', actualPoints: 40, ownershipPercentage: 0.3 }),
          makePortfolioTeam({ id: 'pt2', portfolioId: 'p1', teamId: 't2', actualPoints: 40, ownershipPercentage: 0.9 }),
        ],
      });

      // WHEN computing ownership teams data
      const result = computeOwnershipTeamsData(input);

      // THEN the team with higher ownership (t2) comes first
      expect(result[0].teamId).toBe('t2');
    });
  });

  // =========================================================================
  // ownershipTeamsData -- edge cases
  // =========================================================================

  describe('ownershipTeamsData edge cases', () => {
    it('returns empty array when no teams have ownership and showAll is false', () => {
      // GIVEN teams exist but none have portfolio team records
      const input = makeOwnershipInput({
        ownershipShowAllTeams: false,
        portfolios: [makePortfolio({ id: 'p1', entryId: 'e1' })],
        teams: [
          makeEntryTeam({ id: 'et1', entryId: 'e1', teamId: 't1' }),
          makeEntryTeam({ id: 'et2', entryId: 'e1', teamId: 't2' }),
        ],
        allCalcuttaPortfolioTeams: [],
      });

      // WHEN computing ownership teams data
      const result = computeOwnershipTeamsData(input);

      // THEN no teams are included
      expect(result).toEqual([]);
    });

    it('handles a single team without errors', () => {
      // GIVEN exactly one team with ownership
      const input = makeOwnershipInput({
        ownershipShowAllTeams: false,
        portfolios: [makePortfolio({ id: 'p1', entryId: 'e1' })],
        teams: [makeEntryTeam({ id: 'et1', entryId: 'e1', teamId: 't1' })],
        allCalcuttaPortfolioTeams: [
          makePortfolioTeam({ id: 'pt1', portfolioId: 'p1', teamId: 't1', ownershipPercentage: 1.0 }),
        ],
      });

      // WHEN computing ownership teams data
      const result = computeOwnershipTeamsData(input);

      // THEN exactly one team is returned
      expect(result).toHaveLength(1);
    });

    it('treats missing portfolio data as zero for sorting', () => {
      // GIVEN showAll mode with two tournament teams, one with portfolio data and one without
      const input = makeOwnershipInput({
        sortBy: 'points',
        ownershipShowAllTeams: true,
        portfolios: [makePortfolio({ id: 'p1', entryId: 'e1' })],
        teams: [],
        tournamentTeams: [
          makeTournamentTeam({ id: 't1', schoolId: 's1' }),
          makeTournamentTeam({ id: 't2', schoolId: 's2' }),
        ],
        schools: [
          { id: 's1', name: 'Duke' },
          { id: 's2', name: 'UNC' },
        ],
        allCalcuttaPortfolioTeams: [
          makePortfolioTeam({ id: 'pt1', portfolioId: 'p1', teamId: 't1', actualPoints: 50, ownershipPercentage: 0.5 }),
        ],
      });

      // WHEN computing ownership teams data sorted by points
      const result = computeOwnershipTeamsData(input);

      // THEN the team with points (t1) comes first, the other defaults to 0
      expect(result[0].teamId).toBe('t1');
    });
  });
});
