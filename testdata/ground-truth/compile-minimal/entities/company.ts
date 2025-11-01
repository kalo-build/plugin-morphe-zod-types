import { z } from 'zod'
import { PersonSchema } from './person'

export const CompanySchema = z.object({
  id: z.number(),
  name: z.string().optional(),
  taxID: z.string().optional(),
  personIDs: z.number().array().optional(),
  persons: z.lazy(() => PersonSchema).array().optional()
})

export type Company = z.infer<typeof CompanySchema>

export const CompanyIDPrimarySchema = z.object({
  id: z.number()
})

export type CompanyIDPrimary = z.infer<typeof CompanyIDPrimarySchema>
