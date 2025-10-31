import React, { PropsWithChildren } from 'react';
import { CssBaseline, ThemeProvider } from '@material-ui/core';
import { createTheme } from '@material-ui/core/styles';
import type { BackstageTheme } from '@backstage/theme';

export type AegisThemeId = 'aegis-dark' | 'aegis-light';

type Mode = 'dark' | 'light';

type ThemeConfig = {
  id: AegisThemeId;
  title: string;
  mode: Mode;
  description: string;
};

const typography = {
  fontFamily: "'Inter', 'IBM Plex Sans', 'Helvetica Neue', Arial, sans-serif",
  fontWeightLight: 300,
  fontWeightRegular: 400,
  fontWeightMedium: 500,
  fontWeightBold: 600,
  h1: {
    fontWeight: 600,
    fontSize: '2.5rem',
    letterSpacing: '-0.03em',
  },
  h2: {
    fontWeight: 600,
    fontSize: '2rem',
    letterSpacing: '-0.02em',
  },
  h3: {
    fontWeight: 600,
    fontSize: '1.75rem',
    letterSpacing: '-0.01em',
  },
  h4: {
    fontWeight: 600,
    fontSize: '1.5rem',
  },
  h5: {
    fontWeight: 500,
    fontSize: '1.25rem',
  },
  h6: {
    fontWeight: 500,
    fontSize: '1.125rem',
  },
  subtitle1: {
    fontWeight: 500,
    fontSize: '1rem',
    letterSpacing: '-0.005em',
  },
  subtitle2: {
    fontWeight: 500,
    fontSize: '0.875rem',
    letterSpacing: '0.02em',
    textTransform: 'uppercase',
  },
  body1: {
    fontSize: '1rem',
    lineHeight: 1.65,
  },
  body2: {
    fontSize: '0.875rem',
    lineHeight: 1.6,
  },
  button: {
    fontWeight: 600,
    textTransform: 'none',
    letterSpacing: '0.02em',
  },
  caption: {
    fontSize: '0.75rem',
    letterSpacing: '0.06em',
    textTransform: 'uppercase',
  },
} as const;

const radii = {
  sm: 8,
  card: 12,
  panel: 16,
} as const;

