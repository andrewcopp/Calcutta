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

        {/* Add more admin sections here as needed */}
      </div>
    </div>
  );
};

export default AdminPage; 