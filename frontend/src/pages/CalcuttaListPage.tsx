import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { Calcutta } from '../types/calcutta';
import { calcuttaService } from '../services/calcuttaService';

export function CalcuttaListPage() {
  const [calcuttas, setCalcuttas] = useState<Calcutta[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchCalcuttas = async () => {
      try {
        const data = await calcuttaService.getAllCalcuttas();
        setCalcuttas(data);
        setLoading(false);
      } catch (err) {
        setError('Failed to fetch calcuttas');
        setLoading(false);
      }
    };

    fetchCalcuttas();
  }, []);

  if (loading) {
    return <div>Loading...</div>;
  }

  if (error) {
    return <div className="error">{error}</div>;
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-3xl font-bold">Calcuttas</h1>
        <Link
          to="/calcuttas/create"
          className="bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600"
        >
          Create New Calcutta
        </Link>
      </div>

      <div className="grid gap-4">
        {calcuttas.map((calcutta) => (
          <Link
            key={calcutta.id}
            to={`/calcuttas/${calcutta.id}`}
            className="block p-4 bg-white rounded-lg shadow hover:shadow-md transition-shadow"
          >
            <h2 className="text-xl font-semibold">{calcutta.name}</h2>
            <p className="text-gray-600">Created: {new Date(calcutta.created).toLocaleDateString()}</p>
          </Link>
        ))}
        
        {calcuttas.length === 0 && (
          <div className="text-center py-8 bg-white rounded-lg shadow">
            <p className="text-gray-500 mb-4">No calcuttas found.</p>
            <Link
              to="/calcuttas/create"
              className="text-blue-500 hover:text-blue-700"
            >
              Create your first Calcutta
            </Link>
          </div>
        )}
      </div>
    </div>
  );
} 