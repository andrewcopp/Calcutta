import React from 'react';
import { TournamentTeam } from '../../types/calcutta';
import { Card } from '../../components/ui/Card';

interface TeamStatsCardsProps {
  teams: TournamentTeam[];
}

export const TeamStatsCards: React.FC<TeamStatsCardsProps> = ({ teams }) => {
  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
      <Card className="p-4">
        <h3 className="text-lg font-semibold text-gray-700 mb-2">Total Teams</h3>
        <p className="text-3xl font-bold text-blue-600">{teams.length}</p>
      </Card>
      <Card className="p-4">
        <h3 className="text-lg font-semibold text-gray-700 mb-2">Seed Distribution</h3>
        <div className="space-y-1">
          {Array.from({ length: 16 }, (_, i) => i + 1).map(seed => {
            const count = teams.filter(t => t.seed === seed).length;
            return count > 0 ? (
              <div key={seed} className="flex justify-between text-sm">
                <span>Seed {seed}:</span>
                <span className="font-medium">{count}</span>
              </div>
            ) : null;
          })}
        </div>
      </Card>
      <Card className="p-4">
        <h3 className="text-lg font-semibold text-gray-700 mb-2">Bye Distribution</h3>
        <div className="space-y-1">
          {Array.from({ length: Math.max(...teams.map(t => t.byes)) + 1 }, (_, i) => i).map(byes => {
            const count = teams.filter(t => t.byes === byes).length;
            return count > 0 ? (
              <div key={byes} className="flex justify-between text-sm">
                <span>{byes} {byes === 1 ? 'Bye' : 'Byes'}:</span>
                <span className="font-medium">{count}</span>
              </div>
            ) : null;
          })}
        </div>
      </Card>
      <Card className="p-4">
        <h3 className="text-lg font-semibold text-gray-700 mb-2">Win Distribution</h3>
        <div className="space-y-1">
          {Array.from({ length: Math.max(...teams.map(t => t.wins)) + 1 }, (_, i) => i).map(wins => {
            const count = teams.filter(t => t.wins === wins).length;
            return count > 0 ? (
              <div key={wins} className="flex justify-between text-sm">
                <span>{wins} {wins === 1 ? 'Win' : 'Wins'}:</span>
                <span className="font-medium">{count}</span>
              </div>
            ) : null;
          })}
          <div className="pt-2 mt-2 border-t border-gray-200">
            <div className="flex justify-between text-sm font-semibold">
              <span>Total Wins:</span>
              <span className="text-blue-600">{teams.reduce((sum, team) => sum + team.wins, 0)}</span>
            </div>
          </div>
        </div>
      </Card>
    </div>
  );
};
