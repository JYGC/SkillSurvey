import { createApp } from 'vue';
import { createRouter, createWebHistory } from 'vue-router';
import CarbonVue3 from '@carbon/vue';
import App from './App.vue';
import { routes } from './routes';
import { pointPocketBaseAtAddressInHostConfig } from './store/host-config';

const messageShownWhenSettingsCannotBeRead = 'SkillSurvey cannot start: its settings could not be read.';

function showThatSkillSurveyCannotStartInsteadOfBlankPage(reasonSettingsCouldNotBeRead: unknown): void {
  console.error(reasonSettingsCouldNotBeRead);
  const appElement = document.getElementById('app');
  if (appElement) {
    appElement.textContent = messageShownWhenSettingsCannotBeRead;
  }
}

// The router runs its first navigation as soon as it is installed, so PocketBase needs its address before then.
async function startSkillSurveyOncePocketBaseHasItsAddress(): Promise<void> {
  try {
    await pointPocketBaseAtAddressInHostConfig();
  } catch (reasonSettingsCouldNotBeRead) {
    showThatSkillSurveyCannotStartInsteadOfBlankPage(reasonSettingsCouldNotBeRead);
    return;
  }

  const app = createApp(App);
  const router = createRouter({
    history: createWebHistory(process.env.BASE_URL),
    routes,
  });
  app.use(router);
  app.use(CarbonVue3);
  app.mount('#app');
}

void startSkillSurveyOncePocketBaseHasItsAddress();
