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
            <tr v-for="skillName in skillNames" :key="skillName.iD">
                <td>{{ skillName.name }}</td>
                <td>{{ skillName.skillType.name }}</td>
                <td>{{ skillName.skillNameAliases }}</td>
                <td>
                    <router-link to="/skill-edit/{{ skillName.iD }}">Edit</router-link>
                </td>
            </tr>
        </table>
    </div>
</template>
<script lang="ts">
    import { SkillName } from '@/schemas/skills';
    import { defineComponent } from 'vue';

    export default defineComponent({
        setup() {
            let skillNames: Array<SkillName> = [
                {
                    iD: 1,
                    skillTypeID: -1,
                    skillType: {
                        iD: -1,
                        name: "Backend Language",
                        description: "",
                        skillNames: []
                    },
                    name: "C#",
                    isEnabled: true,
                    skillNameAliases: []
                },
            ];
            return {
                skillNames
            };
        },
    });
</script>
