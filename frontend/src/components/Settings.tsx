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

// OpenAI endpoint presets for quick 1-click configuration
const OPENAI_PRESETS = [
  {
    name: 'NVIDIA NIM',
    baseURL: 'https://integrate.api.nvidia.com/v1',
    model: 'mistralai/mistral-7b-instruct',
  },
  {
    name: 'Groq',
    baseURL: 'https://api.groq.com/openai/v1',
    model: 'llama-3.3-70b-versatile',
  },
  {
    name: 'OpenAI',
    baseURL: 'https://api.openai.com/v1',
    model: 'gpt-4o-mini',
  },
  {
    name: 'LM Studio (Local)',
    baseURL: 'http://localhost:1234/v1',
    model: 'local-model',
  },
  {
    name: 'Ollama API',
    baseURL: 'http://localhost:11434/v1',
    model: 'llama3.2',
  },
];

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
        const [styles, analysis, types, prompts] = await Promise.all([
          onLoadRewriteStyles(),
          onLoadAnalysisStyles(),
          onLoadTextTypes(),
          onLoadCustomPrompts(),
        ]);
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

  // Handle form field changes with functional updates to avoid race conditions/stale closures
  const handleChange = <K extends keyof Config>(key: K, value: Config[K]) => {
    setFormData(prev => ({ ...prev, [key]: value }));
  };

  // Switch between embedded and OpenAI execution modes
  const handleSelectMode = (mode: 'embedded' | 'openai') => {
    setFormData(prev => {
      if (mode === 'openai') {
        return {
          ...prev,
          provider_mode: 'openai',
          useOpenAICompatible: true,
          openAIBaseURL: prev.openAIBaseURL || 'https://integrate.api.nvidia.com/v1',
          openAIModel: prev.openAIModel || 'mistralai/mistral-7b-instruct',
        };
      }
      return {
        ...prev,
        provider_mode: 'embedded',
        useOpenAICompatible: false,
      };
    });
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
      const mode = (formData.provider_mode === 'openai' || formData.useOpenAICompatible) ? 'openai' : 'embedded';
      let version = '';
      if (mode === 'openai') {
        const baseURL = formData.openAIBaseURL || 'https://integrate.api.nvidia.com/v1';
        const model = formData.openAIModel || 'mistralai/mistral-7b-instruct';
        const apiKey = formData.openAIAPIKey || '';
        version = await SettingsAPI.TestConnection(
          baseURL,
          model,
          apiKey,
          true
        );
      } else {
        version = await SettingsAPI.TestConnection('', '', '', false);
      }
      setDetectedVersion(version);
      setConnectionStatus('success');
    } catch (error: any) {
      setDetectedVersion(error?.message || String(error));
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
      <div 
        className={`status-badge ${connectionStatus}`}
        title={detectedVersion}
      >
        {connectionStatus === 'success'
          ? `✓ Connected (${detectedVersion || 'OK'})`
          : `✗ Connection failed: ${detectedVersion || 'Unknown error'}`}
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
  const renderAITab = () => {
    const currentMode = (formData.provider_mode === 'openai' || formData.useOpenAICompatible) ? 'openai' : 'embedded';

    return (
      <div className="tab-content">
        <h3>AI Provider Configuration</h3>

        {/* Provider Mode Selection Cards */}
        <div className="form-group" style={{ marginBottom: '22px' }}>
          <label style={{ fontWeight: 600, display: 'block', marginBottom: '8px' }}>Execution Mode</label>
          <div className="mode-selector-grid">
            <button
              type="button"
              className={`mode-selector-card ${currentMode === 'embedded' ? 'active' : ''}`}
              onClick={() => handleSelectMode('embedded')}
            >
              <div className="mode-selector-card-header">
                <span className="mode-selector-card-title">
                  💻 Embedded llama.cpp
                </span>
                <span className="mode-selector-radio">
                  {currentMode === 'embedded' && <span className="mode-selector-radio-dot" />}
                </span>
              </div>
              <span className="mode-selector-card-desc">
                Local built-in engine (~200MB RAM, 100% offline & private)
              </span>
            </button>

            <button
              type="button"
              className={`mode-selector-card ${currentMode === 'openai' ? 'active' : ''}`}
              onClick={() => handleSelectMode('openai')}
            >
              <div className="mode-selector-card-header">
                <span className="mode-selector-card-title">
                  🌐 OpenAI-Compatible
                </span>
                <span className="mode-selector-radio">
                  {currentMode === 'openai' && <span className="mode-selector-radio-dot" />}
                </span>
              </div>
              <span className="mode-selector-card-desc">
                Cloud or custom API (NVIDIA NIM, Groq, LM Studio, vLLM, OpenAI)
              </span>
            </button>
          </div>
        </div>

        {currentMode === 'embedded' ? (
          <div className="settings-grid">
            {/* Model Card */}
            <div style={{
              gridColumn: '1 / -1',
              padding: '14px 16px',
              backgroundColor: 'rgba(255, 255, 255, 0.05)',
              borderRadius: '8px',
              border: '1px solid rgba(255, 255, 255, 0.1)',
              marginBottom: '10px',
              display: 'flex',
              flexWrap: 'wrap',
              gap: '16px',
              justifyContent: 'space-between',
              alignItems: 'center'
            }}>
              <div>
                <strong style={{ fontSize: '15px' }}>Model: SmolLM2-360M (Q4_K_M GGUF)</strong>
                <div style={{ fontSize: '12px', color: '#a0aec0', marginTop: '4px' }}>
                  Ultra-lightweight in-app inference via embedded <code>llama.cpp</code> (~200 MB RAM)
                </div>
              </div>
              <div style={{ display: 'flex', gap: '8px' }}>
                <span className="badge" style={{ padding: '4px 8px', borderRadius: '4px', background: '#3182ce', fontSize: '11px', fontWeight: 600 }}>CPU Hardware</span>
                <span className="badge" style={{ padding: '4px 8px', borderRadius: '4px', background: '#38a169', fontSize: '11px', fontWeight: 600 }}>Thinking: OFF</span>
                <span className="badge" style={{ padding: '4px 8px', borderRadius: '4px', background: '#805ad5', fontSize: '11px', fontWeight: 600 }}>Streaming: ON</span>
              </div>
            </div>

            {/* CPU Threads */}
            <div className="form-group">
              <label htmlFor="embedded_max_threads">CPU Threads Limit</label>
              <input
                type="number"
                id="embedded_max_threads"
                className="input"
                min="1"
                max="32"
                value={formData.embedded_max_threads ?? 6}
                onChange={(e) => handleChange('embedded_max_threads', parseInt(e.target.value) || 6)}
              />
              <small>Auto-detects CPU cores and caps execution to this limit (recommended: 4–6)</small>
            </div>

            {/* Context Size */}
            <div className="form-group">
              <label htmlFor="embedded_context_size">Context Window (Tokens)</label>
              <select
                id="embedded_context_size"
                className="select-input"
                value={formData.embedded_context_size ?? 4096}
                onChange={(e) => handleChange('embedded_context_size', parseInt(e.target.value) || 4096)}
              >
                <option value={2048}>2048 tokens (2K)</option>
                <option value={3072}>3072 tokens (3K)</option>
                <option value={4096}>4096 tokens (4K - Recommended)</option>
              </select>
              <small>Maximum context length for rewriting and text analysis</small>
            </div>

            {/* Thinking Toggle */}
            <div className="form-group toggle-group" style={{ gridColumn: '1 / -1' }}>
              <div className="toggle-label">
                <span>Disable Thinking (Thinking: OFF)</span>
                <small>Filters out reasoning/thought chains so you get clean, instant rewritten text</small>
              </div>
              <div
                className={`toggle-switch ${formData.disable_thinking !== false ? 'active' : ''}`}
                onClick={() => handleChange('disable_thinking', formData.disable_thinking === false)}
              />
            </div>

            {/* Streaming Toggle */}
            <div className="form-group toggle-group" style={{ gridColumn: '1 / -1' }}>
              <div className="toggle-label">
                <span>Token Streaming (Streaming: ON)</span>
                <small>Real-time token generation for fast responsive UI and ghost typing</small>
              </div>
              <div
                className={`toggle-switch ${!formData.disableStreaming ? 'active' : ''}`}
                onClick={() => handleChange('disableStreaming', !formData.disableStreaming)}
              />
            </div>

            {/* GGUF Model Path */}
            <div className="form-group" style={{ gridColumn: '1 / -1' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '6px' }}>
                <label htmlFor="embedded_model_path" style={{ margin: 0 }}>GGUF Model File Path</label>
                <button
                  type="button"
                  className="btn btn-secondary"
                  style={{ padding: '3px 10px', fontSize: '12px' }}
                  onClick={() => (SettingsAPI as any).OpenModelsFolder?.()}
                >
                  📁 Open Models Folder (%APPDATA%)
                </button>
              </div>
              <input
                type="text"
                id="embedded_model_path"
                className="input"
                value={formData.embedded_model_path || ''}
                onChange={(e) => handleChange('embedded_model_path', e.target.value)}
                placeholder="%APPDATA%\TheCopyfather\models\qwen3-1.7b-q4_k_m.gguf"
              />
              <small>Leave empty or set path. Standard AppData location: <code>%APPDATA%\TheCopyfather\models\</code></small>
            </div>

            {/* llama-server binary path */}
            <div className="form-group" style={{ gridColumn: '1 / -1' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '6px' }}>
                <label htmlFor="embedded_binary_path" style={{ margin: 0 }}>Embedded Engine Path</label>
                <button
                  type="button"
                  className="btn btn-secondary"
                  style={{ padding: '3px 10px', fontSize: '12px' }}
                  onClick={() => (SettingsAPI as any).OpenEngineFolder?.()}
                >
                  📁 Open Engine Folder (%APPDATA%)
                </button>
              </div>
              <input
                type="text"
                id="embedded_binary_path"
                className="input"
                value={formData.embedded_binary_path || ''}
                onChange={(e) => handleChange('embedded_binary_path', e.target.value)}
                placeholder="%APPDATA%\TheCopyfather\engine\llama-server.exe"
              />
              <small>Leave empty or set path. Standard AppData location: <code>%APPDATA%\TheCopyfather\engine\</code></small>
            </div>
          </div>
        ) : (
          <div className="settings-grid">
            {/* OpenAI Provider Overview Card */}
            <div style={{
              gridColumn: '1 / -1',
              padding: '14px 16px',
              backgroundColor: 'rgba(255, 255, 255, 0.05)',
              borderRadius: '8px',
              border: '1px solid rgba(255, 255, 255, 0.1)',
              marginBottom: '10px',
              display: 'flex',
              flexWrap: 'wrap',
              gap: '16px',
              justifyContent: 'space-between',
              alignItems: 'center'
            }}>
              <div>
                <strong style={{ fontSize: '15px' }}>Provider: Remote / Cloud OpenAI-Compatible API</strong>
                <div style={{ fontSize: '12px', color: '#a0aec0', marginTop: '4px' }}>
                  Connect to any standard OpenAI chat/completions endpoint
                </div>
              </div>
              <div style={{ display: 'flex', gap: '8px' }}>
                <span className="badge" style={{ padding: '4px 8px', borderRadius: '4px', background: '#3182ce', fontSize: '11px', fontWeight: 600 }}>Cloud / Local API</span>
                <span className="badge" style={{ padding: '4px 8px', borderRadius: '4px', background: '#38a169', fontSize: '11px', fontWeight: 600 }}>Guardrails: ACTIVE</span>
                <span className="badge" style={{ padding: '4px 8px', borderRadius: '4px', background: '#805ad5', fontSize: '11px', fontWeight: 600 }}>Streaming: {formData.disableStreaming ? 'OFF' : 'ON'}</span>
              </div>
            </div>

            {/* Quick Provider Presets */}
            <div style={{ gridColumn: '1 / -1', marginBottom: '4px' }}>
              <label style={{ fontSize: '12px', fontWeight: 600, color: 'var(--text-secondary)', display: 'block', marginBottom: '6px' }}>
                Quick Endpoint Presets:
              </label>
              <div className="preset-chips-container">
                {OPENAI_PRESETS.map((preset) => (
                  <button
                    key={preset.name}
                    type="button"
                    className="preset-chip"
                    onClick={() => {
                      setFormData(prev => ({
                        ...prev,
                        openAIBaseURL: preset.baseURL,
                        openAIModel: preset.model,
                      }));
                    }}
                    title={`Use ${preset.name} (${preset.baseURL})`}
                  >
                    ⚡ {preset.name}
                  </button>
                ))}
              </div>
            </div>

            <div className="form-group">
              <label htmlFor="openAIBaseURL">API Endpoint URL</label>
              <input
                type="text"
                id="openAIBaseURL"
                className="input"
                value={formData.openAIBaseURL || ''}
                onChange={(e) => handleChange('openAIBaseURL', e.target.value)}
                placeholder="https://integrate.api.nvidia.com/v1"
              />
              <small>OpenAI-compatible API endpoint (e.g., NVIDIA NIM, LM Studio, vLLM, OpenAI)</small>
            </div>

            <div className="form-group">
              <label htmlFor="openAIModel">Model</label>
              <input
                type="text"
                id="openAIModel"
                className="input"
                value={formData.openAIModel || ''}
                onChange={(e) => handleChange('openAIModel', e.target.value)}
                placeholder="mistralai/mistral-7b-instruct"
              />
              <small>Model identifier (e.g. mistralai/mistral-7b-instruct, gpt-4o-mini)</small>
            </div>

            <div className="form-group">
              <label htmlFor="openAIAPIKey">API Key</label>
              <input
                type="password"
                id="openAIAPIKey"
                className="input"
                value={formData.openAIAPIKey || ''}
                onChange={(e) => handleChange('openAIAPIKey', e.target.value)}
                placeholder="Enter your API key (leave blank for local models)"
              />
              <small>Your API key for the remote service (stored securely)</small>
            </div>

            <div className="form-group toggle-group" style={{ gridColumn: '1 / -1' }}>
              <div className="toggle-label">
                <span>Disable Streaming</span>
                <small>Enable if your remote provider does not support SSE streaming</small>
              </div>
              <div
                className={`toggle-switch ${formData.disableStreaming ? 'active' : ''}`}
                onClick={() => handleChange('disableStreaming', !formData.disableStreaming)}
              />
            </div>
          </div>
        )}

        <div className="settings-actions" style={{ marginTop: '20px' }}>
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
  };

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
