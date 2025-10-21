type PortalTheme = 'white' | 'g10' | 'g90' | 'g100';

export interface IUserSettings {
  user: string;
  portalThemes: PortalTheme;
}