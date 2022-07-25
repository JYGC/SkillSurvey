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
            <tr v-for="skillType in skillTypes" :key="skillType.ID">
                <td>{{ skillType.Name }}</td>
                <td>{{ skillType.Description }}</td>
                <td>{{ (skillType.SkillNames !== null) ? skillType.SkillNames.length : 0 }}</td>
                <td>
                    <router-link to="/skill-type-edit/{{ skillType.ID }}">Edit</router-link>
                </td>
            </tr>
        </table>
    </div>
</template>
<script lang="ts">
import { SkillType } from '@/schemas/skills';
import { defineComponent, reactive } from 'vue';

export default defineComponent({
    setup() {
        let skillTypes: Array<SkillType> = reactive([]);
        return {
            skillTypes
        };
    },
    created() {
        // get data from API
        fetch('http://localhost:3000/api/skilltypelist').then(
            response => response.json()
        ).then(data => {
            for (let i: number = 0; i < data.length; i++) {
                this.skillTypes.push(data[i]);
            }
        });
    }
});
</script>
