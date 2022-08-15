<template>
    <div>
        <a href="#" @click.prevent="$router.go(-1)">Back</a>
    </div>
    <div>
        <SkillTypeView v-model="skillType" />
    </div>
    <div>
        <label>Skills of this type:</label>
    </div>
    <div>
        <div v-for="skillName in skillType.SkillNames" :key="skillName.ID">
            <router-link :to="{ name: 'skill-edit', params: { skillid: skillName.ID } }">
                {{skillName.Name}}
            </router-link>
        </div>
        <div>
            <router-link :to="{ name: 'skill-add', params: { skilltypeid: skillType.ID } }">New Skill</router-link>
        </div>
    </div>
    <div>
        <span>
            <button v-on:click="saveSkillType()">Save</button>
            <button v-on:click="deleteSkillType()">Delete</button>
        </span>
    </div>
</template>
<script lang="ts">
import SkillTypeView from '@/components/SkillTypeView.vue';
import { SkillType } from '@/schemas/skills';
import { defineComponent, reactive } from 'vue';
import { useRoute } from 'vue-router';

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
    },
    created() {
        fetch(`http://localhost:3000/skilltype/getbyid?skilltypeid=${ useRoute().params.skilltypeid }`).then(
            response => response.json()
        ).then(data => {
            this.skillType.ID = data.ID;
            this.skillType.Name = data.Name;
            this.skillType.Description = data.Description;
            this.skillType.SkillNames = data.SkillNames;
        });
    },
    methods: {
        saveSkillType(): void {
            fetch('http://localhost:3000/skilltype/save', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(this.skillType)
            }).then(response => response.json()).then(json => {
                console.log(json);
                this.$router.go(-1);
            });
        },
        deleteSkillType(): void {
            fetch('http://localhost:3000/skilltype/delete', {
                method: 'DELETE',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    ID: this.skillType.ID
                })
            }).then(response => response.json()).then(json => {
                console.log(json); // if json is not int, throw error
                this.$router.go(-1);
            });
        }
    }
})
</script>
