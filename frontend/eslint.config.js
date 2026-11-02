const { defineConfigWithVueTs, vueTsConfigs } = require('@vue/eslint-config-typescript');
const pluginVue = require('eslint-plugin-vue');

module.exports = defineConfigWithVueTs(
  { ignores: ['dist/'] },
  pluginVue.configs['flat/essential'],
  vueTsConfigs.recommended,
  {
    files: ['src/views/**/*.vue'],
    rules: {
      'vue/multi-word-component-names': 'off',
    },
  },
  {
    files: ['eslint.config.js', 'vue.config.js'],
    rules: {
      '@typescript-eslint/no-require-imports': 'off',
    },
  },
);
