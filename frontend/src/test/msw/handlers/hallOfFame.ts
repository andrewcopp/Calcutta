import { http, HttpResponse } from 'msw';

const BASE = 'http://localhost:8080/api';

export const hallOfFameHandlers = [
  http.get(`${BASE}/hall-of-fame/best-teams`, () => {
    return HttpResponse.json({
      teams: [
        {
          tournamentName: 'NCAA 2025',
          tournamentYear: 2025,
          calcuttaId: 'calc-1',
          teamId: 'team-1',
          schoolName: 'Duke',
          seed: 1,
          region: 'East',
          teamPoints: 100,
          totalBid: 50,
          calcuttaTotalBid: 500,
          calcuttaTotalPoints: 1000,
          investmentShare: 0.1,
          pointsShare: 0.1,
          rawROI: 2.0,
          normalizedROI: 1.5,
        },
      ],
    });
  }),

  http.get(`${BASE}/hall-of-fame/best-investments`, () => {
    return HttpResponse.json({
      investments: [
        {
          tournamentName: 'NCAA 2025',
          tournamentYear: 2025,
          calcuttaId: 'calc-1',
          entryId: 'entry-1',
          entryName: 'Test Entry',
          teamId: 'team-1',
          schoolName: 'Duke',
          seed: 1,
          investment: 25,
          ownershipPercentage: 0.5,
          rawReturns: 50,
          normalizedReturns: 2.0,
        },
      ],
    });
  }),

  http.get(`${BASE}/hall-of-fame/best-entries`, () => {
    return HttpResponse.json({
      entries: [
        {
          tournamentName: 'NCAA 2025',
          tournamentYear: 2025,
          calcuttaId: 'calc-1',
          entryId: 'entry-1',
          entryName: 'Test Entry',
          totalReturns: 150,
          totalParticipants: 10,
          averageReturns: 100,
          normalizedReturns: 1.5,
        },
      ],
    });
  }),

  http.get(`${BASE}/hall-of-fame/best-careers`, () => {
    return HttpResponse.json({
      careers: [
        {
          entryName: 'Test Player',
          wins: 3,
          years: 5,
          bestFinish: 1,
          podiums: 4,
          inTheMoneys: 5,
          top10s: 5,
          careerEarningsCents: 50000,
          activeInLatestCalcutta: true,
        },
      ],
    });
  }),
];
