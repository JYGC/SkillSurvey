export interface SkillType {
  id: string;
  name: string;
  description: string;
  expand?: { skillNames_via_skillType?: SkillName[] };
}

export interface SkillName {
  id: string;
  skillType: string;          // relation ID
  name: string;
  isEnabled: boolean;
  expand?: {
    skillType?: SkillType;
    skillNameAliases_via_skillName?: SkillNameAlias[];
  };
}

export interface SkillNameAlias {
  id: string;
  skillName: string;          // relation ID
  alias: string;
  expand?: { skillName?: SkillName };
}
