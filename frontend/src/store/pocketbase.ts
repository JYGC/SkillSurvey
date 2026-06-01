import PocketBase from 'pocketbase';

const url = process.env.TEST_PB_URL ?? process.env.VUE_APP_POCKETBASE_URL ?? '';
const pb = new PocketBase(url);

if (typeof document !== 'undefined') {
  pb.authStore.loadFromCookie(document.cookie);
  pb.authStore.onChange(() => {
    document.cookie = pb.authStore.exportToCookie({ httpOnly: false, secure: false });
  });
}

export default pb;
