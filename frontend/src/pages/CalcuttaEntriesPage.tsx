import { useEffect, useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import { CalcuttaEntry } from '../types/calcutta';
import { calcuttaService } from '../services/calcuttaService';

export function CalcuttaEntriesPage() {
  const { calcuttaId } = useParams<{ calcuttaId: string }>();
  const [entries, setEntries] = useState<CalcuttaEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchEntries = async () => {
      if (!calcuttaId) return;
      
      try {
        const data = await calcuttaService.getCalcuttaEntries(calcuttaId);
        setEntries(data);
        setLoading(false);
      } catch (err) {
        setError('Failed to fetch entries');
        setLoading(false);
      }
    };

    fetchEntries();
  }, [calcuttaId]);

  if (loading) {
    return <div>Loading...</div>;
  }

  if (error) {
    return <div className="error">{error}</div>;
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-6">
        <Link to="/" className="text-blue-600 hover:text-blue-800">← Back to Calcuttas</Link>
      </div>
      <h1 className="text-3xl font-bold mb-6">Entries</h1>
      <div className="grid gap-4">
        {entries.map((entry) => (
          <Link
            key={entry.id}
            to={`/calcuttas/${calcuttaId}/entries/${entry.id}`}
            className="block p-4 bg-white rounded-lg shadow hover:shadow-md transition-shadow"
          >
            <h2 className="text-xl font-semibold">{entry.name}</h2>
            <p className="text-gray-600">Created: {new Date(entry.created).toLocaleDateString()}</p>
          </Link>
        ))}
      </div>
    </div>
  );
} 