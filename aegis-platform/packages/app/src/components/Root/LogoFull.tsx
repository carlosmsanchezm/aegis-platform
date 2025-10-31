import { makeStyles, Typography } from '@material-ui/core';

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
  return (
    <span className={classes.root} aria-label="ÆGIS logo">
      <img
        src="/branding/aegis-logo-full.svg"
        alt="ÆGIS emblem"
        className={classes.image}
      />
      <Typography component="span" className={classes.wordmark}>
        ÆGIS
      </Typography>
    </span>
  );
};

export default LogoFull;
