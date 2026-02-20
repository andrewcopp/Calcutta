import React, { useState, useRef, useEffect, useCallback } from 'react';
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
}

export const Combobox: React.FC<ComboboxProps> = ({
  options,
  value,
  onChange,
  onSelect,
  placeholder = 'Search...',
  disabled = false,
  excludeIds,
  className,
}) => {
  const [isOpen, setIsOpen] = useState(false);
  const [highlightIndex, setHighlightIndex] = useState(-1);
  const containerRef = useRef<HTMLDivElement>(null);
  const listRef = useRef<HTMLUListElement>(null);

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
    [onChange, onSelect]
  );

  const handleKeyDown = (e: React.KeyboardEvent) => {
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
        value={value}
        onChange={(e) => {
          onChange(e.target.value);
          setIsOpen(true);
          setHighlightIndex(0);
        }}
        onFocus={() => {
          setIsOpen(true);
          setHighlightIndex(0);
        }}
        onKeyDown={handleKeyDown}
        placeholder={placeholder}
        disabled={disabled}
        className="h-10 w-full rounded-lg border border-border bg-surface px-4 py-2 text-sm text-text outline-none focus:ring-2 focus:ring-primary focus:border-primary disabled:opacity-50"
      />
      {isOpen && filtered.length > 0 && (
        <ul
          ref={listRef}
          className="absolute z-50 mt-1 max-h-60 w-full overflow-auto rounded-lg border border-border bg-surface shadow-lg"
          role="listbox"
        >
          {filtered.map((opt, i) => (
            <li
              key={opt.id}
              role="option"
              aria-selected={i === highlightIndex}
              className={cn(
                'cursor-pointer px-4 py-2 text-sm',
                i === highlightIndex ? 'bg-blue-50 text-blue-700' : 'text-text hover:bg-gray-50'
              )}
              onMouseDown={(e) => {
                e.preventDefault();
                handleSelect(opt.id, opt.label);
              }}
              onMouseEnter={() => setHighlightIndex(i)}
            >
              {opt.label}
            </li>
          ))}
        </ul>
      )}
      {isOpen && filtered.length === 0 && value && (
        <div className="absolute z-50 mt-1 w-full rounded-lg border border-border bg-surface px-4 py-2 text-sm text-gray-500 shadow-lg">
          No results found
        </div>
      )}
    </div>
  );
};
