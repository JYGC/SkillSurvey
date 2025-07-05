<template>
    <div class="row vertical-padding">
        <div class="col-md-12">
            <b-button class="float-start" @click.prevent="$router.go(-1)">Back</b-button>
            <b-button class="float-end margin-left-10" :disabled="isSaveBlocked()" v-on:click="saveSkillType()">Save</b-button>
            <b-button class="float-end" v-b-modal.confirm-delete :disabled="skillTypeModelValue.skillType.SkillNames.length != 0">Delete</b-button>
        </div>
    </div>
    <div class="row">
        <div class="col-md-6">
            <SkillTypeView v-model="skillTypeModelValue" />
        </div>
        <div class="col-md-6">
            <div class="row vertical-padding">
                <div class="col-md-4">
                    <label>Skills of this type:</label>
                </div>
                <div class="col-md-8">
                    <b-nav vertical>
                        <b-nav-item class="new-association vertical-padding" :to="{ name: 'skill-add', params: { skilltypeid: skillTypeModelValue.skillType.ID } }">
                            New Skill
                        </b-nav-item>
                        <b-nav-item
                        class="association"
                          v-for="skillName in skillTypeModelValue.skillType.SkillNames" :key="skillName.ID"
                          :to="{ name: 'skill-edit', params: { skillid: skillName.ID } }">
                            {{skillName.Name}}
                        </b-nav-item>
                    </b-nav>
                </div>
            </div>
        </div>
    </div>
    <b-modal id="confirm-delete" hide-header ok-title="Confirm" ok-variant="danger" @ok="deleteSkillType()">
        <p>Are you sure you want to delete this skill classification?</p>
    </b-modal>
</template>
<script lang="ts">
import SkillTypeView from '@/components/SkillTypeView.vue';
import { SkillType } from '@/schemas/skills';
import { defineComponent, reactive } from 'vue';
import { useRoute } from 'vue-router';

export default defineComponent({
    setup() {
        let skillTypeModelValue: { skillType: SkillType } = reactive({
            skillType: {
                ID: -1,
                Name: "",
                Description: "",
                SkillNames: []
            }
        });
        return {
            skillTypeModelValue
        };
    },
    components: {
        SkillTypeView
    },
    created() {
        fetch(`http://localhost:3000/skilltype/getbyid?skilltypeid=${ useRoute().params.skilltypeid }`).then(
            response => response.json()
        ).then(data => {
            this.skillTypeModelValue.skillType.ID = data.ID;
            this.skillTypeModelValue.skillType.Name = data.Name;
            this.skillTypeModelValue.skillType.Description = data.Description;
            this.skillTypeModelValue.skillType.SkillNames = data.SkillNames;
        });
    },
    methods: {
        saveSkillType(): void {
            fetch('http://localhost:3000/skilltype/save', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(this.skillTypeModelValue.skillType)
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
                    ID: this.skillTypeModelValue.skillType.ID
                })
            }).then(response => response.json()).then(json => {
                console.log(json); // if json is not int, throw error
                this.$router.go(-1);
            });
        },
        isSaveBlocked(): boolean {
            if (this.skillTypeModelValue.skillType.Name.trim() === "") return true;
            if (this.skillTypeModelValue.skillType.Description.trim() === "") return true;
            return false;
        }
    }
})
</script>
