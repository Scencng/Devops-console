import request from '@/mysql/utils/request'

export type ExplorerNodeType = 'database' | 'group' | 'table' | 'view' | 'function' | 'query' | 'backup'
export type ExplorerGroupKind = 'tables' | 'views' | 'functions' | 'queries' | 'backups'

export interface TreeNodeData {
  key: string
  type: ExplorerNodeType
  label: string
  databaseName: string
  tableName?: string
  groupKind?: ExplorerGroupKind
  children?: TreeNodeData[]
  isLeaf?: boolean
  childrenLoaded?: boolean
  loadingChildren?: boolean
}

interface QueryRowsResponse {
  rows?: Array<Record<string, unknown>>
}

interface BackupRecordResponse {
  records?: Array<{
    fileName: string
  }>
}

interface ExplorerDatabaseObjects {
  tables: string[]
  views: string[]
  functions: string[]
  backups: string[]
}

let explorerTreeCache: string[] | null = null
let explorerTreeRequest: Promise<string[]> | null = null
const explorerObjectsCache = new Map<string, ExplorerDatabaseObjects>()
const explorerObjectsRequest = new Map<string, Promise<ExplorerDatabaseObjects>>()

function formatSQLStringLiteral(value: string) {
  return `'${value.replace(/\\/g, '\\\\').replace(/'/g, "\\'")}'`
}

function buildObjectsCacheKey(databaseName: string, includeBackups: boolean) {
  return `${databaseName}:${includeBackups ? '1' : '0'}`
}

function cloneObjects(objects: ExplorerDatabaseObjects): ExplorerDatabaseObjects {
  return {
    tables: [...objects.tables],
    views: [...objects.views],
    functions: [...objects.functions],
    backups: [...objects.backups]
  }
}

async function requestDatabaseObjects(databaseName: string, options?: { includeBackups?: boolean }) {
  const includeBackups = options?.includeBackups ?? false
  const [tables, views, functions, backupResponse] = await Promise.all([
    request.get<string[]>('/api/metadata/tables', {
      params: { db: databaseName }
    }),
    request
      .post<QueryRowsResponse>('/api/query/execute', {
        sql: `
          SELECT TABLE_NAME
          FROM information_schema.VIEWS
          WHERE TABLE_SCHEMA = ${formatSQLStringLiteral(databaseName)}
          ORDER BY TABLE_NAME
        `,
        database: databaseName
      })
      .then((result) =>
        (result.rows ?? [])
          .map((row) => String(row.TABLE_NAME ?? row.table_name ?? '').trim())
          .filter(Boolean)
      )
      .catch(() => [] as string[]),
    request
      .post<QueryRowsResponse>('/api/query/execute', {
        sql: `
          SELECT ROUTINE_NAME
          FROM information_schema.ROUTINES
          WHERE ROUTINE_SCHEMA = ${formatSQLStringLiteral(databaseName)}
            AND ROUTINE_TYPE = 'FUNCTION'
          ORDER BY ROUTINE_NAME
        `,
        database: databaseName
      })
      .then((result) =>
        (result.rows ?? [])
          .map((row) => String(row.ROUTINE_NAME ?? row.routine_name ?? '').trim())
          .filter(Boolean)
      )
      .catch(() => [] as string[]),
    includeBackups
      ? request
          .get<BackupRecordResponse>('/api/backup/list', {
            params: { database: databaseName }
          })
          .catch(() => ({ records: [] as Array<{ fileName: string }> }))
      : Promise.resolve({ records: [] as Array<{ fileName: string }> })
  ])

  const viewSet = new Set(views)
  const plainTables = tables.filter((name) => !viewSet.has(name))

  return {
    tables: plainTables,
    views,
    functions,
    backups: (backupResponse.records ?? []).map((record) => record.fileName)
  }
}

export async function fetchDatabaseObjects(databaseName: string, options?: { includeBackups?: boolean; force?: boolean }) {
  const includeBackups = options?.includeBackups ?? false
  const force = options?.force ?? false
  const cacheKey = buildObjectsCacheKey(databaseName, includeBackups)

  if (!force) {
    const cached = explorerObjectsCache.get(cacheKey)
    if (cached) {
      return cloneObjects(cached)
    }

    const pending = explorerObjectsRequest.get(cacheKey)
    if (pending) {
      return pending.then(cloneObjects)
    }
  }

  const nextRequest = requestDatabaseObjects(databaseName, { includeBackups })
    .then((objects) => {
      explorerObjectsCache.set(cacheKey, cloneObjects(objects))
      return objects
    })
    .finally(() => {
      explorerObjectsRequest.delete(cacheKey)
    })

  explorerObjectsRequest.set(cacheKey, nextRequest)
  return nextRequest.then(cloneObjects)
}

export function createGroupNode(databaseName: string, groupKind: ExplorerGroupKind, children: TreeNodeData[] = []) {
  return {
    key: `db:${databaseName}:${groupKind}`,
    type: 'group',
    label: groupKind,
    databaseName,
    groupKind,
    children
  } as TreeNodeData
}

