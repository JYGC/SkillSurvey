import pb from '@/store/pocketbase';
import { beforeEach } from 'vitest';

beforeEach(() => {
  pb.authStore.clear();
});
