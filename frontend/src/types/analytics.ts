export interface SeedAnalytics {
  seed: number;
  totalPoints: number;
  totalInvestment: number;
  pointsPercentage: number;
  investmentPercentage: number;
  teamCount: number;
  averagePoints: number;
  averageInvestment: number;
  roi: number;
}

export interface RegionAnalytics {
  region: string;
  totalPoints: number;
  totalInvestment: number;
  pointsPercentage: number;
  investmentPercentage: number;
  teamCount: number;
  averagePoints: number;
  averageInvestment: number;
  roi: number;
}

export interface TeamAnalytics {
  schoolId: string;
  schoolName: string;
  totalPoints: number;
  totalInvestment: number;
  appearances: number;
  averagePoints: number;
  averageInvestment: number;
  averageSeed: number;
  roi: number;
}

export interface SeedVarianceAnalytics {
  seed: number;
  investmentStdDev: number;
  pointsStdDev: number;
  investmentMean: number;
  pointsMean: number;
  investmentCV: number;
  pointsCV: number;
  teamCount: number;
  varianceRatio: number;
}

export interface SeedInvestmentPoint {
  seed: number;
  tournamentName: string;
  tournamentYear: number;
  calcuttaId: string;
  teamId: string;
  schoolName: string;
  totalBid: number;
  calcuttaTotalBid: number;
  normalizedBid: number;
}

export interface SeedInvestmentSummary {
  seed: number;
  count: number;
  mean: number;
  stdDev: number;
  min: number;
  q1: number;
  median: number;
  q3: number;
  max: number;
}

export interface SeedInvestmentDistributionResponse {
  points: SeedInvestmentPoint[];
  summaries: SeedInvestmentSummary[];
}

export interface BestInvestment {
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

export interface BestInvestmentsResponse {
  investments: BestInvestment[];
}

export interface AnalyticsResponse {
  seedAnalytics?: SeedAnalytics[];
  regionAnalytics?: RegionAnalytics[];
  teamAnalytics?: TeamAnalytics[];
  seedVarianceAnalytics?: SeedVarianceAnalytics[];
  totalPoints: number;
  totalInvestment: number;
  baselineROI: number;
}
