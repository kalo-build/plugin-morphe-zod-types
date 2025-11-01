import { z } from 'zod'
export enum Nationality {
  DE = 'German',
  FR = 'French',
  US = 'American'
}

export const NationalitySchema = z.nativeEnum(Nationality)

export type NationalityType = z.infer<typeof NationalitySchema>
