import { z } from 'zod'
import { Nationality, NationalitySchema } from '../enums/nationality'
import { CompanySchema } from './company'
import { ContactInfoSchema } from './contact-info'

export const PersonSchema = z.object({
  firstName: z.string().optional(),
  id: z.number(),
  lastName: z.string().optional(),
  nationality: NationalitySchema.optional(),
  companyID: z.number().optional(),
  company: z.lazy(() => CompanySchema).optional(),
  contactInfoID: z.number().optional(),
  contactInfo: z.lazy(() => ContactInfoSchema).optional()
})

export type Person = z.infer<typeof PersonSchema>

export const PersonIDNameSchema = z.object({
  firstName: z.string().optional(),
  lastName: z.string().optional()
})

export type PersonIDName = z.infer<typeof PersonIDNameSchema>

export const PersonIDPrimarySchema = z.object({
  id: z.number()
})

export type PersonIDPrimary = z.infer<typeof PersonIDPrimarySchema>
