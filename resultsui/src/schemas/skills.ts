export interface SkillType {
	ID: number;
	Name: string;
	Description: string;
	SkillNames: Array<SkillName>;
}

export interface SkillName {
	ID: number;
	SkillTypeID: number;
	SkillType: SkillType | null;
	Name: string;
	IsEnabled: boolean;
	SkillNameAliases: Array<SkillNameAlias>;
}

export interface SkillNameAlias {
	ID: number;
	SkillNameID: number;
	SkillName: SkillName | null;
	Alias: string;
}