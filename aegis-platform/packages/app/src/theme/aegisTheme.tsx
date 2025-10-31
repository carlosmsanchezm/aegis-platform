import React, { PropsWithChildren } from 'react';
import { CssBaseline, ThemeProvider } from '@material-ui/core';
import { createTheme } from '@material-ui/core/styles';
import { alpha } from '@material-ui/core/styles/colorManipulator';
import type { BackstageTheme } from '@backstage/theme';

type ThemeMode = 'dark' | 'light';

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
    fontWeight: 600,
    fontSize: '1.25rem',
  },
  h6: {
    fontWeight: 600,
    fontSize: '1.05rem',
  },
  subtitle1: {
    fontWeight: 500,
    letterSpacing: '-0.01em',
  },
  subtitle2: {
    fontWeight: 500,
    letterSpacing: '-0.008em',
  },
  body1: {
    fontSize: '1rem',
    lineHeight: 1.7,
  },
  body2: {
    fontSize: '0.9rem',
    lineHeight: 1.6,
  },
  button: {
    fontWeight: 600,
    letterSpacing: '-0.01em',
    textTransform: 'none',
  },
} as const;

const createPalette = (mode: ThemeMode) => {
  const isDark = mode === 'dark';
  const backgroundDefault = isDark ? '#0E0F12' : '#F7F8FB';
  const backgroundPaper = isDark ? '#181A1F' : '#FFFFFF';
  const textPrimary = isDark ? '#F5F7FA' : '#111827';
  const textSecondary = isDark
    ? 'rgba(226, 232, 240, 0.65)'
    : 'rgba(55, 65, 81, 0.72)';
  const hint = isDark ? 'rgba(148, 163, 184, 0.6)' : 'rgba(100, 116, 139, 0.7)';
  const divider = isDark ? 'rgba(148, 163, 184, 0.16)' : 'rgba(15, 23, 42, 0.1)';
  const navigationBackground = isDark ? '#0B0C10' : '#FFFFFF';
  const navigationIndicator = isDark ? '#9D8BFF' : '#4338CA';
  const navigationHover = isDark
    ? 'rgba(157, 139, 255, 0.12)'
    : 'rgba(79, 70, 229, 0.08)';
  const navigationSubmenu = isDark ? '#14161C' : '#F1F2F7';
  const navigationColor = isDark ? 'rgba(199, 210, 254, 0.85)' : '#475569';
  const navigationSelected = isDark ? '#FFFFFF' : '#111827';
  const bannerInfo = isDark ? '#1F2937' : '#E8F1FF';
  const bannerError = isDark ? '#4C1D1D' : '#FDE8E9';
  const bannerText = isDark ? '#F8FAFC' : '#0F172A';
  const bannerLink = isDark ? '#A5B4FC' : '#4C1D95';
  const bannerWarning = isDark ? '#78350F' : '#FEF08A';
  const errorBackground = isDark
    ? 'rgba(255, 107, 107, 0.16)'
    : 'rgba(220, 38, 38, 0.12)';
  const warningBackground = isDark
    ? 'rgba(245, 158, 11, 0.16)'
    : 'rgba(217, 119, 6, 0.12)';
  const infoBackground = isDark
    ? 'rgba(56, 189, 248, 0.14)'
    : 'rgba(37, 99, 235, 0.12)';

  return {
    type: mode,
    primary: {
      main: isDark ? '#9D8BFF' : '#4338CA',
      light: isDark ? '#B7AFFF' : '#6366F1',
      dark: isDark ? '#6552CC' : '#312E81',
    },
    secondary: {
      main: isDark ? '#58D3FF' : '#0EA5E9',
      light: isDark ? '#7DE1FF' : '#38BDF8',
      dark: isDark ? '#1171A1' : '#0369A1',
    },
    error: {
      main: '#FF6B6B',
    },
    warning: {
      main: '#F59E0B',
    },
    success: {
      main: '#4ADE80',
    },
    background: {
      default: backgroundDefault,
      paper: backgroundPaper,
    },
    text: {
      primary: textPrimary,
      secondary: textSecondary,
      hint,
    },
    divider,
    status: {
      ok: '#22C55E',
      warning: '#F59E0B',
      error: '#FF6B6B',
      running: '#38BDF8',
      pending: '#FACC15',
      aborted: '#94A3B8',
    },
    bursts: {
      fontColor: textPrimary,
      slackChannelText: isDark ? '#BBD1FF' : '#373D4F',
      backgroundColor: {
        default: isDark ? '#11131A' : '#F5F7FF',
      },
      gradient: {
        linear: isDark
          ? 'linear-gradient(135deg, rgba(157, 139, 255, 0.45), rgba(56, 189, 248, 0.35))'
          : 'linear-gradient(135deg, rgba(67, 56, 202, 0.25), rgba(14, 165, 233, 0.2))',
      },
    },
    navigation: {
      background: navigationBackground,
      indicator: navigationIndicator,
      color: navigationColor,
      selectedColor: navigationSelected,
      navItem: {
        hoverBackground: navigationHover,
      },
      submenu: {
        background: navigationSubmenu,
      },
    },
    banner: {
      info: bannerInfo,
      error: bannerError,
      text: bannerText,
      link: bannerLink,
      warning: bannerWarning,
      closeButtonColor: bannerText,
    },
    errorBackground,
    warningBackground,
    infoBackground,
    link: isDark ? '#8B5CF6' : '#4338CA',
    linkHover: isDark ? '#A78BFA' : '#312E81',
    errorText: isDark ? '#FECACA' : '#B91C1C',
    infoText: isDark ? '#BAE6FD' : '#1D4ED8',
    warningText: isDark ? '#FDE68A' : '#B45309',
    gold: '#FACC15',
    border: divider,
    textContrast: isDark ? '#0B0F1A' : '#FFFFFF',
    textVerySubtle: isDark ? 'rgba(148, 163, 184, 0.5)' : 'rgba(71, 85, 105, 0.6)',
    textSubtle: textSecondary,
    highlight: isDark
      ? 'rgba(157, 139, 255, 0.18)'
      : 'rgba(79, 70, 229, 0.16)',
    pinSidebarButton: {
      icon: isDark ? '#0E111A' : '#FFFFFF',
      background: isDark
        ? 'rgba(148, 163, 184, 0.45)'
        : 'rgba(71, 85, 105, 0.18)',
    },
    tabbar: {
      indicator: navigationIndicator,
    },
  } as const;
};

