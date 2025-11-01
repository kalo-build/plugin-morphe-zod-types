import { z } from 'zod'

export const AddressSchema = z.object({
  city: z.string().optional(),
  houseNr: z.string().optional(),
  street: z.string().optional(),
  zipCode: z.string().optional()
})

export type Address = z.infer<typeof AddressSchema>
