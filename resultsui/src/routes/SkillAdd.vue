<template>
    <div>
        <a href="#" @click.prevent="$router.go(-1)">Back</a>
    </div>
    <div>
        <SkillView v-model="skillName" />
    </div>
    <div>
        <button>Save</button>
    </div>
</template>

<script lang="ts">
import SkillView from '@/components/SkillView.vue';
import { SkillName } from '@/schemas/skills';
import { defineComponent, reactive } from 'vue';

export default defineComponent({
    setup() {
        let skillName: SkillName = reactive({
            ID: -1,
            SkillTypeID: 0,
            SkillType: {
                ID: 0,
                Name: "",
                Description: "",
                SkillNames: []
            },
            Name: "",
            IsEnabled: true,
            SkillNameAliases: []
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
