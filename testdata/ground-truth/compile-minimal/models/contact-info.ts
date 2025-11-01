import { z } from 'zod'
import { PersonSchema } from './person'

export const ContactInfoSchema = z.object({
  email: z.string().optional(),
  id: z.number(),
  personID: z.number().optional(),
  person: z.lazy(() => PersonSchema).optional()
})

export type ContactInfo = z.infer<typeof ContactInfoSchema>

export const ContactInfoIDEmailSchema = z.object({
  email: z.string().optional()
})

export type ContactInfoIDEmail = z.infer<typeof ContactInfoIDEmailSchema>

export const ContactInfoIDPrimarySchema = z.object({
  id: z.number()
})

export type ContactInfoIDPrimary = z.infer<typeof ContactInfoIDPrimarySchema>
