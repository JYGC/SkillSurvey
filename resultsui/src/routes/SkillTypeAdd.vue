<template>
    <div>
        <button>Back</button>
    </div>
    <div>
        <SkillTypeView v-model="skillType" />
    </div>
    <div>
        <button>Save</button>
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
        return {
            skillType
        };
    },
    components: {
        SkillTypeView
    }
})
</script>
