import { z } from 'zod'
import { NationalitySchema, type Nationality } from '../enums/nationality'
import { CompanySchema, type Company } from './company'

export interface Person {
  email?: string | undefined;
  id: number;
  lastName?: string | undefined;
  nationality?: Nationality | undefined;
  companyID?: number | undefined;
  company?: Company | undefined
}

export const PersonSchema: z.ZodType<Person> = z.object({
  email: z.string().optional(),
  id: z.number(),
  lastName: z.string().optional(),
  nationality: NationalitySchema.optional(),
  companyID: z.number().optional(),
  company: z.lazy(() => CompanySchema).optional()
})

export const PersonIDPrimarySchema = z.object({
  id: z.number()
})

export type PersonIDPrimary = z.infer<typeof PersonIDPrimarySchema>
