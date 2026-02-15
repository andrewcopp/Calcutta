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
      className={cn('grid grid-cols-12 gap-4 items-center py-3 border-b border-gray-200 hover:bg-gray-50', {
        'bg-blue-50': bidAmount > 0,
      })}
    >
      <div className="col-span-5">
        <div className="font-medium text-gray-900">{schoolName}</div>
      </div>

      <div className="col-span-2 text-center">
        <span className="inline-flex items-center justify-center w-8 h-8 rounded-full bg-gray-100 text-gray-700 text-sm font-medium">
          {seed}
        </span>
      </div>

      <div className="col-span-2 text-center">
        <span className="text-sm text-gray-600">{region}</span>
      </div>

      <div className="col-span-3">
        <div className="relative">
          <div className="absolute inset-y-0 left-3 flex items-center pointer-events-none">
            <span className="text-gray-500 text-sm">$</span>
          </div>
          <Input
            type="number"
            min="0"
            max={maxBid}
            step="0.01"
            value={bidAmount || ''}
            onChange={handleBidChange}
            placeholder="0.00"
            className={cn('pl-7', {
              'border-red-500 focus:ring-red-500': validationError,
            })}
          />
        </div>
        {validationError && <div className="text-xs text-red-600 mt-1">{validationError}</div>}
      </div>
    </div>
  );
}
