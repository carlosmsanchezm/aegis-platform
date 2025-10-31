import { makeStyles } from '@material-ui/core';
import { useTheme } from '@material-ui/core/styles';

const useStyles = makeStyles(theme => ({
  svg: {
    width: 32,
    height: 32,
    color: theme.palette.text.primary,
    transition: 'color 200ms ease',
  },
}));

const LogoIcon = () => {
  const classes = useStyles();
  const theme = useTheme();
  const tone = theme.palette.type === 'dark' ? '#F5F1E8' : '#1B1F2D';

  return (
    <svg
      className={classes.svg}
      viewBox="0 0 36 36"
      role="img"
      aria-hidden="true"
      focusable="false"
    >
      <g
        fill="none"
        stroke={tone}
        strokeWidth={2.2}
        strokeLinecap="round"
        strokeLinejoin="round"
      >
        <path d="M18 6 L6.5 28.5 H29.5 Z" />
        <line x1="18" y1="6" x2="18" y2="20" />
        <line x1="6.5" y1="28.5" x2="18" y2="20" />
        <line x1="29.5" y1="28.5" x2="18" y2="20" />
      </g>
      <circle cx="18" cy="6" r="3" fill={tone} />
      <circle cx="6.5" cy="28.5" r="3" fill={tone} />
      <circle cx="29.5" cy="28.5" r="3" fill={tone} />
      <circle cx="18" cy="20" r="2.6" fill={tone} />
    </svg>
  );
};

export default LogoIcon;
