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
        <router-link v-for="skillName in skillType.SkillNames" :key="skillName.ID" to="/skill-edit/{{skillName.ID}}">
            {{skillName.Name}}
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
            ID: -1,
            Name: "",
            Description: "",
            SkillNames: []
        });
        skillType.SkillNames?.push({
            ID: 2,
            SkillTypeID: skillType.ID,
            SkillType: skillType,
            Name: "C#",
            IsEnabled: true,
            SkillNameAliases: []
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
