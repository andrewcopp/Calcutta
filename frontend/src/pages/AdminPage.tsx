import React from 'react';
import { Link } from 'react-router-dom';

export const AdminPage: React.FC = () => {
  return (
    <div className="container mx-auto px-4 py-8">
      <h1 className="text-3xl font-bold mb-8">Admin Console</h1>
      
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        <Link 
          to="/admin/tournaments" 
          className="p-6 bg-white rounded-lg shadow-md hover:shadow-lg transition-shadow"
        >
          <h2 className="text-xl font-semibold mb-2">Tournaments</h2>
          <p className="text-gray-600">Manage tournaments, teams, and brackets</p>
        </Link>

        <Link 
          to="/admin/bundles" 
          className="p-6 bg-white rounded-lg shadow-md hover:shadow-lg transition-shadow"
        >
          <h2 className="text-xl font-semibold mb-2">Bundles</h2>
          <p className="text-gray-600">Export or import bundle archives</p>
        </Link>

        <Link 
          to="/admin/analytics" 
          className="p-6 bg-white rounded-lg shadow-md hover:shadow-lg transition-shadow"
        >
          <h2 className="text-xl font-semibold mb-2">Analytics</h2>
          <p className="text-gray-600">View historical trends and patterns across all calcuttas</p>
        </Link>

        <Link 
          to="/admin/hall-of-fame" 
          className="p-6 bg-white rounded-lg shadow-md hover:shadow-lg transition-shadow"
        >
          <h2 className="text-xl font-semibold mb-2">Hall of Fame</h2>
          <p className="text-gray-600">Leaderboards for best teams, investments, and entries across all years</p>
        </Link>
      </div>
    </div>
  );
};

export default AdminPage; 