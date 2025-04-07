import React from 'react';
import { Link } from 'react-router-dom';

export function RulesPage() {
  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-6">
        <Link to="/" className="text-blue-600 hover:text-blue-800">← Back to Home</Link>
      </div>
      
      <div className="max-w-4xl mx-auto">
        <h1 className="text-4xl font-bold text-gray-900 mb-8">How Calcutta Works</h1>
        
        <div className="space-y-8">
          <section className="bg-white rounded-lg shadow-lg p-6">
            <h2 className="text-2xl font-semibold text-gray-900 mb-4">What is a Calcutta?</h2>
            <p className="text-gray-600 mb-4">
              A Calcutta is an auction system where participants bid on teams in a tournament.
              The goal is to build a portfolio of teams that will earn the most points based on
              their tournament performance.
            </p>
            <p className="text-gray-600">
              The name "Calcutta" comes from the city in India where this auction system was
              first popularized for horse racing in the 19th century.
            </p>
          </section>

          <section className="bg-white rounded-lg shadow-lg p-6">
            <h2 className="text-2xl font-semibold text-gray-900 mb-4">The Auction Process</h2>
            <ol className="list-decimal list-inside space-y-3 text-gray-600">
              <li>Teams are auctioned off one by one to the highest bidder</li>
              <li>Participants can bid on multiple teams to build their portfolio</li>
              <li>The total amount bid on a team determines its ownership percentage</li>
              <li>If you bid $100 on a team that sold for $1000, you own 10% of that team</li>
            </ol>
          </section>

          <section className="bg-white rounded-lg shadow-lg p-6">
            <h2 className="text-2xl font-semibold text-gray-900 mb-4">Scoring System</h2>
            <div className="space-y-4">
              <p className="text-gray-600">
                Points are awarded based on how far teams advance in the tournament:
              </p>
              <ul className="list-disc list-inside space-y-2 text-gray-600">
                <li>First Four Win or Bye: 0 points</li>
                <li>First Round Win: 50 points</li>
                <li>Round of 32 Win: 150 points</li>
                <li>Sweet 16 Win: 300 points</li>
                <li>Elite 8 Win: 500 points</li>
                <li>Final Four Win: 750 points</li>
                <li>Championship Game Win: 1050 points</li>
              </ul>
              <p className="text-gray-600 mt-4">
                Your final score is the sum of your teams' points multiplied by your
                ownership percentage in each team.
              </p>
            </div>
          </section>

          <section className="bg-white rounded-lg shadow-lg p-6">
            <h2 className="text-2xl font-semibold text-gray-900 mb-4">Strategy Tips</h2>
            <ul className="list-disc list-inside space-y-3 text-gray-600">
              <li>Balance your portfolio between favorites and underdogs</li>
              <li>Consider the potential points vs. the cost of ownership</li>
              <li>Look for teams with favorable paths to the later rounds</li>
              <li>Don't overspend on any single team</li>
              <li>Consider the historical performance of different seeds</li>
            </ul>
          </section>
        </div>
      </div>
    </div>
  );
} 