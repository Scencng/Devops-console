import { computed, ref } from 'vue'
import { defineStore } from 'pinia'

export type Locale = 'zh-CN' | 'en-US'

const LOCALE_STORAGE_KEY = 'mydeploy-locale'

function loadLocale(): Locale {
  if (typeof window === 'undefined') {
    return 'zh-CN'
  }

  const raw = window.localStorage.getItem(LOCALE_STORAGE_KEY)
  return raw === 'en-US' ? 'en-US' : 'zh-CN'
}

export const useLocaleStore = defineStore('locale', () => {
  const locale = ref<Locale>(loadLocale())

  const isChinese = computed(() => locale.value === 'zh-CN')

  function setLocale(nextLocale: Locale) {
    locale.value = nextLocale
    if (typeof window !== 'undefined') {
      window.localStorage.setItem(LOCALE_STORAGE_KEY, nextLocale)
      document.documentElement.lang = nextLocale === 'zh-CN' ? 'zh-CN' : 'en'
    }
  }

  function toggleLocale() {
    setLocale(locale.value === 'zh-CN' ? 'en-US' : 'zh-CN')
  }

  return {
    locale,
    isChinese,
    setLocale,
    toggleLocale
  }
})
