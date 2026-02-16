import { School } from '../types/school';
import { apiClient } from '../api/apiClient';

export const schoolService = {
  async getSchools(): Promise<School[]> {
    return apiClient.get<School[]>('/schools');
  },
};
