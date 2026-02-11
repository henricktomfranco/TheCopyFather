import { useState, useEffect } from 'react'
import { Config } from '../App'
import * as runtime from '../../wailsjs/runtime'
import '../styles/Settings.css'
import logoImage from '../assets/logo.jpg'

interface SettingsProps {
  settings: Config
  onSave: (settings: Config) => void
  onCancel: () => void
}

function Settings({ settings, onSave, onCancel }: SettingsProps) {
  const [formData, setFormData] = useState<Config>(settings)
  const [availableModels, setAvailableModels] = useState<string[]>([])
  const [testingConnection, setTestingConnection] = useState(false)
  const [connectionStatus, setConnectionStatus] = useState<'idle' | 'success' | 'error'>('idle')
  const [detectedVersion, setDetectedVersion] = useState('')
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    loadAvailableModels()
    // Show window and set size for settings
    runtime.WindowSetSize(580, 680)
    runtime.WindowCenter()
    runtime.WindowShow()
  }, [])

  const loadAvailableModels = async () => {
    try {
      // @ts-ignore
      const models = await window.go.main.App.GetAvailableModels()
      setAvailableModels(models || [])
    } catch (error) {
      console.error('Failed to load models:', error)
      setAvailableModels([])
    }
  }

  const testConnection = async () => {
    setTestingConnection(true)
    setConnectionStatus('idle')
    setDetectedVersion('')

    try {
      // @ts-ignore
      const version = await window.go.main.App.TestConnection(
        formData.server_url,
        formData.model,
        formData.api_key || ""
      )
      setDetectedVersion(version)
      setConnectionStatus('success')
    } catch (error) {
      setConnectionStatus('error')
    }

    setTestingConnection(false)
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setSaving(true)
    await onSave(formData)
    setSaving(false)
  }

  const handleChange = (field: keyof Config, value: any) => {
    setFormData(prev => ({ ...prev, [field]: value }))
  }

  return (
    <div className="settings">
      <div className="settings-header">
        <div className="logo-container">
          <img src={logoImage} alt="App Logo" className="app-logo" />
        </div>
        <h2><span>⚙️</span> Settings</h2>
      </div>

      <form onSubmit={handleSubmit} className="settings-form">
        <div className="settings-section">
          <h3>Ollama Node</h3>

          <div className="form-group">
            <label htmlFor="server_url">Endpoint URL</label>
            <input
              type="text"
              id="server_url"
              value={formData.server_url}
              onChange={(e) => handleChange('server_url', e.target.value)}
              placeholder="http://localhost:11434"
            />
            <small>Default local node usually runs on port 11434</small>
          </div>

          <div className="form-group">
            <label htmlFor="model">Model Identifier</label>
            <div className="model-input-group">
              <input
                type="text"
                id="model"
                value={formData.model}
                onChange={(e) => handleChange('model', e.target.value)}
                placeholder="gemma3:1b"
                list="available-models"
              />
              <datalist id="available-models">
                {availableModels.map(model => (
                  <option key={model} value={model} />
                ))}
              </datalist>
            </div>
          </div>

          <button
            type="button"
            className="test-btn"
            onClick={testConnection}
            disabled={testingConnection}
          >
            {testingConnection ? 'Verifying node...' : 'Verify Connection'}
          </button>

          {connectionStatus !== 'idle' && (
            <div className={`status-message ${connectionStatus}`}>
              {connectionStatus === 'success'
                ? `✓ Node connected (Version: ${detectedVersion || 'Unknown'})`
                : '✗ Failed to connect to node'}
            </div>
          )}
        </div>

        <div className="settings-section">
          <h3>Preferences</h3>

          <div className="form-group">
            <label className="checkbox-label">
              <span>Start with Windows</span>
              <input
                type="checkbox"
                checked={formData.auto_start}
                onChange={(e) => handleChange('auto_start', e.target.checked)}
              />
              <div className="switch"></div>
            </label>
          </div>

          <div className="form-group">
            <label className="checkbox-label">
              <span>Clipboard Monitoring</span>
              <input
                type="checkbox"
                checked={formData.monitor_clipboard}
                onChange={(e) => handleChange('monitor_clipboard', e.target.checked)}
              />
              <div className="switch"></div>
            </label>
            <small>Automatically detect text when copied to clipboard</small>
          </div>

          <div className="form-group">
            <label htmlFor="hotkey">Global Activation Key</label>
            <input
              type="text"
              id="hotkey"
              value={formData.hotkey}
              onChange={(e) => handleChange('hotkey', e.target.value)}
              placeholder="ctrl+shift+r"
            />
            <small>Current shortcut for triggering the rewrite flow</small>
          </div>

          <div className="form-group">
            <label htmlFor="default_style">Initial Style</label>
            <select
              id="default_style"
              value={formData.default_style}
              onChange={(e) => handleChange('default_style', e.target.value)}
            >
              <option value="grammar">🛡️ Grammar & Spelling</option>
              <option value="paraphrase">🔄 Paraphrase (Standard)</option>
              <option value="standard">📝 Standard</option>
              <option value="formal">📢 Formal</option>
              <option value="casual">💬 Casual</option>
              <option value="creative">✨ Creative</option>
              <option value="short">📏 Short</option>
              <option value="expand">📖 Expand</option>
            </select>
          </div>

          <div className="form-group">
            <label htmlFor="auto_paste_mode">Auto-Paste Behavior</label>
            <select
              id="auto_paste_mode"
              value={formData.auto_paste_mode || 'ask'}
              onChange={(e) => handleChange('auto_paste_mode', e.target.value)}
            >
              <option value="ask">❓ Ask every time</option>
              <option value="always">✓ Always paste automatically</option>
              <option value="never">✗ Never paste (copy only)</option>
            </select>
            <small>What happens when you click "Replace Selection"</small>
          </div>

          <div className="form-group">
            <label htmlFor="popup_position_mode">Popup Position</label>
            <select
              id="popup_position_mode"
              value={formData.popup_position_mode || 'cursor'}
              onChange={(e) => handleChange('popup_position_mode', e.target.value)}
            >
              <option value="cursor">📍 Near cursor</option>
              <option value="center">🖥️ Screen center</option>
            </select>
            <small>Where the popup window appears</small>
          </div>
        </div>

        <div className="settings-footer">
          <button
            type="button"
            className="cancel-btn"
            onClick={onCancel}
          >
            Go Back
          </button>
          <button
            type="submit"
            className="save-btn"
            disabled={saving}
          >
            {saving ? 'Applying...' : 'Apply Changes'}
          </button>
        </div>
      </form>
    </div>
  )
}

export default Settings
