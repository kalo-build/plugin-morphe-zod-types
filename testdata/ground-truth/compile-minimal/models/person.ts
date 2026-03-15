import { z } from 'zod'
import { NationalitySchema, type Nationality } from '../enums/nationality'
import { CompanySchema, type Company } from './company'
import { ContactInfoSchema, type ContactInfo } from './contact-info'

export interface Person {
  firstName: string;
  id: number;
  lastName: string;
  nationality: Nationality;
  companyID: number;
  company?: Company | undefined;
  contactInfoID: number;
  contactInfo?: ContactInfo | undefined
}

export const PersonSchema: z.ZodType<Person> = z.object({
  firstName: z.string(),
  id: z.number(),
  lastName: z.string(),
  nationality: NationalitySchema,
  companyID: z.number(),
  company: z.lazy(() => CompanySchema).optional(),
  contactInfoID: z.number(),
  contactInfo: z.lazy(() => ContactInfoSchema).optional()
})

export const PersonIDNameSchema = z.object({
  firstName: z.string(),
  lastName: z.string()
})

export type PersonIDName = z.infer<typeof PersonIDNameSchema>

export const PersonIDPrimarySchema = z.object({
  id: z.number()
})

export type PersonIDPrimary = z.infer<typeof PersonIDPrimarySchema>