const buildPalette = (mode: Mode) => {
  const isDark = mode === 'dark';

  return {
    type: mode,
    mode,
    status: {
      ok: '#10b981',
      warning: '#f59e0b',
      error: '#ef4444',
      running: '#38bdf8',
      pending: '#facc15',
      aborted: '#9ca3af',
    },
    bursts: {
      fontColor: isDark ? '#EDEDED' : '#111827',
      slackChannelText: isDark ? '#E0E7FF' : '#312E81',
      backgroundColor: {
        default: isDark ? '#0F1117' : '#EEF2FF',
      },
      gradient: {
        linear: isDark
          ? 'linear-gradient(135deg, rgba(155,135,255,0.65), rgba(34,211,238,0.45))'
          : 'linear-gradient(135deg, rgba(79,70,229,0.55), rgba(14,165,233,0.45))',
      },
    },
    primary: {
      main: isDark ? '#9B87FF' : '#4F46E5',
      light: isDark ? '#BAAFFF' : '#818CF8',
      dark: isDark ? '#6C5DD3' : '#3730A3',
      contrastText: isDark ? '#0F0F10' : '#FFFFFF',
    },
    secondary: {
      main: isDark ? '#22D3EE' : '#0EA5E9',
      light: isDark ? '#67E8F9' : '#38BDF8',
      dark: isDark ? '#0E7490' : '#0369A1',
      contrastText: '#0F172A',
    },
    error: {
      main: '#f87171',
    },
    warning: {
      main: '#f59e0b',
    },
    success: {
      main: '#22c55e',
    },
    info: {
      main: '#38bdf8',
    },
    background: {
      default: isDark ? '#0E0E0E' : '#FBFAF7',
      paper: isDark ? '#1F1F1F' : '#FFFFFF',
    },
    text: {
      primary: isDark ? '#EEEDEE' : '#0B0B0B',
      secondary: isDark ? '#A7A7A7' : '#525252',
      disabled: isDark ? 'rgba(255,255,255,0.32)' : 'rgba(15,15,15,0.32)',
      hint: isDark ? '#8E8E8E' : '#737373',
    },
    divider: isDark ? '#242424' : '#E7E7E7',
    action: {
      hover: isDark ? 'rgba(255,255,255,0.04)' : 'rgba(15,15,15,0.04)',
      selected: isDark ? 'rgba(155,135,255,0.16)' : 'rgba(79,70,229,0.12)',
      focus: isDark ? 'rgba(155,135,255,0.28)' : 'rgba(79,70,229,0.24)',
      active: isDark ? 'rgba(255,255,255,0.64)' : 'rgba(15,15,15,0.72)',
      disabled: isDark ? 'rgba(255,255,255,0.3)' : 'rgba(15,15,15,0.3)',
      disabledBackground: isDark
        ? 'rgba(255,255,255,0.08)'
        : 'rgba(15,15,15,0.06)',
    },
    navigation: {
      background: isDark ? '#131313' : '#EFECE3',
      indicator: isDark ? '#9B87FF' : '#4F46E5',
      color: isDark ? '#D6D6D6' : '#3B3B3B',
      selectedColor: isDark ? '#FFFFFF' : '#111827',
      navItem: {
        hoverBackground: isDark ? 'rgba(255,255,255,0.06)' : 'rgba(79,70,229,0.08)',
      },
      submenu: {
        background: isDark ? '#191919' : '#F8F7F1',
      },
    },
    banner: {
      info: isDark ? '#1E293B' : '#DBEAFE',
      error: isDark ? '#4C0519' : '#FEE2E2',
      text: isDark ? '#E2E8F0' : '#111827',
      link: '#38bdf8',
      warning: '#f59e0b',
      closeButtonColor: isDark ? '#F8FAFC' : '#0F172A',
    },
    link: isDark ? '#67E8F9' : '#2563EB',
    linkHover: isDark ? '#38BDF8' : '#1D4ED8',
    errorText: isDark ? '#FECACA' : '#B91C1C',
    infoText: isDark ? '#BAE6FD' : '#0F172A',
    warningText: isDark ? '#FDE68A' : '#854D0E',
    gold: '#facc15',
    errorBackground: isDark ? 'rgba(248,113,113,0.12)' : 'rgba(248,113,113,0.18)',
    warningBackground: isDark ? 'rgba(245,158,11,0.12)' : 'rgba(245,158,11,0.18)',
    infoBackground: isDark ? 'rgba(56,189,248,0.12)' : 'rgba(56,189,248,0.16)',
    pinSidebarButton: {
      icon: isDark ? '#0F172A' : '#F9FAFB',
      background: isDark ? 'rgba(255,255,255,0.16)' : 'rgba(79,70,229,0.12)',
    },
    tabbar: {
      indicator: isDark ? '#9B87FF' : '#4F46E5',
    },
  } as const;
};

