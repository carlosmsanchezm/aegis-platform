import { Grid } from '@material-ui/core';
import CodeIcon from '@material-ui/icons/Code';
import StorageIcon from '@material-ui/icons/Storage';
import WhatshotIcon from '@material-ui/icons/Whatshot';
import FlashOnIcon from '@material-ui/icons/FlashOn';
import DeveloperModeIcon from '@material-ui/icons/DeveloperMode';
import StarsIcon from '@material-ui/icons/Stars';
import { TemplateCard } from '../components/TemplateCard';
import { TemplateId, WorkspaceTemplate } from '../types';

const iconForTemplate = (kind: WorkspaceTemplate['icon']) => {
  switch (kind) {
    case 'python':
      return <CodeIcon fontSize="large" />;
    case 'data':
      return <StorageIcon fontSize="large" />;
    case 'pytorch':
      return <WhatshotIcon fontSize="large" />;
    case 'spark':
      return <FlashOnIcon fontSize="large" />;
    case 'terminal':
      return <DeveloperModeIcon fontSize="large" />;
    case 'custom':
    default:
      return <StarsIcon fontSize="large" />;
  }
};

export type StepTemplatesProps = {
  templates: WorkspaceTemplate[];
  selectedTemplateId?: TemplateId;
  onSelect: (id: TemplateId) => void;
};

export const StepTemplates = ({ templates, selectedTemplateId, onSelect }: StepTemplatesProps) => (
  <Grid container spacing={3} role="radiogroup" aria-label="Workspace templates">
    {templates.map(template => (
      <Grid item xs={12} md={6} lg={4} key={template.id} role="presentation">
        <TemplateCard
          title={template.name}
          description={template.description}
          icon={iconForTemplate(template.icon)}
          selected={selectedTemplateId === template.id}
          onSelect={() => onSelect(template.id)}
        />
      </Grid>
    ))}
  </Grid>
);
