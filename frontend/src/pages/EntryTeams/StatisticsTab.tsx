import React from 'react';
import { CalcuttaPortfolio, CalcuttaPortfolioTeam } from '../../types/calcutta';
import { PortfolioScores } from './PortfolioScores';

type StatisticsTabProps = {
  portfolios: CalcuttaPortfolio[];
  portfolioTeams: CalcuttaPortfolioTeam[];
};

export function StatisticsTab({ portfolios, portfolioTeams }: StatisticsTabProps) {
  if (portfolios.length === 0) {
    return <p className="text-gray-500 text-sm py-4">No portfolio statistics available yet. Statistics appear after the tournament begins.</p>;
  }
  return <PortfolioScores portfolio={portfolios[0]} teams={portfolioTeams} />;
}
