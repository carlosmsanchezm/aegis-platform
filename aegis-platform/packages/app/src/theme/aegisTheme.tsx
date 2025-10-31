import React, { PropsWithChildren } from 'react';
import { CssBaseline } from '@material-ui/core';
import { ThemeProvider } from '@material-ui/core/styles';
import { createUnifiedTheme } from '@backstage/theme';
import type { BackstageTheme } from '@backstage/theme';

const baseTheme = createUnifiedTheme({
  palette: {
    type: 'dark',
    primary: {
      main: '#7C5CFF',
      contrastText: '#050B1A',
    },
    secondary: {
      main: '#00E5FF',
      contrastText: '#02101F',
    },
    error: {
      main: '#FF5F5F',
    },
    warning: {
      main: '#FFB74D',
    },
    success: {
      main: '#4ADE80',
    },
    info: {
      main: '#38BDF8',
    },
    background: {
      default: '#050B1A',
      paper: '#0D1424',
    },
    text: {
      primary: '#E4F1FF',
      secondary: '#8CA6C1',
      hint: '#5F6E8F',
    },
    divider: 'rgba(142, 163, 208, 0.24)',
    navigation: {
      background: '#040A15',
      indicator: '#00E5FF',
      color: '#7A90B0',
      selectedColor: '#F8FAFF',
      navItem: {
        hoverBackground: 'rgba(0, 229, 255, 0.08)',
      },
      submenu: {
        background: '#070F1E',
      },
    },
    banner: {
      info: '#123D5A',
      error: '#441D2D',
      text: '#F1F7FF',
      link: '#67E8F9',
    },
    errorBackground: 'rgba(255, 95, 95, 0.16)',
    warningBackground: 'rgba(255, 183, 77, 0.16)',
    infoBackground: 'rgba(56, 189, 248, 0.16)',
    successBackground: 'rgba(74, 222, 128, 0.14)',
  },
  typography: {
    fontFamily:
      "'Inter', 'IBM Plex Sans', 'Roboto', 'Helvetica Neue', Helvetica, Arial, sans-serif",
    h1: {
      fontWeight: 700,
      letterSpacing: '0.04em',
      fontSize: '3rem',
    },
    h2: {
      fontWeight: 600,
      letterSpacing: '0.03em',
      fontSize: '2.4rem',
    },
    h3: {
      fontWeight: 600,
      letterSpacing: '0.02em',
      fontSize: '2rem',
    },
    subtitle1: {
      fontWeight: 500,
      letterSpacing: '0.04em',
    },
    body1: {
      fontSize: '1rem',
      lineHeight: 1.6,
    },
    button: {
      textTransform: 'none',
      fontWeight: 600,
      letterSpacing: '0.08em',
    },
  },
  shape: {
    borderRadius: 16,
  },
  defaultPageTheme: 'home',
  pageTheme: {},
  overrides: {
    MuiCssBaseline: {
      '@global': {
        body: {
          backgroundColor: '#050B1A',
          backgroundImage:
            'radial-gradient(circle at 20% 20%, rgba(0, 229, 255, 0.16), transparent 45%), radial-gradient(circle at 80% 0%, rgba(140, 83, 255, 0.24), transparent 40%)',
          color: '#E4F1FF',
          fontFamily:
            "'Inter', 'IBM Plex Sans', 'Roboto', 'Helvetica Neue', Helvetica, Arial, sans-serif",
          minHeight: '100vh',
        },
        a: {
          color: '#67E8F9',
        },
        '*::-webkit-scrollbar': {
          width: 10,
          backgroundColor: '#040A15',
        },
        '*::-webkit-scrollbar-thumb': {
          backgroundColor: '#1F2E46',
          borderRadius: 8,
          border: '2px solid #040A15',
        },
      },
    },
    MuiButton: {
      root: {
        borderRadius: 999,
        padding: '10px 22px',
        fontWeight: 600,
        textTransform: 'none',
        transition: 'transform 0.18s ease, box-shadow 0.18s ease',
        boxShadow: '0 10px 30px rgba(3, 20, 43, 0.35)',
        '&:hover': {
          transform: 'translateY(-1px)',
          boxShadow: '0 18px 40px rgba(0, 229, 255, 0.28)',
        },
      },
      containedPrimary: {
        backgroundImage: 'linear-gradient(120deg, #7C5CFF 0%, #3E8BFF 100%)',
        boxShadow: '0 16px 32px rgba(109, 92, 255, 0.35)',
        color: '#050B1A',
      },
      outlinedPrimary: {
        borderColor: 'rgba(0, 229, 255, 0.6)',
        color: '#00E5FF',
        '&:hover': {
          borderColor: '#00E5FF',
          backgroundColor: 'rgba(0, 229, 255, 0.08)',
        },
      },
    },
    MuiAppBar: {
      colorPrimary: {
        backgroundColor: '#060E1F',
        boxShadow: '0 20px 40px rgba(3, 12, 32, 0.64)',
        borderBottom: '1px solid rgba(103, 232, 249, 0.12)',
      },
    },
    MuiDrawer: {
      paper: {
        backgroundColor: '#040A15',
        borderRight: '1px solid rgba(98, 153, 255, 0.14)',
        backdropFilter: 'blur(18px)',
      },
    },
    MuiPaper: {
      rounded: {
        borderRadius: 20,
      },
      elevation1: {
        background:
          'linear-gradient(160deg, rgba(11, 20, 39, 0.85) 0%, rgba(8, 22, 43, 0.95) 100%)',
        border: '1px solid rgba(103, 232, 249, 0.12)',
        boxShadow: '0 30px 70px rgba(2, 12, 34, 0.65)',
      },
    },
    MuiCard: {
      root: {
        borderRadius: 24,
        background:
          'linear-gradient(180deg, rgba(9, 17, 36, 0.9) 0%, rgba(6, 17, 33, 0.96) 100%)',
        border: '1px solid rgba(124, 92, 255, 0.14)',
        boxShadow: '0 28px 60px rgba(1, 9, 26, 0.6)',
      },
    },
    MuiChip: {
      root: {
        borderRadius: 10,
        backgroundColor: 'rgba(0, 229, 255, 0.12)',
        color: '#67E8F9',
      },
    },
    MuiListItem: {
      button: {
        borderRadius: 12,
        color: '#8CA6C1',
        '&:hover': {
          backgroundColor: 'rgba(124, 92, 255, 0.12)',
          color: '#F0F7FF',
        },
        '&$selected': {
          backgroundColor: 'rgba(0, 229, 255, 0.12)',
          color: '#F8FAFF',
        },
      },
      selected: {},
    },
    MuiTabs: {
      indicator: {
        height: 3,
        borderRadius: 3,
        backgroundImage: 'linear-gradient(90deg, #00E5FF 0%, #7C5CFF 100%)',
      },
    },
    MuiTab: {
      root: {
        minHeight: 48,
        textTransform: 'none',
        fontWeight: 600,
        letterSpacing: '0.05em',
      },
      textColorPrimary: {
        color: '#7A90B0',
        '&$selected': {
          color: '#F8FAFF',
        },
      },
      selected: {},
    },
    MuiTableCell: {
      head: {
        fontWeight: 600,
        letterSpacing: '0.06em',
        color: '#E4F1FF',
        borderBottom: '1px solid rgba(142, 163, 208, 0.24)',
      },
      root: {
        borderBottom: '1px solid rgba(142, 163, 208, 0.12)',
      },
    },
    MuiTooltip: {
      tooltip: {
        backgroundColor: 'rgba(7, 24, 42, 0.95)',
        color: '#E4F1FF',
        borderRadius: 10,
        boxShadow: '0 18px 40px rgba(3, 12, 32, 0.5)',
      },
    },
  },
}) as BackstageTheme;