type AegisPalette = ReturnType<typeof createPalette>;

const createOverrides = (mode: ThemeMode, palette: AegisPalette) => {
  const isDark = mode === 'dark';
  const focusRing = alpha(palette.primary.main, isDark ? 0.45 : 0.35);
  const cardBorder = alpha(palette.divider, 0.9);

  return {
    MuiCssBaseline: {
      '@global': {
        '*': {
          boxSizing: 'border-box',
        },
        ':root': {
          '--aegis-card-radius': '18px',
          '--aegis-card-shadow': isDark
            ? '0px 18px 45px rgba(2, 6, 23, 0.55)'
            : '0px 18px 45px rgba(15, 23, 42, 0.12)',
          '--aegis-ring-color': focusRing,
          '--aegis-border-color': palette.divider,
        },
        body: {
          margin: 0,
          backgroundColor: palette.background.default,
          backgroundImage: isDark
            ? 'radial-gradient(circle at 22% 16%, rgba(157, 139, 255, 0.15), transparent 55%), radial-gradient(circle at 80% 4%, rgba(14, 165, 233, 0.12), transparent 55%)'
            : 'radial-gradient(circle at 12% 0%, rgba(67, 56, 202, 0.1), transparent 45%), radial-gradient(circle at 88% 12%, rgba(14, 165, 233, 0.12), transparent 55%)',
          color: palette.text.primary,
          fontFamily: typography.fontFamily,
          fontFeatureSettings: '"cv02","cv03","cv04","cv11"',
          letterSpacing: '-0.01em',
        },
        a: {
          color: palette.link,
        },
      },
    },
    MuiDrawer: {
      paper: {
        backgroundColor: palette.navigation.background,
        color: palette.navigation.color,
      },
    },
    MuiDivider: {
      root: {
        backgroundColor: alpha(palette.divider, 0.8),
      },
    },
    MuiPaper: {
      rounded: {
        borderRadius: 'var(--aegis-card-radius)',
      },
      elevation1: {
        boxShadow: 'var(--aegis-card-shadow)',
      },
    },
    MuiCard: {
      root: {
        borderRadius: 'var(--aegis-card-radius)',
        border: `1px solid ${cardBorder}`,
        backgroundColor: palette.background.paper,
        boxShadow: 'none',
        transition:
          'border-color 200ms ease, transform 200ms ease, box-shadow 200ms ease',
        '&:hover': {
          boxShadow: 'var(--aegis-card-shadow)',
          transform: 'translateY(-2px)',
        },
      },
    },
    MuiButton: {
      root: {
        borderRadius: 999,
        fontWeight: 600,
        padding: '0.55rem 1.5rem',
      },
      contained: {
        boxShadow: 'none',
      },
      containedPrimary: {
        color: palette.textContrast,
        '&:hover': {
          boxShadow: 'var(--aegis-card-shadow)',
        },
      },
      outlined: {
        borderColor: alpha(palette.divider, 0.8),
        '&:hover': {
          borderColor: palette.primary.main,
          backgroundColor: alpha(palette.primary.main, 0.08),
        },
      },
      text: {
        paddingLeft: '0.5rem',
        paddingRight: '0.5rem',
      },
    },
    MuiStepper: {
      root: {
        backgroundColor: 'transparent',
        padding: '0 0 24px',
      },
    },
    MuiStepConnector: {
      line: {
        borderColor: alpha(palette.divider, 0.7),
      },
    },
    MuiStepLabel: {
      label: {
        color: palette.text.secondary,
        '&$active': {
          color: palette.text.primary,
        },
        '&$completed': {
          color: palette.primary.main,
        },
      },
      active: {},
      completed: {},
      iconContainer: {
        '& $active': {
          color: palette.primary.main,
        },
      },
    },
    MuiOutlinedInput: {
      root: {
        borderRadius: 12,
        '&$focused $notchedOutline': {
          borderColor: palette.primary.main,
          boxShadow: `0 0 0 1px ${focusRing}`,
        },
      },
      focused: {},
      input: {
        padding: '14px 16px',
      },
      notchedOutline: {
        borderColor: alpha(palette.divider, 0.8),
      },
    },
    MuiFormLabel: {
      root: {
        color: palette.text.secondary,
        '&$focused': {
          color: palette.primary.main,
        },
      },
      focused: {},
    },
    MuiSwitch: {
      track: {
        backgroundColor: alpha(palette.primary.main, isDark ? 0.5 : 0.35),
      },
    },
    MuiTooltip: {
      tooltip: {
        borderRadius: 12,
        backgroundColor: alpha(palette.text.primary, isDark ? 0.92 : 0.9),
        color: isDark ? '#0B0F1A' : '#111827',
      },
    },
  } as const;
};

