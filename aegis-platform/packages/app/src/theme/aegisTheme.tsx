import React, { PropsWithChildren } from 'react';
import { CssBaseline, ThemeProvider } from '@material-ui/core';
import { createTheme } from '@material-ui/core/styles';
import { alpha } from '@material-ui/core/styles/colorManipulator';
import type { BackstageTheme } from '@backstage/theme';

type Mode = 'dark' | 'light';

const typography = {
  fontFamily:
    "'Inter', 'SF Pro Display', 'IBM Plex Sans', 'Segoe UI', sans-serif",
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
    fontSize: '0.915rem',
    lineHeight: 1.6,
  },
  subtitle1: {
    fontWeight: 600,
    letterSpacing: '-0.01em',
  },
  button: {
    fontWeight: 600,
    letterSpacing: '-0.01em',
    textTransform: 'none',
  },
} as const;

const createPalette = (mode: Mode) => {
  const isDark = mode === 'dark';
  const baseBackground = isDark ? '#0E0E0E' : '#F5F5F3';
  const paperBackground = isDark ? '#151516' : '#FFFFFF';
  const primaryMain = '#8B5CF6';
  const secondaryMain = '#22D3EE';
  const textPrimary = isDark ? '#F5F5F5' : '#161616';
  const textSecondary = isDark ? '#B3B3B9' : '#3C3C43';
  const divider = isDark ? 'rgba(255,255,255,0.08)' : 'rgba(18,18,23,0.08)';
  const navigationBackground = isDark ? '#090909' : '#F8F8F6';
  const hoverBackground = isDark
    ? 'rgba(139, 92, 246, 0.12)'
    : 'rgba(109, 40, 217, 0.08)';
  const submenuBackground = isDark ? '#151516' : '#F1F1EF';
  const bannerInfo = isDark ? '#1E2533' : '#E7ECFA';
  const bannerError = isDark ? '#371A24' : '#FCE8EB';
  const infoBackground = isDark ? 'rgba(96, 165, 250, 0.16)' : '#E0F2FE';
  const warningBackground = isDark
    ? 'rgba(250, 204, 21, 0.18)'
    : 'rgba(251, 191, 36, 0.2)';
  const errorBackground = isDark
    ? 'rgba(239, 68, 68, 0.18)'
    : 'rgba(248, 113, 113, 0.16)';

  return {
    type: mode,
    primary: {
      main: primaryMain,
      light: '#A78BFA',
      dark: '#6D28D9',
    },
    secondary: {
      main: secondaryMain,
      light: '#38E8FF',
      dark: '#0EA5E9',
    },
    error: {
      main: '#F87171',
    },
    warning: {
      main: '#F59E0B',
    },
    success: {
      main: '#22C55E',
    },
    background: {
      default: baseBackground,
      paper: paperBackground,
    },
    text: {
      primary: textPrimary,
      secondary: textSecondary,
      hint: isDark ? '#6B7280' : '#6D6D76',
    },
    divider,
    navigation: {
      background: navigationBackground,
      indicator: primaryMain,
      color: isDark ? '#C8C8D0' : '#4B4B52',
      selectedColor: textPrimary,
      navItem: {
        hoverBackground,
      },
      submenu: {
        background: submenuBackground,
      },
    },
    banner: {
      info: bannerInfo,
      error: bannerError,
      text: textPrimary,
      link: primaryMain,
    },
    link: primaryMain,
    linkHover: isDark ? '#C4B5FD' : '#7C3AED',
    errorText: isDark ? '#FCA5A5' : '#B91C1C',
    infoText: isDark ? '#BFDBFE' : '#1E3A8A',
    warningText: isDark ? '#FCE7AA' : '#854D0E',
    errorBackground,
    warningBackground,
    infoBackground,
    neutral: {
      main: isDark ? '#A1A1AA' : '#52525B',
    },
    navigationIndicator: primaryMain,
    tabbar: {
      indicator: primaryMain,
    },
    status: {
      ok: '#16A34A',
      warning: '#FACC15',
      error: '#EF4444',
      running: '#38BDF8',
      pending: '#F59E0B',
      aborted: '#6B7280',
    },
    bursts: {
      fontColor: textPrimary,
      slackChannelText: isDark ? '#F0F9FF' : '#0F172A',
      backgroundColor: {
        default: baseBackground,
      },
      gradient: {
        linear: isDark
          ? 'linear-gradient(135deg, rgba(139,92,246,0.28), rgba(34,211,238,0.16))'
          : 'linear-gradient(135deg, rgba(109,40,217,0.16), rgba(14,165,233,0.12))',
      },
    },
    pinSidebarButton: {
      icon: isDark ? '#111827' : '#F9FAFB',
      background: isDark ? 'rgba(148,163,184,0.45)' : 'rgba(17,24,39,0.28)',
    },
  } as BackstageTheme['palette'];
};