const createOverrides = (theme: BackstageTheme, mode: Mode) => {
  const isDark = mode === 'dark';
  const focusContrast = isDark ? '#0E0E0E' : '#FFFFFF';
  const focusColor = isDark ? 'rgba(155,135,255,0.65)' : 'rgba(79,70,229,0.55)';
  const focusShadow = `0 0 0 2px ${focusContrast}, 0 0 0 4px ${focusColor}`;
  const cardBorder = isDark
    ? 'rgba(255,255,255,0.08)'
    : 'rgba(15,23,42,0.08)';
  const panelBorder = isDark
    ? 'rgba(255,255,255,0.12)'
    : 'rgba(15,23,42,0.1)';
  const cardShadow = isDark
    ? '0 20px 48px rgba(0,0,0,0.45)'
    : '0 14px 32px rgba(15,23,42,0.14)';
  const panelShadow = isDark
    ? '0 30px 60px rgba(0,0,0,0.5)'
    : '0 24px 48px rgba(15,23,42,0.16)';

  return {
    MuiCssBaseline: {
      '@global': {
        body: {
          backgroundColor: theme.palette.background.default,
          color: theme.palette.text.primary,
          fontFeatureSettings: '"cv02","cv03","cv04","cv11"',
          letterSpacing: '-0.01em',
        },
        a: {
          color: theme.palette.link,
        },
        '*::-webkit-scrollbar': {
          width: 10,
          height: 10,
        },
        '*::-webkit-scrollbar-thumb': {
          backgroundColor: isDark
            ? 'rgba(255,255,255,0.18)'
            : 'rgba(15,23,42,0.22)',
          borderRadius: 999,
        },
        '*::-webkit-scrollbar-track': {
          backgroundColor: 'transparent',
        },
      },
    },
    MuiPaper: {
      rounded: {
        borderRadius: radii.panel,
      },
      elevation1: {
        borderRadius: radii.panel,
        backgroundColor: theme.palette.background.paper,
        border: `1px solid ${panelBorder}`,
        boxShadow: panelShadow,
      },
    },
    MuiCard: {
      root: {
        borderRadius: radii.card,
        border: `1px solid ${cardBorder}`,
        boxShadow: cardShadow,
        backgroundColor: theme.palette.background.paper,
        padding: theme.spacing(0),
      },
    },
    MuiCardContent: {
      root: {
        padding: theme.spacing(6),
        '&:last-child': {
          paddingBottom: theme.spacing(6),
        },
      },
    },
    MuiButton: {
      root: {
        borderRadius: radii.sm,
        padding: theme.spacing(1.5, 3),
        fontWeight: 600,
        letterSpacing: '0.04em',
        '&:focus-visible': {
          boxShadow: focusShadow,
        },
      },
      outlined: {
        borderWidth: 1,
        '&:hover': {
          borderWidth: 1,
        },
        '&:focus-visible': {
          boxShadow: focusShadow,
        },
      },
      containedPrimary: {
        boxShadow: isDark
          ? '0 16px 32px rgba(155,135,255,0.32)'
          : '0 14px 30px rgba(79,70,229,0.28)',
        '&:hover': {
          boxShadow: isDark
            ? '0 20px 36px rgba(155,135,255,0.4)'
            : '0 16px 32px rgba(79,70,229,0.32)',
        },
      },
      text: {
        '&:focus-visible': {
          boxShadow: focusShadow,
        },
      },
    },
    MuiIconButton: {
      root: {
        borderRadius: radii.card,
        '&:focus-visible': {
          boxShadow: focusShadow,
        },
      },
    },
    MuiOutlinedInput: {
      root: {
        borderRadius: radii.sm,
        transition: 'box-shadow 150ms ease, border-color 150ms ease',
        '&.Mui-focused .MuiOutlinedInput-notchedOutline': {
          borderColor: theme.palette.primary.main,
          boxShadow: focusShadow,
        },
      },
    },
    MuiInputBase: {
      input: {
        fontSize: '0.95rem',
      },
    },
    MuiFormLabel: {
      root: {
        fontWeight: 500,
      },
    },
    MuiStepper: {
      root: {
        padding: theme.spacing(0),
        background: 'transparent',
      },
    },
    MuiStepConnector: {
      line: {
        borderColor: isDark ? '#2F2F2F' : '#D4D4D8',
      },
    },
    MuiStepLabel: {
      label: {
        fontWeight: 500,
        color: theme.palette.text.secondary,
        '&.MuiStepLabel-active': {
          color: theme.palette.text.primary,
        },
        '&.MuiStepLabel-completed': {
          color: theme.palette.text.primary,
        },
      },
    },
    MuiStepIcon: {
      root: {
        color: isDark ? '#3F3F46' : '#D4D4D8',
        '&.MuiStepIcon-active': {
          color: theme.palette.primary.main,
        },
        '&.MuiStepIcon-completed': {
          color: theme.palette.primary.main,
        },
      },
    },
    MuiDrawer: {
      paper: {
        backgroundColor: theme.palette.navigation.background,
        borderRight: `1px solid ${cardBorder}`,
      },
    },
    MuiListItem: {
      root: {
        borderRadius: radii.card,
      },
    },
    MuiDivider: {
      root: {
        backgroundColor: theme.palette.divider,
        opacity: 1,
      },
    },
    MuiTooltip: {
      tooltip: {
        borderRadius: radii.sm,
        fontSize: '0.75rem',
        padding: theme.spacing(1, 1.5),
      },
    },
    MuiDialog: {
      paper: {
        borderRadius: radii.panel,
        border: `1px solid ${panelBorder}`,
      },
    },
  } as const;
};

