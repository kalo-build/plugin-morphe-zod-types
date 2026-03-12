import { z } from 'zod'
import { NationalitySchema, type Nationality } from '../enums/nationality'
import { CompanySchema, type Company } from './company'

export interface Person {
  email: string;
  id: number;
  lastName: string;
  nationality: Nationality;
  companyID: number;
  company: Company
}

export const PersonSchema: z.ZodType<Person> = z.object({
  email: z.string(),
  id: z.number(),
  lastName: z.string(),
  nationality: NationalitySchema,
  companyID: z.number(),
  company: z.lazy(() => CompanySchema)
})

export const PersonIDPrimarySchema = z.object({
  id: z.number()
})

export type PersonIDPrimary = z.infer<typeof PersonIDPrimarySchema>
