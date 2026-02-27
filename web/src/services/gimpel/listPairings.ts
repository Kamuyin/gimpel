// @ts-ignore
/* eslint-disable */
import { request } from '@umijs/max';

/** List pairings GET /api/v1/pairings */
export async function listPairings(options?: { [key: string]: any }) {
  return request<API.ListPairingsResponse>('/api/v1/pairings', {
    method: 'GET',
    ...(options || {}),
  });
}
