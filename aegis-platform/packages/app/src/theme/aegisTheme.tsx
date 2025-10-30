import { PropsWithChildren } from 'react';
import { CssBaseline, ThemeProvider } from '@material-ui/core';
import { createTheme, Theme } from '@material-ui/core/styles';

const neonGradient =
  'linear-gradient(135deg, rgba(0, 245, 255, 0.1) 0%, rgba(79, 23, 255, 0.25) 45%, rgba(255, 45, 149, 0.2) 100%)';

const panelBackdrop =
  'linear-gradient(145deg, rgba(7, 21, 46, 0.95) 0%, rgba(11, 28, 64, 0.9) 45%, rgba(18, 10, 30, 0.9) 100%)';

const baseTheme: Theme = createTheme({
  palette: {
    type: 'dark',
    primary: {
      main: '#00f5ff',
      contrastText: '#04070f',
    },
    secondary: {
      main: '#ff2d95',
    },
    error: {
      main: '#ff3b5f',
    },
    warning: {
      main: '#ffb347',
    },
    success: {
      main: '#4cffa6',
    },
    background: {
      default: '#050912',
      paper: '#091428',
    },
    text: {
      primary: '#e5f4ff',
      secondary: '#7c9ac4',
      hint: '#557099',
    },
    divider: 'rgba(0, 245, 255, 0.2)',
  },
  typography: {
    fontFamily: "'Rajdhani', 'Roboto', sans-serif",
    h1: {
      fontFamily: "'Orbitron', 'Rajdhani', sans-serif",
      letterSpacing: '0.06em',
    },
    h2: {
      fontFamily: "'Orbitron', 'Rajdhani', sans-serif",
      letterSpacing: '0.04em',
    },
    h3: {
      fontFamily: "'Orbitron', 'Rajdhani', sans-serif",
      letterSpacing: '0.03em',
    },
    button: {
      textTransform: 'uppercase',
      letterSpacing: '0.08em',
      fontWeight: 600,
    },
    subtitle1: {
      letterSpacing: '0.04em',
    },
    subtitle2: {
      letterSpacing: '0.05em',
    },
    body1: {
      fontSize: '0.975rem',
    },
    body2: {
      fontSize: '0.875rem',
      letterSpacing: '0.02em',
    },
  },
  shape: {
    borderRadius: 16,
  },
  overrides: {
    MuiCssBaseline: {
      '@global': {
        body: {
          backgroundColor: '#050912',
          backgroundImage:
            'radial-gradient(circle at 20% 20%, rgba(0, 245, 255, 0.08) 0, rgba(5, 9, 18, 0) 60%), ' +
            'radial-gradient(circle at 80% 30%, rgba(255, 45, 149, 0.08) 0, rgba(5, 9, 18, 0) 55%)',
          minHeight: '100vh',
        },
        '#root': {
          minHeight: '100vh',
        },
        '*::-webkit-scrollbar': {
          width: 8,
          height: 8,
        },
        '*::-webkit-scrollbar-thumb': {
          background: 'rgba(0, 245, 255, 0.3)',
          borderRadius: 8,
        },
      },
    },
    MuiPaper: {
      root: {
        backgroundImage: panelBackdrop,
        border: '1px solid rgba(0, 245, 255, 0.12)',
        boxShadow:
          '0 0 20px rgba(0, 245, 255, 0.08), inset 0 0 30px rgba(255, 45, 149, 0.05)',
      },
    },
    MuiButton: {
      root: {
        borderRadius: 999,
        letterSpacing: '0.08em',
      },
      containedPrimary: {
        boxShadow: '0 0 20px rgba(0, 245, 255, 0.35)',
        '&:hover': {
          boxShadow: '0 0 28px rgba(0, 245, 255, 0.5)',
        },
      },
      outlinedPrimary: {
        borderColor: 'rgba(0, 245, 255, 0.6)',
        '&:hover': {
          borderColor: '#00f5ff',
          backgroundColor: 'rgba(0, 245, 255, 0.08)',
        },
      },
    },
    MuiCard: {
      root: {
        background: panelBackdrop,
        border: '1px solid rgba(0, 245, 255, 0.12)',
        boxShadow:
          '0 20px 60px rgba(3, 12, 35, 0.45), inset 0 0 25px rgba(0, 245, 255, 0.05)',
      },
    },
    MuiStepIcon: {
      root: {
        color: 'rgba(124, 154, 196, 0.4)',
        '&$active': {
          color: '#00f5ff',
          filter: 'drop-shadow(0 0 6px rgba(0, 245, 255, 0.7))',
        },
        '&$completed': {
          color: '#4cffa6',
          filter: 'drop-shadow(0 0 6px rgba(76, 255, 166, 0.6))',
        },
      },
    },
    MuiStepLabel: {
      label: {
        color: '#7c9ac4',
        textTransform: 'uppercase',
        letterSpacing: '0.12em',
        fontWeight: 600,
        '&$active': {
          color: '#e5f4ff',
        },
        '&$completed': {
          color: '#4cffa6',
        },
      },
    },
    MuiTableRow: {
      root: {
        transition: 'transform 150ms ease, box-shadow 200ms ease',
        '&:hover': {
          transform: 'translateY(-1px)',
          boxShadow: '0 14px 32px rgba(0, 245, 255, 0.08)',
        },
      },
    },
  },
});

export const aegisDarkTheme = {
  id: 'aegis-dark',
  title: 'ÆGIS Tactical',
  variant: 'dark' as const,
  Provider: ({ children }: PropsWithChildren<{}>) => (
    <ThemeProvider theme={baseTheme}>
      <CssBaseline />
      <div
        style={{
          minHeight: '100vh',
          backgroundImage: `${neonGradient}, radial-gradient(circle at 10% 90%, rgba(79, 23, 255, 0.15) 0%, rgba(5, 9, 18, 0) 65%)`,
          backgroundColor: '#050912',
        }}
      >
        {children}
      </div>
    </ThemeProvider>
  ),
};

export type AegisTheme = typeof aegisDarkTheme;
