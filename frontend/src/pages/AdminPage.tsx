import React from 'react';
import { Link } from 'react-router-dom';
import { Card } from '../components/ui/Card';
import { PageContainer, PageHeader } from '../components/ui/Page';

export const AdminPage: React.FC = () => {
  return (
    <PageContainer>
      <PageHeader title="Admin Console" />

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        <Link 
          to="/admin/tournaments" 
          className="block"
        >
          <Card className="hover:shadow-md transition-shadow">
            <h2 className="text-xl font-semibold mb-2">Tournaments</h2>
            <p className="text-gray-600">Manage tournaments, teams, and brackets</p>
          </Card>
        </Link>

        <Link 
          to="/admin/bundles" 
          className="block"
        >
          <Card className="hover:shadow-md transition-shadow">
            <h2 className="text-xl font-semibold mb-2">Bundles</h2>
            <p className="text-gray-600">Export or import bundle archives</p>
          </Card>
        </Link>

        <Link 
          to="/admin/api-keys" 
          className="block"
        >
          <Card className="hover:shadow-md transition-shadow">
            <h2 className="text-xl font-semibold mb-2">API Keys</h2>
            <p className="text-gray-600">Create, view, and revoke server-to-server API keys</p>
          </Card>
        </Link>

        <Link 
          to="/admin/users" 
          className="block"
        >
          <Card className="hover:shadow-md transition-shadow">
            <h2 className="text-xl font-semibold mb-2">Users</h2>
            <p className="text-gray-600">View users and their roles/permissions</p>
          </Card>
        </Link>

        <Link 
          to="/admin/hall-of-fame" 
          className="block"
        >
          <Card className="hover:shadow-md transition-shadow">
            <h2 className="text-xl font-semibold mb-2">Hall of Fame</h2>
            <p className="text-gray-600">Leaderboards for best teams, investments, and entries across all years</p>
          </Card>
        </Link>
      </div>
    </PageContainer>
  );
};