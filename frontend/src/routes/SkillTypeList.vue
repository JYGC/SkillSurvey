<template>
    <div class="float-end">
        <b-button class="vertical-padding" :to="{ name: 'skill-type-add' }">New Skill Type</b-button>
    </div>
    <div>
        <table>
            <thead>
                <tr>
                    <td>Skill type name</td>
                    <td>Description</td>
                    <td>No. of skills</td>
                    <td></td>
                </tr>
            </thead>
            <tbody>
                <tr v-for="skillType in skillTypes" :key="skillType.ID">
                    <td>{{ skillType.Name }}</td>
                    <td>{{ skillType.Description }}</td>
                    <td>{{ (skillType.SkillNames !== null) ? skillType.SkillNames.length : 0 }}</td>
                    <td>
                        <router-link :to="{ name: 'skill-type-edit', params: { skilltypeid: skillType.ID } }">Edit</router-link>
                    </td>
                </tr>
            </tbody>
        </table>
    </div>
</template>
<script lang="ts" setup>
import { SkillType } from '@/schemas/skills';
import { reactive } from 'vue';
import { sortByProperty } from '../services/arrays';

const getAllSkillTypesUrl = 'http://localhost:3000/skilltype/getall';

let skillTypes: Array<SkillType> = reactive([]);

(async function() {
    // get data from API
    const response = await fetch(getAllSkillTypesUrl);
    const sortedData = sortByProperty<SkillType>(await response.json(), skillType => skillType.Name);
    for (let i: number = 0; i < sortedData.length; i++) skillTypes.push(sortedData[i]);
})();
</script>
