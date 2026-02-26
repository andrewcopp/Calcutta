import { z } from 'zod';
import { SchoolSchema } from '../schemas/school';
import { apiClient } from '../api/apiClient';

export type { School } from '../schemas/school';

export const schoolService = {
  async getSchools() {
    const res = await apiClient.get('/schools', { schema: z.object({ items: z.array(SchoolSchema) }) });
    return res.items;
  },
};
