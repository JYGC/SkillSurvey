import PocketBase from 'pocketbase';

// export class BackendClient implements IBackendClient {
//   private readonly __pb = new PocketBase();
  
//   constructor() {
//     this.__pb = new PocketBase(process.env.VUE_APP_POCKETBASE_URL);
//     this.__pb.authStore.loadFromCookie(document.cookie);
//     this.__pb.authStore.onChange(() => {
//       document.cookie = this.__pb.authStore.exportToCookie({ httpOnly: false, secure: false });
//     });
//   }

//   public get isTokenValid() {
//     return this.__pb.authStore.isValid;
//   }

//   public get token() {
//     return this.__pb.authStore.token;
//   }

//   public loginAsync = async (email: string, password: string) => {
//       await this.__pb.collection('users').authWithPassword(email, password);
//       return this.__pb.authStore.exportToCookie({ httpOnly: false });
//   }

//   public logoutAsync = async () => {
//     await this.__pb.collection('users').authRefresh();
//     this.__pb.authStore.clear();
//   };

//   public authRefresh = async () => await this.__pb.collection('users').authRefresh();

//   public get loggedInUser(): AuthRecord {
//     return this.__pb.authStore.record;
//   };
// }

export const getBackendClient = () => {
  const backendClient = new PocketBase(process.env.VUE_APP_POCKETBASE_URL);
  backendClient.authStore.loadFromCookie(document.cookie);
  backendClient.authStore.onChange(() => {
      document.cookie = backendClient.authStore.exportToCookie({ httpOnly: false, secure: false });
    });
  return backendClient;
};