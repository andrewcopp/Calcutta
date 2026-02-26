import { http, HttpResponse } from 'msw';

const BASE = 'http://localhost:8080/api/v1';

const validPool = {
  id: 'pool-1',
  name: 'Test Pool',
  tournamentId: 'tourn-1',
  ownerId: 'owner-1',
  minTeams: 3,
  maxTeams: 10,
  maxInvestmentCredits: 50,
  budgetCredits: 100,
  createdAt: '2026-01-01T00:00:00Z',
  updatedAt: '2026-01-01T00:00:00Z',
};

const validPortfolio = {
  id: 'portfolio-1',
  name: 'Test Portfolio',
  poolId: 'pool-1',
  status: 'submitted' as const,
  createdAt: '2026-01-01T00:00:00Z',
  updatedAt: '2026-01-01T00:00:00Z',
};

const validInvestment = {
  id: 'inv-1',
  portfolioId: 'portfolio-1',
  teamId: 'team-1',
  credits: 10,
  createdAt: '2026-01-01T00:00:00Z',
  updatedAt: '2026-01-01T00:00:00Z',
};

export const poolHandlers = [
  http.get(`${BASE}/pools`, ({ request }) => {
    const url = new URL(request.url);
    if (url.searchParams.get('include') === 'rankings') {
      return HttpResponse.json({
        items: [
          {
            ...validPool,
            hasPortfolio: true,
            ranking: { rank: 1, totalPortfolios: 5, points: 150 },
          },
        ],
      });
    }
    return HttpResponse.json({ items: [validPool] });
  }),

  http.get(`${BASE}/pools/:id/dashboard`, () => {
    return HttpResponse.json({
      pool: validPool,
      investingOpen: false,
      totalPortfolios: 1,
      portfolios: [validPortfolio],
      investments: [validInvestment],
      ownershipSummaries: [],
      ownershipDetails: [],
      schools: [{ id: 'sch-1', name: 'Duke' }],
      tournamentTeams: [
        {
          id: 'team-1',
          schoolId: 'sch-1',
          tournamentId: 'tourn-1',
          seed: 1,
          region: 'East',
          byes: 0,
          wins: 0,
          isEliminated: false,
          createdAt: '2026-01-01T00:00:00Z',
          updatedAt: '2026-01-01T00:00:00Z',
        },
      ],
      roundStandings: [],
    });
  }),

  http.get(`${BASE}/pools/:id/portfolios/:portfolioId/investments`, () => {
    return HttpResponse.json({ items: [validInvestment] });
  }),

  http.get(`${BASE}/pools/:id/payouts`, () => {
    return HttpResponse.json({
      items: [{ position: 1, amountCents: 10000 }],
    });
  }),

  http.get(`${BASE}/pools/:id`, () => {
    return HttpResponse.json(validPool);
  }),

  http.post(`${BASE}/pools/:id/portfolios`, () => {
    return HttpResponse.json(validPortfolio);
  }),

  http.post(`${BASE}/pools`, () => {
    return HttpResponse.json(validPool);
  }),

  http.patch(`${BASE}/pools/:id`, () => {
    return HttpResponse.json(validPool);
  }),

  http.patch(`${BASE}/pools/:poolId/portfolios/:portfolioId`, () => {
    return HttpResponse.json(validPortfolio);
  }),

  http.put(`${BASE}/pools/:id/payouts`, () => {
    return HttpResponse.json({
      items: [{ position: 1, amountCents: 10000 }],
    });
  }),
];
