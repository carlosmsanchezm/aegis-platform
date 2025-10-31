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
  const baseBackground = isDark ? '#050505' : '#F7F7F7';
  const paperBackground = isDark ? '#0F0F0F' : '#FFFFFF';
  const primaryMain = isDark ? '#FFFFFF' : '#1A1A1A';
  const secondaryMain = isDark ? '#CFCFCF' : '#2B2B2B';
  const textPrimary = isDark ? '#F5F5F5' : '#111111';
  const textSecondary = isDark ? '#B5B5B5' : '#4F4F4F';
  const divider = isDark ? 'rgba(255,255,255,0.12)' : 'rgba(17,17,17,0.12)';
  const navigationBackground = isDark ? '#080808' : '#EFEFEF';
  const hoverBackground = isDark
    ? 'rgba(255,255,255,0.1)'
    : 'rgba(0,0,0,0.06)';
  const submenuBackground = isDark ? '#111111' : '#FFFFFF';
  const bannerInfo = isDark ? '#101010' : '#F1F1F1';
  const bannerError = isDark ? '#141414' : '#E6E6E6';
  const infoBackground = isDark ? 'rgba(255,255,255,0.08)' : 'rgba(0,0,0,0.04)';
  const warningBackground = isDark ? 'rgba(255,255,255,0.06)' : 'rgba(0,0,0,0.07)';
  const errorBackground = isDark ? 'rgba(255,255,255,0.12)' : 'rgba(0,0,0,0.1)';
  const indicator = isDark ? '#FFFFFF' : '#111111';

  return {
    type: mode,
    primary: {
      main: primaryMain,
      light: isDark ? '#FFFFFF' : '#3C3C3C',
      dark: isDark ? '#D0D0D0' : '#000000',
    },
    secondary: {
      main: secondaryMain,
      light: isDark ? '#E0E0E0' : '#3F3F3F',
      dark: isDark ? '#A0A0A0' : '#141414',
    },
    error: {
      main: isDark ? '#E0E0E0' : '#303030',
    },
    warning: {
      main: isDark ? '#D0D0D0' : '#444444',
    },
    success: {
      main: isDark ? '#C4C4C4' : '#3A3A3A',
    },
    background: {
      default: baseBackground,
      paper: paperBackground,
    },
    text: {
      primary: textPrimary,
      secondary: textSecondary,
      hint: isDark ? '#8A8A8A' : '#6D6D6D',
    },
    divider,
    navigation: {
      background: navigationBackground,
      indicator,
      color: textSecondary,
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
      link: textPrimary,
    },
    link: textPrimary,
    linkHover: isDark ? '#FFFFFF' : '#000000',
    errorText: textSecondary,
    infoText: textSecondary,
    warningText: textSecondary,
    errorBackground,
    warningBackground,
    infoBackground,
    neutral: {
      main: textSecondary,
    },
    navigationIndicator: indicator,
    tabbar: {
      indicator,
    },
    status: {
      ok: isDark ? '#E0E0E0' : '#303030',
      warning: isDark ? '#BFBFBF' : '#4A4A4A',
      error: isDark ? '#F0F0F0' : '#202020',
      running: isDark ? '#D6D6D6' : '#2E2E2E',
      pending: isDark ? '#AFAFAF' : '#3A3A3A',
      aborted: isDark ? '#7D7D7D' : '#5C5C5C',
    },
    bursts: {
      fontColor: textPrimary,
      slackChannelText: textPrimary,
      backgroundColor: {
        default: baseBackground,
      },
      gradient: {
        linear: isDark
          ? 'linear-gradient(135deg, rgba(255,255,255,0.12), rgba(255,255,255,0.02))'
          : 'linear-gradient(135deg, rgba(0,0,0,0.06), rgba(0,0,0,0.02))',
      },
    },
    pinSidebarButton: {
      icon: isDark ? '#050505' : '#F5F5F5',
      background: isDark ? 'rgba(255,255,255,0.28)' : 'rgba(0,0,0,0.18)',
    },
  } as unknown as BackstageTheme['palette'];
};

