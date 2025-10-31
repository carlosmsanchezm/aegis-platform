import { PropsWithChildren } from 'react';
import { CssBaseline } from '@material-ui/core';
import { createTheme, ThemeProvider } from '@material-ui/core/styles';
import type { BackstageTheme } from '@backstage/theme';

const backgroundGradient =
  'radial-gradient(120% 120% at 10% 20%, rgba(16, 163, 127, 0.18) 0%, rgba(28, 34, 44, 0) 55%), radial-gradient(120% 120% at 90% 10%, rgba(123, 110, 246, 0.16) 0%, rgba(28, 34, 44, 0) 60%)';

const fontFamily = '"Inter", "IBM Plex Sans", "Helvetica Neue", Arial, sans-serif';

const baseTheme = createTheme({
  palette: {
    type: 'dark',
    primary: { main: '#10A37F' },
    secondary: { main: '#7B6EF6' },
    error: { main: '#F97066' },
    warning: { main: '#F7B84B' },
    info: { main: '#60A5FA' },
    success: { main: '#4ADE80' },
    background: {
      default: '#05060A',
      paper: '#0E1117',
    },
    text: {
      primary: '#F6F7F9',
      secondary: '#B6BCC6',
      hint: '#7A8699',
    },
    divider: 'rgba(255, 255, 255, 0.08)',
    navigation: {
      background: '#05060A',
      indicator: '#10A37F',
      color: '#9BA1B2',
      selectedColor: '#F6F7F9',
      navItem: {
        hoverBackground: 'rgba(16, 163, 127, 0.12)',
      },
      submenu: {
        background: '#0B0E14',
      },
    },
    banner: {
      info: '#113F67',
      error: '#4A1F23',
      text: '#F6F7F9',
      link: '#7B6EF6',
    },
    errorBackground: 'rgba(249, 112, 102, 0.12)',
    warningBackground: 'rgba(247, 184, 75, 0.12)',
    infoBackground: 'rgba(96, 165, 250, 0.12)',
  },
  typography: {
    fontFamily,
    h1: {
      fontWeight: 600,
      fontSize: '2.75rem',
      letterSpacing: '-0.015em',
    },
    h2: {
      fontWeight: 600,
      fontSize: '2.25rem',
      letterSpacing: '-0.01em',
    },
    h3: {
      fontWeight: 600,
      fontSize: '1.75rem',
      letterSpacing: '-0.01em',
    },
    subtitle1: {
      fontWeight: 500,
      fontSize: '1rem',
    },
    body1: {
      fontSize: '0.95rem',
      lineHeight: 1.6,
    },
    body2: {
      fontSize: '0.875rem',
      lineHeight: 1.6,
      color: '#A0A7B3',
    },
    button: {
      textTransform: 'none',
      fontWeight: 600,
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
          backgroundColor: '#05060A',
          backgroundImage: backgroundGradient,
          color: '#F6F7F9',
        },
        a: {
          color: '#10A37F',
        },
        '::-webkit-scrollbar': {
          width: 8,
          height: 8,
        },
        '::-webkit-scrollbar-thumb': {
          backgroundColor: 'rgba(123, 110, 246, 0.45)',
          borderRadius: 999,
        },
      },
    },
    MuiPaper: {
      root: {
        backgroundColor: '#0E1117',
        backgroundImage:
          'linear-gradient(145deg, rgba(15, 19, 27, 0.9), rgba(9, 11, 17, 0.9))',
        border: '1px solid rgba(255, 255, 255, 0.05)',
        boxShadow:
          '0 18px 50px rgba(3, 4, 6, 0.55), inset 0 0 0 1px rgba(255, 255, 255, 0.02)',
      },
    },
    MuiButton: {
      root: {
        borderRadius: 999,
        padding: '10px 20px',
      },
      containedPrimary: {
        boxShadow: 'none',
        '&:hover': {
          backgroundColor: '#0E8C6B',
          boxShadow: '0 10px 30px rgba(16, 163, 127, 0.25)',
        },
      },
      outlinedPrimary: {
        borderColor: 'rgba(16, 163, 127, 0.6)',
        '&:hover': {
          borderColor: '#10A37F',
          backgroundColor: 'rgba(16, 163, 127, 0.1)',
        },
      },
    },
    MuiListItem: {
      root: {
        borderRadius: 10,
        '&$selected': {
          backgroundColor: 'rgba(16, 163, 127, 0.12)',
        },
      },
      button: {
        '&:hover': {
          backgroundColor: 'rgba(123, 110, 246, 0.12)',
        },
      },
    },
    MuiDrawer: {
      paper: {
        backgroundColor: '#05060A',
        borderRight: '1px solid rgba(255, 255, 255, 0.05)',
      },
    },
    MuiAppBar: {
      colorPrimary: {
        backgroundColor: '#05060A',
        boxShadow: '0 1px 0 rgba(255, 255, 255, 0.05)',
      },
    },
    MuiChip: {
      root: {
        backgroundColor: 'rgba(16, 163, 127, 0.12)',
        color: '#10A37F',
        borderRadius: 8,
      },
    },
    MuiTableCell: {
      head: {
        color: '#B6BCC6',
      },
      body: {
        borderBottom: '1px solid rgba(255, 255, 255, 0.04)',
      },
    },
    MuiTabs: {
      indicator: {
        height: 3,
        borderRadius: 3,
        backgroundColor: '#10A37F',
      },
    },
  },
}) as BackstageTheme;

