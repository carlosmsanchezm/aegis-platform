import { PropsWithChildren } from 'react';
import { CssBaseline, ThemeProvider } from '@material-ui/core';
import { createTheme } from '@material-ui/core/styles';
import { alpha } from '@material-ui/core/styles/colorManipulator';
import type { BackstageTheme, PageTheme } from '@backstage/theme';

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
  const baseBackground = isDark ? '#0B0B0C' : '#F7F7F7';
  const paperBackground = isDark ? '#161617' : '#FFFFFF';
  const primaryMain = isDark ? '#E6E6E6' : '#1C1C1C';
  const secondaryMain = isDark ? '#BFBFBF' : '#4F4F4F';
  const textPrimary = isDark ? '#FAFAFA' : '#0F0F0F';
  const textSecondary = isDark ? '#BDBDBD' : '#4C4C4C';
  const divider = isDark ? 'rgba(255,255,255,0.14)' : 'rgba(0,0,0,0.12)';
  const navigationBackground = isDark ? '#090909' : '#F1F1F1';
  const hoverBackground = isDark
    ? 'rgba(255, 255, 255, 0.08)'
    : 'rgba(0, 0, 0, 0.06)';
  const submenuBackground = isDark ? '#161616' : '#E8E8E8';
  const bannerInfo = isDark ? '#1D1D1F' : '#E7E7E7';
  const bannerError = isDark ? '#252526' : '#DCDCDC';
  const infoBackground = isDark ? 'rgba(255, 255, 255, 0.12)' : '#ECECEC';
  const warningBackground = isDark
    ? 'rgba(255, 255, 255, 0.18)'
    : '#E2E2E2';
  const errorBackground = isDark
    ? 'rgba(255, 255, 255, 0.22)'
    : '#D5D5D5';

  return {
    type: mode,
    primary: {
      main: primaryMain,
      light: isDark ? '#F4F4F4' : '#2E2E2E',
      dark: isDark ? '#C7C7C7' : '#111111',
    },
    secondary: {
      main: secondaryMain,
      light: isDark ? '#D8D8D8' : '#636363',
      dark: isDark ? '#9D9D9D' : '#3B3B3B',
    },
    error: {
      main: isDark ? '#9E9E9E' : '#3F3F3F',
    },
    warning: {
      main: isDark ? '#BEBEBE' : '#5C5C5C',
    },
    success: {
      main: isDark ? '#D6D6D6' : '#2F2F2F',
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
      color: isDark ? '#D5D5D5' : '#4F4F4F',
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
      main: isDark ? '#AFAFAF' : '#5A5A5A',
    },
    navigationIndicator: primaryMain,
    tabbar: {
      indicator: primaryMain,
    },
    status: {
      ok: isDark ? '#D9D9D9' : '#2E2E2E',
      warning: isDark ? '#BFBFBF' : '#4B4B4B',
      error: isDark ? '#969696' : '#3A3A3A',
      running: isDark ? '#C9C9C9' : '#3C3C3C',
      pending: isDark ? '#B3B3B3' : '#555555',
      aborted: isDark ? '#8C8C8C' : '#6B6B6B',
    },
    bursts: {
      fontColor: textPrimary,
      slackChannelText: isDark ? '#EAEAEA' : '#1C1C1C',
      backgroundColor: {
        default: baseBackground,
      },
      gradient: {
        linear: isDark
          ? 'linear-gradient(135deg, rgba(255,255,255,0.08), rgba(255,255,255,0.02))'
          : 'linear-gradient(135deg, rgba(0,0,0,0.08), rgba(0,0,0,0.02))',
      },
    },
    pinSidebarButton: {
      icon: isDark ? '#0F0F0F' : '#F5F5F5',
      background: isDark ? 'rgba(255,255,255,0.22)' : 'rgba(0,0,0,0.24)',
    },
  } as unknown as BackstageTheme['palette'];
};