const createGlobalStyles = (theme: BackstageTheme, mode: Mode) => {
  const isDark = mode === 'dark';
  const subtle = isDark ? 'rgba(255,255,255,0.08)' : 'rgba(0,0,0,0.05)';
  const bodyGradient = isDark
    ? 'radial-gradient(circle at 15% 20%, rgba(255,255,255,0.12), transparent 55%), radial-gradient(circle at 85% 12%, rgba(255,255,255,0.06), transparent 55%), #050505'
    : 'radial-gradient(circle at 12% 18%, rgba(0,0,0,0.06), transparent 50%), radial-gradient(circle at 88% 12%, rgba(0,0,0,0.04), transparent 55%), #F7F7F7';
  const sidebarBorder = alpha(theme.palette.text.primary, isDark ? 0.22 : 0.12);

  return {
    ':root': {
      '--aegis-card-surface': theme.palette.background.paper,
      '--aegis-card-border': isDark ? 'rgba(255,255,255,0.1)' : 'rgba(17,17,17,0.08)',
      '--aegis-card-shadow': isDark
        ? '0 18px 45px rgba(0, 0, 0, 0.5)'
        : '0 18px 45px rgba(0, 0, 0, 0.14)',
      '--aegis-muted': isDark ? 'rgba(255,255,255,0.1)' : 'rgba(0,0,0,0.06)',
    },
    body: {
      background: bodyGradient,
      color: theme.palette.text.primary,
      fontFamily: typography.fontFamily,
      letterSpacing: '-0.01em',
      minHeight: '100vh',
    },
    '.BackstageSidebar-root': {
      backgroundColor: theme.palette.navigation.background,
      borderRight: `1px solid ${sidebarBorder}`,
    },
    '.BackstageHeader-header': {
      backgroundColor: 'transparent',
      backgroundImage: 'none !important',
      boxShadow: 'none !important',
      border: 'none',
      padding: theme.spacing(3, 0, 2),
      minHeight: 'auto',
      margin: 0,
    },
    '.BackstageHeader-type, .BackstageHeader-subtitle, .BackstageHeader-breadcrumb': {
      display: 'none',
    },
    '.BackstageHeader-title, .BackstageHeader-label, .BackstageHeader-subtitle': {
      color: theme.palette.text.primary,
      letterSpacing: '-0.02em',
      fontWeight: 600,
    },
    '.BackstageSidebarItem-open': {
      paddingRight: theme.spacing(2.5),
    },
    '.BackstageSidebarItem-label': {
      width: 'auto !important',
      maxWidth: 'none',
      whiteSpace: 'normal',
      overflow: 'visible',
      textOverflow: 'unset',
      lineHeight: 1.35,
      display: 'block',
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
          ? 'linear-gradient(135deg, #FFFFFF, #BEBEBE)'
          : 'linear-gradient(135deg, #1A1A1A, #3A3A3A)',
        color: isDark ? '#050505' : '#F7F7F7',
        '&:hover': {
          backgroundImage: isDark
            ? 'linear-gradient(135deg, #F0F0F0, #CFCFCF)'
            : 'linear-gradient(135deg, #262626, #4B4B4B)',
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
        fill: isDark ? '#050505' : '#F7F7F7',
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
        backgroundColor: isDark ? '#0D0D0D' : '#F9F9F9',
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
    ? 'linear-gradient(120deg, rgba(255,255,255,0.12), rgba(255,255,255,0.04))'
    : 'linear-gradient(120deg, rgba(0,0,0,0.06), rgba(0,0,0,0.02))';
  const paletteColors = (isDark
    ? ['#F5F5F5', '#BEBEBE']
    : ['#1A1A1A', '#4A4A4A']) as string[];
  const fontColor = isDark ? '#F5F5F5' : '#111111';

  return {
    home: {
      colors: paletteColors,
      shape: 'gradient',
      backgroundImage: gradientBase,
      fontColor,
    },
    documentation: {
      colors: paletteColors,
      shape: 'gradient',
      backgroundImage: gradientBase,
      fontColor,
    },
    tool: {
      colors: paletteColors,
      shape: 'gradient',
      backgroundImage: gradientBase,
      fontColor,
    },
    service: {
      colors: paletteColors,
      shape: 'gradient',
      backgroundImage: gradientBase,
      fontColor,
    },
    website: {
      colors: paletteColors,
      shape: 'gradient',
      backgroundImage: gradientBase,
      fontColor,
    },
    library: {
      colors: paletteColors,
      shape: 'gradient',
      backgroundImage: gradientBase,
      fontColor,
    },
    other: {
      colors: paletteColors,
      shape: 'gradient',
      backgroundImage: gradientBase,
      fontColor,
    },
    app: {
      colors: paletteColors,
      shape: 'gradient',
      backgroundImage: gradientBase,
      fontColor,
    },
    apis: {
      colors: paletteColors,
      shape: 'gradient',
      backgroundImage: gradientBase,
      fontColor,
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
