<template>
    <div>
        <label>Skill name:</label>
    </div>
    <div>
        <input type="text" name="skillName" id="skillName" v-model="skillName.Name" />
    </div>
    <div>
        <label>Skill type:</label>
    </div>
    <div>
        <select name="skill-types" id="skill-types" v-model="skillName.SkillTypeID">
            <option v-for="(value, propertyName) in skillTypes" :value="propertyName" v-bind:key="propertyName">{{value}}</option>
        </select>
    </div>
    <div>
        <label>Alternate phrases:</label>
    </div>
    <div>
        <div v-for="alias in skillName?.SkillNameAliases" v-bind:key="alias.ID">
            <input type="text" v-model="alias.Alias" />
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
import { computed, defineComponent, reactive } from 'vue';

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
        let skillTypes: { [id: number]: string } = reactive({
            "0": "-- Select skill type --"
        });
        return {
            skillName,
            newAlias,
            skillTypes
        };
    },
    created() {
        fetch('http://localhost:3000/api/getallidandname').then(
            response => response.json()
        ).then(data => {
            for (let key in data) {
                this.skillTypes[parseInt(key)] = data[key];
            }
        });
    },
    methods: {
        addNewSkillNameAlias(): void {
            if (typeof this.skillName === 'undefined' || this.skillName === null) return;
            let newAliasObject: SkillNameAlias = {
                ID: -1,
                SkillNameID: this.skillName.ID,
                SkillName: this.skillName,
                Alias: this.newAlias
            };
            this.skillName.SkillNameAliases.push(newAliasObject);
            this.newAlias = "";
        },
        deleteNewSkillNameAlias(skillAlias: SkillNameAlias): void {
            this.skillName?.SkillNameAliases.splice(this.skillName?.SkillNameAliases.indexOf(skillAlias), 1);
        }
    }
})
</script>
