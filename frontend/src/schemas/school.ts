import { z } from 'zod';

export const SchoolSchema = z.object({
  id: z.string(),
  name: z.string(),
});

export type School = z.infer<typeof SchoolSchema>;
