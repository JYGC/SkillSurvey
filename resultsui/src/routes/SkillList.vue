<template>
    <div class="float-end">
        <b-button class="vertical-padding" :to="{ name: 'skill-add' }">New Skill</b-button>
    </div>
    <div>
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
                <tr v-for="skillName in skillNames" :key="skillName.ID">
                    <td>{{ skillName.Name }}</td>
                    <td>{{ skillName.SkillType.Name }}</td>
                    <td>
                        <div v-for="alias in getAliasList(skillName.SkillNameAliases)" :key="alias">
                            {{ alias }}
                        </div>
                    </td>
                    <td>
                        <router-link :to="{ name: 'skill-edit', params: { skillid: skillName.ID } }">Edit</router-link>
                    </td>
                </tr>
            </tbody>
        </table>
    </div>
</template>
<script lang="ts">
import { SkillName, SkillNameAlias } from '@/schemas/skills';
import { defineComponent, reactive } from 'vue';

export default defineComponent({
    setup() {
        let skillNames: Array<SkillName> = reactive([]);
        return {
            skillNames
        };
    },
    created() {
        // get data from API
        fetch('http://localhost:3000/skill/getall').then(
            response => response.json()
        ).then(data => {
            for (let i: number = 0; i < data.length; i++) this.skillNames.push(data[i]);
        });
    },
    methods: {
        getAliasList(skillNameAliases: SkillNameAlias[]): string[] {
            if (skillNameAliases === null) return [];
            let aliasList: string[] = skillNameAliases.map(s => s.Alias);
            return aliasList;
        }
    }
});
</script>
