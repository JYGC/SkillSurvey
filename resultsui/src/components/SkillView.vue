<template>
    <div class="col-md-6">
        <div class="row vertical-padding">
            <div class="col-md-3">
                <label class="float-start">Skill name:</label>
            </div>
            <div class="col-md-9">
                <input class="float-start fill-parent" type="text" name="skillName" id="skillName" v-model="modalValueData.skillName.Name" />
            </div>
        </div>
        <div class="row vertical-padding">
            <div class="col-md-3">
                <label class="float-start">Skill type:</label>
            </div>
            <div class="col-md-9">
                <select class="float-start fill-parent" name="skill-types" id="skill-types" v-model="modalValueData.skillName.SkillTypeID" :disabled="typeof forSkillTypeID !== 'undefined'">
                    <option v-for="(value, propertyName) in skillTypes" :value="propertyName" v-bind:key="propertyName">{{value}}</option>
                </select>
            </div>
        </div>
    </div>
    <div class="col-md-6">
        <div class="row vertical-padding">
            <div class="col-md-3">
                <label class="float-start">Alternate phrases:</label>
            </div>
            <div class="col-md-9">
                <div class="row vertical-padding" v-for="alias in modalValueData?.skillName?.SkillNameAliases" v-bind:key="alias.ID">
                    <div class="col-md-12">
                        <input type="text" v-model="alias.Alias" />
                        <b-button v-on:click="deleteNewSkillNameAlias(alias)">Delete</b-button>
                    </div>
                </div>
                <div class="row vertical-padding">
                    <div class="col-md-12">
                        <input type="text" v-model="modalValueData.newAlias" />
                        <b-button v-on:click="addNewSkillNameAlias()" :disabled="isAddNewAliasBlocked()">Add</b-button>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>
<script lang="ts">
import { SkillName, SkillNameAlias } from '@/schemas/skills';
import { computed, defineComponent, reactive } from 'vue';

export default defineComponent({
    props: {
        modelValue: {
            type: Object as () => { skillName: SkillName, newAlias: string }
        },
        forSkillTypeID: Number
    },
    emit: ['update:modelValue'],
    setup(props, { emit }) {
        const modalValueData = computed({
            get: () => props.modelValue,
            set: (value) => emit('update:modelValue', value)
        });
        let skillTypes: { [id: number]: string } = reactive({
            "0": "-- Select skill type --"
        });
        return {
            modalValueData,
            skillTypes
        };
    },
    created() {
        fetch('http://localhost:3000/skilltype/getallidandname').then(
            response => response.json()
        ).then(data => {
            for (let key in data) {
                this.skillTypes[parseInt(key)] = data[key];
            }
            if (typeof this.modalValueData === 'undefined' || typeof this.modalValueData.skillName === 'undefined' || typeof this.forSkillTypeID === 'undefined') return;
            this.modalValueData.skillName.SkillTypeID = this.forSkillTypeID;
        });
    },
    methods: {
        addNewSkillNameAlias(): void {
            if (typeof this.modalValueData === 'undefined' || typeof this.modalValueData.skillName === 'undefined' || this.modalValueData.skillName === null) return;
            let newAliasObject: SkillNameAlias = {
                ID: -1,
                SkillNameID: this.modalValueData.skillName.ID,
                SkillName: null,
                Alias: this.modalValueData.newAlias
            };
            this.modalValueData.skillName.SkillNameAliases.push(newAliasObject);
            this.modalValueData.newAlias = "";
        },
        deleteNewSkillNameAlias(skillAlias: SkillNameAlias): void {
            if (typeof this.modalValueData === 'undefined') return;
            this.modalValueData.skillName?.SkillNameAliases.splice(this.modalValueData.skillName?.SkillNameAliases.indexOf(skillAlias), 1);
        },
        isAddNewAliasBlocked(): boolean {
            if (this.modalValueData?.newAlias.trim().length === 0) return true;
            return false;
        }
    }
})
</script>
