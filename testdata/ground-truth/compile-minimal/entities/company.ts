import { z } from 'zod'
import { PersonSchema, type Person } from './person'

export interface Company {
  id: number;
  name?: string | undefined;
  taxID?: string | undefined;
  personIDs?: number[] | undefined;
  persons?: Person[] | undefined
}

export const CompanySchema: z.ZodType<Company> = z.object({
  id: z.number(),
  name: z.string().optional(),
  taxID: z.string().optional(),
  personIDs: z.number().array().optional(),
  persons: z.lazy(() => PersonSchema).array().optional()
})

export const CompanyIDPrimarySchema = z.object({
  id: z.number()
})

export type CompanyIDPrimary = z.infer<typeof CompanyIDPrimarySchema>
