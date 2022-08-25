<template>
    <div class="float-start">
        <b-button class="vertical-padding" @click.prevent="$router.go(-1)">Back</b-button>
    </div>
    <div>
        <SkillTypeView v-model="skillType" />
    </div>
    <div>
        <b-button v-on:click="addNewSkillType()">Save</b-button>
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
                },
                body: JSON.stringify(this.skillType)
            }).then(response => response.json()).then(json => {
                console.log(json);
                this.$router.go(-1);
            });
        }
    }
})
</script>
