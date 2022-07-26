<template>
    <div>
        <a href="#" @click.prevent="$router.go(-1)">Back</a>
    </div>
    <div>
        <SkillView v-model="skillName" />
    </div>
    <div>
        <span>
            <button>Save</button>
            <button>Delete</button>
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
            ID: -1,
            SkillTypeID: -1,
            SkillType: {
                ID: 0,
                Name: "",
                Description: "",
                SkillNames: []
            },
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
            this.skillName.SkillType = data.SkillType;
            this.skillName.Name = data.Name;
            this.skillName.IsEnabled = data.IsEnabled;
            this.skillName.SkillNameAliases = data.SkillNameAliases;
        });
    },
})
</script>
