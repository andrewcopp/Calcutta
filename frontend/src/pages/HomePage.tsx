import React from 'react';
import { Link } from 'react-router-dom';
import { AuthForm } from '../components/Auth/AuthForm';
import { useUser } from '../contexts/UserContext';

export function HomePage() {
  const { user } = useUser();

  if (!user) {
    return (
      <div className="min-h-screen bg-gradient-to-b from-blue-50 to-white">
        <div className="container mx-auto px-4 py-16">
          <div className="max-w-4xl mx-auto text-center">
            <h1 className="text-5xl font-bold text-gray-900 mb-6">
              Welcome to Calcutta
            </h1>
            <p className="text-xl text-gray-600 mb-8">
              The ultimate NCAA Tournament auction platform where strategy meets excitement.
              Bid on teams, build your portfolio, and compete for glory in the most thrilling
              college basketball tournament of the year.
            </p>
            <AuthForm />
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gradient-to-b from-blue-50 to-white">
      <div className="container mx-auto px-4 py-16">
        <div className="max-w-4xl mx-auto text-center">
          <h1 className="text-5xl font-bold text-gray-900 mb-6">
            Welcome to Calcutta
          </h1>
          <p className="text-xl text-gray-600 mb-8">
            The ultimate NCAA Tournament auction platform where strategy meets excitement.
            Bid on teams, build your portfolio, and compete for glory in the most thrilling
            college basketball tournament of the year.
          </p>
          
          <div className="grid grid-cols-1 md:grid-cols-2 gap-8 mb-16">
            <Link
              to="/rules"
              className="p-6 bg-white rounded-lg shadow-lg hover:shadow-xl transition-shadow"
            >
              <h2 className="text-2xl font-semibold text-blue-600 mb-3">How It Works</h2>
              <p className="text-gray-600">
                Learn about the Calcutta auction system, scoring rules, and strategies
                to build a winning portfolio.
              </p>
            </Link>
            
            <Link
              to="/calcuttas"
              className="p-6 bg-white rounded-lg shadow-lg hover:shadow-xl transition-shadow"
            >
              <h2 className="text-2xl font-semibold text-blue-600 mb-3">Historical Calcuttas</h2>
              <p className="text-gray-600">
                Explore past tournaments, view winning strategies, and see how previous
                participants fared in their Calcutta adventures.
              </p>
            </Link>
          </div>

          <div className="bg-white rounded-lg shadow-lg p-8 mb-8">
            <h2 className="text-2xl font-semibold text-gray-900 mb-4">Ready to Join?</h2>
            <p className="text-gray-600 mb-6">
              Create your entry, bid on teams, and compete for the championship.
              The tournament is waiting for your strategic brilliance!
            </p>
            <Link
              to="/calcuttas"
              className="inline-block bg-blue-600 text-white px-8 py-3 rounded-lg font-semibold hover:bg-blue-700 transition-colors"
            >
              View Active Calcuttas
            </Link>
          </div>
        </div>
      </div>
    </div>
  );
} 