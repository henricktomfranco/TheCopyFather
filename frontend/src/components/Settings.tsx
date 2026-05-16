import { useState, useEffect } from 'react'
import { Config, TextTypeInfo } from '../App'
import * as runtime from '../../wailsjs/runtime'
import * as AppAPI from '../../wailsjs/go/main/App'
import { rewriter as rewriterModels } from '../../wailsjs/go/models'
import '../styles/main.css'
import '../styles/Settings.css'

interface OperationResult {
  success: boolean
  error?: string
}

interface SettingsProps {
  settings: Config
  onSave: (settings: Config) => void
  onCancel: () => void
  onLoadCustomPrompts: () => Promise<Record<string, Record<string, string>>>
  onSaveCustomPrompt: (style: string, textType: string, prompt: string) => Promise<OperationResult>
  onDeleteCustomPrompt: (style: string, textType: string) => Promise<OperationResult>
  onResetAllCustomPrompts: () => Promise<OperationResult>
  onGetDefaultPrompt: (style: string, textType: string) => Promise<string>
  onLoadRewriteStyles: () => Promise<string[]>
  onLoadAnalysisStyles: () => Promise<string[]>
  onLoadTextTypes: () => Promise<rewriterModels.TextTypeInfo[]>
}

interface StyleOption {
  value: string
  label: string
  icon: string
}

