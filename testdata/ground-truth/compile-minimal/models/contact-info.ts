import { z } from 'zod'
import { PersonSchema, type Person } from './person'

export interface ContactInfo {
  email: string;
  id: number;
  personID: number;
  person?: Person | undefined
}

export const ContactInfoSchema: z.ZodType<ContactInfo> = z.object({
  email: z.string(),
  id: z.number(),
  personID: z.number(),
  person: z.lazy(() => PersonSchema).optional()
})

export const ContactInfoIDEmailSchema = z.object({
  email: z.string()
})

export type ContactInfoIDEmail = z.infer<typeof ContactInfoIDEmailSchema>

export const ContactInfoIDPrimarySchema = z.object({
  id: z.number()
})

export type ContactInfoIDPrimary = z.infer<typeof ContactInfoIDPrimarySchema>
