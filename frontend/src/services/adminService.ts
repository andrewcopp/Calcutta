import { School } from '../types/school';
import { apiClient } from '../api/apiClient';

export const adminService = {
  async getAllSchools(): Promise<School[]> {
    return apiClient.get<School[]>('/schools');
  }
}; 