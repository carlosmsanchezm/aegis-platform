import { makeStyles, Typography } from '@material-ui/core';
import { useTheme } from '@material-ui/core/styles';

const useStyles = makeStyles(theme => ({
  root: {
    display: 'flex',
    alignItems: 'center',
    gap: theme.spacing(3),
    padding: theme.spacing(1.75, 0),
    minHeight: theme.spacing(11),
  },
  image: {
    height: 88,
    width: 'auto',
    display: 'block',
  },
  wordmark: {
    fontWeight: 700,
    fontSize: '2.2rem',
    letterSpacing: '0.18em',
    textTransform: 'uppercase',
    color: theme.palette.text.primary,
    whiteSpace: 'nowrap',
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
        style={
          isDark
            ? { filter: 'brightness(0) invert(1)' }
            : undefined
        }
      />
      <Typography component="span" className={classes.wordmark}>
        ÆGIS
      </Typography>
    </span>
  );
};

export default LogoFull;