baseTheme.getPageTheme = ({ themeId }: { themeId: string }) => {
  const pageThemes: Record<string, any> = {
    home: {
      colors: ['#10A37F', '#7B6EF6'],
      shape:
        'radial-gradient(100% 120% at 20% 20%, rgba(16, 163, 127, 0.25), rgba(16, 163, 127, 0))',
      backgroundImage: backgroundGradient,
      fontColor: '#F6F7F9',
    },
    documentation: {
      colors: ['#7B6EF6', '#60A5FA'],
      shape:
        'radial-gradient(140% 100% at 10% 30%, rgba(123, 110, 246, 0.3), rgba(10, 13, 20, 0))',
      backgroundImage:
        'radial-gradient(120% 120% at 0% 50%, rgba(123, 110, 246, 0.2), rgba(5, 6, 10, 0))',
      fontColor: '#E9EAF0',
    },
    tool: {
      colors: ['#10A37F', '#4ADE80'],
      shape:
        'radial-gradient(120% 120% at 80% 20%, rgba(74, 222, 128, 0.3), rgba(5, 6, 10, 0))',
      backgroundImage:
        'radial-gradient(120% 120% at 70% 10%, rgba(16, 163, 127, 0.25), rgba(5, 6, 10, 0))',
      fontColor: '#F6F7F9',
    },
    service: {
      colors: ['#7B6EF6', '#10A37F'],
      shape:
        'linear-gradient(120deg, rgba(123, 110, 246, 0.24), rgba(16, 163, 127, 0.0))',
      backgroundImage:
        'radial-gradient(120% 120% at 90% 60%, rgba(16, 163, 127, 0.18), rgba(5, 6, 10, 0))',
      fontColor: '#F6F7F9',
    },
    website: {
      colors: ['#60A5FA', '#7B6EF6'],
      shape:
        'radial-gradient(120% 120% at 30% 70%, rgba(96, 165, 250, 0.25), rgba(5, 6, 10, 0))',
      backgroundImage:
        'radial-gradient(120% 120% at 95% 5%, rgba(96, 165, 250, 0.18), rgba(5, 6, 10, 0))',
      fontColor: '#F6F7F9',
    },
    library: {
      colors: ['#10A37F', '#7B6EF6'],
      shape:
        'radial-gradient(120% 120% at 50% 50%, rgba(16, 163, 127, 0.25), rgba(5, 6, 10, 0))',
      backgroundImage:
        'radial-gradient(120% 120% at 15% 15%, rgba(123, 110, 246, 0.16), rgba(5, 6, 10, 0))',
      fontColor: '#F6F7F9',
    },
    app: {
      colors: ['#10A37F', '#60A5FA'],
      shape:
        'radial-gradient(120% 120% at 70% 80%, rgba(96, 165, 250, 0.28), rgba(5, 6, 10, 0))',
      backgroundImage:
        'radial-gradient(120% 120% at 40% 0%, rgba(16, 163, 127, 0.25), rgba(5, 6, 10, 0))',
      fontColor: '#F6F7F9',
    },
    apis: {
      colors: ['#7B6EF6', '#10A37F'],
      shape:
        'radial-gradient(120% 120% at 20% 50%, rgba(123, 110, 246, 0.28), rgba(5, 6, 10, 0))',
      backgroundImage:
        'radial-gradient(120% 120% at 70% 10%, rgba(16, 163, 127, 0.2), rgba(5, 6, 10, 0))',
      fontColor: '#F6F7F9',
    },
    other: {
      colors: ['#10A37F', '#7B6EF6'],
      shape:
        'radial-gradient(120% 120% at 50% 50%, rgba(16, 163, 127, 0.2), rgba(5, 6, 10, 0))',
      backgroundImage: backgroundGradient,
      fontColor: '#F6F7F9',
    },
  };

  return pageThemes[themeId] || pageThemes.other;
};

export const aegisTheme = {
  id: 'aegis-dark',
  title: 'ÆGIS Dark',
  variant: 'dark' as const,
  Provider: ({ children }: PropsWithChildren<{}>) => (
    <ThemeProvider theme={baseTheme}>
      <CssBaseline />
      <div
        style={{
          minHeight: '100vh',
          backgroundImage: backgroundGradient,
          backgroundColor: '#05060A',
        }}
      >
        {children}
      </div>
    </ThemeProvider>
  ),
  theme: baseTheme,
};
