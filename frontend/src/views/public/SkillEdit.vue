<template>
    <div>
        <div class="row vertical-padding">
            <div class="col-md-12">
                <b-button class="float-start" @click.prevent="$router.go(-1)">Back</b-button>
                <b-button class="float-end margin-left-10" v-on:click="saveSkill()"
                :disabled="isSaveBlocked()">Save</b-button>
                <b-button class="float-end" v-b-modal.confirm-delete>Delete</b-button>
            </div>
        </div>
        <div class="row">
            <SkillView v-model="skillModalValue" />
        </div>
        <b-modal id="confirm-delete" hide-header ok-title="Confirm" ok-variant="danger" @ok="deleteSkill()">
            <p>Are you sure you want to delete this skill?</p>
        </b-modal>
    </div>
</template>

<script lang="ts" setup>
import SkillView from '@/components/SkillView.vue';
import { SkillName } from '@/schemas/skills';
import { reactive } from 'vue';
import { useRoute, useRouter } from 'vue-router';

let skillModalValue: { skillName: SkillName, newAlias: string } = reactive({
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

fetch(`http://localhost:3000/skill/getbyid?skillid=${ useRoute().params.skillid }`).then(
    response => response.json()
).then(data => {
    skillModalValue.skillName.ID = data.ID;
    skillModalValue.skillName.SkillTypeID = data.SkillTypeID;
    skillModalValue.skillName.Name = data.Name;
    skillModalValue.skillName.IsEnabled = data.IsEnabled;
    skillModalValue.skillName.SkillNameAliases = data.SkillNameAliases;
});

const router = useRouter();

function saveSkill(): void {
    skillModalValue.skillName
    fetch('http://localhost:3000/skill/save', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(skillModalValue.skillName)
    }).then(response => response.json()).then(json => {
        console.log(json);
        router.back();
    });
}

function deleteSkill(): void {
    fetch('http://localhost:3000/skill/delete', {
        method: 'DELETE',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            ID: skillModalValue.skillName.ID
        })
    }).then(response => response.json()).then(json => {
        console.log(json); // if json is not int, throw error
        router.back();
    });
}

function isSaveBlocked(): boolean {
    if (skillModalValue.newAlias.trim().length > 0) return true;
    if (skillModalValue.skillName.Name.trim().length === 0) return true;
    if (skillModalValue.skillName.SkillTypeID === 0) return true;
    return false;
}
</script>
