import { z } from 'zod'

export const AddressSchema = z.object({
  city: z.string(),
  houseNr: z.string(),
  street: z.string(),
  zipCode: z.string()
})

export type Address = z.infer<typeof AddressSchema>
