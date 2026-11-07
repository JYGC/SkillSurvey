import pb from '@/store/pocketbase';

const hostConfigPath = '/host-config.json';

interface HostConfig {
  PocketBaseUrl?: unknown;
}

// Hosted builds leave VUE_APP_POCKETBASE_URL empty so one bundle serves test and production, and the install
// script writes host-config.json beside the frontend instead. Dev and E2E builds set the address at build time.
export async function pointPocketBaseAtAddressInHostConfig(): Promise<void> {
  if (process.env.VUE_APP_POCKETBASE_URL) {
    return;
  }

  const hostConfigResponse = await fetch(hostConfigPath, { cache: 'no-store' });
  if (!hostConfigResponse.ok) {
    throw new Error(`${hostConfigPath} could not be read: HTTP ${hostConfigResponse.status}`);
  }

  const hostConfig = (await hostConfigResponse.json()) as HostConfig;
  if (typeof hostConfig.PocketBaseUrl !== 'string' || hostConfig.PocketBaseUrl === '') {
    throw new Error(`${hostConfigPath} has no PocketBaseUrl`);
  }

  pb.baseURL = hostConfig.PocketBaseUrl;
}
