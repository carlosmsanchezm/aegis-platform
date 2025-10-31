import { ReactNode } from 'react';
import { Card, CardActionArea, CardContent, Typography } from '@material-ui/core';
import { makeStyles } from '@material-ui/core/styles';
import CheckCircleIcon from '@material-ui/icons/CheckCircle';
import clsx from 'clsx';

const useStyles = makeStyles(theme => ({
  root: {
    position: 'relative',
    height: '100%',
    borderRadius: theme.shape.borderRadius,
    border: `1px solid rgba(0,0,0,${theme.palette.type === 'dark' ? 0.15 : 0.08})`,
    backgroundColor: theme.palette.background.paper,
    transition: 'border-color 200ms ease, box-shadow 200ms ease, transform 200ms ease',
  },
  selected: {
    borderColor: theme.palette.primary.main,
    boxShadow: `0 0 0 2px ${theme.palette.primary.main}`,
    transform: 'translateY(-2px)',
  },
  action: {
    height: '100%',
    alignItems: 'stretch',
  },
  content: {
    display: 'flex',
    flexDirection: 'column',
    alignItems: 'flex-start',
    gap: theme.spacing(1.5),
    padding: theme.spacing(3),
  },
  iconWrapper: {
    display: 'inline-flex',
    alignItems: 'center',
    justifyContent: 'center',
    width: 48,
    height: 48,
    borderRadius: theme.shape.borderRadius,
    background:
      theme.palette.type === 'dark'
        ? 'rgba(155,135,255,0.12)'
        : 'rgba(79,70,229,0.12)',
    color: theme.palette.primary.main,
  },
  title: {
    fontWeight: 600,
    letterSpacing: '-0.01em',
  },
  description: {
    color: theme.palette.text.secondary,
    lineHeight: 1.55,
  },
  meta: {
    marginTop: 'auto',
    fontWeight: 500,
    color: theme.palette.text.secondary,
  },
  check: {
    position: 'absolute',
    top: theme.spacing(2),
    right: theme.spacing(2),
    color: theme.palette.primary.main,
  },
  compactContent: {
    padding: theme.spacing(2.5),
    gap: theme.spacing(1),
  },
  compactTitle: {
    fontSize: '1rem',
  },
  compactDescription: {
    fontSize: '0.875rem',
  },
}));

export type TemplateCardProps = {
  title: string;
  description: string;
  icon: ReactNode;
  meta?: ReactNode;
  selected?: boolean;
  disabled?: boolean;
  layout?: 'regular' | 'compact';
  onSelect?: () => void;
};

export const TemplateCard = ({
  title,
  description,
  icon,
  meta,
  selected = false,
  disabled = false,
  layout = 'regular',
  onSelect,
}: TemplateCardProps) => {
  const classes = useStyles();

  return (
    <Card className={clsx(classes.root, { [classes.selected]: selected })} elevation={0}>
      <CardActionArea
        className={classes.action}
        onClick={onSelect}
        disabled={disabled}
        aria-pressed={selected}
      >
        <CardContent
          className={clsx(classes.content, {
            [classes.compactContent]: layout === 'compact',
          })}
        >
          <span className={classes.iconWrapper}>{icon}</span>
          <div>
            <Typography
              variant="h6"
              className={clsx(classes.title, {
                [classes.compactTitle]: layout === 'compact',
              })}
            >
              {title}
            </Typography>
            <Typography
              variant="body2"
              className={clsx(classes.description, {
                [classes.compactDescription]: layout === 'compact',
              })}
            >
              {description}
            </Typography>
          </div>
          {meta ? (
            <Typography variant="body2" className={classes.meta}>
              {meta}
            </Typography>
          ) : null}
        </CardContent>
      </CardActionArea>
      {selected ? <CheckCircleIcon className={classes.check} /> : null}
    </Card>
  );
};
