<template>
    <div class="row vertical-padding">
        <div class="col-md-12">
            <b-button class="float-start" @click.prevent="$router.go(-1)">Back</b-button>
            <b-button class="float-end" v-on:click="addNewSKill()"
              :disabled="isAddBlocked()">Add</b-button>
        </div>
    </div>
    <div class="row">
        <SkillView v-model="skillViewModalValueData" :forSkillTypeID="forSkillTypeID" />
    </div>
</template>

<script lang="ts">
import SkillView from '@/components/SkillView.vue';
import { SkillName } from '@/schemas/skills';
import { defineComponent, reactive } from 'vue';
import { useRoute } from 'vue-router';

export default defineComponent({
    setup() {
        let skillViewModalValueData: { skillName: SkillName, newAlias: string } = reactive({
            skillName: {
                ID: 0,
                SkillTypeID: 0,
                SkillType: null,
                Name: "",
                IsEnabled: true,
                SkillNameAliases: []
            },
            newAlias: ""
        });
        let forSkillTypeID = useRoute().params.skilltypeid;
        return {
            skillViewModalValueData,
            forSkillTypeID
        };
    },
    components: {
        SkillView
    },
    methods: {
        addNewSKill(): void {
            fetch('http://localhost:3000/skill/add', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(this.skillViewModalValueData.skillName)
            }).then(response => response.json()).then(json => {
                console.log(json);
                this.$router.go(-1);
            });
        },
        isAddBlocked(): boolean {
            if (this.skillViewModalValueData.newAlias.trim().length > 0) return true;
            if (this.skillViewModalValueData.skillName.Name.trim().length === 0) return true;
            if (this.skillViewModalValueData.skillName.SkillTypeID === 0) return true;
            return false;
        }
    }
})
</script>
