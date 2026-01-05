import { useRef } from 'react';

export function TabsNav<T extends string>({
  tabs,
  activeTab,
  onTabChange,
}: {
  tabs: readonly { id: T; label: string }[];
  activeTab: T;
  onTabChange: (id: T) => void;
}) {
  const tabRefs = useRef<Record<string, HTMLButtonElement | null>>({});

  const focusTab = (id: T) => {
    const el = tabRefs.current[String(id)];
    if (el) el.focus();
  };

  const handleKeyDown: React.KeyboardEventHandler<HTMLDivElement> = (e) => {
    const currentIndex = tabs.findIndex((t) => t.id === activeTab);
    if (currentIndex < 0) return;

    let nextIndex: number | null = null;
    if (e.key === 'ArrowRight') nextIndex = (currentIndex + 1) % tabs.length;
    else if (e.key === 'ArrowLeft') nextIndex = (currentIndex - 1 + tabs.length) % tabs.length;
    else if (e.key === 'Home') nextIndex = 0;
    else if (e.key === 'End') nextIndex = tabs.length - 1;

    if (nextIndex == null) return;

    e.preventDefault();
    const nextTab = tabs[nextIndex];
    onTabChange(nextTab.id);
    focusTab(nextTab.id);
  };

  return (
    <div
      className="mb-8 flex gap-2 border-b border-gray-200"
      role="tablist"
      aria-orientation="horizontal"
      onKeyDown={handleKeyDown}
    >
      {tabs.map((tab) => (
        <button
          key={tab.id}
          type="button"
          onClick={() => onTabChange(tab.id)}
          ref={(el) => {
            tabRefs.current[String(tab.id)] = el;
          }}
          role="tab"
          aria-selected={activeTab === tab.id}
          tabIndex={activeTab === tab.id ? 0 : -1}
          className={`px-4 py-2 -mb-px border-b-2 font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 focus-visible:ring-offset-2 ${
            activeTab === tab.id ? 'border-blue-600 text-blue-600' : 'border-transparent text-gray-600 hover:text-gray-900'
          }`}
        >
          {tab.label}
        </button>
      ))}
    </div>
  );
}
