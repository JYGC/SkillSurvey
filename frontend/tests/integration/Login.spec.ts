import { describe, it, expect, beforeEach } from 'vitest';
import { mount, flushPromises } from '@vue/test-utils';
import { createRouter, createMemoryHistory } from 'vue-router';
import Login from '@/views/public/Login.vue';
import pb from '@/store/pocketbase';
import { SEED_USER_EMAIL, SEED_USER_PASSWORD } from '../setup/seed';

function makeRouter() {
  return createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/', component: { template: '<div/>' } },
      { path: '/login', component: Login },
      { path: '/user/profile', component: { template: '<div/>' } },
    ],
  });
}

const stubs = { CvFluidForm: true, CvTextInput: { template: '<input />' }, CvButton: { template: '<button @click="$emit(\'click\')"><slot /></button>' }, CvLink: true };

describe('Login', () => {
  beforeEach(() => {
    pb.authStore.clear();
  });

  it('valid credentials navigate to /user/profile', async () => {
    const router = makeRouter();
    await router.push('/login');
    const wrapper = mount(Login, { global: { plugins: [router], stubs } });

    await wrapper.find('input[type="text"], input:not([type])').setValue(SEED_USER_EMAIL);
    await wrapper.findAll('input')[1]?.setValue(SEED_USER_PASSWORD);
    await wrapper.find('button').trigger('click');
    await flushPromises();

    expect(router.currentRoute.value.path).toBe('/user/profile');
  });

  it('invalid credentials show error paragraph, no navigation', async () => {
    const router = makeRouter();
    await router.push('/login');
    const wrapper = mount(Login, { global: { plugins: [router], stubs } });

    await wrapper.find('input:not([type]), input[type="text"]').setValue(SEED_USER_EMAIL);
    await wrapper.findAll('input')[1]?.setValue('wrongpassword');
    await wrapper.find('button').trigger('click');
    await flushPromises();

    expect(router.currentRoute.value.path).toBe('/login');
    expect(wrapper.find('p[data-testid="login-error"]').exists()).toBe(true);
  });
});
