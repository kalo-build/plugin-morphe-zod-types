import { z } from 'zod'
import { NationalitySchema, type Nationality } from '../enums/nationality'
import { CompanySchema, type Company } from './company'
import { ContactInfoSchema, type ContactInfo } from './contact-info'

export interface Person {
  firstName?: string | undefined;
  id: number;
  lastName?: string | undefined;
  nationality?: Nationality | undefined;
  companyID?: number | undefined;
  company?: Company | undefined;
  contactInfoID?: number | undefined;
  contactInfo?: ContactInfo | undefined
}

export const PersonSchema: z.ZodType<Person> = z.object({
  firstName: z.string().optional(),
  id: z.number(),
  lastName: z.string().optional(),
  nationality: NationalitySchema.optional(),
  companyID: z.number().optional(),
  company: z.lazy(() => CompanySchema).optional(),
  contactInfoID: z.number().optional(),
  contactInfo: z.lazy(() => ContactInfoSchema).optional()
})

export const PersonIDNameSchema = z.object({
  firstName: z.string().optional(),
  lastName: z.string().optional()
})

export type PersonIDName = z.infer<typeof PersonIDNameSchema>

export const PersonIDPrimarySchema = z.object({
  id: z.number()
})

export type PersonIDPrimary = z.infer<typeof PersonIDPrimarySchema>
