import React, { RefObject } from 'react'
import { rewriter as rewriterModels } from '../../wailsjs/go/models'

export interface StyleData {
  value: string
  label: string
  icon: string
  desc: string
}

interface StyleSelectorProps {
  dropdownRef: RefObject<HTMLDivElement>
  textTypeDropdownRef: RefObject<HTMLDivElement>
  dropdownOpen: boolean
  setDropdownOpen: (open: boolean) => void
  textTypeDropdownOpen: boolean
  setTextTypeDropdownOpen: (open: boolean) => void
  mainMode: 'rewrite' | 'sliders' | 'analyze'
  rewriteStyle: string
  analysisStyle: string
  REWRITE_STYLES: StyleData[]
  ANALYSIS_STYLES: StyleData[]
  currentRewriteStyleData: StyleData
  currentAnalysisStyleData: StyleData
  handleRewriteStyleChange: (val: string) => void
  handleAnalysisStyleChange: (val: string) => void
  selectedTextType: string
  availableTextTypes: rewriterModels.TextTypeInfo[]
  isUserOverride: boolean
  handleTextTypeChange: (val: string) => void
}

export function StyleSelector({
  dropdownRef,
  textTypeDropdownRef,
  dropdownOpen,
  setDropdownOpen,
  textTypeDropdownOpen,
  setTextTypeDropdownOpen,
  mainMode,
  rewriteStyle,
  analysisStyle,
  REWRITE_STYLES,
  ANALYSIS_STYLES,
  currentRewriteStyleData,
  currentAnalysisStyleData,
  handleRewriteStyleChange,
  handleAnalysisStyleChange,
  selectedTextType,
  availableTextTypes,
  isUserOverride,
  handleTextTypeChange
}: StyleSelectorProps) {
  const stylesList = mainMode === 'rewrite' ? REWRITE_STYLES : ANALYSIS_STYLES
  const currentStyleValue = mainMode === 'rewrite' ? rewriteStyle : analysisStyle

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (!dropdownOpen && (e.key === 'ArrowDown' || e.key === 'ArrowUp' || e.key === 'Enter')) {
      e.preventDefault()
      setDropdownOpen(true)
      return
    }

    if (dropdownOpen) {
      const currentIndex = stylesList.findIndex(s => s.value === currentStyleValue)
      if (e.key === 'ArrowDown') {
        e.preventDefault()
        const nextIndex = (currentIndex + 1) % stylesList.length
        const nextStyle = stylesList[nextIndex].value
        if (mainMode === 'rewrite') handleRewriteStyleChange(nextStyle)
        else handleAnalysisStyleChange(nextStyle)
      } else if (e.key === 'ArrowUp') {
        e.preventDefault()
        const prevIndex = (currentIndex - 1 + stylesList.length) % stylesList.length
        const prevStyle = stylesList[prevIndex].value
        if (mainMode === 'rewrite') handleRewriteStyleChange(prevStyle)
        else handleAnalysisStyleChange(prevStyle)
      } else if (e.key === 'Enter' || e.key === 'Escape') {
        e.preventDefault()
        setDropdownOpen(false)
      }
    }
  }

  return (
    <div className="style-section" ref={dropdownRef}>
      <div className="style-dropdown">
        <button
          className={`style-trigger ${dropdownOpen ? 'open' : ''}`}
          onClick={() => setDropdownOpen(!dropdownOpen)}
          onKeyDown={handleKeyDown}
          aria-haspopup="listbox"
          aria-expanded={dropdownOpen}
          title="Select style (Use Arrow keys to cycle)"
        >
          <span className="style-icon">
            {mainMode === 'rewrite' ? currentRewriteStyleData.icon : currentAnalysisStyleData.icon}
          </span>
          <span className="style-label">
            {mainMode === 'rewrite' ? currentRewriteStyleData.label : currentAnalysisStyleData.label}
          </span>
          <span className="chevron">▼</span>
        </button>

        {dropdownOpen && (
          <div className="style-menu">
            {mainMode === 'rewrite' ? (
              REWRITE_STYLES.map(s => (
                <button
                  key={s.value}
                  className={`style-option ${s.value === rewriteStyle ? 'active' : ''}`}
                  onClick={() => handleRewriteStyleChange(s.value)}
                >
                  <span className="option-icon">{s.icon}</span>
                  <div className="option-content">
                    <span className="option-label">{s.label}</span>
                    <span className="option-desc">{s.desc}</span>
                  </div>
                </button>
              ))
            ) : (
              ANALYSIS_STYLES.map(s => (
                <button
                  key={s.value}
                  className={`style-option ${s.value === analysisStyle ? 'active' : ''}`}
                  onClick={() => handleAnalysisStyleChange(s.value)}
                >
                  <span className="option-icon">{s.icon}</span>
                  <div className="option-content">
                    <span className="option-label">{s.label}</span>
                    <span className="option-desc">{s.desc}</span>
                  </div>
                </button>
              ))
            )}
          </div>
        )}
      </div>

      {selectedTextType && (
        <div className="text-type-badge" ref={textTypeDropdownRef}>
          <button
            className="text-type-trigger"
            onClick={() => setTextTypeDropdownOpen(!textTypeDropdownOpen)}
            title={isUserOverride ? 'You selected this type' : 'Auto-detected'}
          >
            <span>{availableTextTypes.find(t => t.Type === selectedTextType)?.Icon || '📝'}</span>
            <span>{availableTextTypes.find(t => t.Type === selectedTextType)?.Label || 'Text'}</span>
          </button>
          {textTypeDropdownOpen && (
            <div className="text-type-menu">
              {availableTextTypes.map(t => (
                <button
                  key={t.Type}
                  className={`text-type-option ${t.Type === selectedTextType ? 'active' : ''}`}
                  onClick={() => handleTextTypeChange(t.Type)}
                >
                  <span className="option-icon">{t.Icon}</span>
                  <span className="option-label">{t.Label}</span>
                </button>
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  )
}
