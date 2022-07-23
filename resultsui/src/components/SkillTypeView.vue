<template>
    <div>
        <label>Skill Type:</label>
    </div>
    <div>
        <input type="text" name="skillType" id="skillType" v-model="skillType.name" />
    </div>
    <div>
        <label>Description:</label>
    </div>
    <div>
        <textarea class="description" v-model="skillType.description" placeholder="Skill type description"></textarea>
    </div>
</template>
<script lang="ts">
import { SkillType } from '@/schemas/skills';
import { defineComponent } from 'vue';

export default defineComponent({
    name: 'skill-type-view',
    props: {
        'skillType': Object as () => SkillType
    }
})
</script>
