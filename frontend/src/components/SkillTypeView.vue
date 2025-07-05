<template>
    <div class="row vertical-padding">
        <div class="col-md-3">
            <label class="float-start">Skill Type:</label>
        </div>
        <div class="col-md-9">
            <input class="float-start fill-parent" type="text" name="skillType" id="skillType" v-model="modalValueData.skillType.Name" />
        </div>
    </div>
    <div class="row vertical-padding">
        <div class="col-md-3">
            <label class="float-start">Description:</label>
        </div>
        <div class="col-md-9">
            <textarea class="description float-start fill-parent" v-model="modalValueData.skillType.Description" placeholder="Skill type description"></textarea>
        </div>
    </div>
</template>
<script lang="ts">
import { SkillType } from '@/schemas/skills';
import { computed, defineComponent } from 'vue';

export default defineComponent({
    props: {
        modelValue: {
            type: Object as () => { skillType: SkillType }
        }
    },
    emit: ['update:modelValue'],
    setup(props, { emit }) {
        const modalValueData = computed({
            get: () => props.modelValue,
            set: (value) => emit('update:modelValue', value)
        });
        return {
            modalValueData
        };
    }
})
</script>
