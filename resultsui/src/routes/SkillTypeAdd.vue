<template>
    <div>
        <a href="#" @click.prevent="$router.go(-1)">Back</a>
    </div>
    <div>
        <SkillTypeView v-model="skillType" />
    </div>
    <div>
        <button v-on:click="addNewSkillType()">Save</button>
    </div>
</template>
<script lang="ts">
import SkillTypeView from '@/components/SkillTypeView.vue';
import { SkillType } from '@/schemas/skills';
import { defineComponent, reactive } from 'vue';

export default defineComponent({
    setup() {
        let skillType: SkillType = reactive({
            ID: 0,
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
    },
    methods: {
        addNewSkillType(): void {
            fetch('http://localhost:3000/skilltype/add', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                }
            });
        }
    }
})
</script>
