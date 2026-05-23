import { useSettingsStore } from '../../store/settingsStore'
import { useEngineStore } from '../../store/engineStore'
import * as App from '../../../wailsjs/go/ipc/App'

interface Props {
  onBack: () => void
}

export function SettingsPanel({ onBack }: Props) {
  const {
    autoStart, autoFixClipboard, useZWJClusters, toggleHotkey, runAtStartup,
    setAutoStart, setAutoFixClipboard, setUseZWJClusters, setToggleHotkey, setRunAtStartup,
  } = useSettingsStore()
  const { mapping } = useEngineStore()

  const handleSave = () => {
    App.SaveConfig({
      mapping_profile: mapping,
      auto_start: autoStart,
      auto_fix_clipboard: autoFixClipboard,
      use_zwj_clusters: useZWJClusters,
      toggle_hotkey: toggleHotkey,
      run_at_startup: runAtStartup,
    })
    onBack()
  }

  return (
    <div className="flex flex-col h-screen bg-[#0F0F12] text-[#F0F0F5] p-4 gap-4">
      <div className="flex items-center gap-3">
        <button onClick={onBack} className="text-[#8888AA] hover:text-[#F0F0F5] text-sm">← Back</button>
        <span className="text-sm font-semibold">Settings</span>
      </div>

      <div className="flex flex-col gap-3">
        <ToggleSetting
          label="Start engine on launch"
          value={autoStart}
          onChange={setAutoStart}
        />
        <ToggleSetting
          label="Auto-fix Sinhala in clipboard"
          value={autoFixClipboard}
          onChange={setAutoFixClipboard}
        />
        <ToggleSetting
          label="Use ZWJ for cluster forms (rakar/yansaya)"
          value={useZWJClusters}
          onChange={setUseZWJClusters}
        />
        <ToggleSetting
          label="Start with Windows"
          value={runAtStartup}
          onChange={setRunAtStartup}
        />

        <div>
          <label className="text-xs text-[#8888AA] block mb-1">Toggle hotkey</label>
          <input
            value={toggleHotkey}
            onChange={(e) => setToggleHotkey(e.target.value)}
            className="w-full bg-[#1A1A20] border border-[#2A2A38] rounded-lg px-3 py-1.5
                       text-xs text-[#F0F0F5] focus:outline-none focus:border-[#4F8EF7]"
          />
        </div>
      </div>

      <div className="mt-auto">
        <button
          onClick={handleSave}
          className="w-full py-2 text-xs rounded-lg bg-[#4F8EF7] hover:bg-[#4F8EF7]/80
                     text-white font-semibold transition-colors"
        >
          Save Settings
        </button>
      </div>
    </div>
  )
}

function ToggleSetting({
  label, value, onChange,
}: {
  label: string
  value: boolean
  onChange: (v: boolean) => void
}) {
  return (
    <div className="flex items-center justify-between bg-[#1A1A20] border border-[#2A2A38] rounded-lg px-4 py-3">
      <span className="text-xs text-[#F0F0F5]">{label}</span>
      <button
        onClick={() => onChange(!value)}
        className={`relative w-10 h-5 rounded-full transition-colors duration-200 ${
          value ? 'bg-[#4F8EF7]' : 'bg-[#2A2A38]'
        }`}
      >
        <span
          className={`absolute top-0.5 w-4 h-4 bg-white rounded-full shadow transition-transform duration-200 ${
            value ? 'translate-x-5' : 'translate-x-0.5'
          }`}
        />
      </button>
    </div>
  )
}