const createGlobalStyles = (theme: BackstageTheme, mode: Mode) => {
  const isDark = mode === 'dark';
  const subtle = isDark ? 'rgba(255,255,255,0.04)' : 'rgba(17,24,39,0.05)';
  const bodyGradient = isDark
    ? 'radial-gradient(circle at 10% 10%, rgba(139, 92, 246, 0.18), transparent 50%), radial-gradient(circle at 85% 15%, rgba(34, 211, 238, 0.18), transparent 55%), #0E0E0E'
    : 'radial-gradient(circle at 5% 10%, rgba(109, 40, 217, 0.12), transparent 45%), radial-gradient(circle at 90% 15%, rgba(14, 165, 233, 0.1), transparent 55%), #F5F5F3';

  return {
    ':root': {
      '--aegis-card-surface': theme.palette.background.paper,
      '--aegis-card-border': isDark ? 'rgba(255,255,255,0.08)' : 'rgba(15, 23, 42, 0.08)',
      '--aegis-card-shadow': isDark
        ? '0 18px 45px rgba(0, 0, 0, 0.5)'
        : '0 18px 45px rgba(15, 23, 42, 0.18)',
      '--aegis-muted': isDark ? 'rgba(255,255,255,0.08)' : 'rgba(15,23,42,0.06)',
    },
    body: {
      background: bodyGradient,
      color: theme.palette.text.primary,
      fontFamily: typography.fontFamily,
      letterSpacing: '-0.01em',
      minHeight: '100vh',
    },
    a: {
      color: theme.palette.link,
    },
    hr: {
      borderColor: subtle,
    },
  };
};

const createOverrides = (theme: BackstageTheme, mode: Mode) => {
  const isDark = mode === 'dark';
  const outline = alpha(theme.palette.primary.main, isDark ? 0.12 : 0.22);

  return {
    MuiCssBaseline: {
      '@global': createGlobalStyles(theme, mode),
    },
    MuiPaper: {
      rounded: {
        borderRadius: theme.shape.borderRadius,
      },
      elevation1: {
        backgroundColor: 'var(--aegis-card-surface)',
        border: '1px solid var(--aegis-card-border)',
        boxShadow: 'var(--aegis-card-shadow)',
      },
    },
    MuiCard: {
      root: {
        backgroundColor: 'var(--aegis-card-surface)',
        borderRadius: theme.shape.borderRadius,
        border: '1px solid var(--aegis-card-border)',
        boxShadow: '0 12px 40px rgba(0,0,0,0.22)',
        transition: 'transform 150ms ease, box-shadow 150ms ease, border-color 150ms ease',
        '&:hover': {
          boxShadow: '0 18px 50px rgba(15,23,42,0.28)',
          borderColor: outline,
        },
      },
    },
    MuiButton: {
      root: {
        borderRadius: 999,
      },
      containedPrimary: {
        backgroundImage: isDark
          ? 'linear-gradient(135deg, #8B5CF6, #22D3EE)'
          : 'linear-gradient(135deg, #6D28D9, #0EA5E9)',
        color: '#0B0B10',
        '&:hover': {
          backgroundImage: isDark
            ? 'linear-gradient(135deg, #A78BFA, #38E8FF)'
            : 'linear-gradient(135deg, #7C3AED, #22D3EE)',
        },
      },
      outlined: {
        borderColor: alpha(theme.palette.text.primary, 0.22),
        '&:hover': {
          borderColor: theme.palette.primary.main,
          backgroundColor: alpha(theme.palette.primary.main, 0.08),
        },
      },
    },
    MuiStepIcon: {
      root: {
        color: alpha(theme.palette.text.secondary, 0.35),
        '&$active': {
          color: theme.palette.primary.main,
        },
        '&$completed': {
          color: theme.palette.primary.main,
        },
      },
      text: {
        fill: isDark ? '#050505' : '#F9FAFB',
        fontWeight: 600,
      },
    },
    MuiStepConnector: {
      line: {
        borderColor: alpha(theme.palette.text.secondary, 0.2),
      },
    },
    MuiOutlinedInput: {
      root: {
        borderRadius: 14,
        backgroundColor: isDark ? '#111112' : '#FBFBFA',
        '& $notchedOutline': {
          borderColor: 'var(--aegis-card-border)',
        },
        '&:hover $notchedOutline': {
          borderColor: alpha(theme.palette.primary.main, 0.35),
        },
        '&$focused $notchedOutline': {
          borderColor: theme.palette.primary.main,
        },
      },
      input: {
        paddingTop: 14,
        paddingBottom: 14,
      },
    },
    MuiTypography: {
      gutterBottom: {
        marginBottom: theme.spacing(1.25),
      },
    },
    MuiTabs: {
      indicator: {
        height: 3,
        borderRadius: 999,
        backgroundColor: theme.palette.primary.main,
      },
    },
  };
};

