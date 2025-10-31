import React, { PropsWithChildren } from 'react';
import { CssBaseline, ThemeProvider } from '@material-ui/core';
import { createTheme } from '@material-ui/core/styles';
import type { BackstageTheme } from '@backstage/theme';

const palette = {
  type: 'dark',
  mode: 'dark',
  status: {
    ok: '#10b981',
    warning: '#f59e0b',
    error: '#f43f5e',
    running: '#38bdf8',
    pending: '#facc15',
    aborted: '#64748b',
  },
  bursts: {
    fontColor: '#f8fafc',
    slackChannelText: '#cbd5f5',
    backgroundColor: {
      default: '#0f172a',
    },
    gradient: {
      linear:
        'linear-gradient(135deg, rgba(16,185,129,0.6), rgba(14,165,233,0.6))',
    },
  },
  primary: {
    main: '#10b981',
    light: '#34d399',
    dark: '#0f766e',
  },
  secondary: {
    main: '#6366f1',
    light: '#8b5cf6',
    dark: '#4338ca',
  },
  error: {
    main: '#f43f5e',
  },
  warning: {
    main: '#f59e0b',
  },
  success: {
    main: '#22c55e',
  },
  background: {
    default: '#040711',
    paper: '#0b1220',
  },
  border: 'rgba(148, 163, 184, 0.18)',
  text: {
    primary: '#f8fafc',
    secondary: '#94a3b8',
    hint: '#64748b',
  },
  textContrast: '#0f172a',
  textVerySubtle: '#475569',
  textSubtle: '#94a3b8',
  highlight: 'rgba(14, 165, 233, 0.18)',
  divider: 'rgba(148, 163, 184, 0.12)',
  navigation: {
    background: '#050a16',
    indicator: '#10b981',
    color: '#cbd5f5',
    selectedColor: '#f8fafc',
    navItem: {
      hoverBackground: 'rgba(16, 185, 129, 0.08)',
    },
    submenu: {
      background: '#0d1629',
    },
  },
  banner: {
    info: '#1e293b',
    error: '#4c0519',
    text: '#e2e8f0',
    link: '#38bdf8',
    warning: '#f59e0b',
    closeButtonColor: '#f8fafc',
  },
  link: '#38bdf8',
  linkHover: '#0ea5e9',
  errorText: '#fecdd3',
  infoText: '#bae6fd',
  warningText: '#fef08a',
  gold: '#facc15',
  errorBackground: 'rgba(244, 63, 94, 0.12)',
  warningBackground: 'rgba(245, 158, 11, 0.12)',
  infoBackground: 'rgba(56, 189, 248, 0.12)',
  pinSidebarButton: {
    icon: '#0f172a',
    background: 'rgba(148, 163, 184, 0.45)',
  },
  tabbar: {
    indicator: '#10b981',
  },
} as const;

const typography = {
  fontFamily: "'Inter', 'SF Pro Display', 'Helvetica Neue', Arial, sans-serif",
  fontWeightLight: 300,
  fontWeightRegular: 400,
  fontWeightMedium: 500,
  fontWeightBold: 700,
  h1: {
    fontWeight: 600,
    fontSize: '2.75rem',
    letterSpacing: '-0.03em',
  },
  h2: {
    fontWeight: 600,
    fontSize: '2.25rem',
    letterSpacing: '-0.025em',
  },
  h3: {
    fontWeight: 600,
    fontSize: '1.875rem',
    letterSpacing: '-0.02em',
  },
  h4: {
    fontWeight: 600,
    fontSize: '1.5rem',
  },
  h5: {
    fontWeight: 500,
    fontSize: '1.25rem',
  },
  body1: {
    fontSize: '1rem',
    lineHeight: 1.7,
  },
  body2: {
    fontSize: '0.875rem',
    lineHeight: 1.6,
  },
  subtitle1: {
    fontWeight: 500,
    letterSpacing: '-0.01em',
  },
  button: {
    fontWeight: 600,
    letterSpacing: '-0.01em',
    textTransform: 'none',
  },
} as const;

