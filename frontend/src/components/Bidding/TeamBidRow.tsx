import React from 'react';
import { Input } from '../ui/Input';
import { cn } from '../../lib/cn';

type TeamBidRowProps = {
  teamId: string;
  schoolName: string;
  seed: number;
  region: string;
  bidAmount: number;
  maxBid: number;
  onBidChange: (teamId: string, bid: number) => void;
  validationError?: string;
};

export function TeamBidRow({
  teamId,
  schoolName,
  seed,
  region,
  bidAmount,
  maxBid,
  onBidChange,
  validationError,
}: TeamBidRowProps) {
  const handleBidChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value;
    if (value === '') {
      onBidChange(teamId, 0);
      return;
    }

    const numValue = parseFloat(value);
    if (!isNaN(numValue) && numValue >= 0) {
      onBidChange(teamId, numValue);
    }
  };

  return (
    <div
      className={cn(
        'py-3 border-b border-gray-200 hover:bg-gray-50',
        'sm:grid sm:grid-cols-12 sm:gap-4 sm:items-center',
        {
          'bg-blue-50': bidAmount > 0,
        }
      )}
    >
      {/* Mobile: name + seed/region on one line */}
      <div className="flex items-center justify-between sm:contents">
        <div className="sm:col-span-5">
          <div className="font-medium text-gray-900">{schoolName}</div>
        </div>

        <div className="flex items-center gap-2 sm:contents">
          <div className="sm:col-span-2 sm:text-center">
            <span className="inline-flex items-center justify-center w-8 h-8 rounded-full bg-gray-100 text-gray-700 text-sm font-medium">
              {seed}
            </span>
          </div>

          <div className="sm:col-span-2 sm:text-center">
            <span className="text-sm text-gray-600">{region}</span>
          </div>
        </div>
      </div>

      {/* Bid input - full width on mobile, col-span-3 on desktop */}
      <div className="mt-2 sm:mt-0 sm:col-span-3">
        <div className="relative">
          <Input
            type="number"
            min="0"
            max={maxBid}
            step="1"
            value={bidAmount || ''}
            onChange={handleBidChange}
            placeholder="0"
            className={cn('pr-10', {
              'border-red-500 focus:ring-red-500': validationError,
            })}
          />
          <div className="absolute inset-y-0 right-3 flex items-center pointer-events-none">
            <span className="text-gray-500 text-sm">pts</span>
          </div>
        </div>
        {validationError && <div className="text-xs text-red-600 mt-1">{validationError}</div>}
      </div>
    </div>
  );
}
