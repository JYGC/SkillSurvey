<template>
    <div>
        <button>Back</button>
    </div>
    <div>
        <SkillView v-model="skillName" />
    </div>
    <div>
        <button>Save</button>
    </div>
</template>

<script lang="ts">
import SkillView from '@/components/SkillView.vue'
import { SkillName } from '@/schemas/skills';
import { defineComponent, reactive } from 'vue'

export default defineComponent({
    setup() {
        let skillName: SkillName = reactive({
            iD: -1,
            skillTypeID: -1,
            skillType: {
                iD: -1,
                name: "",
                description: "",
                skillNames: []
            },
            name: "",
            isEnabled: true,
            skillNameAliases: []
        });
        return {
            skillName
        };
    },
    components: {
        SkillView
    }
})
</script>
