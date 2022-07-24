export interface SkillType {
	iD: number
	name: string
	description: string
	skillNames: Array<SkillName>
}

export interface SkillName {
	iD: number
	skillTypeID: number
	skillType: SkillType
	name: string
	isEnabled: boolean
	skillNameAliases: Array<SkillNameAlias>
}

export interface SkillNameAlias {
	iD: number
	skillNameID: number | undefined
	skillName: SkillName | undefined
	alias: string
}