const overrides = {
  MuiCssBaseline: {
    '@global': {
      body: {
        background:
          'radial-gradient(circle at 20% 20%, rgba(56, 189, 248, 0.12), transparent 55%), radial-gradient(circle at 80% 10%, rgba(129, 140, 248, 0.1), transparent 55%), #040711',
        color: '#f8fafc',
        fontFeatureSettings: '"cv02","cv03","cv04","cv11"',
        letterSpacing: '-0.01em',
      },
      a: {
        color: '#34d399',
      },
      '*::-webkit-scrollbar': {
        width: 8,
        height: 8,
      },
      '*::-webkit-scrollbar-thumb': {
        backgroundColor: 'rgba(148, 163, 184, 0.25)',
        borderRadius: 999,
      },
      '*::-webkit-scrollbar-track': {
        backgroundColor: 'transparent',
      },
    },
  },
  MuiPaper: {
    root: {
      backgroundColor: 'rgba(11, 18, 32, 0.9)',
      backdropFilter: 'blur(12px)',
      border: '1px solid rgba(148, 163, 184, 0.16)',
      borderRadius: 20,
      boxShadow:
        '0 20px 45px rgba(15, 23, 42, 0.35), inset 0 1px 0 rgba(255, 255, 255, 0.02)',
    },
  },
  MuiButton: {
    root: {
      borderRadius: 999,
      textTransform: 'none',
      fontWeight: 600,
      padding: '10px 20px',
      letterSpacing: '-0.01em',
    },
    containedPrimary: {
      background:
        'linear-gradient(135deg, rgba(16, 185, 129, 0.92), rgba(14, 165, 233, 0.92))',
      boxShadow: '0 12px 30px rgba(14, 165, 233, 0.25)',
      '&:hover': {
        background:
          'linear-gradient(135deg, rgba(16, 185, 129, 1), rgba(14, 165, 233, 1))',
        boxShadow: '0 16px 36px rgba(16, 185, 129, 0.25)',
      },
    },
    outlined: {
      borderColor: 'rgba(148, 163, 184, 0.32)',
      '&:hover': {
        borderColor: 'rgba(148, 163, 184, 0.6)',
        backgroundColor: 'rgba(148, 163, 184, 0.05)',
      },
    },
  },
  MuiDrawer: {
    paper: {
      backgroundColor: 'rgba(5, 10, 22, 0.85)',
      backdropFilter: 'blur(14px)',
      borderRight: '1px solid rgba(148, 163, 184, 0.18)',
    },
  },
  MuiListItem: {
    root: {
      borderRadius: 12,
      margin: '4px 12px',
    },
  },
  MuiCard: {
    root: {
      borderRadius: 20,
      border: '1px solid rgba(148, 163, 184, 0.14)',
      background:
        'linear-gradient(160deg, rgba(15,23,42,0.96) 0%, rgba(15,23,42,0.65) 100%)',
      boxShadow: '0 14px 30px rgba(8, 15, 31, 0.45)',
    },
  },
  MuiAppBar: {
    colorPrimary: {
      backgroundColor: 'rgba(4, 7, 17, 0.9)',
      backdropFilter: 'blur(16px)',
      borderBottom: '1px solid rgba(148, 163, 184, 0.12)',
    },
  },
  MuiChip: {
    root: {
      backgroundColor: 'rgba(148, 163, 184, 0.12)',
      color: '#94a3b8',
    },
    colorPrimary: {
      backgroundColor: 'rgba(16, 185, 129, 0.16)',
      color: '#10b981',
    },
    colorSecondary: {
      backgroundColor: 'rgba(99, 102, 241, 0.16)',
      color: '#6366f1',
    },
  },
} as const;

const props = {
  MuiButton: {
    disableElevation: true,
  },
} as const;

const baseTheme = createTheme({
  palette: palette as any,
  typography: typography as any,
  shape: {
    borderRadius: 18,
  },
  overrides: overrides as any,
  props: props as any,
}) as BackstageTheme;

baseTheme.defaultPageTheme = 'home';

