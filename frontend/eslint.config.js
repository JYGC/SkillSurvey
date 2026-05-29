const { defineConfigWithVueTs, vueTsConfigs } = require('@vue/eslint-config-typescript');
const pluginVue = require('eslint-plugin-vue');

module.exports = defineConfigWithVueTs(
  pluginVue.configs['flat/essential'],
  vueTsConfigs.recommended,
  {
    files: ['src/views/**/*.vue'],
    rules: {
      'vue/multi-word-component-names': 'off',
    },
  },
);
