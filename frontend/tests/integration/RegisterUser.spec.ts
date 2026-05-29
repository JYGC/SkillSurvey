import { describe, it, expect, vi, beforeEach } from 'vitest';
import { mount, flushPromises } from '@vue/test-utils';
import { createRouter, createMemoryHistory } from 'vue-router';
import RegisterUser from '@/views/public/RegisterUser.vue';

const mockRegister = vi.fn();

vi.mock('@/repositories/auth.repository', () => ({
  authRepository: {
    register: mockRegister,
  },
}));

function makeRouter() {
  return createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/', component: { template: '<div/>' } },
      { path: '/register', component: RegisterUser },
      { path: '/login', component: { template: '<div/>' } },
    ],
  });
}

const stubs = {
  CvFluidForm: true,
  CvTextInput: { template: '<input />' },
  CvButton: { template: '<button @click="$emit(\'click\')"><slot /></button>' },
  CvLink: true,
};

describe('RegisterUser', () => {
  beforeEach(() => {
    vi.resetAllMocks();
  });

  it('password mismatch: error shown, register not called', async () => {
    const router = makeRouter();
    await router.push('/register');
    const wrapper = mount(RegisterUser, { global: { plugins: [router], stubs } });

    const inputs = wrapper.findAll('input');
    await inputs[0].setValue('Test User');
    await inputs[1].setValue('new@example.com');
    await inputs[2].setValue('Password1!');
    await inputs[3].setValue('DifferentPassword!');
    await wrapper.find('button').trigger('click');
    await flushPromises();

    expect(mockRegister).not.toHaveBeenCalled();
    expect(wrapper.find('p[data-testid="register-error"]').exists()).toBe(true);
    expect(router.currentRoute.value.path).toBe('/register');
  });

  it('valid form: register called and navigates to /', async () => {
    mockRegister.mockResolvedValue({ id: 'newuser' });
    const router = makeRouter();
    await router.push('/register');
    const wrapper = mount(RegisterUser, { global: { plugins: [router], stubs } });

    const inputs = wrapper.findAll('input');
    await inputs[0].setValue('Test User');
    await inputs[1].setValue('new@example.com');
    await inputs[2].setValue('Password1!');
    await inputs[3].setValue('Password1!');
    await wrapper.find('button').trigger('click');
    await flushPromises();

    expect(mockRegister).toHaveBeenCalledWith('Test User', 'new@example.com', 'Password1!', 'Password1!');
    expect(router.currentRoute.value.path).toBe('/');
  });

  it('register fails: error shown, no navigation', async () => {
    mockRegister.mockRejectedValue(new Error('Registration failed'));
    const router = makeRouter();
    await router.push('/register');
    const wrapper = mount(RegisterUser, { global: { plugins: [router], stubs } });

    const inputs = wrapper.findAll('input');
    await inputs[0].setValue('Test User');
    await inputs[1].setValue('new@example.com');
    await inputs[2].setValue('Password1!');
    await inputs[3].setValue('Password1!');
    await wrapper.find('button').trigger('click');
    await flushPromises();

    expect(wrapper.find('p[data-testid="register-error"]').exists()).toBe(true);
    expect(router.currentRoute.value.path).toBe('/register');
  });
});