baseTheme.getPageTheme = ({ themeId }: { themeId: string }) => {
  const pageThemes: Record<string, any> = {
    home: {
      colors: ['#10b981', '#0ea5e9'],
      shape:
        'linear-gradient(135deg, rgba(16, 185, 129, 0.25), rgba(14, 165, 233, 0.2))',
      backgroundImage:
        'radial-gradient(circle at 20% 20%, rgba(16, 185, 129, 0.15), transparent 55%), radial-gradient(circle at 80% 15%, rgba(14, 165, 233, 0.15), transparent 55%)',
      fontColor: '#f8fafc',
    },
    documentation: {
      colors: ['#38bdf8', '#6366f1'],
      shape:
        'linear-gradient(135deg, rgba(56, 189, 248, 0.22), rgba(99, 102, 241, 0.18))',
      backgroundImage:
        'radial-gradient(circle at 15% 10%, rgba(56, 189, 248, 0.18), transparent 55%), radial-gradient(circle at 70% 30%, rgba(99, 102, 241, 0.18), transparent 60%)',
      fontColor: '#f8fafc',
    },
    tool: {
      colors: ['#22c55e', '#0ea5e9'],
      shape:
        'linear-gradient(135deg, rgba(34, 197, 94, 0.25), rgba(14, 165, 233, 0.22))',
      backgroundImage:
        'radial-gradient(circle at 30% 40%, rgba(34, 197, 94, 0.2), transparent 60%), radial-gradient(circle at 80% 60%, rgba(14, 165, 233, 0.18), transparent 60%)',
      fontColor: '#f8fafc',
    },
    service: {
      colors: ['#f59e0b', '#6366f1'],
      shape:
        'linear-gradient(135deg, rgba(245, 158, 11, 0.24), rgba(99, 102, 241, 0.2))',
      backgroundImage:
        'radial-gradient(circle at 25% 30%, rgba(245, 158, 11, 0.18), transparent 60%), radial-gradient(circle at 75% 70%, rgba(99, 102, 241, 0.18), transparent 65%)',
      fontColor: '#f8fafc',
    },
    website: {
      colors: ['#10b981', '#64748b'],
      shape:
        'linear-gradient(135deg, rgba(16, 185, 129, 0.24), rgba(100, 116, 139, 0.18))',
      backgroundImage:
        'radial-gradient(circle at 35% 20%, rgba(16, 185, 129, 0.18), transparent 60%), radial-gradient(circle at 65% 70%, rgba(100, 116, 139, 0.2), transparent 65%)',
      fontColor: '#f8fafc',
    },
    library: {
      colors: ['#6366f1', '#14b8a6'],
      shape:
        'linear-gradient(135deg, rgba(99, 102, 241, 0.24), rgba(20, 184, 166, 0.22))',
      backgroundImage:
        'radial-gradient(circle at 25% 25%, rgba(99, 102, 241, 0.18), transparent 60%), radial-gradient(circle at 70% 65%, rgba(20, 184, 166, 0.18), transparent 65%)',
      fontColor: '#f8fafc',
    },
    other: {
      colors: ['#0ea5e9', '#6366f1'],
      shape:
        'linear-gradient(135deg, rgba(14, 165, 233, 0.24), rgba(99, 102, 241, 0.22))',
      backgroundImage:
        'radial-gradient(circle at 20% 30%, rgba(14, 165, 233, 0.18), transparent 60%), radial-gradient(circle at 75% 50%, rgba(99, 102, 241, 0.2), transparent 65%)',
      fontColor: '#f8fafc',
    },
    app: {
      colors: ['#22c55e', '#f97316'],
      shape:
        'linear-gradient(135deg, rgba(34, 197, 94, 0.24), rgba(249, 115, 22, 0.22))',
      backgroundImage:
        'radial-gradient(circle at 25% 25%, rgba(34, 197, 94, 0.2), transparent 60%), radial-gradient(circle at 70% 65%, rgba(249, 115, 22, 0.18), transparent 65%)',
      fontColor: '#f8fafc',
    },
    apis: {
      colors: ['#38bdf8', '#22c55e'],
      shape:
        'linear-gradient(135deg, rgba(56, 189, 248, 0.24), rgba(34, 197, 94, 0.22))',
      backgroundImage:
        'radial-gradient(circle at 30% 30%, rgba(56, 189, 248, 0.2), transparent 60%), radial-gradient(circle at 70% 60%, rgba(34, 197, 94, 0.18), transparent 65%)',
      fontColor: '#f8fafc',
    },
  };

  return pageThemes[themeId] || pageThemes.other;
};

export const aegisTheme = {
  id: 'aegis-dark',
  title: 'ÆGIS Flux',
  variant: 'dark' as const,
  Provider: ({ children }: PropsWithChildren<{}>) => (
    <ThemeProvider theme={baseTheme}>
      <CssBaseline />
      {children}
    </ThemeProvider>
  ),
  theme: baseTheme,
};

export default aegisTheme;
