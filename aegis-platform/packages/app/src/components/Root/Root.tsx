import { PropsWithChildren, ReactNode } from 'react';
import { makeStyles } from '@material-ui/core';
import AddCircleOutlineIcon from '@material-ui/icons/AddCircleOutline';
import BuildIcon from '@material-ui/icons/Build';
import CloudQueueIcon from '@material-ui/icons/CloudQueue';
import DashboardIcon from '@material-ui/icons/Dashboard';
import DescriptionIcon from '@material-ui/icons/Description';
import ExtensionIcon from '@material-ui/icons/Extension';
import HomeIcon from '@material-ui/icons/Home';
import LaptopMacIcon from '@material-ui/icons/LaptopMac';
import LockIcon from '@material-ui/icons/Lock';
import SearchIcon from '@material-ui/icons/Search';
import SecurityIcon from '@material-ui/icons/Security';
import SettingsIcon from '@material-ui/icons/Settings';
import TimelineIcon from '@material-ui/icons/Timeline';
import LogoFull from './LogoFull';
import LogoIcon from './LogoIcon';
import {
  Settings as SidebarSettings,
  UserSettingsSignInAvatar,
} from '@backstage/plugin-user-settings';
import { SidebarSearchModal } from '@backstage/plugin-search';
import {
  Sidebar,
  sidebarConfig,
  SidebarDivider,
  SidebarGroup,
  SidebarItem,
  SidebarPage,
  SidebarScrollWrapper,
  SidebarSpace,
  useSidebarOpenState,
  Link,
} from '@backstage/core-components';
import { MyGroupsSidebarItem } from '@backstage/plugin-org';
import { NotificationsSidebarItem } from '@backstage/plugin-notifications';

const useSidebarLogoStyles = makeStyles({
  root: {
    width: sidebarConfig.drawerWidthClosed,
    height: 3 * sidebarConfig.logoHeight,
    display: 'flex',
    flexFlow: 'row nowrap',
    alignItems: 'center',
    marginBottom: -14,
  },
  link: {
    width: sidebarConfig.drawerWidthClosed,
    marginLeft: 24,
  },
});

const useNavSectionStyles = makeStyles(theme => ({
  section: {
    margin: theme.spacing(0.5, 0, 3),
  },
  header: {
    margin: theme.spacing(0, 2, 1),
    fontSize: 11,
    fontWeight: 600,
    letterSpacing: '0.12em',
    textTransform: 'uppercase',
    color: theme.palette.text.secondary,
    opacity: 0.76,
  },
  items: {
    display: 'flex',
    flexDirection: 'column',
    gap: theme.spacing(0.5),
    padding: theme.spacing(0, 1),
  },
  searchTrigger: {
    margin: theme.spacing(1, 2, 3),
  },
}));

const SidebarLogo = () => {
  const classes = useSidebarLogoStyles();
  const { isOpen } = useSidebarOpenState();

  return (
    <div className={classes.root}>
      <Link to="/" underline="none" className={classes.link} aria-label="Home">
        {isOpen ? <LogoFull /> : <LogoIcon />}
      </Link>
    </div>
  );
};

type NavSectionProps = {
  label: string;
  icon: ReactNode;
  children: ReactNode;
};

const NavSection = ({ label, icon, children }: NavSectionProps) => {
  const classes = useNavSectionStyles();

  return (
    <SidebarGroup label={label} icon={icon} value={label}>
      <div className={classes.section}>
        <div className={classes.header}>{label}</div>
        <div className={classes.items}>{children}</div>
      </div>
    </SidebarGroup>
  );
};

export const Root = ({ children }: PropsWithChildren<{}>) => {
  const navClasses = useNavSectionStyles();

  return (
    <SidebarPage>
      <Sidebar>
        <SidebarLogo />

        <SidebarGroup label="Search" icon={<SearchIcon />} to="/search">
          <div className={navClasses.searchTrigger}>
            <SidebarSearchModal />
          </div>
        </SidebarGroup>

        <SidebarDivider />

        <NavSection label="Create" icon={<AddCircleOutlineIcon />}>
          <SidebarItem
            icon={LockIcon}
            to="aegis/workspaces/create"
            text="Launch Secure Workspace"
          />
          <SidebarItem icon={BuildIcon} to="aegis" text="Launch Workload" />
        </NavSection>

        <SidebarDivider />

        <SidebarScrollWrapper>
          <NavSection label="Manage" icon={<DashboardIcon />}>
            <SidebarItem icon={LaptopMacIcon} to="aegis/workloads" text="Workspaces" />
            <SidebarItem icon={CloudQueueIcon} to="aegis/clusters" text="Clusters" />
            <SidebarItem icon={TimelineIcon} to="aegis/telemetry" text="Telemetry" />
            <SidebarItem icon={SecurityIcon} to="aegis/posture" text="Live Posture" />
            <MyGroupsSidebarItem singularTitle="My Group" pluralTitle="My Groups" />
          </NavSection>

          <SidebarDivider />

          <NavSection label="Admin" icon={<SettingsIcon />}>
            <SidebarItem icon={HomeIcon} to="catalog" text="Catalog" />
            <SidebarItem icon={ExtensionIcon} to="api-docs" text="APIs" />
            <SidebarItem icon={DescriptionIcon} to="docs" text="Docs" />
            <NotificationsSidebarItem />
            <SidebarItem icon={SettingsIcon} to="settings" text="Settings" />
          </NavSection>
        </SidebarScrollWrapper>

        <SidebarSpace />
        <SidebarDivider />

        <SidebarGroup
          label="Profile"
          icon={<UserSettingsSignInAvatar />}
          to="/settings"
        >
          <SidebarSettings />
        </SidebarGroup>
      </Sidebar>
      {children}
    </SidebarPage>
  );
};