export function TabsNav<T extends string>({
  tabs,
  activeTab,
  onTabChange,
}: {
  tabs: readonly { id: T; label: string }[];
  activeTab: T;
  onTabChange: (id: T) => void;
}) {
  return (
    <div className="mb-8 flex gap-2 border-b border-gray-200">
      {tabs.map((tab) => (
        <button
          key={tab.id}
          type="button"
          onClick={() => onTabChange(tab.id)}
          className={`px-4 py-2 -mb-px border-b-2 font-medium transition-colors ${
            activeTab === tab.id ? 'border-blue-600 text-blue-600' : 'border-transparent text-gray-600 hover:text-gray-900'
          }`}
        >
          {tab.label}
        </button>
      ))}
    </div>
  );
}
