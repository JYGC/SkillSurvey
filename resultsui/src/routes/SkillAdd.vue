<template>
    <div>
        <skill-view v-bind:skillName="skillName" ref="skillView" />
    </div>
    <div>
        {{skillName.name}}
    </div>
    <div>
        <button v-on:click="alert(1)">Call Ch</button>
    </div>
</template>

<script lang="ts">
import SkillView from '@/components/SkillView.vue'
import { SkillName } from '@/schemas/skills';
import { defineComponent } from 'vue'

export default defineComponent({
    name: 'skill-add',
    components: {
        SkillView
    },
    data() {
        return {
            skillName: Object as () => SkillName
        };
    },
    methods: {
        // callCh(): void {
        //     this.$refs.skillView.testCallChild();
        // }
    }
})
</script>
