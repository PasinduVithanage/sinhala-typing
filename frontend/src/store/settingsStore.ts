import { create } from 'zustand'

interface SettingsState {
  autoStart: boolean
  autoFixClipboard: boolean
  useZWJClusters: boolean
  toggleHotkey: string
  runAtStartup: boolean
  setAutoStart: (v: boolean) => void
  setAutoFixClipboard: (v: boolean) => void
  setUseZWJClusters: (v: boolean) => void
  setToggleHotkey: (v: string) => void
  setRunAtStartup: (v: boolean) => void
}

export const useSettingsStore = create<SettingsState>((set) => ({
  autoStart: false,
  autoFixClipboard: false,
  useZWJClusters: true,
  toggleHotkey: 'Ctrl+Alt+S',
  runAtStartup: false,
  setAutoStart: (autoStart) => set({ autoStart }),
  setAutoFixClipboard: (autoFixClipboard) => set({ autoFixClipboard }),
  setUseZWJClusters: (useZWJClusters) => set({ useZWJClusters }),
  setToggleHotkey: (toggleHotkey) => set({ toggleHotkey }),
  setRunAtStartup: (runAtStartup) => set({ runAtStartup }),
}))
