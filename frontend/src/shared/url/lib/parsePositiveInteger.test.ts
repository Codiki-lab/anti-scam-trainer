import { describe, expect, it } from 'vitest'
import { parsePositiveInteger } from './parsePositiveInteger'

describe('parsePositiveInteger', () => {
  it.each([
    ['1', 1],
    ['42', 42],
    [undefined, null],
    [null, null],
    ['', null],
    ['demo', null],
    ['0', null],
    ['-1', null],
    ['1.5', null],
  ])('parses %s as %s', (value, expected) => {
    expect(parsePositiveInteger(value)).toBe(expected)
  })
})
