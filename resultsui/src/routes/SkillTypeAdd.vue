<template>
    <div class="col-md-6">
        <div class="row vertical-padding">
            <div class="col-md-12">
                <b-button class="float-start" @click.prevent="$router.go(-1)">Back</b-button>
                <b-button class="float-end" :disabled="isAddBlocked()" v-on:click="addNewSkillType()">Add</b-button>
            </div>
        </div>
        <div class="row">
            <div class="float-start">
                <SkillTypeView v-model="skillTypeModelValue" />
            </div>
        </div>
    </div>
</template>

<script lang="ts">
import SkillTypeView from '@/components/SkillTypeView.vue';
import { SkillType } from '@/schemas/skills';
import { defineComponent, reactive } from 'vue';

export default defineComponent({
    setup() {
        let skillTypeModelValue: { skillType: SkillType } = reactive({
            skillType: {
                ID: 0,
                Name: "",
                Description: "",
                SkillNames: []
            }
        });
        return {
            skillTypeModelValue
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
                body: JSON.stringify(this.skillTypeModelValue.skillType)
            }).then(response => response.json()).then(json => {
                console.log(json);
                this.$router.go(-1);
            });
        },
        isAddBlocked(): boolean {
            if (this.skillTypeModelValue.skillType.Name.trim() === "") return true;
            if (this.skillTypeModelValue.skillType.Description.trim() === "") return true;
            return false;
        }
    }
})
</script>
