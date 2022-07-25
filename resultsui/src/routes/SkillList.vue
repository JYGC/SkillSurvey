<template>
    <div>
        <button to="/skill-add">New Skill</button>
    </div>
    <div>
        <table>
            <tr>
                <td>Skill name</td>
                <td>Skill type</td>
                <td>Skill aliases</td>
                <td></td>
            </tr>
            <tr v-for="skillName in skillNames" :key="skillName.ID">
                <td>{{ skillName.Name }}</td>
                <td>{{ skillName.SkillType.Name }}</td>
                <td>{{ getTopFiveAliases(skillName.SkillNameAliases) }}</td>
                <td>
                    <router-link to="/skill-edit/{{ skillName.ID }}">Edit</router-link>
                </td>
            </tr>
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
        fetch('http://localhost:3000/api/skilllist').then(
            response => response.json()
        ).then(data => {
            for (let i: number = 0; i < data.length; i++) {
                this.skillNames.push(data[i]);
            }
        });
    },
    methods: {
        getTopFiveAliases(skillNameAliases: SkillNameAlias[]): string {
            console.log(skillNameAliases);
            if (skillNameAliases !== null) return skillNameAliases.map(s => s.Alias).join(", ");
            return "";
        }
    }
});
</script>
