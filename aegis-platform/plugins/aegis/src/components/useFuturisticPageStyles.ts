import { makeStyles } from '@material-ui/core/styles';
import { alpha } from '@material-ui/core/styles/colorManipulator';

export const useFuturisticPageStyles = makeStyles(theme => ({
  page: {
    position: 'relative',
    overflow: 'hidden',
    minHeight: '100%',
  },
  background: {
    position: 'absolute',
    inset: '-30% -15% auto',
    height: '160%',
    background: `radial-gradient(circle at 10% 20%, ${alpha(
      theme.palette.primary.main,
      0.25,
    )} 0, transparent 55%), radial-gradient(circle at 80% 10%, ${alpha(
      theme.palette.secondary.main,
      0.22,
    )} 0, transparent 60%), radial-gradient(circle at 50% 80%, ${alpha(
      '#4b9eff',
      0.2,
    )} 0, transparent 55%)`,
    filter: 'blur(40px)',
    opacity: 0.75,
    zIndex: 0,
    pointerEvents: 'none',
  },
  header: {
    position: 'relative',
    zIndex: 1,
    background: `linear-gradient(145deg, ${alpha('#07162b', 0.92)}, ${alpha(
      '#040b17',
      0.88,
    )})`,
    borderBottom: `1px solid ${alpha(theme.palette.primary.main, 0.25)}`,
    boxShadow: '0 36px 90px rgba(20, 255, 220, 0.18)',
  },
  headerTitle: {
    letterSpacing: '0.3em',
    color: theme.palette.primary.light,
    textShadow: '0 0 24px rgba(83, 251, 224, 0.3)',
  },
  headerSubtitle: {
    color: alpha('#dfe9ff', 0.78),
    letterSpacing: '0.12em',
    fontFamily: `'IBM Plex Mono', monospace`,
  },
  headerActions: {
    marginTop: theme.spacing(2),
    display: 'flex',
    gap: theme.spacing(2),
    flexWrap: 'wrap',
  },
  content: {
    position: 'relative',
    zIndex: 1,
    marginTop: theme.spacing(-5),
  },
  holoPanel: {
    background: `linear-gradient(160deg, ${alpha('#0d1f3b', 0.94)}, ${alpha(
      '#050b16',
      0.88,
    )})`,
    border: `1px solid ${alpha(theme.palette.primary.main, 0.35)}`,
    boxShadow: '0 38px 110px rgba(30, 255, 229, 0.16)',
    borderRadius: 20,
    padding: theme.spacing(4),
  },
  selectionCard: {
    background: `linear-gradient(160deg, ${alpha('#071a33', 0.92)}, ${alpha(
      '#040b16',
      0.88,
    )})`,
    border: `1px solid ${alpha('#4b9eff', 0.28)}`,
    boxShadow: '0 25px 70px rgba(75, 158, 255, 0.16)',
    borderRadius: 18,
    transition: 'transform 220ms ease, box-shadow 220ms ease, border 220ms ease',
    '&:hover': {
      transform: 'translateY(-4px)',
      boxShadow: '0 45px 120px rgba(75, 158, 255, 0.22)',
    },
  },
  selectionCardActive: {
    borderColor: alpha(theme.palette.primary.main, 0.9),
    boxShadow: '0 0 45px rgba(83, 251, 224, 0.45)',
  },
  selectionCardTitle: {
    letterSpacing: '0.18em',
    textTransform: 'uppercase',
    color: alpha('#dfe9ff', 0.85),
  },
  selectionCardSubtitle: {
    color: alpha('#dfe9ff', 0.65),
    fontFamily: `'IBM Plex Mono', monospace`,
    letterSpacing: '0.1em',
  },
  inlineStat: {
    display: 'inline-flex',
    alignItems: 'center',
    gap: theme.spacing(1.5),
    letterSpacing: '0.12em',
    textTransform: 'uppercase',
    fontFamily: `'IBM Plex Mono', monospace`,
    color: theme.palette.primary.light,
  },
  callout: {
    marginTop: theme.spacing(3),
    padding: theme.spacing(2.5, 3),
    borderLeft: `3px solid ${alpha(theme.palette.secondary.main, 0.7)}`,
    background: alpha('#0a1325', 0.9),
    color: alpha('#dfe9ff', 0.85),
    fontFamily: `'IBM Plex Mono', monospace`,
    letterSpacing: '0.08em',
  },
  stepper: {
    background: 'transparent',
    padding: theme.spacing(2, 0),
    '& .MuiStepLabel-label': {
      color: alpha('#dfe9ff', 0.72),
      textTransform: 'uppercase',
      letterSpacing: '0.16em',
      fontSize: '0.75rem',
    },
    '& .MuiStepIcon-root': {
      color: alpha(theme.palette.primary.main, 0.35),
      '&.MuiStepIcon-active': {
        color: theme.palette.primary.main,
      },
      '&.MuiStepIcon-completed': {
        color: theme.palette.secondary.main,
      },
    },
  },
  tablePaper: {
    background: `linear-gradient(155deg, ${alpha('#07203e', 0.95)}, ${alpha(
      '#030811',
      0.92,
    )})`,
    border: `1px solid ${alpha('#4b9eff', 0.35)}`,
    boxShadow: '0 32px 95px rgba(75, 158, 255, 0.16)',
    borderRadius: 20,
    padding: theme.spacing(2.5),
  },
  sectionTitle: {
    letterSpacing: '0.22em',
    color: alpha('#dfe9ff', 0.85),
    textTransform: 'uppercase',
    marginBottom: theme.spacing(2.5),
  },
}));
