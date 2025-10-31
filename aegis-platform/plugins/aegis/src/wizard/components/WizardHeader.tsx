import { Typography } from '@material-ui/core';
import { makeStyles } from '@material-ui/core/styles';

const useStyles = makeStyles(theme => ({
  root: {
    marginBottom: theme.spacing(4),
  },
  eyebrow: {
    fontSize: '0.75rem',
    letterSpacing: '0.18em',
    textTransform: 'uppercase',
    color: theme.palette.text.secondary,
    marginBottom: theme.spacing(1),
    display: 'inline-block',
  },
  title: {
    fontWeight: 600,
    marginBottom: theme.spacing(1),
  },
  subtitle: {
    color: theme.palette.text.secondary,
    maxWidth: 640,
  },
}));

export type WizardHeaderProps = {
  title: string;
  subtitle?: string;
  stepLabel?: string;
};

export const WizardHeader = ({ title, subtitle, stepLabel }: WizardHeaderProps) => {
  const classes = useStyles();

  return (
    <div className={classes.root}>
      {stepLabel ? (
        <Typography variant="caption" component="span" className={classes.eyebrow}>
          {stepLabel}
        </Typography>
      ) : null}
      <Typography variant="h3" className={classes.title}>
        {title}
      </Typography>
      {subtitle ? (
        <Typography variant="body1" className={classes.subtitle}>
          {subtitle}
        </Typography>
      ) : null}
    </div>
  );
};
