import { useState, useRef, useEffect, useCallback, useId, type ReactNode, type KeyboardEvent } from 'react';
import { cn } from '../../lib/cn';

export interface ComboboxOption {
  id: string;
  label: string;
}

interface ComboboxProps {
  options: ComboboxOption[];
  value: string;
  onChange: (value: string) => void;
  onSelect: (id: string) => void;
  placeholder?: string;
  disabled?: boolean;
  excludeIds?: Set<string>;
  className?: string;
  onBlur?: () => void;
  validationState?: 'none' | 'valid' | 'error';
  renderOption?: (option: ComboboxOption, isHighlighted: boolean) => ReactNode;
}

export function Combobox({
  options,
  value,
  onChange,
  onSelect,
  placeholder = 'Search...',
  disabled = false,
  excludeIds,
  className,
  onBlur,
  validationState,
  renderOption,
}: ComboboxProps) {
  const [isOpen, setIsOpen] = useState(false);
  const [highlightIndex, setHighlightIndex] = useState(-1);
  const [dropAbove, setDropAbove] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);
  const listRef = useRef<HTMLUListElement>(null);
  const instanceId = useId();
  const listboxId = `${instanceId}-listbox`;
  const getOptionId = (index: number) => `${instanceId}-option-${index}`;

  const filtered = options.filter((opt) => {
    if (excludeIds?.has(opt.id)) return false;
    if (!value) return true;
    return opt.label.toLowerCase().includes(value.toLowerCase());
  });

  const handleSelect = useCallback(
    (id: string, label: string) => {
      onChange(label);
      onSelect(id);
      setIsOpen(false);
      setHighlightIndex(-1);
    },
    [onChange, onSelect],
  );

  const handleKeyDown = (e: KeyboardEvent) => {
    if (!isOpen) {
      if (e.key === 'ArrowDown' || e.key === 'Enter') {
        setIsOpen(true);
        setHighlightIndex(0);
        e.preventDefault();
      }
      return;
    }

    switch (e.key) {
      case 'ArrowDown':
        e.preventDefault();
        setHighlightIndex((prev) => Math.min(prev + 1, filtered.length - 1));
        break;
      case 'ArrowUp':
        e.preventDefault();
        setHighlightIndex((prev) => Math.max(prev - 1, 0));
        break;
      case 'Enter':
        e.preventDefault();
        if (highlightIndex >= 0 && highlightIndex < filtered.length) {
          handleSelect(filtered[highlightIndex].id, filtered[highlightIndex].label);
        }
        break;
      case 'Escape':
        e.preventDefault();
        setIsOpen(false);
        setHighlightIndex(-1);
        break;
    }
  };

  // Scroll highlighted item into view
  useEffect(() => {
    if (highlightIndex >= 0 && listRef.current) {
      const item = listRef.current.children[highlightIndex] as HTMLElement;
      item?.scrollIntoView({ block: 'nearest' });
    }
  }, [highlightIndex]);

  // Close on outside click
  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setIsOpen(false);
        setHighlightIndex(-1);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  return (
    <div ref={containerRef} className={cn('relative', className)}>
      <input
        type="text"
        role="combobox"
        aria-expanded={isOpen && filtered.length > 0}
        aria-controls={listboxId}
        aria-activedescendant={highlightIndex >= 0 ? getOptionId(highlightIndex) : undefined}
        aria-autocomplete="list"
        aria-haspopup="listbox"
        value={value}
        onChange={(e) => {
          onChange(e.target.value);
          setIsOpen(true);
          setHighlightIndex(0);
        }}
        onFocus={() => {
          if (containerRef.current) {
            const rect = containerRef.current.getBoundingClientRect();
            const spaceBelow = window.innerHeight - rect.bottom;
            setDropAbove(spaceBelow < 280);
          }
          setIsOpen(true);
          setHighlightIndex(0);
        }}
        onBlur={() => {
          if (onBlur) {
            setTimeout(onBlur, 150);
          }
        }}
        onKeyDown={handleKeyDown}
        placeholder={placeholder}
        disabled={disabled}
        className={cn(
          'h-10 w-full rounded-lg border bg-card px-4 py-2 text-sm text-foreground outline-none disabled:opacity-50',
          validationState === 'valid'
            ? 'border-green-500 ring-2 ring-green-200'
            : validationState === 'error'
              ? 'border-red-500 ring-2 ring-red-200'
              : 'border-border focus:ring-2 focus:ring-primary focus:border-primary',
        )}
      />
      {isOpen && filtered.length > 0 && (
        <ul
          ref={listRef}
          className={cn(
            'absolute z-50 max-h-60 w-full overflow-auto rounded-lg border border-border bg-card shadow-lg',
            dropAbove ? 'bottom-full mb-1' : 'mt-1',
          )}
          id={listboxId}
          role="listbox"
        >
          {filtered.map((opt, i) => (
            <li
              key={opt.id}
              id={getOptionId(i)}
              role="option"
              aria-selected={i === highlightIndex}
              className={cn(
                'cursor-pointer px-4 py-2 text-sm',
                i === highlightIndex ? 'bg-primary/10 text-primary' : 'text-foreground hover:bg-accent',
              )}
              onMouseDown={(e) => {
                e.preventDefault();
                handleSelect(opt.id, opt.label);
              }}
              onMouseEnter={() => setHighlightIndex(i)}
            >
              {renderOption ? renderOption(opt, i === highlightIndex) : opt.label}
            </li>
          ))}
        </ul>
      )}
      {isOpen && filtered.length === 0 && value && (
        <div
          className={cn(
            'absolute z-50 w-full rounded-lg border border-border bg-card px-4 py-2 text-sm text-muted-foreground shadow-lg',
            dropAbove ? 'bottom-full mb-1' : 'mt-1',
          )}
        >
          No results found
        </div>
      )}
    </div>
  );
}
