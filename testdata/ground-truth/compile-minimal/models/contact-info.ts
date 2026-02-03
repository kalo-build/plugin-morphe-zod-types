import { z } from 'zod'
import { PersonSchema, type Person } from './person'

export interface ContactInfo {
  email?: string | undefined;
  id: number;
  personID?: number | undefined;
  person?: Person | undefined
}

export const ContactInfoSchema: z.ZodType<ContactInfo> = z.object({
  email: z.string().optional(),
  id: z.number(),
  personID: z.number().optional(),
  person: z.lazy(() => PersonSchema).optional()
})

export const ContactInfoIDEmailSchema = z.object({
  email: z.string().optional()
})

export type ContactInfoIDEmail = z.infer<typeof ContactInfoIDEmailSchema>

export const ContactInfoIDPrimarySchema = z.object({
  id: z.number()
})

export type ContactInfoIDPrimary = z.infer<typeof ContactInfoIDPrimarySchema>
