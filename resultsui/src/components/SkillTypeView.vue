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
import { computed, defineComponent } from 'vue';

export default defineComponent({
    props: {
        modelValue: {
            type: Object as () => SkillType
        }
    },
    emit: ['update:modelValue'],
    setup(props, { emit }) {
        const skillType = computed({
            get: () => props.modelValue,
            set: (value) => emit('update:modelValue', value)
        });
        return {
            skillType
        }
    }
})
</script>
