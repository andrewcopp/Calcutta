import { Tournament } from '../types/tournament';

const API_URL = 'http://localhost:8080/api';

export const fetchTournaments = async (): Promise<Tournament[]> => {
  try {
    const response = await fetch(`${API_URL}/tournaments`);
    if (!response.ok) {
      throw new Error(`HTTP error! Status: ${response.status}`);
    }
    const data = await response.json();
    return data;
  } catch (error) {
    console.error('Error fetching tournaments:', error);
    throw error;
  }
}; 