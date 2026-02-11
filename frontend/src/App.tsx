import { useState, useEffect } from 'react'
import Popup from './components/Popup'
import Settings from './components/Settings'
import Welcome from './components/Welcome'
import * as runtime from '../wailsjs/runtime'
import './styles/main.css'

export type View = 'popup' | 'settings' | 'welcome'

export interface RewriteOption {
  style: string
  text: string
  error?: string
}

export interface Config {
  server_url: string
  model: string
  api_key?: string
  default_style: string
  auto_start: boolean
  hotkey: string
  monitor_clipboard: boolean
  first_run: boolean
  auto_paste_mode?: string
  popup_position_mode?: string
}

function App() {
  const [currentView, setCurrentView] = useState<View>('settings')
  const [selectedText, setSelectedText] = useState('')
  const [selectionTrigger, setSelectionTrigger] = useState(0)
  const [settings, setSettings] = useState<Config | null>(null)

  useEffect(() => {
    // Listen for text selection event from backend
    runtime.EventsOn('text:selected', (text: string) => {
      console.log('Frontend received text:', text)
      setSelectedText(text)
      setSelectionTrigger(prev => prev + 1)
      setCurrentView(prev => prev === 'welcome' ? 'welcome' : 'popup')
    })

    // Listen for show settings event from system tray
    runtime.EventsOn('window:showsettings', () => {
      setCurrentView('settings')
      loadSettings()
    })

    // Load initial settings
    loadSettings()

    return () => {
      runtime.EventsOff('text:selected')
      runtime.EventsOff('window:showsettings')
    }
  }, [])

  const loadSettings = async () => {
    try {
      // @ts-ignore
      const config = await window.go.main.App.GetSettings()
      setSettings(config)
      if (config.first_run) {
        setCurrentView('welcome')
      }
    } catch (error) {
      console.error('Failed to load settings:', error)
    }
  }

  const handleSelectRewrite = async (text: string) => {
    try {
      // @ts-ignore
      await window.go.main.App.ApplyRewrite(text)
      runtime.WindowHide()
    } catch (error) {
      console.error('Failed to apply rewrite:', error)
    }
  }

  const handleSaveSettings = async (newSettings: Config) => {
    try {
      // @ts-ignore
      await window.go.main.App.SaveSettings(newSettings)
      setSettings(newSettings)
      // Just keep current view or return to popup if text was selected
      if (selectedText) {
        setCurrentView('popup')
      }
    } catch (error) {
      console.error('Failed to save settings:', error)
    }
  }

  const handleAcceptFirstRun = async () => {
    if (!settings) return
    const newSettings = { ...settings, first_run: false }
    await handleSaveSettings(newSettings)
    setCurrentView('settings') // Go to settings first to confirm URL
  }

  const handleClose = () => {
    runtime.WindowHide()
  }

  const handleOpenSettings = () => {
    setCurrentView('settings')
  }

  return (
    <div className="app">
      {currentView === 'welcome' && (
        <Welcome onAccept={handleAcceptFirstRun} />
      )}

      {currentView === 'popup' && settings && (
        <Popup
          key={`popup-${selectionTrigger}`}
          originalText={selectedText}
          onSelect={handleSelectRewrite}
          onClose={handleClose}
          onSettings={handleOpenSettings}
          defaultStyle={settings.default_style}
        />
      )}

      {currentView === 'settings' && settings && (
        <Settings
          settings={settings}
          onSave={handleSaveSettings}
          onCancel={() => {
            if (selectedText) {
              setCurrentView('popup')
            } else {
              handleClose()
            }
          }}
        />
      )}
    </div>
  )
}

export default App
