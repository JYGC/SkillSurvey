export interface SkillType {
	ID: number;
	Name: string;
	Description: string;
	SkillNames: Array<SkillName>;
}

export interface SkillName {
	ID: number;
	SkillTypeID: number;
	SkillType: SkillType;
	Name: string;
	IsEnabled: boolean;
	SkillNameAliases: Array<SkillNameAlias>;
}

export interface SkillNameAlias {
	ID: number;
	SkillNameID: number;
	SkillName: SkillName;
	Alias: string;
}