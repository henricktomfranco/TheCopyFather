import React, { useState, useEffect } from 'react';
import * as SettingsAPI from '../../wailsjs/go/main/SettingsService';
import { Config, TextTypeInfo, StyleInfo } from '../App';
import './SettingsNew.css';

interface SettingsProps {
  settings: Config;
  onSave: (settings: Config) => Promise<void>;
  onCancel: () => void;
  onLoadCustomPrompts: () => Promise<Record<string, Record<string, string>>>;
  onSaveCustomPrompt: (style: string, textType: string, prompt: string) => Promise<{ success: boolean; error?: string }>;
  onDeleteCustomPrompt: (style: string, textType: string) => Promise<{ success: boolean; error?: string }>;
  onResetAllCustomPrompts: () => Promise<{ success: boolean; error?: string }>;
  onGetDefaultPrompt: (style: string, textType: string) => Promise<string>;
  onLoadRewriteStyles: () => Promise<string[]>;
  onLoadAnalysisStyles: () => Promise<string[]>;
  onLoadTextTypes: () => Promise<TextTypeInfo[]>;
}

// Tab types
type TabType = 'general' | 'ai' | 'updates' | 'advanced';

const SettingsNew: React.FC<SettingsProps> = ({
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
  onLoadTextTypes,
}) => {
  // State for form data
  const [formData, setFormData] = useState<Config>(settings);
  const [activeTab, setActiveTab] = useState<TabType>('general');
  const [availableModels, setAvailableModels] = useState<string[]>([]);
  const [rewriteStyles, setRewriteStyles] = useState<string[]>([]);
  const [analysisStyles, setAnalysisStyles] = useState<string[]>([]);
  const [textTypes, setTextTypes] = useState<TextTypeInfo[]>([]);
  const [customPrompts, setCustomPrompts] = useState<Record<string, Record<string, string>>>({});
  const [testingConnection, setTestingConnection] = useState(false);
  const [connectionStatus, setConnectionStatus] = useState<'idle' | 'success' | 'error'>('idle');
  const [detectedVersion, setDetectedVersion] = useState('');
  const [selectedStyle, setSelectedStyle] = useState<string>('');
  const [selectedTextType, setSelectedTextType] = useState<string>('');
  const [customPromptText, setCustomPromptText] = useState('');
  const [defaultPromptText, setDefaultPromptText] = useState('');
  const [isLoading, setIsLoading] = useState(true);

  // Load data on mount
  useEffect(() => {
    const loadData = async () => {
      try {
        const [models, styles, analysis, types, prompts] = await Promise.all([
          SettingsAPI.GetAvailableModels().catch(() => []),
          onLoadRewriteStyles(),
          onLoadAnalysisStyles(),
          onLoadTextTypes(),
          onLoadCustomPrompts(),
        ]);
        setAvailableModels(models || []);
        setRewriteStyles(styles || []);
        setAnalysisStyles(analysis || []);
        setTextTypes(types || []);
        setCustomPrompts(prompts || {});
      } catch (error) {
        console.error('Failed to load settings data:', error);
      } finally {
        setIsLoading(false);
      }
    };
    loadData();
  }, []);

  // Update form data when settings change
  useEffect(() => {
    setFormData(settings);
  }, [settings]);

  // Handle form field changes
  const handleChange = <K extends keyof Config>(key: K, value: Config[K]) => {
    setFormData({ ...formData, [key]: value });
  };

  // Handle saving settings
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    await onSave(formData);
  };

  // Test AI connection
  const testConnection = async () => {
    setTestingConnection(true);
    setConnectionStatus('idle');
    setDetectedVersion('');

    try {
const version = await SettingsAPI.TestConnection(
formData.useOpenAICompatible ? (formData.openAIBaseURL ?? '') : (formData.server_url ?? ''),
formData.useOpenAICompatible ? (formData.openAIModel ?? '') : (formData.model ?? ''),
formData.useOpenAICompatible ? (formData.openAIAPIKey ?? '') : (formData.api_key ?? ''),
formData.useOpenAICompatible ?? false
);
      setDetectedVersion(version);
      setConnectionStatus('success');
    } catch (error) {
      setConnectionStatus('error');
    }

    setTestingConnection(false);
  };

  // Load custom prompt for editing
  const loadCustomPrompt = async (style: string, textType: string) => {
    setSelectedStyle(style);
    setSelectedTextType(textType);
    try {
      const prompt = await onGetDefaultPrompt(style, textType);
      const custom = customPrompts[style]?.[textType] || '';
      setDefaultPromptText(prompt);
      setCustomPromptText(custom || prompt);
    } catch (error) {
      console.error('Failed to load prompt:', error);
    }
  };

  // Save custom prompt
  const saveCustomPrompt = async () => {
    if (!selectedStyle || !selectedTextType) return;
    const result = await onSaveCustomPrompt(selectedStyle, selectedTextType, customPromptText);
    if (result.success) {
      setCustomPrompts({
        ...customPrompts,
        [selectedStyle]: {
          ...customPrompts[selectedStyle],
          [selectedTextType]: customPromptText,
        },
      });
    }
  };

  // Delete custom prompt
  const deleteCustomPrompt = async (style: string, textType: string) => {
    const result = await onDeleteCustomPrompt(style, textType);
    if (result.success) {
      const newPrompts = { ...customPrompts };
      delete newPrompts[style]?.[textType];
      if (Object.keys(newPrompts[style] || {}).length === 0) {
        delete newPrompts[style];
      }
      setCustomPrompts(newPrompts);
      if (selectedStyle === style && selectedTextType === textType) {
        setSelectedStyle('');
        setSelectedTextType('');
        setCustomPromptText('');
        setDefaultPromptText('');
      }
    }
  };

  // Reset all custom prompts
  const resetAllCustomPrompts = async () => {
    const result = await onResetAllCustomPrompts();
    if (result.success) {
      setCustomPrompts({});
      setSelectedStyle('');
      setSelectedTextType('');
      setCustomPromptText('');
      setDefaultPromptText('');
    }
  };

  // Render connection status
  const renderConnectionStatus = () => {
    if (connectionStatus === 'idle') return null;
    return (
      <div className={`status-badge ${connectionStatus}`}>
        {connectionStatus === 'success'
          ? `✓ Connected (v${detectedVersion || 'Unknown'})`
          : '✗ Connection failed'}
      </div>
    );
  };

  // Render tab navigation
  const renderTabNavigation = () => (
    <div className="settings-tabs">
      <button
        type="button"
        className={`tab ${activeTab === 'general' ? 'active' : ''}`}
        onClick={() => setActiveTab('general')}
      >
        General
      </button>
      <button
        type="button"
        className={`tab ${activeTab === 'ai' ? 'active' : ''}`}
        onClick={() => setActiveTab('ai')}
      >
        AI Provider
      </button>
      <button
        type="button"
        className={`tab ${activeTab === 'updates' ? 'active' : ''}`}
        onClick={() => setActiveTab('updates')}
      >
        Updates
      </button>
      <button
        type="button"
        className={`tab ${activeTab === 'advanced' ? 'active' : ''}`}
        onClick={() => setActiveTab('advanced')}
      >
        Advanced
      </button>
    </div>
  );

  // Render General tab
  const renderGeneralTab = () => (
    <div className="tab-content">
      <h3>General Settings</h3>
      
      <div className="settings-grid">
        {/* Hotkey */}
        <div className="form-group">
          <label htmlFor="hotkey">Global Hotkey</label>
          <input
            type="text"
            id="hotkey"
            className="input"
            value={formData.hotkey || 'Ctrl+Shift+R'}
            onChange={(e) => handleChange('hotkey', e.target.value)}
            placeholder="Ctrl+Shift+R"
          />
          <small>Press this hotkey to show the rewrite popup</small>
        </div>

        {/* Default Style */}
        <div className="form-group">
          <label htmlFor="default_style">Default Rewrite Style</label>
          <select
            id="default_style"
            className="select-input"
            value={formData.default_style}
            onChange={(e) => handleChange('default_style', e.target.value)}
          >
            {rewriteStyles.map((style) => (
              <option key={style} value={style}>
                {styleInfoData[style]?.Label || style}
              </option>
            ))}
          </select>
          <small>Default style for rewriting text</small>
        </div>

        {/* Auto Start */}
        <div className="form-group toggle-group">
          <div className="toggle-label">
            <span>Start on Boot</span>
            <small>Automatically start TheCopyFather when your computer boots</small>
          </div>
          <div
            className={`toggle-switch ${formData.auto_start ? 'active' : ''}`}
            onClick={() => handleChange('auto_start', !formData.auto_start)}
          />
        </div>

        {/* Mini Mode */}
        <div className="form-group toggle-group">
          <div className="toggle-label">
            <span>Mini Mode</span>
            <small>Show a compact popup instead of the full window</small>
          </div>
          <div
            className={`toggle-switch ${formData.mini_mode ? 'active' : ''}`}
            onClick={() => handleChange('mini_mode', !formData.mini_mode)}
          />
        </div>

        {/* Monitor Clipboard */}
        <div className="form-group toggle-group">
          <div className="toggle-label">
            <span>Monitor Clipboard</span>
            <small>Automatically detect text in your clipboard</small>
          </div>
          <div
            className={`toggle-switch ${formData.monitor_clipboard ? 'active' : ''}`}
            onClick={() => handleChange('monitor_clipboard', !formData.monitor_clipboard)}
          />
        </div>

        {/* Auto Paste Mode */}
        <div className="form-group">
          <label htmlFor="auto_paste_mode">Auto-Paste Behavior</label>
          <select
            id="auto_paste_mode"
            className="select-input"
            value={formData.auto_paste_mode || 'ask'}
            onChange={(e) => handleChange('auto_paste_mode', e.target.value)}
          >
            <option value="ask">Ask before pasting</option>
            <option value="always">Always paste</option>
            <option value="never">Never paste</option>
          </select>
          <small>What to do when you select a rewrite</small>
        </div>

        {/* Popup Position Mode */}
        <div className="form-group">
          <label htmlFor="popup_position_mode">Popup Position</label>
          <select
            id="popup_position_mode"
            className="select-input"
            value={formData.popup_position_mode || 'cursor'}
            onChange={(e) => handleChange('popup_position_mode', e.target.value)}
          >
            <option value="cursor">At cursor position</option>
            <option value="center">Center of screen</option>
            <option value="top">Top of screen</option>
          </select>
          <small>Where the rewrite popup appears</small>
        </div>

        {/* Auto Minimize on Copy */}
        <div className="form-group toggle-group">
          <div className="toggle-label">
            <span>Auto-Minimize on Copy</span>
            <small>Automatically hide the window after copying a rewrite</small>
          </div>
          <div
            className={`toggle-switch ${formData.auto_minimize_on_copy ? 'active' : ''}`}
            onClick={() => handleChange('auto_minimize_on_copy', !formData.auto_minimize_on_copy)}
          />
        </div>
      </div>
    </div>
  );

  // Render AI Provider tab
  const renderAITab = () => (
    <div className="tab-content">
      <h3>AI Provider Configuration</h3>
      
      <div className="settings-section">
        <div className="toggle-group">
          <div className="toggle-label">
            <span>Use OpenAI-Compatible API</span>
            <small>Enable to use NVIDIA NIM, LM Studio, or other OpenAI-compatible endpoints</small>
          </div>
          <div
            className={`toggle-switch ${formData.useOpenAICompatible ? 'active' : ''}`}
            onClick={() => handleChange('useOpenAICompatible', !formData.useOpenAICompatible)}
          />
        </div>
      </div>

      <div className="settings-section">
        <div className="toggle-group">
          <div className="toggle-label">
            <span>Disable Streaming</span>
            <small>Some providers/models don't support streaming. Enable this to use non-streaming mode.</small>
          </div>
          <div
            className={`toggle-switch ${formData.disableStreaming ? 'active' : ''}`}
            onClick={() => handleChange('disableStreaming', !formData.disableStreaming)}
          />
        </div>
      </div>

      {!formData.useOpenAICompatible ? (
        <div className="settings-grid">
          <div className="form-group">
            <label htmlFor="server_url">Ollama Endpoint URL</label>
            <input
              type="text"
              id="server_url"
              className="input"
              value={formData.server_url}
              onChange={(e) => handleChange('server_url', e.target.value)}
              placeholder="http://localhost:11434"
            />
            <small>Local Ollama server URL (default: http://localhost:11434)</small>
          </div>

          <div className="form-group">
            <label htmlFor="model">Ollama Model</label>
            <div className="model-input-group">
              <input
                type="text"
                id="model"
                className="input"
                value={formData.model}
                onChange={(e) => handleChange('model', e.target.value)}
                placeholder="gemma3:1b"
                list="ollama-models"
              />
              <datalist id="ollama-models">
                {availableModels.map((model) => (
                  <option key={model} value={model} />
                ))}
              </datalist>
            </div>
            <small>Model to use for rewriting (e.g., gemma3:1b, llama3:8b)</small>
          </div>

          <div className="form-group">
            <label htmlFor="api_key">Ollama API Key (Optional)</label>
            <input
              type="password"
              id="api_key"
              className="input"
              value={formData.api_key || ''}
              onChange={(e) => handleChange('api_key', e.target.value)}
              placeholder="Leave empty for local Ollama"
            />
            <small>Only needed for remote Ollama instances with authentication</small>
          </div>
        </div>
      ) : (
        <div className="settings-grid">
          <div className="form-group">
            <label htmlFor="openAIBaseURL">API Endpoint URL</label>
            <input
              type="text"
              id="openAIBaseURL"
              className="input"
              value={formData.openAIBaseURL}
              onChange={(e) => handleChange('openAIBaseURL', e.target.value)}
              placeholder="https://integrate.api.nvidia.com/v1"
            />
            <small>OpenAI-compatible API endpoint (e.g., NVIDIA NIM, LM Studio)</small>
          </div>

          <div className="form-group">
            <label htmlFor="openAIModel">Model</label>
            <input
              type="text"
              id="openAIModel"
              className="input"
              value={formData.openAIModel}
              onChange={(e) => handleChange('openAIModel', e.target.value)}
              placeholder="mistralai/mistral-7b-instruct"
            />
            <small>Model identifier (e.g., mistralai/mistral-7b-instruct)</small>
          </div>

          <div className="form-group">
            <label htmlFor="openAIAPIKey">API Key</label>
            <input
              type="password"
              id="openAIAPIKey"
              className="input"
              value={formData.openAIAPIKey || ''}
              onChange={(e) => handleChange('openAIAPIKey', e.target.value)}
              placeholder="Enter your API key"
            />
            <small>Your API key for the OpenAI-compatible service</small>
          </div>
        </div>
      )}

      <div className="settings-actions">
        <button
          type="button"
          className="btn btn-secondary"
          onClick={testConnection}
          disabled={testingConnection}
        >
          {testingConnection ? 'Testing...' : 'Test Connection'}
        </button>
        {renderConnectionStatus()}
      </div>
    </div>
  );

  // Render Updates tab
  const renderUpdatesTab = () => (
    <div className="tab-content">
      <h3>Update Settings</h3>
      
      <div className="settings-grid">
        <div className="form-group toggle-group">
          <div className="toggle-label">
            <span>Auto-Update</span>
            <small>Automatically check for and download updates</small>
          </div>
          <div
            className={`toggle-switch ${formData.autoUpdateEnabled ? 'active' : ''}`}
            onClick={() => handleChange('autoUpdateEnabled', !formData.autoUpdateEnabled)}
          />
        </div>

        <div className="form-group">
          <label htmlFor="updateChannel">Update Channel</label>
          <select
            id="updateChannel"
            className="select-input"
            value={formData.updateChannel || 'stable'}
            onChange={(e) => handleChange('updateChannel', e.target.value)}
          >
            <option value="stable">Stable</option>
            <option value="beta">Beta</option>
          </select>
          <small>Choose between stable releases or beta versions</small>
        </div>

        <div className="form-group">
          <label>Current Version</label>
          <div className="version-info">
            <span className="version-badge">{formData.currentVersion || '1.0.0'}</span>
          </div>
          <small>Your current version of TheCopyFather</small>
        </div>
      </div>
    </div>
  );

  // Render Advanced tab
  const renderAdvancedTab = () => (
    <div className="tab-content">
      <h3>Advanced Settings</h3>
      
      <div className="settings-section">
        <h4>Custom Prompts</h4>
        <p className="section-description">
          Customize the prompts used for different rewrite styles and text types.
        </p>

        <div className="prompts-grid">
          <div className="prompts-selectors">
            <div className="form-group">
              <label>Rewrite Style</label>
              <select
                className="select-input"
                value={selectedStyle}
                onChange={(e) => {
                  setSelectedStyle(e.target.value);
                  setSelectedTextType('');
                  setCustomPromptText('');
                  setDefaultPromptText('');
                }}
              >
                <option value="">Select a style...</option>
                {rewriteStyles.map((style) => (
                  <option key={style} value={style}>
                    {styleInfoData[style]?.Label || style}
                  </option>
                ))}
              </select>
            </div>

            {selectedStyle && (
              <div className="form-group">
                <label>Text Type</label>
                <select
                  className="select-input"
                  value={selectedTextType}
                  onChange={(e) => {
                    setSelectedTextType(e.target.value);
                    loadCustomPrompt(selectedStyle, e.target.value);
                  }}
                >
                  <option value="">Select a text type...</option>
                  {textTypes.map((type) => (
                    <option key={type.Type} value={type.Type}>
                      {type.Label}
                    </option>
                  ))}
                </select>
              </div>
            )}
          </div>

          {selectedStyle && selectedTextType && (
            <div className="prompt-editor">
              <div className="form-group">
                <label>Default Prompt</label>
                <textarea
                  className="textarea"
                  value={defaultPromptText}
                  readOnly
                  rows={4}
                />
              </div>

              <div className="form-group">
                <label>Custom Prompt</label>
                <textarea
                  className="textarea"
                  value={customPromptText}
                  onChange={(e) => setCustomPromptText(e.target.value)}
                  rows={4}
                  placeholder="Enter your custom prompt..."
                />
                <small>Leave empty to use the default prompt</small>
              </div>

              <div className="prompt-actions">
                <button
                  type="button"
                  className="btn btn-primary"
                  onClick={saveCustomPrompt}
                  disabled={!customPromptText}
                >
                  Save
                </button>
                <button
                  type="button"
                  className="btn btn-secondary"
                  onClick={() => loadCustomPrompt(selectedStyle, selectedTextType)}
                >
                  Reset
                </button>
                <button
                  type="button"
                  className="btn btn-danger"
                  onClick={() => deleteCustomPrompt(selectedStyle, selectedTextType)}
                >
                  Delete
                </button>
              </div>
            </div>
          )}
        </div>

        {Object.keys(customPrompts).length > 0 && (
          <div className="custom-prompts-list">
            <h5>Saved Custom Prompts</h5>
            <div className="prompts-table">
              {Object.entries(customPrompts).map(([style, typePrompts]) => (
                <div key={style} className="prompt-style-group">
                  <div className="style-header">
                    <span>{styleInfoData[style]?.Label || style}</span>
                  </div>
                  {Object.entries(typePrompts).map(([textType, prompt]) => (
                    <div key={textType} className="prompt-item">
                      <span className="prompt-text-type">
                        {textTypes.find(t => t.Type === textType)?.Label || textType}:
                      </span>
                      <span className="prompt-preview">{prompt.substring(0, 50)}...</span>
                      <button
                        type="button"
                        className="btn-icon"
                        onClick={() => {
                          setSelectedStyle(style);
                          setSelectedTextType(textType);
                          loadCustomPrompt(style, textType);
                        }}
                      >
                        ✏️
                      </button>
                      <button
                        type="button"
                        className="btn-icon"
                        onClick={() => deleteCustomPrompt(style, textType)}
                      >
                        🗑️
                      </button>
                    </div>
                  ))}
                </div>
              ))}
            </div>
            <button
              type="button"
              className="btn btn-danger"
              onClick={resetAllCustomPrompts}
            >
              Reset All Custom Prompts
            </button>
          </div>
        )}
      </div>

      <div className="settings-section">
        <h4>Debug & Logging</h4>
        <p className="section-description">
          Options for debugging and troubleshooting.
        </p>

        <div className="settings-grid">
          <div className="form-group toggle-group">
            <div className="toggle-label">
              <span>Enable Debug Logging</span>
              <small>Log additional information for troubleshooting</small>
            </div>
            <div
              className={`toggle-switch ${formData.first_run ? 'active' : ''}`}
              onClick={() => handleChange('first_run', !formData.first_run)}
            />
          </div>
        </div>
      </div>
    </div>
  );

  // Render the active tab
  const renderActiveTab = () => {
    switch (activeTab) {
      case 'general':
        return renderGeneralTab();
      case 'ai':
        return renderAITab();
      case 'updates':
        return renderUpdatesTab();
      case 'advanced':
        return renderAdvancedTab();
      default:
        return renderGeneralTab();
    }
  };

  if (isLoading) {
    return (
      <div className="settings-container">
        <div className="loading">Loading settings...</div>
      </div>
    );
  }

  return (
    <div className="settings-container">
      <header className="settings-header">
        <h2>Settings</h2>
        <p className="settings-subtitle">Configure TheCopyFather to your preferences</p>
      </header>

      {renderTabNavigation()}

      <form onSubmit={handleSubmit} className="settings-form">
        {renderActiveTab()}

        <div className="settings-footer">
          <button type="button" className="btn btn-secondary" onClick={onCancel}>
            Cancel
          </button>
          <button type="submit" className="btn btn-primary">
            Save Settings
          </button>
        </div>
      </form>
    </div>
  );
};

// StyleInfo data for the rewriter module
const styleInfoData: Record<string, StyleInfo> = {
grammar: { Label: 'Grammar & Spelling', Icon: '🛡️', Description: 'Corrects errors and improves flow' },
paraphrase: { Label: 'Paraphrase', Icon: '🔄', Description: 'Rewrite with different structure' },
standard: { Label: 'Standard', Icon: '📝', Description: 'Balanced and natural' },
formal: { Label: 'Formal', Icon: '📢', Description: 'Professional tone' },
casual: { Label: 'Casual', Icon: '💬', Description: 'Conversational' },
creative: { Label: 'Creative', Icon: '✨', Description: 'Expressive and vivid' },
short: { Label: 'Short', Icon: '📏', Description: 'Concise and brief' },
expand: { Label: 'Expand', Icon: '📖', Description: 'More detail and depth' },
summarize: { Label: 'TL;DR Summary', Icon: '📋', Description: 'Concise 2-3 sentence summary' },
bullets: { Label: 'Key Points', Icon: '•••', Description: 'Extract 3-5 main bullet points' },
insights: { Label: 'Key Insights', Icon: '💡', Description: 'Important facts and arguments' },
};

export default SettingsNew;
