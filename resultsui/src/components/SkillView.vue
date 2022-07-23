<template>
    <div>
        <label>Skill name:</label>
    </div>
    <div>
        <input type="text" name="skillName" id="skillName" v-model="name" />
    </div>
    <!-- <div>
        <label>Alternate phrases:</label>
    </div>
    <div>
        <span v-for="skillNameAlias in skillName.skillNameAliases" v-bind:key="skillNameAlias.iD">
            <input type="text" v-model="skillNameAlias.name" />
            <button>Delete</button>
        </span>
        <span>
            <input type="text" v-model="newAlias" />
            <button v-on:click="addNewSkillNameAlias()">Add</button>
        </span>
    </div> -->
</template>
<script lang="ts">
import { SkillNameAlias } from '@/schemas/skills';
import { defineComponent } from 'vue';

export default defineComponent({
    name: 'skill-view',
    data() {
        return {
            name: ""//,
            //skillNameAliases: Array<SkillNameAlias>()
        };
    },
    methods: {
        // addNewSkillNameAlias(): void {
        //     if (typeof this.skillName !== 'undefined' && this.skillName !== null)
        //     {
        //         let newSkillNameAlias: SkillNameAlias = {
        //             iD: -1,
        //             skillNameID: this.skillName.iD,
        //             skillName: this.skillName,
        //             alias: this.newAlias
        //         };
        //         this.skillNameAliases.push(newSkillNameAlias);
        //     }
        // },
        // getSkillNameAliases(): Array<SkillNameAlias> {
        //     return this.skillNameAliases;
        // },
        testCallChild(): void {
            console.log(1);
        }
    }
})
</script>
