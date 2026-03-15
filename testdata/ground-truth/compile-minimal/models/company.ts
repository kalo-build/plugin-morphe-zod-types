import { z } from 'zod'
import { PersonSchema, type Person } from './person'

export interface Company {
  id: number;
  name: string;
  taxID: string;
  personIDs: number[];
  persons?: Person[] | undefined
}

export const CompanySchema: z.ZodType<Company> = z.object({
  id: z.number(),
  name: z.string(),
  taxID: z.string(),
  personIDs: z.number().array(),
  persons: z.lazy(() => PersonSchema).array().optional()
})

export const CompanyIDNameSchema = z.object({
  name: z.string()
})

export type CompanyIDName = z.infer<typeof CompanyIDNameSchema>

export const CompanyIDPrimarySchema = z.object({
  id: z.number()
})

export type CompanyIDPrimary = z.infer<typeof CompanyIDPrimarySchema>
