import { createTheme, darkTheme, genPageTheme, pageTheme } from '@backstage/theme';
import { alpha } from '@material-ui/core/styles/colorManipulator';

const neonCyan = '#53fbe0';
const neonMagenta = '#ff3ef5';
const deepNavy = '#050b16';

const holographicOverlay = `radial-gradient(circle at 20% 20%, ${alpha(
  neonCyan,
  0.22,
)} 0, transparent 55%), radial-gradient(circle at 80% 10%, ${alpha(
  neonMagenta,
  0.18,
)} 0, transparent 50%), radial-gradient(circle at 50% 80%, ${alpha(
  '#3b6dff',
  0.18,
)} 0, transparent 55%)`;

const theme = createTheme({
  palette: {
    ...darkTheme.palette,
    primary: {
      light: '#7afff1',
      main: neonCyan,
      dark: '#1dd1b5',
      contrastText: '#010203',
    },
    secondary: {
      light: '#ff69ff',
      main: neonMagenta,
      dark: '#b100b7',
      contrastText: '#010203',
    },
    background: {
      default: deepNavy,
      paper: alpha('#0d152b', 0.85),
    },
    error: {
      light: '#ff8080',
      main: '#ff4d4d',
      dark: '#c20030',
      contrastText: '#ffffff',
    },
    warning: {
      light: '#ffd666',
      main: '#ffb800',
      dark: '#ff8f00',
      contrastText: '#0a101f',
    },
    success: {
      light: '#66ffcc',
      main: '#33ffad',
      dark: '#00c97c',
      contrastText: '#021011',
    },
    info: {
      light: '#76b9ff',
      main: '#4b9eff',
      dark: '#1c6bd1',
      contrastText: '#010203',
    },
  },
  typography: {
    ...darkTheme.typography,
    fontFamily: `'Space Grotesk', 'IBM Plex Sans', 'Roboto', 'Helvetica', 'Arial', sans-serif`,
    h1: {
      fontFamily: `'Space Grotesk', sans-serif`,
      fontWeight: 700,
      letterSpacing: '0.14em',
      textTransform: 'uppercase',
    },
    h2: {
      fontFamily: `'Space Grotesk', sans-serif`,
      fontWeight: 600,
      letterSpacing: '0.12em',
      textTransform: 'uppercase',
    },
    h3: {
      fontFamily: `'Space Grotesk', sans-serif`,
      fontWeight: 600,
      letterSpacing: '0.08em',
    },
    button: {
      textTransform: 'uppercase',
      letterSpacing: '0.2em',
      fontWeight: 600,
    },
    subtitle1: {
      fontFamily: `'IBM Plex Mono', monospace`,
      letterSpacing: '0.08em',
    },
    subtitle2: {
      fontFamily: `'IBM Plex Mono', monospace`,
      letterSpacing: '0.06em',
    },
    overline: {
      fontFamily: `'IBM Plex Mono', monospace`,
      fontWeight: 600,
      letterSpacing: '0.3em',
    },
  },
  shape: {
    borderRadius: 14,
  },
  overrides: {
    MuiCssBaseline: {
      '@global': {
        body: {
          backgroundColor: deepNavy,
          backgroundImage: `${holographicOverlay}, radial-gradient(circle at 0% 0%, ${alpha(
            '#09204a',
            0.6,
          )} 0, transparent 70%)`,
          backgroundAttachment: 'fixed',
          color: '#dfe9ff',
          minHeight: '100vh',
        },
        '#root': {
          backgroundColor: 'transparent',
        },
      },
    },
    MuiPaper: {
      root: {
        backgroundImage: 'linear-gradient(135deg, rgba(15,35,65,0.95), rgba(4,12,24,0.9))',
        border: `1px solid ${alpha('#3ef7ff', 0.25)}`,
        boxShadow: '0 15px 45px rgba(20, 220, 240, 0.15)',
        backdropFilter: 'blur(16px)',
      },
      rounded: {
        borderRadius: 18,
      },
    },
    MuiCard: {
      root: {
        background: 'linear-gradient(140deg, rgba(13,30,55,0.95), rgba(4,10,24,0.92))',
        border: `1px solid ${alpha('#53fbe0', 0.35)}`,
        boxShadow: '0 30px 65px rgba(15, 255, 218, 0.1)',
      },
    },
    MuiButton: {
      root: {
        borderRadius: 999,
        padding: '10px 32px',
        backdropFilter: 'blur(8px)',
      },
      containedPrimary: {
        boxShadow: '0 0 25px rgba(83, 251, 224, 0.45)',
      },
      outlinedPrimary: {
        borderColor: alpha(neonCyan, 0.65),
        color: neonCyan,
        '&:hover': {
          borderColor: neonCyan,
          boxShadow: '0 0 20px rgba(83, 251, 224, 0.4)',
          backgroundColor: alpha(neonCyan, 0.08),
        },
      },
    },
    MuiTableCell: {
      root: {
        borderBottom: `1px solid ${alpha('#3ef7ff', 0.15)}`,
      },
      head: {
        fontFamily: `'IBM Plex Mono', monospace`,
        letterSpacing: '0.18em',
        color: '#8af9ff',
      },
    },
    MuiTypography: {
      colorTextSecondary: {
        color: alpha('#dfe9ff', 0.72),
      },
    },
    MuiTabs: {
      indicator: {
        height: 4,
        borderRadius: 4,
        background: `linear-gradient(90deg, ${neonCyan}, ${neonMagenta})`,
      },
    },
    MuiTab: {
      textColorPrimary: {
        '&.Mui-selected': {
          color: neonCyan,
        },
      },
    },
  },
});

theme.pageTheme = {
  ...pageTheme,
  aegisMission: genPageTheme({
    colors: ['#040c18', '#051124'],
    shape: 'round',
  }),
  aegisWorkspace: genPageTheme({
    colors: ['#07162b', '#0b1d37'],
    shape: 'round',
  }),
  aegisIntel: genPageTheme({
    colors: ['#070f22', '#111c33'],
    shape: 'wave',
  }),
};

export const aegisFuturisticTheme = theme;
