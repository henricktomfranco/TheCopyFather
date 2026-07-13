import React, { Component, ErrorInfo, ReactNode } from 'react';

interface Props {
  children: ReactNode;
}

interface State {
  hasError: boolean;
  error: Error | null;
}

class ErrorBoundary extends Component<Props, State> {
  public state: State = {
    hasError: false,
    error: null
  };

  public static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error };
  }

  public componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error('Uncaught error:', error, errorInfo);
  }

  public render() {
    if (this.state.hasError) {
      return (
        <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', height: '100vh', padding: '20px', textAlign: 'center', backgroundColor: '#1e1e1e', color: '#e0e0e0', fontFamily: 'system-ui, -apple-system, sans-serif' }}>
          <div style={{ padding: '24px', borderRadius: '8px', backgroundColor: '#2d2d2d', border: '1px solid #ff4444', maxWidth: '400px' }}>
            <span style={{ fontSize: '48px', display: 'block', marginBottom: '16px' }}>⚠️</span>
            <h2 style={{ marginTop: 0, color: '#ff4444' }}>Something went wrong.</h2>
            <p style={{ marginBottom: '24px', color: '#a0a0a0' }}>{this.state.error?.message || 'An unexpected error occurred in the UI.'}</p>
            <button
              style={{ padding: '8px 16px', backgroundColor: '#333', color: '#fff', border: '1px solid #555', borderRadius: '4px', cursor: 'pointer' }}
              onClick={() => window.location.reload()}
            >
              Reload Application
            </button>
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}

export default ErrorBoundary;
