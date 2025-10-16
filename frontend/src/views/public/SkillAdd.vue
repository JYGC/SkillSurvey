<template>
    <div>
        <div class="row vertical-padding">
            <div class="col-md-12">
                <b-button class="float-start" @click.prevent="$router.go(-1)">Back</b-button>
                <b-button class="float-end" v-on:click="addNewSKill()"
                :disabled="isAddBlocked()">Add</b-button>
            </div>
        </div>
        <div class="row">
            <SkillView v-model="skillModalValue" :forSkillTypeID="forSkillTypeID" />
        </div>
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

let forSkillTypeID = useRoute().params.skilltypeid;

const router = useRouter();

function addNewSKill(): void {
    fetch('http://localhost:3000/skill/add', {
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

function isAddBlocked(): boolean {
    if (skillModalValue.newAlias.trim().length > 0) return true;
    if (skillModalValue.skillName.Name.trim().length === 0) return true;
    if (skillModalValue.skillName.SkillTypeID === 0) return true;
    return false;
}
</script>
