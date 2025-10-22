export type PortalTheme = 'white' | 'g10' | 'g90' | 'g100';

export interface IUserSettings {
  id: string;
  user: string;
  portalThemes: PortalTheme;
}