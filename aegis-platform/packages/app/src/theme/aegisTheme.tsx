import React, { PropsWithChildren } from 'react';
import { ThemeProvider, CssBaseline } from '@material-ui/core';
import { createUnifiedTheme } from '@backstage/theme';
import type { BackstageTheme } from '@backstage/theme';

const baseTheme = createUnifiedTheme({
  defaultPageTheme: 'home',
  palette: {
    type: 'dark',
    primary: { main: '#3D9BFF' },
    secondary: { main: '#6F4BFF' },
    error: { main: '#FF4D67' },
    warning: { main: '#FFB347' },
    success: { main: '#4DE2A1' },
    background: {
      default: '#030712',
      paper: 'rgba(9, 16, 35, 0.88)',
    },
    text: {
      primary: '#E6F1FF',
      secondary: '#9BA9C9',
      hint: '#6D7A9A',
    },
    divider: 'rgba(93, 117, 167, 0.4)',
    navigation: {
      background: '#040918',
      indicator: '#4AE8FF',
      color: '#92A2C5',
      selectedColor: '#F4FBFF',
      navItem: {
        hoverBackground: 'rgba(74, 232, 255, 0.08)',
      },
      submenu: {
        background: 'rgba(4, 12, 28, 0.95)',
      },
    },
    banner: {
      info: '#133B5C',
      error: '#4A1F2F',
      text: '#D6ECFF',
      link: '#7FDBFF',
    },
    errorBackground: 'rgba(255, 77, 103, 0.12)',
    warningBackground: 'rgba(255, 179, 71, 0.12)',
    infoBackground: 'rgba(74, 232, 255, 0.12)',
  },
  fontFamily: '"Inter", "Space Grotesk", "Roboto", "Helvetica", "Arial", sans-serif',
  typography: {
    h1: {
      fontWeight: 600,
      letterSpacing: '-0.5px',
    },
    h2: {
      fontWeight: 600,
      letterSpacing: '-0.3px',
    },
    h3: {
      fontWeight: 600,
    },
    button: {
      fontWeight: 600,
      letterSpacing: '0.08em',
      textTransform: 'uppercase',
    },
    subtitle1: {
      color: '#9BA9C9',
    },
    subtitle2: {
      color: '#6D7A9A',
    },
    body1: {
      color: '#DCE6FF',
    },
    body2: {
      color: '#B5C2E3',
    },
  },
  shape: {
    borderRadius: 14,
  },
  overrides: {
    MuiCssBaseline: {
      '@global': {
        body: {
          background:
            'radial-gradient(circle at 20% 20%, rgba(74, 232, 255, 0.09), transparent 55%), radial-gradient(circle at 80% 0%, rgba(155, 90, 255, 0.12), transparent 50%), linear-gradient(180deg, #03050F 0%, #02030A 100%)',
          backgroundAttachment: 'fixed',
          color: '#E6F1FF',
          minHeight: '100vh',
        },
        '#root': {
          backgroundColor: 'transparent',
        },
        a: {
          color: '#4AE8FF',
        },
        '::-webkit-scrollbar': {
          width: 10,
          height: 10,
        },
        '::-webkit-scrollbar-thumb': {
          backgroundColor: 'rgba(74, 232, 255, 0.24)',
          borderRadius: 8,
        },
        '::-webkit-scrollbar-track': {
          backgroundColor: 'rgba(12, 20, 45, 0.4)',
        },
      },
    },
    MuiPaper: {
      root: {
        backgroundColor: 'rgba(6, 14, 32, 0.86)',
        backdropFilter: 'blur(22px)',
        border: '1px solid rgba(74, 232, 255, 0.12)',
      },
      elevation1: {
        boxShadow: '0 24px 48px rgba(2, 13, 37, 0.45)',
      },
    },
    MuiCard: {
      root: {
        borderRadius: 18,
        background:
          'linear-gradient(180deg, rgba(14, 28, 64, 0.9) 0%, rgba(7, 18, 42, 0.88) 100%)',
        border: '1px solid rgba(106, 178, 255, 0.16)',
        boxShadow: '0 32px 54px rgba(2, 12, 38, 0.38)',
      },
    },
    MuiDrawer: {
      paper: {
        backgroundColor: '#040918',
        borderRight: '1px solid rgba(74, 232, 255, 0.1)',
        backdropFilter: 'blur(18px)',
      },
    },
    MuiAppBar: {
      colorPrimary: {
        backgroundImage:
          'linear-gradient(90deg, rgba(23, 46, 91, 0.9) 0%, rgba(36, 17, 74, 0.9) 100%)',
        boxShadow: '0 10px 24px rgba(4, 15, 40, 0.4)',
      },
    },
    MuiButton: {
      root: {
        borderRadius: 12,
        textTransform: 'none',
        fontWeight: 600,
        padding: '10px 20px',
        letterSpacing: '0.04em',
        transition: 'all 200ms ease',
        backgroundImage:
          'linear-gradient(90deg, rgba(61, 155, 255, 0.18), rgba(111, 75, 255, 0.18))',
        border: '1px solid rgba(61, 155, 255, 0.4)',
        '&:hover': {
          backgroundImage:
            'linear-gradient(90deg, rgba(74, 232, 255, 0.3), rgba(136, 94, 255, 0.28))',
          boxShadow: '0 16px 30px rgba(17, 42, 94, 0.48)',
        },
      },
      containedPrimary: {
        backgroundImage:
          'linear-gradient(90deg, #3D9BFF 0%, #6F4BFF 100%)',
        boxShadow: '0 14px 26px rgba(42, 110, 208, 0.45)',
        '&:hover': {
          backgroundImage:
            'linear-gradient(90deg, #4AE8FF 0%, #8E6BFF 100%)',
        },
      },
      outlinedPrimary: {
        borderColor: 'rgba(74, 232, 255, 0.5)',
        '&:hover': {
          borderColor: '#4AE8FF',
          backgroundColor: 'rgba(74, 232, 255, 0.08)',
        },
      },
    },
    MuiChip: {
      root: {
        backgroundColor: 'rgba(74, 232, 255, 0.12)',
        color: '#C9E7FF',
        border: '1px solid rgba(74, 232, 255, 0.3)',
      },
    },
    MuiTabs: {
      indicator: {
        height: 3,
        borderRadius: 3,
        background:
          'linear-gradient(90deg, #4AE8FF 0%, #8E6BFF 100%)',
      },
    },
    MuiTab: {
      root: {
        textTransform: 'none',
        fontWeight: 600,
        minHeight: 48,
      },
      textColorInherit: {
        color: '#9BA9C9',
        '&$selected': {
          color: '#FFFFFF',
        },
      },
    },
    MuiListItem: {
      button: {
        borderRadius: 10,
        margin: '2px 6px',
        '&:hover': {
          backgroundColor: 'rgba(74, 232, 255, 0.12)',
        },
        '&$selected': {
          backgroundColor: 'rgba(74, 232, 255, 0.22)',
          boxShadow: 'inset 0 0 0 1px rgba(74, 232, 255, 0.4)',
        },
      },
    },
    MuiListItemIcon: {
      root: {
        color: '#6F89C9',
        minWidth: 36,
      },
    },
    MuiTableCell: {
      root: {
        borderBottom: '1px solid rgba(93, 117, 167, 0.3)',
      },
      head: {
        color: '#F0F6FF',
        fontWeight: 600,
      },
    },
    MuiTooltip: {
      tooltip: {
        backgroundColor: 'rgba(13, 27, 55, 0.92)',
        color: '#E6F1FF',
        borderRadius: 10,
        border: '1px solid rgba(74, 232, 255, 0.2)',
      },
    },
  },
}) as BackstageTheme;

