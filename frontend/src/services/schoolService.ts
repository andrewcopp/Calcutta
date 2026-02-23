import { z } from 'zod';
import { SchoolSchema } from '../schemas/school';
import { apiClient } from '../api/apiClient';

export type { School } from '../schemas/school';

export const schoolService = {
  async getSchools() {
    return apiClient.get('/schools', { schema: z.array(SchoolSchema) });
  },
};
