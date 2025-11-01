import { z } from 'zod'
export enum UniversalNumber {
  Euler = 2.7182818285,
  Pi = 3.1415926535
}

export const UniversalNumberSchema = z.nativeEnum(UniversalNumber)

export type UniversalNumberType = z.infer<typeof UniversalNumberSchema>
