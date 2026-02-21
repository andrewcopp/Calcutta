import React from 'react';
import { formatDate } from '../utils/format';

interface BiddingOverlayProps {
  tournamentStartingAt: string;
  children: React.ReactNode;
}

export function BiddingOverlay({ tournamentStartingAt, children }: BiddingOverlayProps) {
  return (
    <div className="relative">
      <div className="blur-sm pointer-events-none select-none" aria-hidden="true">
        {children}
      </div>
      <div className="absolute inset-0 flex items-center justify-center">
        <div className="bg-white/90 border border-gray-200 rounded-lg shadow-lg px-8 py-6 text-center max-w-sm">
          <div className="text-3xl mb-3">&#128274;</div>
          <p className="text-lg font-semibold text-gray-900 mb-1">Results Locked</p>
          <p className="text-sm text-gray-600">
            Entries unlock {formatDate(tournamentStartingAt, true)}
          </p>
        </div>
      </div>
    </div>
  );
}
