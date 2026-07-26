import path from 'node:path'
import { fileURLToPath } from 'node:url'
import ExcelJS from 'exceljs'
import { describe, expect, it } from 'vitest'

const testDir = path.dirname(fileURLToPath(import.meta.url))

describe('generated Luban workbooks', () => {
  it('keeps the standard four-row header in #MapNode.xlsx', async () => {
    const workbook = new ExcelJS.Workbook()
    await workbook.xlsx.readFile(path.resolve(testDir, '..', '..', '..', 'configs', 'Datas', '#MapNode.xlsx'))
    const sheet = workbook.worksheets[0]

    expect(sheet.getRow(1).getCell(1).value).toBe('##var')
    expect(sheet.getRow(2).getCell(1).value).toBe('##type')
    expect(sheet.getRow(3).getCell(1).value).toBe('##group')
    expect(sheet.getRow(4).getCell(1).value).toBe('##')
    expect(sheet.getRow(1).values).toContain('area_usages')
    expect(sheet.getRow(1).values).toContain('points')
    expect(sheet.getRow(5).getCell(1).value).toBeNull()
  })
})