const createPageThemes = (mode: Mode) => {
  const isDark = mode === 'dark';
  const gradientBase = isDark
    ? 'linear-gradient(120deg, rgba(139,92,246,0.28), rgba(14,165,233,0.18))'
    : 'linear-gradient(120deg, rgba(109,40,217,0.22), rgba(14,165,233,0.16))';

  return {
    home: {
      colors: ['#8B5CF6', '#22D3EE'],
      shape: 'gradient',
      backgroundImage: gradientBase,
      fontColor: isDark ? '#F5F5F5' : '#111827',
    },
    documentation: {
      colors: ['#22D3EE', '#34D399'],
      shape: 'gradient',
      backgroundImage: gradientBase,
      fontColor: isDark ? '#F5F5F5' : '#0F172A',
    },
    tool: {
      colors: ['#8B5CF6', '#F472B6'],
      shape: 'gradient',
      backgroundImage: gradientBase,
      fontColor: isDark ? '#F9FAFB' : '#0B0D12',
    },
    service: {
      colors: ['#22C55E', '#22D3EE'],
      shape: 'gradient',
      backgroundImage: gradientBase,
      fontColor: isDark ? '#F5F5F5' : '#111827',
    },
    website: {
      colors: ['#F59E0B', '#8B5CF6'],
      shape: 'gradient',
      backgroundImage: gradientBase,
      fontColor: isDark ? '#F5F5F5' : '#0F172A',
    },
    library: {
      colors: ['#6366F1', '#22D3EE'],
      shape: 'gradient',
      backgroundImage: gradientBase,
      fontColor: isDark ? '#F5F5F5' : '#111827',
    },
    other: {
      colors: ['#8B5CF6', '#22D3EE'],
      shape: 'gradient',
      backgroundImage: gradientBase,
      fontColor: isDark ? '#F5F5F5' : '#0F172A',
    },
    app: {
      colors: ['#0EA5E9', '#8B5CF6'],
      shape: 'gradient',
      backgroundImage: gradientBase,
      fontColor: isDark ? '#F5F5F5' : '#111827',
    },
    apis: {
      colors: ['#22D3EE', '#F472B6'],
      shape: 'gradient',
      backgroundImage: gradientBase,
      fontColor: isDark ? '#F5F5F5' : '#111827',
    },
  } as const;
};

const createProvider = (theme: BackstageTheme) =>
  ({ children }: PropsWithChildren<{}>) => (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <div style={{ minHeight: '100vh' }}>{children}</div>
    </ThemeProvider>
  );

const buildAegisTheme = (mode: Mode) => {
  const palette = createPalette(mode);

  const baseTheme = createTheme({
    palette,
    typography,
    shape: {
      borderRadius: 20,
    },
    props: {
      MuiButton: {
        disableElevation: true,
      },
      MuiCard: {
        raised: false,
      },
    },
  }) as BackstageTheme;

  baseTheme.overrides = createOverrides(baseTheme, mode);

  baseTheme.getPageTheme = ({ themeId }: { themeId: string }) => {
    const pageThemes = createPageThemes(mode);
    return pageThemes[themeId as keyof typeof pageThemes] ?? pageThemes.other;
  };

  return {
    id: mode === 'dark' ? 'aegis-dark' : 'aegis-light',
    title: mode === 'dark' ? 'ÆGIS Dark' : 'ÆGIS Light',
    variant: mode,
    Provider: createProvider(baseTheme),
    theme: baseTheme,
  };
};

export const aegisDarkTheme = buildAegisTheme('dark');
export const aegisLightTheme = buildAegisTheme('light');
