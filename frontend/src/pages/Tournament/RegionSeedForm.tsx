import { Button } from '../../components/ui/Button';
import { Combobox, type ComboboxOption } from '../../components/ui/Combobox';
import { School } from '../../types/school';
import { bracketOrder } from '../../utils/bracketOrder';

interface TeamSlot {
  schoolId: string;
  searchText: string;
}

interface RegionSeedFormProps {
  region: string;
  regionState: Record<number, TeamSlot[]>;
  schoolOptions: ComboboxOption[];
  schools: School[];
  usedSchoolIds: Set<string>;
  updateSlot: (region: string, seed: number, slotIndex: number, update: Partial<TeamSlot>) => void;
  addPlayIn: (region: string, seed: number) => void;
  removePlayIn: (region: string, seed: number, slotIndex: number) => void;
  onSlotBlur?: (region: string, seed: number, slotIndex: number) => void;
  slotValidation?: Record<string, 'none' | 'valid' | 'error'>;
}

function slotKey(seed: number, slotIndex: number): string {
  return `${seed}-${slotIndex}`;
}

export function RegionSeedForm({
  region,
  regionState,
  schoolOptions,
  schools,
  usedSchoolIds,
  updateSlot,
  addPlayIn,
  removePlayIn,
  onSlotBlur,
  slotValidation,
}: RegionSeedFormProps) {
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
                    onBlur={onSlotBlur ? () => onSlotBlur(region, seed, slotIndex) : undefined}
                    validationState={slotValidation?.[slotKey(seed, slotIndex)] ?? 'none'}
                  />
                  <div className="w-8 shrink-0">
                    {slots.length === 1 ? (
                      <Button
                        type="button"
                        variant="ghost"
                        size="sm"
                        onClick={() => addPlayIn(region, seed)}
                        title="Add play-in team"
                      >
                        +
                      </Button>
                    ) : (
                      <Button
                        type="button"
                        variant="ghost"
                        size="sm"
                        onClick={() => removePlayIn(region, seed, slotIndex)}
                        className="text-red-500 hover:text-red-700"
                      >
                        -
                      </Button>
                    )}
                  </div>
                </div>
              ))}
            </div>
          </div>
        );
      })}
    </div>
  );
}
