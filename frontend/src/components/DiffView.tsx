import { useState, useEffect, useRef } from 'react'
import * as AppAPI from '../../wailsjs/go/main/App'
import { rewriter as rewriterModels } from '../../wailsjs/go/models'
import '../styles/DiffView.css'

interface DiffViewProps {
  originalText: string
  rewrittenText: string
  onClose: () => void
}

export default function DiffView({ originalText, rewrittenText, onClose }: DiffViewProps) {
  const [diffResult, setDiffResult] = useState<rewriterModels.DiffResult | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [showSideBySide, setShowSideBySide] = useState(false)
  const [diffHtml, setDiffHtml] = useState<string>('')
  const containerRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const computeDiff = async () => {
      if (!originalText || !rewrittenText) {
        setDiffResult(null)
        setLoading(false)
        return
      }

      try {
        const result = await AppAPI.ComputeDiff(originalText, rewrittenText)
        setDiffResult(result)
        setDiffHtml(result.html)
        setLoading(false)
      } catch (err) {
        console.error('Failed to compute diff:', err)
        setError('Failed to compute diff')
        setLoading(false)
      }
    }

    computeDiff()
  }, [originalText, rewrittenText])

  // Enhanced diff HTML with better styling
  useEffect(() => {
    if (diffResult?.html) {
      // Enhance the HTML with better CSS classes and styling
      const enhancedHtml = diffResult.html
        .replace(/<ins>/g, '<ins class="diff-insert">')
        .replace(/<del>/g, '<del class="diff-delete">')
        .replace(/<span class="diff_inline">/g, '<span class="diff-inline">')
      setDiffHtml(enhancedHtml)
    }
  }, [diffResult])

  const toggleViewMode = () => {
    setShowSideBySide(!showSideBySide)
  }

  const getSideBySideView = () => {
    const leftLines = originalText.split('\n')
    const rightLines = rewrittenText.split('\n')
    const maxLines = Math.max(leftLines.length, rightLines.length)

    return (
      <div className="side-by-side-container">
        <div className="side-by-side-panel original-panel">
          <div className="panel-header">Original</div>
          <div className="panel-content">
            {leftLines.map((line, i) => (
              <div key={`left-${i}`} className="line">
                <span className="line-number">{i + 1}</span>
                <span className="line-text">{line || ' '}</span>
              </div>
            ))}
            {Array.from({ length: maxLines - leftLines.length }).map((_, i) => (
              <div key={`left-empty-${leftLines.length + i}`} className="line empty-line">
                <span className="line-number">{leftLines.length + i + 1}</span>
              </div>
            ))}
          </div>
        </div>
        <div className="side-by-side-panel rewritten-panel">
          <div className="panel-header">Rewritten</div>
          <div className="panel-content">
            {rightLines.map((line, i) => (
              <div key={`right-${i}`} className="line">
                <span className="line-number">{i + 1}</span>
                <span className="line-text">{line || ' '}</span>
              </div>
            ))}
            {Array.from({ length: maxLines - rightLines.length }).map((_, i) => (
              <div key={`right-empty-${rightLines.length + i}`} className="line empty-line">
                <span className="line-number">{rightLines.length + i + 1}</span>
              </div>
            ))}
          </div>
        </div>
      </div>
    )
  }

  const getInlineDiffView = () => {
    if (!diffResult?.has_diff) {
      return (
        <div className="no-changes">
          <div className="no-changes-icon">✓</div>
          <div className="no-changes-text">No changes detected</div>
          <div className="no-changes-subtext">The original and rewritten text are identical</div>
        </div>
      )
    }

    return (
      <div className="inline-diff-container">
        <div className="diff-stats">
          <span className="stat additions">
            +{diffResult.additions} addition{diffResult.additions !== 1 ? 's' : ''}
          </span>
          <span className="stat deletions">
            -{diffResult.deletions} deletion{diffResult.deletions !== 1 ? 's' : ''}
          </span>
        </div>
        <div 
          className="diff-content"
          dangerouslySetInnerHTML={{ __html: diffHtml }}
        />
      </div>
    )
  }

  if (loading) {
    return (
      <div className="diff-view loading">
        <div className="loading-spinner"></div>
        <div>Computing diff...</div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="diff-view error">
        <div className="error-icon">⚠️</div>
        <div className="error-message">{error}</div>
        <button className="btn-retry" onClick={() => window.location.reload()}>Retry</button>
      </div>
    )
  }

  return (
    <div className="diff-view" ref={containerRef}>
      <div className="diff-header">
        <div className="diff-title">
          <span className="diff-icon">🔍</span>
          <span>Text Comparison</span>
        </div>
        <div className="diff-actions">
          <button 
            className={`view-toggle ${!showSideBySide ? 'active' : ''}`}
            onClick={toggleViewMode}
            title="Inline diff view"
          >
            📄
          </button>
          <button 
            className={`view-toggle ${showSideBySide ? 'active' : ''}`}
            onClick={toggleViewMode}
            title="Side by side view"
          >
            ⇄
          </button>
          <button className="close-btn" onClick={onClose} title="Close">
            ✕
          </button>
        </div>
      </div>
      
      <div className="diff-body">
        {showSideBySide ? getSideBySideView() : getInlineDiffView()}
      </div>
      
      <div className="diff-footer">
        <div className="footer-info">
          <span>Original: {originalText.length} chars</span>
          <span className="separator">|</span>
          <span>Rewritten: {rewrittenText.length} chars</span>
          {diffResult && (
            <>
              <span className="separator">|</span>
              <span>Changed: {Math.abs(originalText.length - rewrittenText.length)} chars</span>
            </>
          )}
        </div>
      </div>
    </div>
  )
}
