import React from 'react';
import { BracketGame, BracketTeam } from '../types/bracket';

interface BracketGameCardProps {
  game: BracketGame;
  onSelectWinner: (gameId: string, teamId: string) => void;
  onUnselectWinner: (gameId: string) => void;
  isLoading?: boolean;
}

export const BracketGameCard: React.FC<BracketGameCardProps> = ({
  game,
  onSelectWinner,
  onUnselectWinner,
  isLoading = false,
}) => {
  const handleTeamClick = (team: BracketTeam) => {
    if (isLoading) return;
    
    if (game.winner?.teamId === team.teamId) {
      onUnselectWinner(game.gameId);
    } else if (game.canSelect) {
      onSelectWinner(game.gameId, team.teamId);
    }
  };

  const getTeamClassName = (team?: BracketTeam) => {
    if (!team) return 'bg-gray-100 text-gray-400 cursor-not-allowed';
    
    const isWinner = game.winner?.teamId === team.teamId;
    const isLoser = game.winner && game.winner.teamId !== team.teamId;
    const canSelect = game.canSelect || isWinner;
    
    let className = 'flex items-center justify-between p-3 rounded transition-all ';
    
    if (isWinner) {
      className += 'bg-green-100 border-2 border-green-500 cursor-pointer hover:bg-green-200';
    } else if (isLoser) {
      className += 'bg-gray-100 text-gray-500 line-through';
    } else if (canSelect) {
      className += 'bg-white border-2 border-gray-300 cursor-pointer hover:bg-blue-50 hover:border-blue-400';
    } else {
      className += 'bg-gray-50 border-2 border-gray-200';
    }
    
    if (isLoading) {
      className += ' opacity-50 cursor-wait';
    }
    
    return className;
  };

  const renderTeam = (team?: BracketTeam, position?: 'top' | 'bottom') => {
    if (!team) {
      return (
        <div className={getTeamClassName(undefined)}>
          <span className="text-sm">TBD</span>
        </div>
      );
    }

    return (
      <div
        className={getTeamClassName(team)}
        onClick={() => handleTeamClick(team)}
      >
        <div className="flex items-center gap-2">
          <span className="font-semibold text-gray-700 w-6">
            {team.seed}
          </span>
          <span className="font-medium">{team.name}</span>
        </div>
        {game.winner?.teamId === team.teamId && (
          <span className="text-green-600 font-bold">✓</span>
        )}
      </div>
    );
  };

  return (
    <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-4 hover:shadow-md transition-shadow">
      <div className="flex flex-col gap-2">
        {renderTeam(game.team1, 'top')}
        <div className="text-center text-xs text-gray-400 font-medium">vs</div>
        {renderTeam(game.team2, 'bottom')}
      </div>
      
      {game.region && game.round !== 'final_four' && game.round !== 'championship' && (
        <div className="mt-2 text-xs text-gray-500 text-center">
          {game.region}
        </div>
      )}
    </div>
  );
};