baseTheme.getPageTheme = ({ themeId }: { themeId: string }) => {
  const pageThemes: Record<string, any> = {
    home: {
      colors: ['#050B1A', '#0D1424'],
      shape: 'linear-gradient(135deg, rgba(0, 229, 255, 0.4), rgba(124, 92, 255, 0.4))',
      backgroundImage:
        'linear-gradient(135deg, rgba(0, 229, 255, 0.2) 0%, rgba(124, 92, 255, 0.25) 50%, rgba(26, 42, 108, 0.4) 100%)',
      fontColor: '#F8FAFF',
    },
    documentation: {
      colors: ['#071129', '#102040'],
      shape: 'linear-gradient(160deg, rgba(62, 139, 255, 0.45), rgba(0, 229, 255, 0.2))',
      backgroundImage:
        'linear-gradient(120deg, rgba(62, 139, 255, 0.18) 0%, rgba(5, 11, 26, 0.9) 100%)',
      fontColor: '#E4F1FF',
    },
    tool: {
      colors: ['#050B1A', '#112235'],
      shape: 'linear-gradient(140deg, rgba(124, 92, 255, 0.5), rgba(0, 229, 255, 0.3))',
      backgroundImage:
        'linear-gradient(160deg, rgba(7, 24, 56, 0.95) 0%, rgba(11, 33, 64, 0.9) 100%)',
      fontColor: '#F8FAFF',
    },
    service: {
      colors: ['#081027', '#14253B'],
      shape: 'linear-gradient(150deg, rgba(67, 56, 202, 0.6), rgba(56, 189, 248, 0.28))',
      backgroundImage:
        'linear-gradient(130deg, rgba(8, 26, 52, 0.94) 0%, rgba(6, 17, 33, 0.96) 100%)',
      fontColor: '#F1F7FF',
    },
    website: {
      colors: ['#060E1F', '#101C32'],
      shape: 'linear-gradient(135deg, rgba(103, 232, 249, 0.3), rgba(37, 99, 235, 0.3))',
      backgroundImage:
        'linear-gradient(160deg, rgba(7, 18, 38, 0.92) 0%, rgba(15, 39, 73, 0.86) 100%)',
      fontColor: '#F1F5F9',
    },
    library: {
      colors: ['#070F21', '#14233A'],
      shape: 'linear-gradient(160deg, rgba(124, 92, 255, 0.45), rgba(100, 210, 255, 0.28))',
      backgroundImage:
        'linear-gradient(150deg, rgba(5, 18, 37, 0.95) 0%, rgba(11, 25, 49, 0.88) 100%)',
      fontColor: '#E5F1FF',
    },
    other: {
      colors: ['#050B1A', '#0D1424'],
      shape: 'linear-gradient(150deg, rgba(0, 229, 255, 0.4), rgba(124, 92, 255, 0.35))',
      backgroundImage:
        'linear-gradient(135deg, rgba(8, 29, 56, 0.92) 0%, rgba(16, 35, 64, 0.86) 100%)',
      fontColor: '#E4F1FF',
    },
    app: {
      colors: ['#060E20', '#13233A'],
      shape: 'linear-gradient(140deg, rgba(62, 139, 255, 0.5), rgba(0, 229, 255, 0.32))',
      backgroundImage:
        'linear-gradient(150deg, rgba(8, 21, 44, 0.94) 0%, rgba(13, 32, 58, 0.88) 100%)',
      fontColor: '#F8FAFF',
    },
    apis: {
      colors: ['#060E1F', '#0E1B32'],
      shape: 'linear-gradient(150deg, rgba(124, 92, 255, 0.42), rgba(56, 189, 248, 0.26))',
      backgroundImage:
        'linear-gradient(160deg, rgba(6, 17, 33, 0.94) 0%, rgba(12, 28, 54, 0.86) 100%)',
      fontColor: '#F1F7FF',
    },
  };

  return pageThemes[themeId] || pageThemes.other;
};

export const aegisTheme = {
  id: 'aegis-nebula',
  title: 'ÆGIS Nebula',
  variant: 'dark' as const,
  theme: baseTheme,
  Provider: ({ children }: PropsWithChildren<{}>) => (
    <ThemeProvider theme={baseTheme}>
      <CssBaseline />
      {children}
    </ThemeProvider>
  ),
};

