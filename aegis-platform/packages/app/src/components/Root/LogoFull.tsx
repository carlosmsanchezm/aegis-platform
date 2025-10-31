import { makeStyles, Typography } from '@material-ui/core';
import { useTheme } from '@material-ui/core/styles';

const useStyles = makeStyles(theme => ({
  root: {
    display: 'flex',
    alignItems: 'center',
    gap: theme.spacing(2.5),
    padding: theme.spacing(1.5, 0),
    minHeight: theme.spacing(10),
  },
  image: {
    height: 72,
    width: 'auto',
    display: 'block',
  },
  wordmark: {
    fontWeight: 700,
    fontSize: '1.8rem',
    letterSpacing: '0.22em',
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
        src="/branding/aegis-logo-full.svg"
        alt="ÆGIS emblem"
        className={classes.image}
        style={!isDark ? { filter: 'invert(1) brightness(1.1)' } : undefined}
      />
      <Typography component="span" className={classes.wordmark}>
        ÆGIS
      </Typography>
    </span>
  );
};

export default LogoFull;
