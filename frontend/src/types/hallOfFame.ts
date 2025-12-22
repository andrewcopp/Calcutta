export interface InvestmentLeaderboardRow {
  tournamentName: string;
  tournamentYear: number;
  calcuttaId: string;
  entryId: string;
  entryName: string;
  teamId: string;
  schoolName: string;
  seed: number;
  investment: number;
  ownershipPercentage: number;
  rawReturns: number;
  normalizedReturns: number;
}

export interface InvestmentLeaderboardResponse {
  investments: InvestmentLeaderboardRow[];
}

export interface BestTeam {
  tournamentName: string;
  tournamentYear: number;
  calcuttaId: string;
  teamId: string;
  schoolName: string;
  seed: number;
  region: string;
  teamPoints: number;
  totalBid: number;
  calcuttaTotalBid: number;
  calcuttaTotalPoints: number;
  investmentShare: number;
  pointsShare: number;
  rawROI: number;
  normalizedROI: number;
}

export interface BestTeamsResponse {
  teams: BestTeam[];
}

export interface EntryLeaderboardRow {
  tournamentName: string;
  tournamentYear: number;
  calcuttaId: string;
  entryId: string;
  entryName: string;
  totalReturns: number;
  totalParticipants: number;
  averageReturns: number;
  normalizedReturns: number;
}

export interface EntryLeaderboardResponse {
  entries: EntryLeaderboardRow[];
}

export interface CareerLeaderboardRow {
  entryName: string;
  wins: number;
  years: number;
  bestFinish: number;
  podiums: number;
  inTheMoneys: number;
  top10s: number;
  careerEarningsCents: number;
}

export interface CareerLeaderboardResponse {
  careers: CareerLeaderboardRow[];
}
