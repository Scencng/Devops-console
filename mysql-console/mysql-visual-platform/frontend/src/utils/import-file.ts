export async function readImportMatrix(file: File, noWorksheetMessage: string) {
  const xlsx = await import('xlsx')
  const extension = file.name.split('.').pop()?.toLowerCase() ?? ''

  if (extension === 'csv') {
    return parseCSVMatrix(await file.text())
  }

  const buffer = await file.arrayBuffer()
  const workbook = xlsx.read(buffer, { type: 'array', raw: false, cellDates: false })
  const firstSheetName = workbook.SheetNames[0]
  if (!firstSheetName) {
    throw new Error(noWorksheetMessage)
  }

  const worksheet = workbook.Sheets[firstSheetName]
  return xlsx.utils.sheet_to_json<unknown[]>(worksheet, {
    header: 1,
    defval: null,
    raw: false,
    blankrows: false
  })
}

export function normalizeImportedColumnName(header: string, index: number, usedNames: Set<string>) {
  const cleanedHeader = header.replace(/^\uFEFF/, '').trim()
  const baseName = sanitizeImportedIdentifier(cleanedHeader || `column_${index + 1}`)
  let candidate = baseName || `column_${index + 1}`
  let suffix = 2

  while (usedNames.has(candidate.toLowerCase())) {
    candidate = `${baseName || `column_${index + 1}`}_${suffix}`
    suffix += 1
  }

  usedNames.add(candidate.toLowerCase())
  return candidate
}

export function sanitizeImportedIdentifier(value: string) {
  return value
    .replace(/[`"'\\]/g, '')
    .replace(/\s+/g, '_')
    .replace(/[^\p{L}\p{N}_$]/gu, '_')
    .replace(/^_+|_+$/g, '')
}

export function normalizeImportedCellValue(value: unknown): unknown {
  if (value == null) {
    return null
  }

  if (typeof value === 'number' || typeof value === 'boolean') {
    return value
  }

  const text = normalizeImportedTextValue(value)
  if (!text) {
    return null
  }

  if (/^-?\d+$/.test(text)) {
    if (/^0\d+/.test(text) || text.length >= 11) {
      return text
    }

    const parsed = Number.parseInt(text, 10)
    return Number.isNaN(parsed) ? text : parsed
  }

  if (/^-?\d+\.\d+$/.test(text)) {
    const parsed = Number.parseFloat(text)
    return Number.isNaN(parsed) ? text : parsed
  }

  if (/^(true|false)$/i.test(text)) {
    return text.toLowerCase() === 'true'
  }

  return text
}

export function normalizeImportedTextValue(value: unknown) {
  const compact = String(value ?? '')
    .replace(/\u00A0/g, ' ')
    .replace(/\u3000/g, ' ')
    .replace(/[\r\n\t]+/g, ' ')
    .trim()
    .replace(/[^\S\r\n]+/g, ' ')

  if (!compact) {
    return ''
  }

  let normalized = compact
  let previous = ''
  while (normalized !== previous) {
    previous = normalized
    normalized = normalized.replace(/([\u3400-\u9fff])\s+(?=[\u3400-\u9fff])/gu, '$1')
  }

  return normalized
}

export function parseCSVMatrix(source: string) {
  const rows: string[][] = []
  let currentRow: string[] = []
  let currentCell = ''
  let inQuotes = false

  for (let index = 0; index < source.length; index += 1) {
    const char = source[index]
    const nextChar = source[index + 1]

    if (char === '"') {
      if (inQuotes && nextChar === '"') {
        currentCell += '"'
        index += 1
      } else {
        inQuotes = !inQuotes
      }
      continue
    }

    if (char === ',' && !inQuotes) {
      currentRow.push(currentCell)
      currentCell = ''
      continue
    }

    if ((char === '\n' || char === '\r') && !inQuotes) {
      if (char === '\r' && nextChar === '\n') {
        index += 1
      }
      currentRow.push(currentCell)
      rows.push(currentRow)
      currentRow = []
      currentCell = ''
      continue
    }

    currentCell += char
  }

  if (currentCell.length > 0 || currentRow.length > 0) {
    currentRow.push(currentCell)
    rows.push(currentRow)
  }

  return rows
}