const createProps = () => ({
  MuiButton: {
    disableElevation: true,
  },
  MuiTooltip: {
    arrow: true,
  },
}) as const;

const pageThemes = {
  home: {
    colors: ['#9B87FF', '#22D3EE'],
    shape:
      'radial-gradient(circle at 20% 20%, rgba(155,135,255,0.18), transparent 60%), radial-gradient(circle at 80% 15%, rgba(34,211,238,0.18), transparent 60%)',
    backgroundImage:
      'radial-gradient(circle at 10% 30%, rgba(155,135,255,0.24), transparent 65%), radial-gradient(circle at 80% 70%, rgba(34,211,238,0.2), transparent 65%)',
    fontColor: '#F4F4F5',
  },
  documentation: {
    colors: ['#4F46E5', '#0EA5E9'],
    shape:
      'radial-gradient(circle at 35% 25%, rgba(79,70,229,0.18), transparent 65%), radial-gradient(circle at 75% 60%, rgba(14,165,233,0.18), transparent 65%)',
    backgroundImage:
      'radial-gradient(circle at 15% 15%, rgba(79,70,229,0.12), transparent 60%), radial-gradient(circle at 80% 45%, rgba(14,165,233,0.12), transparent 60%)',
    fontColor: '#0B0B0B',
  },
  tool: {
    colors: ['#22c55e', '#0ea5e9'],
    shape:
      'radial-gradient(circle at 30% 20%, rgba(34,197,94,0.16), transparent 60%), radial-gradient(circle at 80% 70%, rgba(14,165,233,0.16), transparent 60%)',
    backgroundImage:
      'radial-gradient(circle at 15% 50%, rgba(34,197,94,0.2), transparent 65%), radial-gradient(circle at 75% 40%, rgba(14,165,233,0.2), transparent 65%)',
    fontColor: '#0F172A',
  },
  service: {
    colors: ['#f59e0b', '#6366f1'],
    shape:
      'radial-gradient(circle at 30% 30%, rgba(245,158,11,0.18), transparent 60%), radial-gradient(circle at 75% 70%, rgba(99,102,241,0.18), transparent 60%)',
    backgroundImage:
      'radial-gradient(circle at 20% 20%, rgba(245,158,11,0.12), transparent 55%), radial-gradient(circle at 80% 60%, rgba(99,102,241,0.12), transparent 60%)',
    fontColor: '#111827',
  },
  website: {
    colors: ['#9B87FF', '#38BDF8'],
    shape:
      'radial-gradient(circle at 35% 30%, rgba(155,135,255,0.16), transparent 60%), radial-gradient(circle at 70% 60%, rgba(56,189,248,0.16), transparent 60%)',
    backgroundImage:
      'radial-gradient(circle at 25% 25%, rgba(155,135,255,0.12), transparent 55%), radial-gradient(circle at 80% 55%, rgba(56,189,248,0.12), transparent 55%)',
    fontColor: '#0B1120',
  },
  library: {
    colors: ['#6366F1', '#14B8A6'],
    shape:
      'radial-gradient(circle at 20% 40%, rgba(99,102,241,0.18), transparent 65%), radial-gradient(circle at 70% 60%, rgba(20,184,166,0.18), transparent 65%)',
    backgroundImage:
      'radial-gradient(circle at 30% 25%, rgba(99,102,241,0.12), transparent 60%), radial-gradient(circle at 80% 65%, rgba(20,184,166,0.12), transparent 60%)',
    fontColor: '#0B1120',
  },
  apis: {
    colors: ['#38BDF8', '#22C55E'],
    shape:
      'radial-gradient(circle at 20% 35%, rgba(56,189,248,0.18), transparent 60%), radial-gradient(circle at 70% 65%, rgba(34,197,94,0.18), transparent 60%)',
    backgroundImage:
      'radial-gradient(circle at 25% 25%, rgba(56,189,248,0.12), transparent 55%), radial-gradient(circle at 80% 60%, rgba(34,197,94,0.12), transparent 55%)',
    fontColor: '#0B1120',
  },
  other: {
    colors: ['#4F46E5', '#22D3EE'],
    shape:
      'radial-gradient(circle at 35% 20%, rgba(79,70,229,0.18), transparent 65%), radial-gradient(circle at 80% 65%, rgba(34,211,238,0.18), transparent 65%)',
    backgroundImage:
      'radial-gradient(circle at 20% 35%, rgba(79,70,229,0.12), transparent 60%), radial-gradient(circle at 75% 55%, rgba(34,211,238,0.12), transparent 60%)',
    fontColor: '#0B1120',
  },
} as const;

