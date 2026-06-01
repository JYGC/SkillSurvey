import { describe, it, expect, vi, beforeEach } from 'vitest';
import { mount, flushPromises } from '@vue/test-utils';
import { createRouter, createMemoryHistory } from 'vue-router';
import Login from '@/views/public/Login.vue';

const mockLogin = vi.hoisted(() => vi.fn());

vi.mock('@/composables/use-auth', () => ({
  useAuth: () => ({ login: mockLogin }),
}));

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

const stubs = {
  CvFluidForm: { template: '<div><slot /></div>' },
  CvTextInput: {
    props: { modelValue: String, type: String, label: String },
    emits: ['update:modelValue'],
    template: '<input :value="modelValue" @input="$emit(\'update:modelValue\', $event.target.value)" />',
  },
  CvButton: { template: '<button @click="$emit(\'click\')"><slot /></button>' },
  CvLink: true,
};

describe('Login', () => {
  beforeEach(() => {
    vi.resetAllMocks();
  });

  it('valid credentials navigate to /user/profile', async () => {
    mockLogin.mockResolvedValue({});
    const router = makeRouter();
    await router.push('/login');
    const wrapper = mount(Login, { global: { plugins: [router], stubs } });

    const inputs = wrapper.findAll('input');
    await inputs[0].setValue('user@example.com');
    await inputs[1].setValue('password');
    await wrapper.find('button').trigger('click');
    await flushPromises();

    expect(mockLogin).toHaveBeenCalledWith('user@example.com', 'password');
    expect(router.currentRoute.value.path).toBe('/user/profile');
  });

  it('invalid credentials show error paragraph, no navigation', async () => {
    mockLogin.mockRejectedValue(new Error('Failed to authenticate'));
    const router = makeRouter();
    await router.push('/login');
    const wrapper = mount(Login, { global: { plugins: [router], stubs } });

    const inputs = wrapper.findAll('input');
    await inputs[0].setValue('user@example.com');
    await inputs[1].setValue('wrongpassword');
    await wrapper.find('button').trigger('click');
    await flushPromises();

    expect(router.currentRoute.value.path).toBe('/login');
    expect(wrapper.find('p[data-testid="login-error"]').exists()).toBe(true);
  });
});
