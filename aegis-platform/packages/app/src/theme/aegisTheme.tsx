import React, { PropsWithChildren } from 'react';
import { CssBaseline, ThemeProvider } from '@material-ui/core';
import { createTheme } from '@material-ui/core/styles';
import type { BackstageTheme } from '@backstage/theme';

const baseTheme = createTheme({
  palette: {
    type: 'dark',
    primary: {
      main: '#10a37f',
    },
    secondary: {
      main: '#5b6ef5',
    },
    error: {
      main: '#f16063',
    },
    warning: {
      main: '#f4b740',
    },
    success: {
      main: '#3dd598',
    },
    background: {
      default: '#050b13',
      paper: 'rgba(11, 23, 42, 0.85)',
    },
    text: {
      primary: '#f7f7f8',
      secondary: '#b6c2d9',
      hint: '#6b7a99',
    },
    divider: 'rgba(91, 112, 146, 0.3)',
    navigation: {
      background: '#050b13',
      indicator: '#10a37f',
      color: '#b6c2d9',
      selectedColor: '#f7f7f8',
      navItem: {
        hoverBackground: 'rgba(16, 163, 127, 0.12)',
      },
      submenu: {
        background: 'rgba(7, 15, 26, 0.95)',
      },
    },
    banner: {
      info: '#14243c',
      error: '#341c1c',
      text: '#f7f7f8',
      link: '#10a37f',
    },
    errorBackground: 'rgba(241, 96, 99, 0.12)',
    warningBackground: 'rgba(244, 183, 64, 0.12)',
    infoBackground: 'rgba(64, 156, 255, 0.12)',
  },
  typography: {
    fontFamily:
      'Inter, "SF Pro Display", "Segoe UI", "Helvetica Neue", Arial, sans-serif',
    h1: {
      fontWeight: 600,
      fontSize: '2.5rem',
      letterSpacing: '-0.02em',
    },
    h2: {
      fontWeight: 600,
      fontSize: '2rem',
      letterSpacing: '-0.015em',
    },
    h3: {
      fontWeight: 600,
      fontSize: '1.75rem',
    },
    subtitle1: {
      color: '#b6c2d9',
    },
    body1: {
      color: '#d6dde9',
      fontSize: '0.95rem',
      lineHeight: 1.6,
    },
    button: {
      textTransform: 'none',
      fontWeight: 500,
      letterSpacing: '0.01em',
    },
  },
  shape: {
    borderRadius: 12,
  },
  overrides: {
    MuiCssBaseline: {
      '@global': {
        body: {
          background: 'radial-gradient(circle at 0% 0%, rgba(25,43,68,0.8), transparent 45%),\n            radial-gradient(circle at 100% 0%, rgba(12,45,64,0.65), transparent 40%),\n            linear-gradient(140deg, #050b13 0%, #0d192c 50%, #050b13 100%)',
          color: '#f7f7f8',
          minHeight: '100vh',
        },
        a: {
          color: '#10a37f',
          transition: 'color 150ms ease',
        },
        'a:hover': {
          color: '#5b6ef5',
        },
      },
    },
    MuiPaper: {
      root: {
        backgroundColor: 'rgba(11, 23, 42, 0.85)',
        backdropFilter: 'blur(18px)',
        border: '1px solid rgba(91, 112, 146, 0.15)',
        boxShadow:
          '0 20px 45px rgba(5, 11, 19, 0.45), 0 1px 0 rgba(255, 255, 255, 0.03) inset',
      },
    },
    MuiButton: {
      root: {
        borderRadius: 999,
        padding: '10px 18px',
        fontWeight: 600,
      },
      containedPrimary: {
        background: 'linear-gradient(135deg, #10a37f 0%, #5b6ef5 100%)',
        boxShadow: '0 10px 25px rgba(16, 163, 127, 0.25)',
        '&:hover': {
          background: 'linear-gradient(135deg, #0c8a69 0%, #4a59d8 100%)',
          boxShadow: '0 14px 32px rgba(16, 163, 127, 0.35)',
        },
      },
      outlined: {
        borderColor: 'rgba(91, 112, 146, 0.4)',
        '&:hover': {
          borderColor: '#10a37f',
          backgroundColor: 'rgba(16, 163, 127, 0.08)',
        },
      },
    },
    MuiAppBar: {
      colorDefault: {
        backgroundColor: 'rgba(7, 15, 26, 0.8)',
        backdropFilter: 'blur(18px)',
        borderBottom: '1px solid rgba(91, 112, 146, 0.2)',
      },
    },
    MuiCard: {
      root: {
        borderRadius: 18,
        border: '1px solid rgba(91, 112, 146, 0.15)',
        boxShadow: '0 24px 60px rgba(5, 11, 19, 0.35)',
      },
    },
    MuiListItem: {
      button: {
        borderRadius: 12,
      },
    },
    MuiTabs: {
      indicator: {
        height: 3,
        borderRadius: 3,
        backgroundImage:
          'linear-gradient(90deg, rgba(16,163,127,1) 0%, rgba(91,110,245,1) 100%)',
      },
    },
    MuiChip: {
      root: {
        backgroundColor: 'rgba(91, 112, 146, 0.18)',
        color: '#f7f7f8',
        borderRadius: 999,
      },
    },
  },
}) as BackstageTheme;

