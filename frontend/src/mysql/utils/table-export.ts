export type ExportRow = Record<string, unknown>

function normalizeExportCell(value: unknown) {
  if (value == null) {
    return ''
  }

  if (typeof value === 'string' && /^-?\d+\.0+$/.test(value)) {
    return value.replace(/\.0+$/, '')
  }

  return value
}

export async function downloadExcel(dataset: ExportRow[], filename: string) {
  const xlsx = await import('xlsx')
  const exportRows = dataset.map((row) =>
    Object.fromEntries(
      Object.entries(row).map(([key, value]) => [key, normalizeExportCell(value)])
    )
  )

  const worksheet = xlsx.utils.json_to_sheet(exportRows)
  const workbook = xlsx.utils.book_new()
  xlsx.utils.book_append_sheet(workbook, worksheet, 'TableData')
  const workbookArray = xlsx.write(workbook, { bookType: 'xlsx', type: 'array' })
  downloadBlob(
    new Blob([workbookArray], {
      type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet'
    }),
    filename
  )
}

function downloadBlob(blob: Blob, filename: string) {
  const link = document.createElement('a')
  const url = URL.createObjectURL(blob)
  link.href = url
  link.download = filename
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  URL.revokeObjectURL(url)
}

