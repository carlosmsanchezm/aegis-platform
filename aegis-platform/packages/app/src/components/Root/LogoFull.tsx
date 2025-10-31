import { makeStyles, Typography } from '@material-ui/core';
import { useTheme } from '@material-ui/core/styles';
import { alpha } from '@material-ui/core/styles/colorManipulator';
import { useSidebarOpenState } from '@backstage/core-components';

const useStyles = makeStyles(theme => ({
  root: {
    display: 'flex',
    alignItems: 'center',
    gap: theme.spacing(2),
    padding: theme.spacing(1.5, 0),
    minHeight: theme.spacing(11),
  },
  emblem: {
    height: theme.spacing(8),
    width: theme.spacing(8),
    flexShrink: 0,
  },
  wordmark: {
    fontWeight: 700,
    fontSize: '1.85rem',
    letterSpacing: '0.16em',
    textTransform: 'uppercase',
    color: theme.palette.text.primary,
    whiteSpace: 'nowrap',
    flexShrink: 1,
    minWidth: 0,
  },
  ae: {
    marginRight: theme.spacing(0.6),
  },
}));

const LogoFull = () => {
  const classes = useStyles();
  const theme = useTheme();
  const paletteMode = (theme.palette as any)?.mode ?? theme.palette.type;
  const isDark = paletteMode === 'dark';
  const { isOpen } = useSidebarOpenState();

  const stroke = theme.palette.text.primary;
  const accent = isDark
    ? alpha(theme.palette.text.primary, 0.12)
    : alpha(theme.palette.text.primary, 0.18);
  const centerFill = theme.palette.text.primary;

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
          fill={accent}
          stroke={stroke}
          strokeWidth={2.4}
          strokeLinejoin="round"
        />
        <circle cx={20} cy={16.2} r={2.6} fill={centerFill} />
        <circle cx={11.2} cy={29.8} r={2.8} fill="none" stroke={stroke} strokeWidth={2.4} />
        <circle cx={28.8} cy={29.8} r={2.8} fill="none" stroke={stroke} strokeWidth={2.4} />
      </svg>
      {isOpen ? (
        <Typography component="span" className={classes.wordmark}>
          <span className={classes.ae}>Æ</span>GIS
        </Typography>
      ) : null}
    </span>
  );
};

export default LogoFull;
