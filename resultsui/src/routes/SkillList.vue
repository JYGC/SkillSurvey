<template>
    <div>
        <div class="float-end">
            <b-button class="vertical-padding" :to="{ name: 'skill-add' }">New Skill</b-button>
        </div>
        <table>
            <thead>
                <tr>
                    <td>Skill name</td>
                    <td>Skill type</td>
                    <td>Skill aliases</td>
                    <td></td>
                </tr>
            </thead>
            <tbody>
                <tr v-for="skillAndAlias in skillsAndAliases" :key="skillAndAlias.Skill.ID">
                    <td>{{ skillAndAlias.Skill.Name }}</td>
                    <td>{{ skillAndAlias.Skill.SkillType?.Name }}</td>
                    <td>
                        <div v-for="alias in skillAndAlias.Aliases" :key="alias">
                            {{ alias }}
                        </div>
                    </td>
                    <td>
                        <router-link :to="{ name: 'skill-edit', params: { skillid: skillAndAlias.Skill.ID } }">Edit</router-link>
                    </td>
                </tr>
            </tbody>
        </table>
    </div>
</template>
<script lang="ts" setup>
import { SkillName, SkillNameAlias } from '@/schemas/skills';
import { reactive } from 'vue';
import { sortByProperty } from '../services/arrays';

const getAllSkillsUrl = 'http://localhost:3000/skill/getall';

let skillsAndAliases: Array<{Skill: SkillName, Aliases: string[]}> = reactive([]);

(async function() {
    // get data from API
    const response = await fetch(getAllSkillsUrl);
    const sortedData = sortByProperty<SkillName>(await response.json(), skill => skill.Name);
    for (let i: number = 0; i < sortedData.length; i++)
        skillsAndAliases.push({
            Skill: sortedData[i],
            Aliases: sortByProperty<string>(getAliasList(sortedData[i].SkillNameAliases), alias => alias)
        });
})();

function getAliasList(skillNameAliases: SkillNameAlias[]): string[] {
    if (skillNameAliases === null) return [];
    let aliasList: string[] = skillNameAliases.map(s => s.Alias);
    return aliasList;
}
</script>
