interface ActionFooterProps {
  loading: boolean
  error: string | null
  result: string
  copied: boolean
  handleCopy: () => void
  onShowDiff?: (rewritten: string) => void
  handleReplace: () => void
  showPasteDialog: boolean
  dontAskAgain: boolean
  setDontAskAgain: (val: boolean) => void
  handlePasteCancel: () => void
  handlePasteConfirm: () => void
}

export function ActionFooter({
  loading,
  error,
  result,
  copied,
  handleCopy,
  onShowDiff,
  handleReplace,
  showPasteDialog,
  dontAskAgain,
  setDontAskAgain,
  handlePasteCancel,
  handlePasteConfirm
}: ActionFooterProps) {
  return (
    <>
      <footer className="popup-footer">
        <div className="footer-left">
          <button
            className="footer-btn"
            onClick={handleCopy}
            disabled={loading || !!error || !result}
          >
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round"><rect x="9" y="9" width="13" height="13" rx="2" ry="2"/><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/></svg>
            {copied ? 'Copied' : 'Copy'}
          </button>
          {onShowDiff && (
            <button
              className="footer-btn"
              onClick={() => onShowDiff(result)}
              disabled={loading || !!error || !result}
              title="View differences"
            >
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round"><polyline points="16 3 21 3 21 8"/><line x1="4" y1="20" x2="21" y2="3"/><polyline points="21 16 21 21 16 21"/><line x1="15" y1="15" x2="21" y2="21"/></svg>
              Diff
            </button>
          )}
        </div>
        <button
          className="btn-replace"
          onClick={handleReplace}
          disabled={loading || !!error || !result}
        >
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round"><polyline points="17 1 21 5 17 9"/><path d="M3 11V9a4 4 0 0 1 4-4h14"/><polyline points="7 23 3 19 7 15"/><path d="M21 13v2a4 4 0 0 1-4 4H3"/></svg>
          Replace
        </button>
      </footer>

      {showPasteDialog && (
        <div className="dialog-overlay">
          <div className="dialog">
            <div className="dialog-header">
              <h3>📋 Replace Text?</h3>
              <p>Automatically paste the rewritten text?</p>
            </div>
            <div className="dialog-content">
              <label className="checkbox-label">
                <input
                  type="checkbox"
                  checked={dontAskAgain}
                  onChange={(e) => setDontAskAgain(e.target.checked)}
                />
                <span>Don't ask again</span>
              </label>
            </div>
            <div className="dialog-footer">
              <button className="btn btn-secondary" onClick={handlePasteCancel}>
                Copy Only
              </button>
              <button className="btn btn-primary" onClick={handlePasteConfirm}>
                ✓ Paste
              </button>
            </div>
          </div>
        </div>
      )}
    </>
  )
}
