# Plan for User Experience Improvements

## Goal: Improve user experience with more informative error messages and troubleshooting suggestions

## Areas to Improve:

### 1. Hotkey Registration Error (initWindowsComponents)
Current Message: "Failed to register hotkey '%s'. It might be in use by another app.\nError: %v"
Improvement: Add more specific troubleshooting steps:
   - Suggest trying a different hotkey combination
   - Mention that some applications (like games, screen recorders) may grab hotkeys globally
   - Suggest restarting the application after changing hotkey

### 2. Hotkey Registration Error (SaveSettings)
Current Message: "Failed to register hotkey '%s': %v"
Improvement: Same as above - add troubleshooting steps for hotkey conflicts

### 3. Clipboard Restoration Error (onHotkeyTriggered retry logic)
Current Message: "Failed to restore original clipboard content.\nError: %v"
Improvement: Add context about what this means and suggest manual restoration if needed

### 4. Paste Failure (ApplyRewriteAndPaste)
Current: Only logs error, no user feedback
Improvement: Add error dialog when SimulatePaste fails with troubleshooting suggestions:
   - Explain that the application couldn't paste the text
   - Suggest clicking the paste button manually or using Ctrl+V
   - Mention that some applications may block automated paste for security

### 5. General Error Dialog Improvements
For all error dialogs, consider:
   - Adding a "Details" expandable section for technical users (if supported by runtime)
   - Using consistent formatting and tone
   - Including relevant context (what operation failed, what the user can do)

## Implementation Approach:
- Modify the MessageDialog calls in the identified locations
- Keep the same dialog type (ErrorDialog) but improve the message content
- For paste failures, add a new MessageDialog call in ApplyRewriteAndPaste

## Files to Modify:
- main.go: initWindowsComponents function (hotkey registration)
- main.go: SaveSettings function (hotkey registration after settings change)
- main.go: onHotkeyTriggered function (clipboard restoration retry)
- main.go: ApplyRewriteAndPaste function (add paste failure dialog)

## Testing:
- Verify that error dialogs appear correctly when simulated failures occur
- Ensure messages are clear and actionable
- Confirm that existing functionality remains intact