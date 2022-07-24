<template>
    <div>
        <button>Back</button>
    </div>
    <div>
        <SkillTypeView v-model="skillType" />
    </div>
    <div>
        <label>Skills of this type:</label>
    </div>
    <div>
        <router-link v-for="skillName in skillType.skillNames" :key="skillName.iD" to="/skill-edit/{{skillName.iD}}">
            {{skillName.name}}
        </router-link>
    </div>
    <div>
        <span>
            <button>Save</button>
            <button>Delete</button>
        </span>
    </div>
</template>
<script lang="ts">
import SkillTypeView from '@/components/SkillTypeView.vue';
import { SkillType } from '@/schemas/skills';
import { defineComponent, reactive } from 'vue';

export default defineComponent({
    setup() {
        let skillType: SkillType = reactive({
            iD: -1,
            name: "",
            description: "",
            skillNames: []
        });
        skillType.skillNames.push({
            iD: 2,
            skillTypeID: skillType.iD,
            skillType: skillType,
            name: "C#",
            isEnabled: true,
            skillNameAliases: []
        });
        return {
            skillType
        };
    },
    components: {
        SkillTypeView
    }
})
</script>