export function createDatabaseNode(databaseName: string) {
  return {
    key: `db:${databaseName}`,
    type: 'database',
    label: databaseName,
    databaseName,
    children: [
      createGroupNode(databaseName, 'tables'),
      createGroupNode(databaseName, 'views'),
      createGroupNode(databaseName, 'functions'),
      createGroupNode(databaseName, 'queries'),
      createGroupNode(databaseName, 'backups')
    ],
    childrenLoaded: false,
    loadingChildren: false
  } as TreeNodeData
}

function mapTreeObjectChildren(databaseName: string, groupKind: ExplorerGroupKind, itemNames: string[]) {
  const typeMap: Record<ExplorerGroupKind, ExplorerNodeType> = {
    tables: 'table',
    views: 'view',
    functions: 'function',
    queries: 'query',
    backups: 'backup'
  }

  const nodeType = typeMap[groupKind]

  return itemNames.map((itemName) => ({
    key: `db:${databaseName}:${nodeType}:${itemName}`,
    type: nodeType,
    label: itemName,
    databaseName,
    tableName: itemName,
    isLeaf: true
  } as TreeNodeData))
}

export async function ensureDatabaseChildrenLoaded(databaseNode: TreeNodeData, options?: { includeBackups?: boolean; force?: boolean }) {
  if (databaseNode.type !== 'database') {
    return
  }

  const includeBackups = options?.includeBackups ?? false
  const force = options?.force ?? false
  const backupsGroup = databaseNode.children?.find((node) => node.type === 'group' && node.groupKind === 'backups')
  const backupsLoaded = Boolean(backupsGroup?.childrenLoaded)

  if (!force) {
    if (databaseNode.loadingChildren) {
      return
    }

    if (databaseNode.childrenLoaded && (!includeBackups || backupsLoaded)) {
      return
    }
  }

  databaseNode.loadingChildren = true

  try {
    const objects = await fetchDatabaseObjects(databaseNode.databaseName, { includeBackups, force })
    const groupMap = new Map<ExplorerGroupKind, TreeNodeData>()

    for (const child of databaseNode.children ?? []) {
      if (child.type === 'group' && child.groupKind) {
        groupMap.set(child.groupKind, child)
      }
    }

    const tablesGroup = groupMap.get('tables') ?? createGroupNode(databaseNode.databaseName, 'tables')
    tablesGroup.children = mapTreeObjectChildren(databaseNode.databaseName, 'tables', objects.tables)
    tablesGroup.childrenLoaded = true

    const viewsGroup = groupMap.get('views') ?? createGroupNode(databaseNode.databaseName, 'views')
    viewsGroup.children = mapTreeObjectChildren(databaseNode.databaseName, 'views', objects.views)
    viewsGroup.childrenLoaded = true

    const functionsGroup = groupMap.get('functions') ?? createGroupNode(databaseNode.databaseName, 'functions')
    functionsGroup.children = mapTreeObjectChildren(databaseNode.databaseName, 'functions', objects.functions)
    functionsGroup.childrenLoaded = true

    const queriesGroup = groupMap.get('queries') ?? createGroupNode(databaseNode.databaseName, 'queries')
    queriesGroup.children = []
    queriesGroup.childrenLoaded = true

    const nextChildren = [tablesGroup, viewsGroup, functionsGroup, queriesGroup]
    const resolvedBackupsGroup = backupsGroup ?? createGroupNode(databaseNode.databaseName, 'backups')
    if (includeBackups || force) {
      resolvedBackupsGroup.children = mapTreeObjectChildren(databaseNode.databaseName, 'backups', objects.backups)
      resolvedBackupsGroup.childrenLoaded = true
    }
    nextChildren.push(resolvedBackupsGroup)

    databaseNode.children = nextChildren
    databaseNode.childrenLoaded = true
  } finally {
    databaseNode.loadingChildren = false
  }
}

async function fetchExplorerTree(force = false) {
  if (!force) {
    if (explorerTreeCache) {
      return [...explorerTreeCache]
    }

    if (explorerTreeRequest) {
      return explorerTreeRequest.then((items) => [...items])
    }
  }

  explorerTreeRequest = request
    .get<string[]>('/api/metadata/databases')
    .then((databases) => {
      explorerTreeCache = [...databases]
      return databases
    })
    .finally(() => {
      explorerTreeRequest = null
    })

  return explorerTreeRequest.then((items) => [...items])
}

export async function loadExplorerTree(options?: { force?: boolean }) {
  const databases = await fetchExplorerTree(options?.force ?? false)
  return databases.map((databaseName) => createDatabaseNode(databaseName))
}

export function preloadExplorerTree() {
  return fetchExplorerTree(false)
}

export function preloadDatabaseObjects(databaseName: string, options?: { includeBackups?: boolean }) {
  return fetchDatabaseObjects(databaseName, options)
}

export function findNodeByKey(nodes: TreeNodeData[], key: string): TreeNodeData | null {
  for (const node of nodes) {
    if (node.key === key) {
      return node
    }

    const children = node.children ?? []
    const matched = findNodeByKey(children, key)
    if (matched) {
      return matched
    }
  }

  return null
}
