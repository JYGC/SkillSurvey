<template>
    <div>
        <label>Skill name:</label>
    </div>
    <div>
        <input type="text" name="skillName" id="skillName" v-model="skillName.name" />
    </div>
    <div>
        <label>Skill type:</label>
    </div>
    <div>
        <select name="skill-types" id="skill-types">
            <option v-for="skillType in skillTypes" value="{{skillType}}" v-bind:key="skillType">{{skillType}}</option>
        </select>
    </div>
    <div>
        <label>Alternate phrases:</label>
    </div>
    <div>
        <div v-for="alias in skillName?.skillNameAliases" v-bind:key="alias.iD">
            <input type="text" v-model="alias.alias" />
            <button v-on:click="deleteNewSkillNameAlias(alias)">Delete</button>
        </div>
        <div>
            <input type="text" v-model="newAlias" />
            <button v-on:click="addNewSkillNameAlias()">Add</button>
        </div>
    </div>
</template>
<script lang="ts">
import { SkillName, SkillNameAlias } from '@/schemas/skills';
import { computed, defineComponent } from 'vue';

export default defineComponent({
    props: {
        modelValue: {
            type: Object as () => SkillName
        }
    },
    emit: ['update:modelValue'],
    setup(props, { emit }) {
        const skillName = computed({
            get: () => props.modelValue,
            set: (value) => emit('update:modelValue', value)
        });
        let newAlias: string = "";
        let skillTypes: Array<string> = ['frontend', 'backend', 'middleware'];
        return {
            skillName,
            newAlias,
            skillTypes
        };
    },
    methods: {
        addNewSkillNameAlias(): void {
            let newAliasObject: SkillNameAlias = {
                iD: -1,
                skillNameID: this.skillName?.iD,
                skillName: this.skillName,
                alias: this.newAlias
            };
            this.skillName?.skillNameAliases.push(newAliasObject);
            this.newAlias = "";
        },
        deleteNewSkillNameAlias(skillAlias: SkillNameAlias): void {
            this.skillName?.skillNameAliases.splice(this.skillName?.skillNameAliases.indexOf(skillAlias), 1);
        }
    }
})
</script>
