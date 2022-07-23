export interface SkillType {
	iD: number
	name: string
	skillNames: Array<SkillName>
}

export interface SkillName {
	iD: number
	skillTypeID: number
	skillType: Array<SkillType>
	name: string
	isEnabled: boolean
	skillNameAliases: Array<SkillNameAlias>
}

export interface SkillNameAlias {
	iD: number
	skillNameID: number
	skillName: SkillName
	alias: string
}