const buildTheme = (config: ThemeConfig): BackstageTheme => {
  const palette = buildPalette(config.mode);

  const theme = createTheme({
    palette: palette as any,
    typography: typography as any,
    shape: {
      borderRadius: radii.card,
    },
    spacing: 4,
    overrides: {},
    props: createProps() as any,
  }) as BackstageTheme;

  theme.overrides = {
    ...theme.overrides,
    ...(createOverrides(theme, config.mode) as any),
  };

  theme.zIndex.drawer = 1200;

  theme.defaultPageTheme = 'home';
  theme.getPageTheme = ({ themeId }: { themeId: string }) =>
    (pageThemes as Record<string, any>)[themeId] || pageThemes.other;

  return theme;
};

const createThemeProvider = (
  config: ThemeConfig,
): { Provider: React.ComponentType<PropsWithChildren<{}>>; theme: BackstageTheme } => {
  const theme = buildTheme(config);

  const Provider = ({ children }: PropsWithChildren<{}>) => (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      {children}
    </ThemeProvider>
  );

  return { Provider, theme };
};

const darkConfig: ThemeConfig = {
  id: 'aegis-dark',
  title: 'ÆGIS Nightfall',
  mode: 'dark',
  description: 'Primary dark theme tuned for control plane work.',
};

const lightConfig: ThemeConfig = {
  id: 'aegis-light',
  title: 'ÆGIS Dawn',
  mode: 'light',
  description: 'Bright theme for daylight and accessibility needs.',
};

const dark = createThemeProvider(darkConfig);
const light = createThemeProvider(lightConfig);

export const aegisDarkTheme = {
  id: darkConfig.id,
  title: darkConfig.title,
  variant: 'dark' as const,
  Provider: dark.Provider,
  theme: dark.theme,
};

export const aegisLightTheme = {
  id: lightConfig.id,
  title: lightConfig.title,
  variant: 'light' as const,
  Provider: light.Provider,
  theme: light.theme,
};

export const aegisThemes = [aegisDarkTheme, aegisLightTheme];

export default aegisDarkTheme;
