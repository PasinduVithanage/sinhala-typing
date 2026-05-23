import { useState, useEffect } from 'react'
import { TrayPanel } from './components/TrayPanel/TrayPanel'
import { SettingsPanel } from './components/Settings/SettingsPanel'
import { useEngine } from './hooks/useEngine'
import { useSettingsStore } from './store/settingsStore'
import * as AppAPI from '../wailsjs/go/ipc/App'
import { EventsOn } from '../wailsjs/runtime/runtime'
import './style.css'

type AppView = 'main' | 'settings'

function App() {
  useEngine()
  const [view, setView] = useState<AppView>('main')
  const { setAutoStart, setAutoFixClipboard, setUseZWJClusters, setToggleHotkey, setRunAtStartup } = useSettingsStore()

  useEffect(() => {
    AppAPI.GetConfig().then((cfg) => {
      setAutoStart(cfg.auto_start)
      setAutoFixClipboard(cfg.auto_fix_clipboard)
      setUseZWJClusters(cfg.use_zwj_clusters)
      setToggleHotkey(cfg.toggle_hotkey)
      setRunAtStartup(cfg.run_at_startup)
    }).catch(() => {})

    EventsOn('ui:show-settings', () => setView('settings'))
  }, [])

  if (view === 'settings') {
    return <SettingsPanel onBack={() => setView('main')} />
  }

  return <TrayPanel onSettings={() => setView('settings')} />
}

export default App