const createPageThemes = (mode: ThemeMode) => {
  const isDark = mode === 'dark';
  const fontColor = isDark ? '#F8FAFC' : '#111827';

  return {
    home: {
      colors: ['#9D8BFF', '#0EA5E9'],
      shape:
        'linear-gradient(135deg, rgba(157, 139, 255, 0.42), rgba(14, 165, 233, 0.32))',
      backgroundImage:
        'radial-gradient(circle at 18% 18%, rgba(157, 139, 255, 0.28), transparent 60%), radial-gradient(circle at 82% 24%, rgba(14, 165, 233, 0.24), transparent 65%)',
      fontColor,
    },
    documentation: {
      colors: ['#4338CA', '#0EA5E9'],
      shape:
        'linear-gradient(135deg, rgba(67, 56, 202, 0.32), rgba(14, 165, 233, 0.26))',
      backgroundImage:
        'radial-gradient(circle at 25% 30%, rgba(67, 56, 202, 0.24), transparent 60%), radial-gradient(circle at 70% 65%, rgba(14, 165, 233, 0.24), transparent 65%)',
      fontColor,
    },
    tool: {
      colors: ['#22C55E', '#0EA5E9'],
      shape:
        'linear-gradient(135deg, rgba(34, 197, 94, 0.32), rgba(14, 165, 233, 0.24))',
      backgroundImage:
        'radial-gradient(circle at 30% 40%, rgba(34, 197, 94, 0.24), transparent 60%), radial-gradient(circle at 80% 60%, rgba(14, 165, 233, 0.22), transparent 60%)',
      fontColor,
    },
    service: {
      colors: ['#F59E0B', '#6366F1'],
      shape:
        'linear-gradient(135deg, rgba(245, 158, 11, 0.32), rgba(99, 102, 241, 0.24))',
      backgroundImage:
        'radial-gradient(circle at 25% 30%, rgba(245, 158, 11, 0.24), transparent 60%), radial-gradient(circle at 75% 70%, rgba(99, 102, 241, 0.24), transparent 65%)',
      fontColor,
    },
    website: {
      colors: ['#10B981', '#64748B'],
      shape:
        'linear-gradient(135deg, rgba(16, 185, 129, 0.28), rgba(100, 116, 139, 0.22))',
      backgroundImage:
        'radial-gradient(circle at 35% 20%, rgba(16, 185, 129, 0.24), transparent 60%), radial-gradient(circle at 65% 70%, rgba(100, 116, 139, 0.22), transparent 65%)',
      fontColor,
    },
    library: {
      colors: ['#6366F1', '#14B8A6'],
      shape:
        'linear-gradient(135deg, rgba(99, 102, 241, 0.32), rgba(20, 184, 166, 0.24))',
      backgroundImage:
        'radial-gradient(circle at 25% 25%, rgba(99, 102, 241, 0.24), transparent 60%), radial-gradient(circle at 70% 65%, rgba(20, 184, 166, 0.22), transparent 65%)',
      fontColor,
    },
    other: {
      colors: ['#0EA5E9', '#6366F1'],
      shape:
        'linear-gradient(135deg, rgba(14, 165, 233, 0.32), rgba(99, 102, 241, 0.26))',
      backgroundImage:
        'radial-gradient(circle at 20% 30%, rgba(14, 165, 233, 0.24), transparent 60%), radial-gradient(circle at 75% 50%, rgba(99, 102, 241, 0.24), transparent 65%)',
      fontColor,
    },
    app: {
      colors: ['#22C55E', '#F97316'],
      shape:
        'linear-gradient(135deg, rgba(34, 197, 94, 0.32), rgba(249, 115, 22, 0.24))',
      backgroundImage:
        'radial-gradient(circle at 25% 25%, rgba(34, 197, 94, 0.24), transparent 60%), radial-gradient(circle at 70% 65%, rgba(249, 115, 22, 0.22), transparent 65%)',
      fontColor,
    },
    apis: {
      colors: ['#38BDF8', '#22C55E'],
      shape:
        'linear-gradient(135deg, rgba(56, 189, 248, 0.32), rgba(34, 197, 94, 0.24))',
      backgroundImage:
        'radial-gradient(circle at 30% 30%, rgba(56, 189, 248, 0.24), transparent 60%), radial-gradient(circle at 70% 60%, rgba(34, 197, 94, 0.22), transparent 65%)',
      fontColor,
    },
  } as Record<string, any>;
};

