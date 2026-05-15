import { useState, useEffect, useCallback, useRef } from 'react'
import * as runtime from '../../wailsjs/runtime'
import * as AppAPI from '../../wailsjs/go/main/App'
import '../styles/main.css'
import '../styles/MiniMode.css'

interface MiniModeProps {
  originalText: string
  currentStyle: string
  onExpand: () => void
  onClose: () => void
  onStyleChange: (style: string) => void
  onQuickRewrite: () => void
  availableStyles: Array<{ value: string; label: string; icon: string }>
  isGenerating: boolean
}

export default function MiniMode({
  originalText,
  currentStyle,
  onExpand,
  onClose,
  onStyleChange,
  onQuickRewrite,
  availableStyles,
  isGenerating
}: MiniModeProps) {
  const [showStyleDropdown, setShowStyleDropdown] = useState(false)
  const [position, setPosition] = useState({ x: 0, y: 0 })
  const [isDragging, setIsDragging] = useState(false)
  const dragOffset = useRef({ x: 0, y: 0 })
  const miniRef = useRef<HTMLDivElement>(null)
  const dropdownRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const initPosition = async () => {
      try {
        const pos = await AppAPI.GetCursorPosition()
        if (pos && miniRef.current) {
          const offsetX = 20
          const offsetY = -80
          let newX = pos.x + offsetX
          let newY = pos.y + offsetY

          const rect = miniRef.current.getBoundingClientRect()
          const screenWidth = window.screen.width
          const screenHeight = window.screen.height

          if (newX + 280 > screenWidth) newX = screenWidth - 290
          if (newY < 0) newY = 10
          if (newY + 100 > screenHeight) newY = screenHeight - 110
          if (newX < 0) newX = 10

          setPosition({ x: newX, y: newY })
        }
      } catch (e) {
        setPosition({
          x: window.screen.width / 2 - 140,
          y: window.screen.height / 2 - 50
        })
      }
    }
    initPosition()
  }, [])

  const handleMouseDown = (e: React.MouseEvent) => {
    if ((e.target as HTMLElement).closest('.minimode-btn, .minimode-dropdown')) {
      return
    }
    setIsDragging(true)
    dragOffset.current = {
      x: e.clientX - position.x,
      y: e.clientY - position.y
    }
  }

  const handleMouseMove = useCallback((e: MouseEvent) => {
    if (!isDragging) return
    setPosition({
      x: e.clientX - dragOffset.current.x,
      y: e.clientY - dragOffset.current.y
    })
  }, [isDragging])

  const handleMouseUp = useCallback(() => {
    setIsDragging(false)
  }, [])

  useEffect(() => {
    if (isDragging) {
      window.addEventListener('mousemove', handleMouseMove)
      window.addEventListener('mouseup', handleMouseUp)
    }
    return () => {
      window.removeEventListener('mousemove', handleMouseMove)
      window.removeEventListener('mouseup', handleMouseUp)
    }
  }, [isDragging, handleMouseMove, handleMouseUp])

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setShowStyleDropdown(false)
      }
    }
    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [])

  useEffect(() => {
    runtime.WindowSetSize(340, 140)
    runtime.WindowShow()

    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        onClose()
      } else if (e.key === 'Enter' && !e.shiftKey) {
        e.preventDefault()
        onQuickRewrite()
      } else if (e.key === ' ' && e.ctrlKey) {
        e.preventDefault()
        onExpand()
      }
    }
    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [onClose, onExpand, onQuickRewrite])

  const truncateText = (text: string) => {
    if (text.length <= 35) return text
    return text.substring(0, 35) + '...'
  }

  const currentStyleData = availableStyles.find(s => s.value === currentStyle) || availableStyles[0]

  return (
    <div
      ref={miniRef}
      className={`mini-mode-v2 ${isDragging ? 'dragging' : ''}`}
      style={{
        left: position.x,
        top: position.y,
        position: 'fixed'
      }}
      onMouseDown={handleMouseDown}
    >
      {/* Header */}
      <div className="mini-header-v2">
        <div className="drag-indicator">
          <span className="drag-dots">⋮⋮</span>
        </div>
        <div className="original-preview" title={originalText}>
          {truncateText(originalText)}
        </div>
        <button className="mini-close-btn" onClick={onClose} title="Close">
          ✕
        </button>
      </div>

      {/* Actions */}
      <div className="mini-actions-v2">
        {/* Style Selector */}
        <div className="mini-style-selector" ref={dropdownRef}>
          <button
            className="mini-style-btn"
            onClick={() => setShowStyleDropdown(!showStyleDropdown)}
            title="Change style"
          >
            <span className="style-icon">{currentStyleData?.icon}</span>
            <span className="style-label-mini">{currentStyleData?.label}</span>
            <span className={`mini-chevron ${showStyleDropdown ? 'open' : ''}`}>▼</span>
          </button>

          {showStyleDropdown && (
            <div className="mini-dropdown">
              {availableStyles.map(style => (
                <button
                  key={style.value}
                  className={`mini-dropdown-item ${style.value === currentStyle ? 'active' : ''}`}
                  onClick={() => {
                    onStyleChange(style.value)
                    setShowStyleDropdown(false)
                  }}
                >
                  <span className="item-icon">{style.icon}</span>
                  <span className="item-label">{style.label}</span>
                </button>
              ))}
            </div>
          )}
        </div>

        {/* Quick Rewrite */}
        <button
          className={`mini-rewrite-btn ${isGenerating ? 'generating' : ''}`}
          onClick={onQuickRewrite}
          disabled={isGenerating}
          title="Rewrite"
        >
          {isGenerating ? (
            <span className="mini-spinner"></span>
          ) : (
            <>
              <span>🔄</span>
              <span>Rewrite</span>
            </>
          )}
        </button>

        {/* Expand */}
        <button
          className="mini-expand-btn"
          onClick={onExpand}
          title="Expand"
        >
          <span>⛶</span>
        </button>
      </div>

      {/* Status */}
      <div className="mini-status-v2">
        <span className="status-hint">
          {isGenerating ? 'Generating...' : 'Press Enter to rewrite'}
        </span>
      </div>
    </div>
  )
}
