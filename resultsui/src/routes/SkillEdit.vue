<template>
    <div>
        <a href="#" @click.prevent="$router.go(-1)">Back</a>
    </div>
    <div>
        <SkillView v-model="skillName" />
    </div>
    <div>
        <span>
            <button v-on:click="editSkill()">Save</button>
            <button v-on:click="deleteKill()">Delete</button>
        </span>
    </div>
</template>

<script lang="ts">
import SkillView from '@/components/SkillView.vue';
import { SkillName } from '@/schemas/skills';
import { defineComponent, reactive } from 'vue';
import { useRoute } from 'vue-router';

export default defineComponent({
    setup() {
        let skillName: SkillName = reactive({
            ID: 0,
            SkillTypeID: 0,
            SkillType: null,
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
    },
    created() {
        fetch(`http://localhost:3000/skill/getbyid?skillid=${ useRoute().params.skillid }`).then(
            response => response.json()
        ).then(data => {
            this.skillName.ID = data.ID;
            this.skillName.SkillTypeID = data.SkillTypeID;
            this.skillName.Name = data.Name;
            this.skillName.IsEnabled = data.IsEnabled;
            this.skillName.SkillNameAliases = data.SkillNameAliases;
        });
    },
    methods: {
        editSkill(): void {
            this.skillName
            console.log(this.skillName);
            fetch('http://localhost:3000/skill/save', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(this.skillName)
            }).then(response => response.json()).then(json => {
                console.log(json);
                this.$router.go(-1);
            });
        },
        deleteKill(): void {
            fetch('http://localhost:3000/skill/delete', {
                method: 'DELETE',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    ID: this.skillName.ID
                })
            }).then(response => response.json()).then(json => {
                console.log(json); // if json is not int, throw error
                this.$router.go(-1);
            });
        }
    }
})
</script>
