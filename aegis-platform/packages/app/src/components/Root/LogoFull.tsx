import { makeStyles } from '@material-ui/core';
import { useTheme } from '@material-ui/core/styles';

const useStyles = makeStyles(theme => ({
  svg: {
    width: 'auto',
    height: 34,
    transition: 'fill 200ms ease, color 200ms ease',
    color: theme.palette.text.primary,
  },
}));

const LogoFull = () => {
  const classes = useStyles();
  const theme = useTheme();
  const tone = theme.palette.type === 'dark' ? '#F5F1E8' : '#1B1F2D';
  const accent = tone;

  return (
    <svg
      className={classes.svg}
      viewBox="0 0 160 40"
      role="img"
      aria-labelledby="aegisLogoTitle"
      focusable="false"
    >
      <title id="aegisLogoTitle">ÆGIS</title>
      <g
        fill="none"
        stroke={accent}
        strokeWidth={2.2}
        strokeLinecap="round"
        strokeLinejoin="round"
      >
        <path d="M18 8 L8 28 H28 Z" />
      </g>
      <circle cx="18" cy="8" r="3" fill={accent} />
      <circle cx="8" cy="28" r="3" fill={accent} />
      <circle cx="28" cy="28" r="3" fill={accent} />
      <circle cx="18" cy="21" r="2.6" fill={accent} />
      <line x1="18" y1="8" x2="18" y2="21" stroke={accent} strokeWidth={2.2} />
      <line x1="8" y1="28" x2="18" y2="21" stroke={accent} strokeWidth={2.2} />
      <line x1="28" y1="28" x2="18" y2="21" stroke={accent} strokeWidth={2.2} />
      <text
        x="44"
        y="26"
        fill={tone}
        fontFamily="'Inter', 'SF Pro Display', 'Helvetica Neue', Arial, sans-serif"
        fontWeight={700}
        fontSize="18"
        letterSpacing="0.08em"
      >
        ÆGIS
      </text>
    </svg>
  );
};

export default LogoFull;
