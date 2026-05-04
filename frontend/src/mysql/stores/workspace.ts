import { ref } from 'vue'
import { defineStore } from 'pinia'

export const useWorkspaceStore = defineStore('mysqlWorkspace', () => {
  const activeDatabase = ref('')
  const activeTable = ref('')

  function setActiveDatabase(databaseName: string) {
    activeDatabase.value = databaseName
  }

  function setActiveTable(databaseName: string, tableName: string) {
    activeDatabase.value = databaseName
    activeTable.value = tableName
  }

  function clearActiveTable() {
    activeTable.value = ''
  }

  function resetWorkspace() {
    activeDatabase.value = ''
    activeTable.value = ''
  }

  return {
    activeDatabase,
    activeTable,
    setActiveDatabase,
    setActiveTable,
    clearActiveTable,
    resetWorkspace
  }
})

