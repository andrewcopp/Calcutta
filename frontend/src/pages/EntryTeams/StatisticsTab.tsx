import React from 'react';
import { CalcuttaPortfolio, CalcuttaPortfolioTeam } from '../../types/calcutta';

type StatisticsTabProps = {
  portfolios: CalcuttaPortfolio[];
  portfolioTeams: CalcuttaPortfolioTeam[];
  PortfolioScoresComponent: React.FC<{ portfolio: CalcuttaPortfolio; teams: CalcuttaPortfolioTeam[] }>;
};

export function StatisticsTab({ portfolios, portfolioTeams, PortfolioScoresComponent }: StatisticsTabProps) {
  return <>{portfolios.length > 0 && <PortfolioScoresComponent portfolio={portfolios[0]} teams={portfolioTeams} />}</>;
}
