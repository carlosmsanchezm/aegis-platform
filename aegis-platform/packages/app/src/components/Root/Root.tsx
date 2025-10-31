import { PropsWithChildren } from 'react';
import { makeStyles } from '@material-ui/core';
import DashboardIcon from '@material-ui/icons/DashboardOutlined';
import TimelineIcon from '@material-ui/icons/Timeline';
import SecurityIcon from '@material-ui/icons/Security';
import CloudQueueIcon from '@material-ui/icons/CloudQueue';
import LaptopIcon from '@material-ui/icons/LaptopMac';
import MemoryIcon from '@material-ui/icons/Memory';
import ExtensionIcon from '@material-ui/icons/Extension';
import LibraryBooks from '@material-ui/icons/LibraryBooks';
import MenuBookIcon from '@material-ui/icons/MenuBook';
import CreateComponentIcon from '@material-ui/icons/AddCircleOutline';
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
  SidebarSpace,
  useSidebarOpenState,
  Link,
} from '@backstage/core-components';
import SearchIcon from '@material-ui/icons/Search';
import { MyGroupsSidebarItem } from '@backstage/plugin-org';
import GroupIcon from '@material-ui/icons/People';
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

export const Root = ({ children }: PropsWithChildren<{}>) => (
  <SidebarPage>
    <Sidebar>
      <SidebarLogo />
      <SidebarGroup label="Search" icon={<SearchIcon />} to="/search">
        <SidebarSearchModal />
      </SidebarGroup>
      <SidebarDivider />
      <SidebarGroup label="Mission Control" icon={<DashboardIcon />}>
        <SidebarItem icon={DashboardIcon} to="catalog" text="Dashboard" />
        <SidebarItem icon={TimelineIcon} to="catalog-graph" text="Telemetry" />
        <SidebarItem icon={SecurityIcon} to="notifications" text="Live Posture" />
        <SidebarItem icon={CloudQueueIcon} to="aegis/workloads" text="Clusters" />
        <SidebarItem icon={LaptopIcon} to="aegis/workspaces/create" text="Secure Workspaces" />
        <SidebarItem icon={MemoryIcon} to="aegis" text="GPU Broker" />
      </SidebarGroup>
      <SidebarDivider />
      <SidebarGroup label="Resources" icon={<MenuBookIcon />}>
        <SidebarItem icon={ExtensionIcon} to="api-docs" text="APIs" />
        <SidebarItem icon={LibraryBooks} to="docs" text="Knowledge Base" />
        <SidebarItem icon={CreateComponentIcon} to="create" text="Automations" />
        <MyGroupsSidebarItem
          singularTitle="Mission Partner"
          pluralTitle="Mission Partners"
          icon={GroupIcon}
        />
      </SidebarGroup>
      <SidebarSpace />
      <SidebarDivider />
      <NotificationsSidebarItem />
      <SidebarDivider />
      <SidebarGroup
        label="Settings"
        icon={<UserSettingsSignInAvatar />}
        to="/settings"
      >
        <SidebarSettings />
      </SidebarGroup>
    </Sidebar>
    {children}
  </SidebarPage>
);