baseTheme.getPageTheme = ({ themeId }: { themeId: string }) => {
  const pageThemes: Record<string, any> = {
    home: {
      colors: ['#10a37f', '#5b6ef5'],
      shape:
        'linear-gradient(135deg, rgba(16,163,127,0.4), rgba(91,110,245,0.3))',
      backgroundImage:
        'radial-gradient(circle at 10% 20%, rgba(16,163,127,0.25), transparent 55%),\n         radial-gradient(circle at 80% 10%, rgba(91,110,245,0.22), transparent 65%)',
      fontColor: '#f7f7f8',
    },
    documentation: {
      colors: ['#5b6ef5', '#8f9bff'],
      shape: 'linear-gradient(135deg, rgba(91,110,245,0.45), rgba(143,155,255,0.35))',
      backgroundImage:
        'radial-gradient(circle at 20% 20%, rgba(91,110,245,0.2), transparent 60%)',
      fontColor: '#f7f7f8',
    },
    tool: {
      colors: ['#10a37f', '#3dd598'],
      shape: 'linear-gradient(135deg, rgba(16,163,127,0.4), rgba(61,213,152,0.3))',
      backgroundImage:
        'radial-gradient(circle at 15% 30%, rgba(16,163,127,0.25), transparent 55%)',
      fontColor: '#f7f7f8',
    },
    service: {
      colors: ['#5b6ef5', '#10a37f'],
      shape: 'linear-gradient(135deg, rgba(91,110,245,0.4), rgba(16,163,127,0.35))',
      backgroundImage:
        'radial-gradient(circle at 70% 25%, rgba(91,110,245,0.25), transparent 60%)',
      fontColor: '#f7f7f8',
    },
    website: {
      colors: ['#8f9bff', '#5b6ef5'],
      shape: 'linear-gradient(120deg, rgba(143,155,255,0.4), rgba(91,110,245,0.3))',
      backgroundImage:
        'radial-gradient(circle at 75% 15%, rgba(143,155,255,0.25), transparent 65%)',
      fontColor: '#f7f7f8',
    },
    library: {
      colors: ['#3dd598', '#10a37f'],
      shape: 'linear-gradient(110deg, rgba(61,213,152,0.4), rgba(16,163,127,0.35))',
      backgroundImage:
        'radial-gradient(circle at 15% 70%, rgba(61,213,152,0.25), transparent 60%)',
      fontColor: '#f7f7f8',
    },
    app: {
      colors: ['#10a37f', '#5b6ef5'],
      shape: 'linear-gradient(115deg, rgba(16,163,127,0.4), rgba(91,110,245,0.4))',
      backgroundImage:
        'radial-gradient(circle at 85% 70%, rgba(16,163,127,0.25), transparent 60%)',
      fontColor: '#f7f7f8',
    },
    apis: {
      colors: ['#5b6ef5', '#10a37f'],
      shape: 'linear-gradient(125deg, rgba(91,110,245,0.4), rgba(16,163,127,0.35))',
      backgroundImage:
        'radial-gradient(circle at 40% 60%, rgba(91,110,245,0.25), transparent 55%)',
      fontColor: '#f7f7f8',
    },
    other: {
      colors: ['#5b6ef5', '#10a37f'],
      shape: 'linear-gradient(135deg, rgba(91,110,245,0.3), rgba(16,163,127,0.3))',
      backgroundImage:
        'radial-gradient(circle at 50% 50%, rgba(91,110,245,0.2), transparent 60%)',
      fontColor: '#f7f7f8',
    },
  };

  return pageThemes[themeId] || pageThemes.other;
};

export const aegisDarkTheme = {
  id: 'aegis-dark',
  title: 'ÆGIS Dark',
  variant: 'dark' as const,
  Provider: ({ children }: PropsWithChildren<{}>) => (
    <ThemeProvider theme={baseTheme}>
      <CssBaseline />
      <div
        style={{
          minHeight: '100vh',
          background:
            'radial-gradient(circle at 0% 0%, rgba(25, 43, 68, 0.6), transparent 45%),\n            radial-gradient(circle at 100% 0%, rgba(12, 45, 64, 0.45), transparent 45%),\n            linear-gradient(140deg, rgba(5, 11, 19, 0.96) 0%, rgba(11, 23, 42, 0.92) 55%, rgba(5, 11, 19, 0.96) 100%)',
        }}
      >
        {children}
      </div>
    </ThemeProvider>
  ),
  theme: baseTheme,
};