export default function Settings({
  settings,
  onSave,
  onCancel,
  onLoadCustomPrompts,
  onSaveCustomPrompt,
  onDeleteCustomPrompt,
  onResetAllCustomPrompts,
  onGetDefaultPrompt,
  onLoadRewriteStyles,
  onLoadAnalysisStyles,
  onLoadTextTypes
}: SettingsProps) {
  const [formData, setFormData] = useState<Config>(settings)
  const [availableModels, setAvailableModels] = useState<string[]>([])
  const [testingConnection, setTestingConnection] = useState(false)
  const [connectionStatus, setConnectionStatus] = useState<'idle' | 'success' | 'error'>('idle')
  const [detectedVersion, setDetectedVersion] = useState('')
  const [saving, setSaving] = useState(false)

  const [customPrompts, setCustomPrompts] = useState<Record<string, Record<string, string>>>({})
  const [rewriteStyles, setRewriteStyles] = useState<StyleOption[]>([])
  const [analysisStyles, setAnalysisStyles] = useState<StyleOption[]>([])
  const [textTypes, setTextTypes] = useState<rewriterModels.TextTypeInfo[]>([])
  const [selectedStyle, setSelectedStyle] = useState('')
  const [selectedTextType, setSelectedTextType] = useState('')
  const [customPromptText, setCustomPromptText] = useState('')
  const [hasCustomPrompt, setHasCustomPrompt] = useState(false)
  const [loadingPrompt, setLoadingPrompt] = useState(false)

  const [feedback, setFeedback] = useState<{ type: 'success' | 'error'; message: string } | null>(null)

  const showFeedback = (type: 'success' | 'error', message: string) => {
    setFeedback({ type, message })
    setTimeout(() => setFeedback(null), 3000)
  }

  useEffect(() => {
    loadAvailableModels()
    loadSettingsData()
    runtime.WindowSetSize(500, 650)
    runtime.WindowCenter()
    runtime.WindowShow()
  }, [])

  const loadAvailableModels = async () => {
    try {
      const models = await AppAPI.GetAvailableModels()
      setAvailableModels(models || [])
    } catch (error) {
      console.error('Failed to load models:', error)
      setAvailableModels([])
    }
  }

  const loadSettingsData = async () => {
    try {
      const [prompts, rwStyles, anStyles, types] = await Promise.all([
        onLoadCustomPrompts(),
        onLoadRewriteStyles(),
        onLoadAnalysisStyles(),
        onLoadTextTypes()
      ])

      setCustomPrompts(prompts || {})

      const styleInfoPromises = rwStyles.map(style =>
        AppAPI.GetStyleInfo(style).catch(() => null)
      )
      const rwStyleInfos = await Promise.all(styleInfoPromises)

      setRewriteStyles(rwStyles.map((style, i) => {
        const info = rwStyleInfos[i]
        return {
          value: style,
          label: info && typeof info !== 'boolean' ? info.label || style : style,
          icon: info && typeof info !== 'boolean' ? info.icon || '📝' : '📝'
        }
      }))

      const anStyleInfoPromises = anStyles.map(style =>
        AppAPI.GetStyleInfo(style).catch(() => null)
      )
      const anStyleInfos = await Promise.all(anStyleInfoPromises)

      setAnalysisStyles(anStyles.map((style, i) => {
        const info = anStyleInfos[i]
        return {
          value: style,
          label: info && typeof info !== 'boolean' ? info.label || style : style,
          icon: info && typeof info !== 'boolean' ? info.icon || '📋' : '📋'
        }
      }))

      setTextTypes(types || [])

      if (rwStyles.length > 0) {
        setSelectedStyle(rwStyles[0])
      }
      if (types.length > 0) {
        setSelectedTextType(types[0].Type)
      }
    } catch (error) {
      console.error('Failed to load settings data:', error)
    }
  }

  const loadPromptForSelection = async () => {
    if (!selectedStyle || !selectedTextType) return

    setLoadingPrompt(true)
    try {
      const key = `${selectedStyle}.${selectedTextType}`
      const custom = customPrompts[selectedStyle]?.[selectedTextType]
      const isCustom = custom !== undefined && custom !== ''

      setHasCustomPrompt(isCustom)
      setCustomPromptText(isCustom ? custom : '')
    } finally {
      setLoadingPrompt(false)
    }
  }

  useEffect(() => {
    loadPromptForSelection()
  }, [selectedStyle, selectedTextType, customPrompts])

  const handleSaveCustomPrompt = async () => {
    if (!selectedStyle || !selectedTextType) return

    const result = await onSaveCustomPrompt(selectedStyle, selectedTextType, customPromptText)

    if (result.success) {
      const updated = { ...customPrompts }
      if (!updated[selectedStyle]) {
        updated[selectedStyle] = {}
      }
      updated[selectedStyle][selectedTextType] = customPromptText
      setCustomPrompts(updated)
      setHasCustomPrompt(true)
      showFeedback('success', 'Custom prompt saved')
    } else {
      showFeedback('error', result.error || 'Failed to save')
    }
  }

  const handleDeleteCustomPrompt = async () => {
    if (!selectedStyle || !selectedTextType) return

    const result = await onDeleteCustomPrompt(selectedStyle, selectedTextType)

    if (result.success) {
      const updated = { ...customPrompts }
      if (updated[selectedStyle]) {
        delete updated[selectedStyle][selectedTextType]
        if (Object.keys(updated[selectedStyle]).length === 0) {
          delete updated[selectedStyle]
        }
      }
      setCustomPrompts(updated)
      showFeedback('success', 'Reset to default')
    } else {
      showFeedback('error', result.error || 'Failed to reset')
    }

    const defaultPrompt = await onGetDefaultPrompt(selectedStyle, selectedTextType)
    setCustomPromptText(defaultPrompt)
    setHasCustomPrompt(false)
  }

  const handleLoadDefault = async () => {
    if (!selectedStyle || !selectedTextType) return
    const defaultPrompt = await onGetDefaultPrompt(selectedStyle, selectedTextType)
    setCustomPromptText(defaultPrompt)
  }

  const testConnection = async () => {
    setTestingConnection(true)
    setConnectionStatus('idle')
    setDetectedVersion('')

    try {
      const version = await AppAPI.TestConnection(
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
    <div className="settings-v2">
      <header className="settings-header">
        <h2><span className="header-icon">⚙️</span> Settings</h2>
      </header>

      <div className="settings-content">
        {/* Connection Section */}
        <section className="settings-section">
          <h3>Ollama Connection</h3>
          
          <div className="form-group">
            <label htmlFor="server_url">Endpoint URL</label>
            <input
              type="text"
              id="server_url"
              className="input"
              value={formData.server_url}
              onChange={(e) => handleChange('server_url', e.target.value)}
              placeholder="http://localhost:11434"
            />
            <small>Default local endpoint runs on port 11434</small>
          </div>

          <div className="form-group">
            <label htmlFor="model">Model</label>
            <div className="model-input-group">
              <input
                type="text"
                id="model"
                className="input"
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

          <div className="connection-test">
            <button
              type="button"
              className="test-btn"
              onClick={testConnection}
              disabled={testingConnection}
            >
              {testingConnection ? 'Testing...' : 'Test Connection'}
            </button>

            {connectionStatus !== 'idle' && (
              <div className={`status-message ${connectionStatus}`}>
                {connectionStatus === 'success'
                  ? `✓ Connected (v${detectedVersion || 'Unknown'})`
                  : '✗ Connection failed'}
              </div>
            )}
          </div>
        </section>

        {/* Preferences Section */}
        <section className="settings-section">
          <h3>Preferences</h3>

          <div className="toggle-group">
            <div className="toggle-label">
              <span>Start with Windows</span>
              <small>Launch automatically on system startup</small>
            </div>
            <div
              className={`toggle-switch ${formData.auto_start ? 'active' : ''}`}
              onClick={() => handleChange('auto_start', !formData.auto_start)}
            />
          </div>

          <div className="form-group" style={{ marginTop: 'var(--spacing-md)' }}>
            <div className="toggle-group">
              <div className="toggle-label">
                <span>Clipboard Monitoring</span>
                <small>Detect text when copied</small>
              </div>
              <div
                className={`toggle-switch ${formData.monitor_clipboard ? 'active' : ''}`}
                onClick={() => handleChange('monitor_clipboard', !formData.monitor_clipboard)}
              />
            </div>
          </div>

          <div className="form-group" style={{ marginTop: 'var(--spacing-md)' }}>
            <div className="toggle-group">
              <div className="toggle-label">
                <span>Compact Mode</span>
                <small>Mini floating widget</small>
              </div>
              <div
                className={`toggle-switch ${formData.mini_mode ? 'active' : ''}`}
                onClick={() => handleChange('mini_mode', !formData.mini_mode)}
              />
            </div>
          </div>

          <div className="form-group" style={{ marginTop: 'var(--spacing-md)' }}>
            <div className="toggle-group">
              <div className="toggle-label">
                <span>Auto-Minimize on Copy</span>
                <small>Minimize window after copying rewritten text</small>
              </div>
              <div
                className={`toggle-switch ${formData.auto_minimize_on_copy ? 'active' : ''}`}
                onClick={() => handleChange('auto_minimize_on_copy', !formData.auto_minimize_on_copy)}
              />
            </div>
          </div>

          <div className="form-group" style={{ marginTop: 'var(--spacing-lg)' }}>
            <label htmlFor="hotkey">Hotkey</label>
            <input
              type="text"
              id="hotkey"
              className="input"
              value={formData.hotkey}
              onChange={(e) => handleChange('hotkey', e.target.value)}
              placeholder="ctrl+shift+r"
            />
          </div>

          <div className="form-group">
            <label htmlFor="default_style">Default Style</label>
            <select
              id="default_style"
              className="select-input"
              value={formData.default_style}
              onChange={(e) => handleChange('default_style', e.target.value)}
            >
              {rewriteStyles.map(style => (
                <option key={style.value} value={style.value}>
                  {style.icon} {style.label}
                </option>
              ))}
              {analysisStyles.map(style => (
                <option key={style.value} value={style.value}>
                  {style.icon} {style.label}
                </option>
              ))}
            </select>
          </div>

          <div className="form-group">
            <label htmlFor="auto_paste_mode">Auto-Paste Behavior</label>
            <select
              id="auto_paste_mode"
              className="select-input"
              value={formData.auto_paste_mode || 'ask'}
              onChange={(e) => handleChange('auto_paste_mode', e.target.value)}
            >
              <option value="ask">❓ Ask every time</option>
              <option value="always">✓ Always paste automatically</option>
              <option value="never">✗ Never paste (copy only)</option>
            </select>
          </div>
        </section>

        {/* Custom Prompts Section */}
        <section className="settings-section">
          <h3>Custom Prompts</h3>
          
          <div className="prompts-grid">
            <div className="form-group">
              <label htmlFor="prompt_style">Style</label>
              <select
                id="prompt_style"
                className="select-input"
                value={selectedStyle}
                onChange={(e) => setSelectedStyle(e.target.value)}
              >
                <optgroup label="Rewrite">
                  {rewriteStyles.map(style => (
                    <option key={style.value} value={style.value}>
                      {style.icon} {style.label}
                    </option>
                  ))}
                </optgroup>
                <optgroup label="Analyze">
                  {analysisStyles.map(style => (
                    <option key={style.value} value={style.value}>
                      {style.icon} {style.label}
                    </option>
                  ))}
                </optgroup>
              </select>
            </div>

            <div className="form-group">
              <label htmlFor="prompt_text_type">Text Type</label>
              <select
                id="prompt_text_type"
                className="select-input"
                value={selectedTextType}
                onChange={(e) => setSelectedTextType(e.target.value)}
              >
                {textTypes.map(type => (
                  <option key={type.Type} value={type.Type}>
                    {type.Icon} {type.Label}
                  </option>
                ))}
              </select>
            </div>
          </div>

          <div className="form-group">
            <label htmlFor="custom_prompt">
              Prompt {hasCustomPrompt && <span className="info-badge">Custom</span>}
            </label>
            <textarea
              id="custom_prompt"
              className="prompt-textarea"
              value={customPromptText}
              onChange={(e) => setCustomPromptText(e.target.value)}
              placeholder="Leave empty to use default..."
              rows={8}
            />
          </div>

          <div className="prompt-actions">
            <button
              type="button"
              className="btn-load-default"
              onClick={handleLoadDefault}
              disabled={loadingPrompt}
            >
              Load Default
            </button>
            <button
              type="button"
              className="btn-reset"
              onClick={handleDeleteCustomPrompt}
              disabled={!hasCustomPrompt || loadingPrompt}
            >
              Reset
            </button>
            <button
              type="button"
              className="btn-save-prompt"
              onClick={handleSaveCustomPrompt}
              disabled={loadingPrompt}
            >
              Save
            </button>
          </div>

          <div className="prompt-actions" style={{ marginTop: 'var(--spacing-md)', borderTop: '1px solid var(--border-subtle)', paddingTop: 'var(--spacing-md)' }}>
            <button
              type="button"
              className="btn-reset-all"
              onClick={async () => {
                if (confirm('Reset ALL custom prompts to defaults? This cannot be undone.')) {
                  const result = await onResetAllCustomPrompts()
                  if (result.success) {
                    const prompts = await onLoadCustomPrompts()
                    setCustomPrompts(prompts || {})
                    showFeedback('success', 'All prompts reset to defaults')
                  } else {
                    showFeedback('error', result.error || 'Failed to reset')
                  }
                }
              }}
            >
              Reset All Prompts to Defaults
            </button>
          </div>
        </section>
      </div>

      {/* Footer */}
      <footer className="settings-footer-v2">
        <button
          type="button"
          className="btn btn-secondary"
          onClick={onCancel}
        >
          Cancel
        </button>
        <button
          type="submit"
          className="btn btn-primary"
          onClick={handleSubmit as any}
          disabled={saving}
        >
          {saving ? 'Saving...' : 'Save Changes'}
        </button>
      </footer>

      {/* Feedback Toast */}
      {feedback && (
        <div className={`feedback-toast ${feedback.type}`}>
          {feedback.message}
        </div>
      )}
    </div>
  )
}
