import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import pb from '@/store/pocketbase';
import { pointPocketBaseAtAddressInHostConfig } from '@/store/host-config';

const pocketBaseAddressInHostConfig = 'http://pocketbase.example.test:8090';
const pocketBaseAddressBeforeHostConfigIsRead = 'http://address-before-host-config-was-read.example.test';
const startPageThatHttpdServesForEveryMissingFile = '<!DOCTYPE html><html><body><div id="app"></div></body></html>';

function answerEveryFetchWith(response: Response) {
  const fetchThatAnswersWithResponse = vi.fn().mockResolvedValue(response);
  vi.stubGlobal('fetch', fetchThatAnswersWithResponse);
  return fetchThatAnswersWithResponse;
}

function jsonResponse(body: unknown): Response {
  return new Response(JSON.stringify(body), { status: 200, headers: { 'Content-Type': 'application/json' } });
}

describe('pointPocketBaseAtAddressInHostConfig', () => {
  beforeEach(() => {
    pb.baseURL = pocketBaseAddressBeforeHostConfigIsRead;
    vi.stubEnv('VUE_APP_POCKETBASE_URL', '');
  });

  afterEach(() => {
    vi.unstubAllEnvs();
    vi.unstubAllGlobals();
  });

  it('points PocketBase at the address in host-config.json, asking for a fresh copy rather than a cached one', async () => {
    const fetchHostConfig = answerEveryFetchWith(jsonResponse({ PocketBaseUrl: pocketBaseAddressInHostConfig }));

    await pointPocketBaseAtAddressInHostConfig();

    expect(fetchHostConfig).toHaveBeenCalledTimes(1);
    expect(fetchHostConfig).toHaveBeenCalledWith('/host-config.json', { cache: 'no-store' });
    expect(pb.baseURL).toBe(pocketBaseAddressInHostConfig);
  });

  it('throws and leaves PocketBase alone when host-config.json is missing', async () => {
    answerEveryFetchWith(new Response('Not Found', { status: 404 }));

    await expect(pointPocketBaseAtAddressInHostConfig()).rejects.toThrow();
    expect(pb.baseURL).toBe(pocketBaseAddressBeforeHostConfigIsRead);
  });

  it('throws and leaves PocketBase alone when httpd answers with the start page because host-config.json is missing', async () => {
    answerEveryFetchWith(
      new Response(startPageThatHttpdServesForEveryMissingFile, { status: 200, headers: { 'Content-Type': 'text/html' } }),
    );

    await expect(pointPocketBaseAtAddressInHostConfig()).rejects.toThrow();
    expect(pb.baseURL).toBe(pocketBaseAddressBeforeHostConfigIsRead);
  });

  it('throws and leaves PocketBase alone when host-config.json has no PocketBaseUrl', async () => {
    answerEveryFetchWith(jsonResponse({}));

    await expect(pointPocketBaseAtAddressInHostConfig()).rejects.toThrow();
    expect(pb.baseURL).toBe(pocketBaseAddressBeforeHostConfigIsRead);
  });

  it('throws and leaves PocketBase alone when host-config.json has an empty PocketBaseUrl', async () => {
    answerEveryFetchWith(jsonResponse({ PocketBaseUrl: '' }));

    await expect(pointPocketBaseAtAddressInHostConfig()).rejects.toThrow();
    expect(pb.baseURL).toBe(pocketBaseAddressBeforeHostConfigIsRead);
  });

  it('throws and leaves PocketBase alone when host-config.json cannot be fetched at all', async () => {
    vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new TypeError('Failed to fetch')));

    await expect(pointPocketBaseAtAddressInHostConfig()).rejects.toThrow();
    expect(pb.baseURL).toBe(pocketBaseAddressBeforeHostConfigIsRead);
  });

  it('fetches nothing and leaves PocketBase alone when the build set VUE_APP_POCKETBASE_URL, as dev and E2E builds do', async () => {
    vi.stubEnv('VUE_APP_POCKETBASE_URL', 'http://dev-backend.example.test:8091');
    const fetchThatMustNotBeCalled = answerEveryFetchWith(jsonResponse({ PocketBaseUrl: pocketBaseAddressInHostConfig }));

    await pointPocketBaseAtAddressInHostConfig();

    expect(fetchThatMustNotBeCalled).not.toHaveBeenCalled();
    expect(pb.baseURL).toBe(pocketBaseAddressBeforeHostConfigIsRead);
  });
});
