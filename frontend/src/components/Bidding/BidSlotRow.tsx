import React, { useRef, useEffect } from 'react';
import { Combobox } from '../ui/Combobox';
import { Input } from '../ui/Input';
import { cn } from '../../lib/cn';
import type { InvestmentSlot, TeamComboboxOption, TeamWithSchool } from '../../hooks/useInvesting';

interface BidSlotRowProps {
  slotIndex: number;
  slot: InvestmentSlot;
  teamOptions: TeamComboboxOption[];
  usedTeamIds: Set<string>;
  teams: TeamWithSchool[];
  maxBidPoints: number;
  minBid: number;
  onSelect: (slotIndex: number, teamId: string) => void;
  onClear: (slotIndex: number) => void;
  onSearchChange: (slotIndex: number, text: string) => void;
  onBidChange: (slotIndex: number, bid: number) => void;
  isOptional: boolean;
}

function renderTeamOption(option: { id: string; label: string }, isHighlighted: boolean) {
  const teamOption = option as TeamComboboxOption;
  return (
    <span className={isHighlighted ? 'font-medium' : ''}>
      {teamOption.label} ({teamOption.region} - {teamOption.seed})
    </span>
  );
}

function BidSlotRowComponent({
  slotIndex,
  slot,
  teamOptions,
  usedTeamIds,
  teams,
  maxBidPoints,
  minBid,
  onSelect,
  onClear,
  onSearchChange,
  onBidChange,
  isOptional,
}: BidSlotRowProps) {
  const bidInputRef = useRef<HTMLInputElement>(null);
  const prevTeamIdRef = useRef(slot.teamId);

  // Auto-focus bid input when a team is selected
  useEffect(() => {
    if (slot.teamId && !prevTeamIdRef.current) {
      bidInputRef.current?.focus();
    }
    prevTeamIdRef.current = slot.teamId;
  }, [slot.teamId]);

  const isFilled = Boolean(slot.teamId);
  const team = isFilled ? teams.find((t) => t.id === slot.teamId) : null;

  const handleBidChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value;
    if (value === '') {
      onBidChange(slotIndex, 0);
      return;
    }
    const numValue = parseFloat(value);
    if (!isNaN(numValue) && numValue >= 0) {
      onBidChange(slotIndex, numValue);
    }
  };

  const validationError =
    slot.investmentAmount > 0
      ? slot.investmentAmount < minBid
        ? `Min ${minBid} credits`
        : slot.investmentAmount > maxBidPoints
          ? `Max ${maxBidPoints} credits`
          : undefined
      : undefined;

  return (
    <div className="py-3 flex items-start gap-3">
      {/* Slot number */}
      <div className="flex items-center justify-center w-7 h-7 rounded-full bg-primary/10 text-primary text-sm font-semibold shrink-0 mt-1.5">
        {slotIndex + 1}
      </div>

      {isFilled ? (
        /* Filled slot */
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-3 flex-wrap">
            <div className="flex items-center gap-2 min-w-0">
              <span className="font-medium text-foreground truncate">
                {team?.school?.name ?? 'Unknown'} ({team?.region} - {team?.seed})
              </span>
            </div>

            <div className="flex items-center gap-2 ml-auto">
              <div className="relative w-24">
                <Input
                  ref={bidInputRef}
                  type="number"
                  min="0"
                  max={maxBidPoints}
                  step="1"
                  value={slot.investmentAmount || ''}
                  onChange={handleBidChange}
                  placeholder="0"
                  className={cn('pr-10 text-right', {
                    'border-red-500 focus:ring-red-500': validationError,
                  })}
                />
                <div className="absolute inset-y-0 right-3 flex items-center pointer-events-none">
                  <span className="text-muted-foreground text-sm">credits</span>
                </div>
              </div>

              <button
                type="button"
                onClick={() => onClear(slotIndex)}
                className="p-1.5 text-muted-foreground/60 hover:text-destructive hover:bg-destructive/10 rounded transition-colors"
                title="Remove team"
              >
                <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" strokeWidth="2" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>
          </div>
          {validationError && <div className="text-xs text-red-600 mt-1 ml-9">{validationError}</div>}
        </div>
      ) : (
        /* Empty slot */
        <div className="flex-1 flex items-center gap-2">
          <Combobox
            options={teamOptions}
            value={slot.searchText}
            onChange={(text) => onSearchChange(slotIndex, text)}
            onSelect={(teamId) => onSelect(slotIndex, teamId)}
            placeholder="Search for a team..."
            excludeIds={usedTeamIds}
            className="flex-1"
            renderOption={renderTeamOption}
          />
          {isOptional && <span className="text-xs text-muted-foreground/60 italic shrink-0">optional</span>}
        </div>
      )}
    </div>
  );
}

export const BidSlotRow = React.memo(BidSlotRowComponent);
