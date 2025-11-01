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
  image: {
    height: theme.spacing(12),
    width: 'auto',
    maxWidth: '100%',
    flexShrink: 0,
    display: 'block',
    objectFit: 'contain',
  },
  wordmark: {
    fontWeight: 700,
    fontSize: '2.35rem',
    letterSpacing: '0.2em',
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
  const { isOpen } = useSidebarOpenState();
  const logoSrc = isDark
    ? '/branding/aegis-logo-full-dark.svg'
    : '/branding/aegis-logo-full.svg';

  return (
    <span className={classes.root} aria-label="ÆGIS logo">
      <img
        src={logoSrc}
        alt="ÆGIS emblem"
        className={classes.image}
      />
      {isOpen ? (
        <Typography component="span" className={classes.wordmark}>
          ÆGIS
        </Typography>
      ) : null}
    </span>
  );
};

export default LogoFull;
