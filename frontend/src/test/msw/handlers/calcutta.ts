import { http, HttpResponse } from 'msw';

const BASE = 'http://localhost:8080/api/v1';

const validCalcutta = {
  id: 'calc-1',
  name: 'Test Pool',
  tournamentId: 'tourn-1',
  ownerId: 'owner-1',
  minTeams: 3,
  maxTeams: 10,
  maxBidPoints: 50,
  budgetPoints: 100,
  createdAt: '2026-01-01T00:00:00Z',
  updatedAt: '2026-01-01T00:00:00Z',
};

const validEntry = {
  id: 'entry-1',
  name: 'Test Entry',
  calcuttaId: 'calc-1',
  status: 'accepted' as const,
  createdAt: '2026-01-01T00:00:00Z',
  updatedAt: '2026-01-01T00:00:00Z',
};

const validEntryTeam = {
  id: 'et-1',
  entryId: 'entry-1',
  teamId: 'team-1',
  bidPoints: 10,
  createdAt: '2026-01-01T00:00:00Z',
  updatedAt: '2026-01-01T00:00:00Z',
};

export const calcuttaHandlers = [
  http.get(`${BASE}/calcuttas/list-with-rankings`, () => {
    return HttpResponse.json({
      items: [
        {
          ...validCalcutta,
          hasEntry: true,
          ranking: { rank: 1, totalEntries: 5, points: 150 },
        },
      ],
    });
  }),

  http.get(`${BASE}/calcuttas/:id/dashboard`, () => {
    return HttpResponse.json({
      calcutta: validCalcutta,
      biddingOpen: false,
      totalEntries: 1,
      entries: [validEntry],
      entryTeams: [validEntryTeam],
      portfolios: [],
      portfolioTeams: [],
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

  http.get(`${BASE}/calcuttas/:id/entries/:entryId/teams`, () => {
    return HttpResponse.json({ items: [validEntryTeam] });
  }),

  http.get(`${BASE}/calcuttas/:id/payouts`, () => {
    return HttpResponse.json({
      items: [{ position: 1, amountCents: 10000 }],
    });
  }),

  http.get(`${BASE}/calcuttas/:id`, () => {
    return HttpResponse.json(validCalcutta);
  }),

  http.get(`${BASE}/calcuttas`, () => {
    return HttpResponse.json({ items: [validCalcutta] });
  }),

  http.post(`${BASE}/calcuttas/:id/entries`, () => {
    return HttpResponse.json(validEntry);
  }),

  http.post(`${BASE}/calcuttas`, () => {
    return HttpResponse.json(validCalcutta);
  }),

  http.patch(`${BASE}/calcuttas/:id`, () => {
    return HttpResponse.json(validCalcutta);
  }),

  http.patch(`${BASE}/entries/:id`, () => {
    return HttpResponse.json(validEntry);
  }),

  http.put(`${BASE}/calcuttas/:id/payouts`, () => {
    return HttpResponse.json({
      items: [{ position: 1, amountCents: 10000 }],
    });
  }),
];