const createGlobalStyles = (theme: BackstageTheme, mode: Mode) => {
  const isDark = mode === 'dark';
  const subtle = isDark ? 'rgba(255,255,255,0.06)' : 'rgba(0,0,0,0.08)';
  const bodyGradient = isDark
    ? 'radial-gradient(circle at 20% 12%, rgba(255,255,255,0.06), transparent 55%), radial-gradient(circle at 82% 10%, rgba(255,255,255,0.04), transparent 60%), #0B0B0C'
    : 'radial-gradient(circle at 14% 10%, rgba(0,0,0,0.06), transparent 48%), radial-gradient(circle at 86% 12%, rgba(0,0,0,0.04), transparent 55%), #F7F7F7';

  return {
    ':root': {
      '--aegis-card-surface': theme.palette.background.paper,
      '--aegis-card-border': isDark ? 'rgba(255,255,255,0.1)' : 'rgba(0,0,0,0.08)',
      '--aegis-card-shadow': isDark
        ? '0 18px 45px rgba(0, 0, 0, 0.45)'
        : '0 18px 45px rgba(0, 0, 0, 0.12)',
      '--aegis-muted': isDark ? 'rgba(255,255,255,0.08)' : 'rgba(0,0,0,0.06)',
    },
    body: {
      background: bodyGradient,
      color: theme.palette.text.primary,
      fontFamily: typography.fontFamily,
      letterSpacing: '-0.01em',
      minHeight: '100vh',
    },
    '.BackstageHeader': {
      backgroundColor: 'transparent',
      backgroundImage: 'none !important',
      boxShadow: 'none',
      borderBottom: 'none',
      marginBottom: theme.spacing(1),
    },
    '.BackstageHeader-header': {
      padding: theme.spacing(1.75, 3, 0.75),
      minHeight: 'auto',
      borderBottom: 'none',
    },
    '.BackstageHeader-title': {
      color: theme.palette.text.primary,
      letterSpacing: '-0.015em',
    },
    '.BackstageHeader-label, .BackstageHeader-subtitle': {
      display: 'none',
    },
    '.BackstageSidebarItem': {
      overflow: 'visible',
    },
    '.BackstageSidebarItem-open': {
      paddingRight: theme.spacing(2.5),
    },
    '.BackstageSidebarItem .MuiListItemText-root': {
      overflow: 'visible',
    },
    '.BackstageSidebarItem-label': {
      width: 'auto !important',
      maxWidth: 'none',
      whiteSpace: 'normal',
      overflow: 'visible',
      textOverflow: 'unset',
      lineHeight: 1.35,
      display: 'block',
      flex: '0 0 auto',
    },
    '.BackstageSidebarItem .MuiTypography-root': {
      whiteSpace: 'normal',
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
          ? 'linear-gradient(135deg, #FFFFFF, #C7C7C7)'
          : 'linear-gradient(135deg, #1A1A1A, #343434)',
        color: isDark ? '#0B0B0B' : '#FAFAFA',
        '&:hover': {
          backgroundImage: isDark
            ? 'linear-gradient(135deg, #F0F0F0, #B8B8B8)'
            : 'linear-gradient(135deg, #242424, #3F3F3F)',
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
    ? 'linear-gradient(120deg, rgba(255,255,255,0.08), rgba(255,255,255,0.02))'
    : 'linear-gradient(120deg, rgba(0,0,0,0.08), rgba(0,0,0,0.02))';

  const palette = {
    accent: isDark ? '#F3F3F3' : '#1A1A1A',
    subtle: isDark ? '#BDBDBD' : '#5B5B5B',
  } as const;

  return {
    home: {
      colors: [palette.accent, palette.subtle] as string[],
      shape: 'gradient',
      backgroundImage: gradientBase,
      fontColor: isDark ? '#FAFAFA' : '#0F0F0F',
    },
    documentation: {
      colors: [palette.accent, palette.subtle] as string[],
      shape: 'gradient',
      backgroundImage: gradientBase,
      fontColor: isDark ? '#F0F0F0' : '#121212',
    },
    tool: {
      colors: [palette.accent, palette.subtle] as string[],
      shape: 'gradient',
      backgroundImage: gradientBase,
      fontColor: isDark ? '#F7F7F7' : '#111111',
    },
    service: {
      colors: [palette.accent, palette.subtle] as string[],
      shape: 'gradient',
      backgroundImage: gradientBase,
      fontColor: isDark ? '#F8F8F8' : '#101010',
    },
    website: {
      colors: [palette.accent, palette.subtle] as string[],
      shape: 'gradient',
      backgroundImage: gradientBase,
      fontColor: isDark ? '#F5F5F5' : '#101010',
    },
    library: {
      colors: [palette.accent, palette.subtle] as string[],
      shape: 'gradient',
      backgroundImage: gradientBase,
      fontColor: isDark ? '#F6F6F6' : '#101010',
    },
    other: {
      colors: [palette.accent, palette.subtle] as string[],
      shape: 'gradient',
      backgroundImage: gradientBase,
      fontColor: isDark ? '#F4F4F4' : '#101010',
    },
    app: {
      colors: [palette.accent, palette.subtle] as string[],
      shape: 'gradient',
      backgroundImage: gradientBase,
      fontColor: isDark ? '#FAFAFA' : '#0F0F0F',
    },
    apis: {
      colors: [palette.accent, palette.subtle] as string[],
      shape: 'gradient',
      backgroundImage: gradientBase,
      fontColor: isDark ? '#F9F9F9' : '#0F0F0F',
    },
  } as Record<string, PageTheme>;
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