const createAegisTheme = (mode: ThemeMode) => {
  const palette = createPalette(mode);
  const baseTheme = createTheme({
    palette,
    typography,
    overrides: createOverrides(mode, palette),
  }) as BackstageTheme;

  baseTheme.defaultPageTheme = 'home';
  baseTheme.getPageTheme = ({ themeId }: { themeId: string }) => {
    const pageThemes = createPageThemes(mode);
    return pageThemes[themeId] || pageThemes.other;
  };

  return baseTheme;
};

const withProvider = (
  theme: BackstageTheme,
  id: string,
  title: string,
  variant: ThemeMode,
) => ({
  id,
  title,
  variant: variant as const,
  Provider: ({ children }: PropsWithChildren<{}>) => (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      {/* Surface gradient lives here so individual pages stay focused. */}
      {children}
    </ThemeProvider>
  ),
  theme,
});

const aegisDarkBase = createAegisTheme('dark');
const aegisLightBase = createAegisTheme('light');

export const aegisDarkTheme = withProvider(
  aegisDarkBase,
  'aegis-dark',
  'ÆGIS Flux (Dark)',
  'dark',
);

export const aegisLightTheme = withProvider(
  aegisLightBase,
  'aegis-light',
  'ÆGIS Flux (Light)',
  'light',
);

// Backwards compatibility for existing imports expecting a single theme.
export const aegisTheme = aegisDarkTheme;

export default aegisDarkTheme;
