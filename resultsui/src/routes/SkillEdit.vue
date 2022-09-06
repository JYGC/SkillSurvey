<template>
    <div class="row vertical-padding">
        <div class="col-md-12">
            <b-button class="float-start" @click.prevent="$router.go(-1)">Back</b-button>
            <b-button class="float-end margin-left-10" v-on:click="saveSkill()"
              :disabled="isSaveBlocked()">Save</b-button>
            <b-button class="float-end" v-b-modal.confirm-delete>Delete</b-button>
        </div>
    </div>
    <div class="row">
        <SkillView v-model="skillViewModalValueData" />
    </div>
    <b-modal id="confirm-delete" hide-header ok-title="Confirm" ok-variant="danger" @ok="deleteSkill()">
        <p>Are you sure you want to delete this skill?</p>
    </b-modal>
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
        return {
            skillViewModalValueData
        };
    },
    components: {
        SkillView
    },
    created() {
        fetch(`http://localhost:3000/skill/getbyid?skillid=${ useRoute().params.skillid }`).then(
            response => response.json()
        ).then(data => {
            this.skillViewModalValueData.skillName.ID = data.ID;
            this.skillViewModalValueData.skillName.SkillTypeID = data.SkillTypeID;
            this.skillViewModalValueData.skillName.Name = data.Name;
            this.skillViewModalValueData.skillName.IsEnabled = data.IsEnabled;
            this.skillViewModalValueData.skillName.SkillNameAliases = data.SkillNameAliases;
        });
    },
    methods: {
        saveSkill(): void {
            this.skillViewModalValueData.skillName
            console.log(this.skillViewModalValueData.skillName);
            fetch('http://localhost:3000/skill/save', {
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
        deleteSkill(): void {
            fetch('http://localhost:3000/skill/delete', {
                method: 'DELETE',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    ID: this.skillViewModalValueData.skillName.ID
                })
            }).then(response => response.json()).then(json => {
                console.log(json); // if json is not int, throw error
                this.$router.go(-1);
            });
        },
        isSaveBlocked(): boolean {
            if (this.skillViewModalValueData.newAlias.trim().length > 0) return true;
            if (this.skillViewModalValueData.skillName.Name.trim().length === 0) return true;
            if (this.skillViewModalValueData.skillName.SkillTypeID === 0) return true;
            return false;
        }
    }
})
</script>
