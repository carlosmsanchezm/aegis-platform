import { makeStyles, Typography } from '@material-ui/core';
import { useTheme } from '@material-ui/core/styles';

const useStyles = makeStyles(theme => ({
  root: {
    display: 'flex',
    alignItems: 'center',
    gap: theme.spacing(1.75),
    padding: theme.spacing(1, 0),
    minHeight: theme.spacing(7),
  },
  image: {
    height: 48,
    width: 'auto',
    display: 'block',
  },
  wordmark: {
    fontWeight: 700,
    fontSize: '1.35rem',
    letterSpacing: '0.18em',
    textTransform: 'uppercase',
    color: theme.palette.text.primary,
  },
}));

const LogoFull = () => {
  const classes = useStyles();
  const theme = useTheme();
  const paletteMode = (theme.palette as any)?.mode ?? theme.palette.type;
  const isDark = paletteMode === 'dark';

  return (
    <span className={classes.root} aria-label="ÆGIS logo">
      <img
        src="/branding/aegis-logo-full.png"
        alt="ÆGIS emblem"
        className={classes.image}
        style={isDark ? { filter: 'invert(1) brightness(1.1)' } : undefined}
      />
      <Typography component="span" className={classes.wordmark}>
        ÆGIS
      </Typography>
    </span>
  );
};

export default LogoFull;
