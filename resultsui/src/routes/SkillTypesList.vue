<template>
    <div>
        <button to="/skill-type-add">New Skill Type</button>
    </div>
    <div>
        <table>
            <tr>
                <td>Skill type name</td>
                <td>Description</td>
                <td>No. of skills</td>
                <td></td>
            </tr>
            <tr v-for="skillType in skillTypes" :key="skillType.iD">
                <td>{{ skillType.name }}</td>
                <td>{{ skillType.description }}</td>
                <td>{{ skillType.skillNames.length }}</td>
                <td>
                    <router-link to="/skill-type-edit/{{ skillType.iD }}">Edit</router-link>
                </td>
            </tr>
        </table>
    </div>
</template>
<script lang="ts">
import { SkillType } from '@/schemas/skills'
import { defineComponent } from 'vue'

export default defineComponent({
    setup() {
        let skillTypes: Array<SkillType> = [];
        skillTypes.push({
            iD: 1,
            name: "Backend",
            description: "",
            skillNames: []
        });
        skillTypes[0].skillNames = [
            {
                iD: 2,
                skillTypeID: skillTypes[0].iD,
                skillType: skillTypes[0],
                name: "C#",
                isEnabled: true,
                skillNameAliases: []
            },
            {
                iD: 3,
                skillTypeID: skillTypes[0].iD,
                skillType: skillTypes[0],
                name: "Python",
                isEnabled: true,
                skillNameAliases: []
            },
        ];
        return {
            skillTypes
        };
    },
});
</script>
