import React from 'react';
import { Button } from '../../components/ui/Button';
import { Combobox, type ComboboxOption } from '../../components/ui/Combobox';
import { School } from '../../types/school';
import { bracketOrder } from '../../utils/bracketOrder';

interface TeamSlot {
  schoolId: string;
  searchText: string;
}

type Region = 'East' | 'West' | 'South' | 'Midwest';

interface RegionSeedFormProps {
  region: Region;
  regionState: Record<number, TeamSlot[]>;
  schoolOptions: ComboboxOption[];
  schools: School[];
  usedSchoolIds: Set<string>;
  updateSlot: (region: Region, seed: number, slotIndex: number, update: Partial<TeamSlot>) => void;
  addPlayIn: (region: Region, seed: number) => void;
  removePlayIn: (region: Region, seed: number) => void;
}

export const RegionSeedForm: React.FC<RegionSeedFormProps> = ({
  region,
  regionState,
  schoolOptions,
  schools,
  usedSchoolIds,
  updateSlot,
  addPlayIn,
  removePlayIn,
}) => {
  return (
    <div className="space-y-3">
      {bracketOrder(16).map((seed) => {
        const slots = regionState[seed];
        return (
          <div key={seed} className="flex items-start gap-3">
            <div className="flex items-center justify-center w-10 h-10 rounded-lg bg-gray-100 text-sm font-bold text-gray-700 shrink-0">
              {seed}
            </div>
            <div className="flex-1 space-y-2">
              {slots.map((slot, slotIndex) => (
                <div key={slotIndex} className="flex items-center gap-2">
                  <Combobox
                    options={schoolOptions}
                    value={slot.searchText}
                    onChange={(text) => updateSlot(region, seed, slotIndex, { searchText: text, schoolId: '' })}
                    onSelect={(schoolId) => {
                      const school = schools.find((s) => s.id === schoolId);
                      updateSlot(region, seed, slotIndex, { schoolId, searchText: school?.name || '' });
                    }}
                    placeholder="Search for a school..."
                    excludeIds={usedSchoolIds}
                    className="flex-1"
                  />
                  {slotIndex === 1 && (
                    <Button
                      type="button"
                      variant="ghost"
                      size="sm"
                      onClick={() => removePlayIn(region, seed)}
                      className="text-red-500 hover:text-red-700 shrink-0"
                    >
                      -
                    </Button>
                  )}
                </div>
              ))}
            </div>
            {slots.length === 1 && (
              <Button
                type="button"
                variant="ghost"
                size="sm"
                onClick={() => addPlayIn(region, seed)}
                className="shrink-0"
                title="Add play-in team"
              >
                +
              </Button>
            )}
          </div>
        );
      })}
    </div>
  );
};
