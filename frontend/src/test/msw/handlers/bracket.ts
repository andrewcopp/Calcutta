import { http, HttpResponse } from 'msw';

const BASE = 'http://localhost:8080/api';

const validBracket = {
  tournamentId: 'tourn-1',
  regions: ['East', 'West', 'South', 'Midwest'],
  games: [
    {
      gameId: 'game-1',
      round: 'round_of_64',
      region: 'East',
      team1: { teamId: 't-1', schoolId: 'sch-1', name: 'Duke', seed: 1, region: 'East' },
      team2: { teamId: 't-2', schoolId: 'sch-2', name: 'UNC', seed: 16, region: 'East' },
      sortOrder: 1,
      canSelect: true,
    },
  ],
};

export const bracketHandlers = [
  http.get(`${BASE}/tournaments/:id/bracket`, () => {
    return HttpResponse.json(validBracket);
  }),

  http.post(`${BASE}/tournaments/:id/bracket/games/:gameId/winner`, () => {
    return HttpResponse.json({
      ...validBracket,
      games: [
        {
          ...validBracket.games[0],
          winner: validBracket.games[0].team1,
          canSelect: false,
        },
      ],
    });
  }),

  http.delete(`${BASE}/tournaments/:id/bracket/games/:gameId/winner`, () => {
    return HttpResponse.json(validBracket);
  }),

  http.get(`${BASE}/tournaments/:id/bracket/validate`, () => {
    return HttpResponse.json({ valid: true, errors: [] });
  }),
];
