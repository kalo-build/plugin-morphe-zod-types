import { z } from 'zod'
import { Nationality, NationalitySchema } from '../enums/nationality'
import { CompanySchema } from './company'

export const PersonSchema = z.object({
  email: z.string().optional(),
  id: z.number(),
  lastName: z.string().optional(),
  nationality: NationalitySchema.optional(),
  companyID: z.number().optional(),
  company: z.lazy(() => CompanySchema).optional()
})

export type Person = z.infer<typeof PersonSchema>

export const PersonIDPrimarySchema = z.object({
  id: z.number()
})

export type PersonIDPrimary = z.infer<typeof PersonIDPrimarySchema>
