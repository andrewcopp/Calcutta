import { cn } from '../../lib/cn';
import { SEED_FILTER_OPTIONS } from '../../hooks/useBidding';
import type { SeedFilter as SeedFilterType } from '../../hooks/useBidding';

interface SeedFilterProps {
  seedFilter: SeedFilterType;
  onSeedFilterChange: (filter: SeedFilterType) => void;
  unbidOnly: boolean;
  onUnbidOnlyChange: (value: boolean) => void;
}

export function SeedFilter({
  seedFilter,
  onSeedFilterChange,
  unbidOnly,
  onUnbidOnlyChange,
}: SeedFilterProps) {
  return (
    <div className="flex flex-wrap items-center gap-3 mb-4">
      <div className="flex items-center gap-2">
        <span className="text-sm text-gray-600">Seeds:</span>
        {SEED_FILTER_OPTIONS.map((filter) => (
          <button
            key={filter}
            onClick={() => onSeedFilterChange(filter)}
            className={cn(
              'px-3 py-1 text-sm rounded-md transition-colors',
              seedFilter === filter
                ? 'bg-blue-600 text-white'
                : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
            )}
          >
            {filter === 'all' ? 'All' : filter}
          </button>
        ))}
      </div>
      <label className="flex items-center gap-2 text-sm text-gray-600 cursor-pointer">
        <input
          type="checkbox"
          checked={unbidOnly}
          onChange={(e) => onUnbidOnlyChange(e.target.checked)}
          className="h-4 w-4 rounded border-gray-300"
        />
        Unbid only
      </label>
    </div>
  );
}