baseTheme.getPageTheme = ({ themeId }: { themeId: string }) => {
  const pageThemes: Record<string, any> = {
    home: {
      colors: ['#3D9BFF', '#6F4BFF'],
      shape: 'linear-gradient(135deg, rgba(61, 155, 255, 0.28), rgba(111, 75, 255, 0.22))',
      backgroundImage:
        'radial-gradient(circle at 15% 20%, rgba(74, 232, 255, 0.25), transparent 60%), radial-gradient(circle at 80% 0%, rgba(155, 90, 255, 0.35), transparent 50%), linear-gradient(180deg, #03050F 0%, #040b1a 100%)',
      fontColor: '#E6F1FF',
    },
    documentation: {
      colors: ['#4AE8FF', '#3D9BFF'],
      shape: 'linear-gradient(135deg, rgba(20, 64, 132, 0.38), rgba(28, 82, 160, 0.26))',
      backgroundImage:
        'radial-gradient(circle at 10% 0%, rgba(34, 142, 230, 0.32), transparent 55%), linear-gradient(180deg, #03050F 0%, #06122A 100%)',
      fontColor: '#E6F1FF',
    },
    tool: {
      colors: ['#6F4BFF', '#A75CFF'],
      shape: 'linear-gradient(135deg, rgba(56, 28, 118, 0.4), rgba(126, 56, 230, 0.34))',
      backgroundImage:
        'radial-gradient(circle at 80% 0%, rgba(155, 90, 255, 0.32), transparent 45%), linear-gradient(180deg, #03050F 0%, #090e1f 100%)',
      fontColor: '#F2EAFF',
    },
    service: {
      colors: ['#4DE2A1', '#3D9BFF'],
      shape: 'linear-gradient(135deg, rgba(36, 110, 156, 0.34), rgba(34, 156, 130, 0.32))',
      backgroundImage:
        'radial-gradient(circle at 30% 90%, rgba(41, 162, 122, 0.36), transparent 55%), linear-gradient(180deg, #03050F 0%, #05142a 100%)',
      fontColor: '#E6F1FF',
    },
    website: {
      colors: ['#3D9BFF', '#FF6B9E'],
      shape: 'linear-gradient(135deg, rgba(61, 155, 255, 0.32), rgba(255, 107, 158, 0.3))',
      backgroundImage:
        'radial-gradient(circle at 40% 40%, rgba(255, 107, 158, 0.32), transparent 50%), linear-gradient(180deg, #03050F 0%, #090f21 100%)',
      fontColor: '#FFEFF7',
    },
    library: {
      colors: ['#6F4BFF', '#4AE8FF'],
      shape: 'linear-gradient(135deg, rgba(75, 44, 150, 0.3), rgba(28, 112, 180, 0.32))',
      backgroundImage:
        'radial-gradient(circle at 75% 40%, rgba(124, 74, 242, 0.28), transparent 52%), linear-gradient(180deg, #03050F 0%, #051023 100%)',
      fontColor: '#F3EBFF',
    },
    other: {
      colors: ['#3D9BFF', '#4DE2A1'],
      shape: 'linear-gradient(135deg, rgba(42, 110, 208, 0.3), rgba(34, 156, 130, 0.28))',
      backgroundImage:
        'radial-gradient(circle at 60% 10%, rgba(74, 232, 255, 0.28), transparent 50%), linear-gradient(180deg, #03050F 0%, #040c1b 100%)',
      fontColor: '#E6F1FF',
    },
    app: {
      colors: ['#3D9BFF', '#6F4BFF'],
      shape: 'linear-gradient(135deg, rgba(23, 82, 160, 0.34), rgba(52, 29, 115, 0.32))',
      backgroundImage:
        'radial-gradient(circle at 15% 80%, rgba(74, 232, 255, 0.25), transparent 40%), linear-gradient(180deg, #03050F 0%, #050d20 100%)',
      fontColor: '#E6F1FF',
    },
    apis: {
      colors: ['#4AE8FF', '#A75CFF'],
      shape: 'linear-gradient(135deg, rgba(28, 112, 180, 0.34), rgba(126, 56, 230, 0.3))',
      backgroundImage:
        'radial-gradient(circle at 20% 0%, rgba(74, 232, 255, 0.28), transparent 50%), radial-gradient(circle at 90% 60%, rgba(167, 92, 255, 0.32), transparent 45%), linear-gradient(180deg, #03050F 0%, #060e1f 100%)',
      fontColor: '#F3F6FF',
    },
  };

  return pageThemes[themeId] || pageThemes.other;
};

export const aegisTheme = {
  id: 'aegis-nebula',
  title: 'ÆGIS Nebula',
  variant: 'dark' as const,
  Provider: ({ children }: PropsWithChildren<{}>) => (
    <ThemeProvider theme={baseTheme}>
      <CssBaseline />
      {children}
    </ThemeProvider>
  ),
  theme: baseTheme,
};

