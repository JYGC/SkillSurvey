<template>
    <div>
        <a href="#" @click.prevent="$router.go(-1)" ref="lnkBack">Back</a>
    </div>
    <div>
        <SkillView v-model="skillName" />
    </div>
    <div>
        <button v-on:click="addNewSKill()">Save</button>
    </div>
</template>

<script lang="ts">
import SkillView from '@/components/SkillView.vue';
import { SkillName } from '@/schemas/skills';
import { defineComponent, reactive } from 'vue';

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
    methods: {
        addNewSKill(): void {
            fetch('http://localhost:3000/skill/add', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(this.skillName)
            }).then(response => response.json()).then(json => {
                console.log(json);
                this.$router.go(-1);
            });
        }
    }
})
</script>
