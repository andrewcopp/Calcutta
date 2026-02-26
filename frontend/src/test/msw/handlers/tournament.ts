import { http, HttpResponse } from 'msw';

const BASE = 'http://localhost:8080/api/v1';

const validTournament = {
  id: 'tourn-1',
  name: 'NCAA 2025',
  rounds: 6,
  startingAt: '2025-03-20T00:00:00Z',
  createdAt: '2025-01-01T00:00:00Z',
  updatedAt: '2025-01-01T00:00:00Z',
};

const validTeam = {
  id: 'team-1',
  schoolId: 'sch-1',
  tournamentId: 'tourn-1',
  seed: 1,
  region: 'East',
  byes: 0,
  wins: 0,
  isEliminated: false,
  createdAt: '2025-01-01T00:00:00Z',
  updatedAt: '2025-01-01T00:00:00Z',
  school: { id: 'sch-1', name: 'Duke' },
};

export const tournamentHandlers = [
  http.get(`${BASE}/tournaments`, ({ request }) => {
    const url = new URL(request.url);
    if (url.pathname.endsWith('/tournaments')) {
      return HttpResponse.json({ items: [validTournament] });
    }
    return HttpResponse.json(validTournament);
  }),

  http.get(`${BASE}/tournaments/:id`, () => {
    return HttpResponse.json(validTournament);
  }),

  http.get(`${BASE}/tournaments/:id/teams`, () => {
    return HttpResponse.json({ items: [validTeam] });
  }),

  http.post(`${BASE}/tournaments`, () => {
    return HttpResponse.json(validTournament);
  }),

  http.patch(`${BASE}/tournaments/:id`, () => {
    return HttpResponse.json(validTournament);
  }),

  http.put(`${BASE}/tournaments/:id/teams`, () => {
    return HttpResponse.json({ items: [validTeam] });
  }),

  http.get(`${BASE}/competitions`, () => {
    return HttpResponse.json({ items: [{ id: 'comp-1', name: 'NCAA' }] });
  }),

  http.get(`${BASE}/seasons`, () => {
    return HttpResponse.json({ items: [{ id: 'season-1', year: 2025 }] });
  }),

  http.get(`${BASE}/tournaments/:id/moderators`, () => {
    return HttpResponse.json({
      items: [{ id: 'mod-1', email: 'mod@test.com', firstName: 'Mod', lastName: 'User' }],
    });
  }),

  http.post(`${BASE}/tournaments/:id/moderators`, () => {
    return HttpResponse.json({
      moderator: { id: 'mod-2', email: 'new@test.com', firstName: 'New', lastName: 'Mod' },
    });
  }),

  http.delete(`${BASE}/tournaments/:id/moderators/:userId`, () => {
    return new HttpResponse(null, { status: 204 });
  }),

  http.put(`${BASE}/tournaments/:id/kenpom`, () => {
    return HttpResponse.json({
      items: [
        {
          ...validTeam,
          kenPom: { netRtg: 25.5, oRtg: 118.0, dRtg: 92.5, adjT: 68.0 },
        },
      ],
    });
  }),
];
