import { makeStyles, Typography } from '@material-ui/core';
import { useTheme } from '@material-ui/core/styles';
import { useSidebarOpenState } from '@backstage/core-components';

const useStyles = makeStyles(theme => ({
  root: {
    display: 'flex',
    alignItems: 'center',
    gap: theme.spacing(2.5),
    padding: theme.spacing(1.5, 0),
    minHeight: theme.spacing(12),
  },
  emblem: {
    height: theme.spacing(7.5),
    width: theme.spacing(7.5),
    flexShrink: 0,
    display: 'block',
  },
  wordmarkWrapper: {
    display: 'flex',
    alignItems: 'center',
    minWidth: 0,
  },
  wordmark: {
    fontWeight: 700,
    fontSize: '2.1rem',
    letterSpacing: '0.22em',
    textTransform: 'uppercase',
    whiteSpace: 'nowrap',
    color: theme.palette.text.primary,
  },
  accent: {
    color: theme.palette.text.secondary,
  },
}));

const LogoFull = () => {
  const classes = useStyles();
  const theme = useTheme();
  const { isOpen } = useSidebarOpenState();
  const paletteMode = (theme.palette as any)?.mode ?? theme.palette.type;
  const isDark = paletteMode === 'dark';
  const primary = theme.palette.text.primary;
  const secondary = isDark ? '#BDBDBD' : '#4A4A4A';

  return (
    <span className={classes.root} aria-label="ÆGIS logo">
      <svg
        className={classes.emblem}
        viewBox="0 0 40 40"
        role="presentation"
        aria-hidden
        focusable="false"
      >
        <path
          d="M20 5.5l12.5 21.65H7.5L20 5.5z"
          fill="none"
          stroke={primary}
          strokeWidth={2.4}
          strokeLinejoin="round"
        />
        <circle cx={20} cy={16.2} r={2.6} fill={primary} />
        <circle
          cx={11.2}
          cy={29.8}
          r={2.8}
          fill="none"
          stroke={secondary}
          strokeWidth={2.4}
        />
        <circle
          cx={28.8}
          cy={29.8}
          r={2.8}
          fill="none"
          stroke={secondary}
          strokeWidth={2.4}
        />
      </svg>
      {isOpen ? (
        <span className={classes.wordmarkWrapper}>
          <Typography component="span" className={classes.wordmark}>
            Æ<span className={classes.accent}>GIS</span>
          </Typography>
        </span>
      ) : null}
    </span>
  );
};

export default LogoFull;
