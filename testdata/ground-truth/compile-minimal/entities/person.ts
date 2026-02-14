import { z } from 'zod'
import { NationalitySchema, type Nationality } from '../enums/nationality'
import { CompanySchema, type Company } from './company'

export interface Person {
  email: string;
  id: number;
  lastName: string;
  nationality: Nationality;
  companyID?: number | undefined;
  company?: Company | undefined
}

export const PersonSchema: z.ZodType<Person> = z.object({
  email: z.string(),
  id: z.number(),
  lastName: z.string(),
  nationality: NationalitySchema,
  companyID: z.number().optional(),
  company: z.lazy(() => CompanySchema).optional()
})

export const PersonIDPrimarySchema = z.object({
  id: z.number()
})

export type PersonIDPrimary = z.infer<typeof PersonIDPrimarySchema>
